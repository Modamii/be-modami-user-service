package service

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/modami/user-service/internal/domain"
	"github.com/modami/user-service/internal/port"
	"github.com/modami/user-service/pkg/apperror"
	"github.com/modami/user-service/pkg/pagination"
)

type ReviewService struct {
	reviewRepo port.ReviewRepository
	userRepo   port.UserRepository
	sellerRepo port.SellerProfileRepository
	cache      port.CacheService
	publisher  port.EventPublisher
}

func NewReviewService(
	reviewRepo port.ReviewRepository,
	userRepo port.UserRepository,
	sellerRepo port.SellerProfileRepository,
	cache port.CacheService,
	publisher port.EventPublisher,
) *ReviewService {
	return &ReviewService{
		reviewRepo: reviewRepo,
		userRepo:   userRepo,
		sellerRepo: sellerRepo,
		cache:      cache,
		publisher:  publisher,
	}
}

func (s *ReviewService) CreateReview(
	ctx context.Context,
	reviewerID, revieweeID uuid.UUID,
	orderID uuid.UUID,
	rating int,
	comment string,
	role domain.ReviewRole,
	isAnonymous bool,
) (*domain.Review, error) {
	// Check if review already exists for this order
	existing, err := s.reviewRepo.GetByOrderID(ctx, orderID)
	if err != nil && err != apperror.ErrNotFound {
		return nil, err
	}
	if existing != nil {
		return nil, apperror.ErrReviewAlreadyExists
	}

	now := time.Now()
	review := &domain.Review{
		ID:          uuid.New(),
		ReviewerID:  reviewerID,
		RevieweeID:  revieweeID,
		OrderID:     orderID,
		Rating:      rating,
		Comment:     comment,
		Role:        role,
		IsAnonymous: isAnonymous,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.reviewRepo.Create(ctx, review); err != nil {
		return nil, err
	}

	if err := s.reviewRepo.UpsertRatingSummary(ctx, revieweeID, rating); err != nil {
		return nil, err
	}

	_ = s.cache.DeleteRatingSummary(ctx, revieweeID)

	_ = s.recalcTrustScore(ctx, revieweeID)

	_ = s.publisher.PublishUserReviewCreated(ctx, &domain.UserReviewCreatedEvent{
		ReviewerID: reviewerID,
		RevieweeID: revieweeID,
		OrderID:    orderID,
		Rating:     rating,
	})

	return review, nil
}

func (s *ReviewService) ListReviews(ctx context.Context, revieweeID uuid.UUID, limit int, cursorStr string) ([]*domain.Review, string, error) {
	if limit <= 0 {
		limit = 20
	}
	cursor, err := pagination.DecodeCursor(cursorStr)
	if err != nil {
		return nil, "", err
	}

	reviews, err := s.reviewRepo.ListByReviewee(ctx, revieweeID, limit+1, cursor)
	if err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(reviews) > limit {
		nextCursor = pagination.EncodeCursor(reviews[limit-1].CreatedAt)
		reviews = reviews[:limit]
	}
	return reviews, nextCursor, nil
}

func (s *ReviewService) GetRatingSummary(ctx context.Context, userID uuid.UUID) (*domain.RatingSummary, error) {
	cached, err := s.cache.GetRatingSummary(ctx, userID)
	if err == nil && cached != nil {
		return cached, nil
	}

	summary, err := s.reviewRepo.GetRatingSummary(ctx, userID)
	if err != nil {
		return nil, err
	}

	_ = s.cache.SetRatingSummary(ctx, userID, summary)
	return summary, nil
}

func (s *ReviewService) recalcTrustScore(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	summary, err := s.reviewRepo.GetRatingSummary(ctx, userID)
	if err != nil {
		return err
	}

	kycVerified := 0.0
	sellerProfile, err := s.sellerRepo.GetByUserID(ctx, userID)
	if err == nil && sellerProfile != nil && sellerProfile.KYCStatus == domain.KYCStatusApproved {
		kycVerified = 1.0
	}

	emailVerified := 0.0
	if user.EmailVerified {
		emailVerified = 1.0
	}

	accountAgeMonths := time.Since(user.CreatedAt).Hours() / 24 / 30
	totalReviews := float64(summary.TotalReviews)

	// Trust score formula
	score := (summary.AvgRating*0.4 +
		kycVerified*1.0 +
		math.Min(totalReviews/50, 1)*0.5 +
		emailVerified*0.5 +
		math.Min(accountAgeMonths/24, 1)*0.5) / 2.9 * 5.0

	return s.userRepo.UpdateTrustScore(ctx, userID, score)
}
