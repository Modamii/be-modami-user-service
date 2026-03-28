package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/modami/user-service/internal/domain"
	"github.com/modami/user-service/internal/dto"
	"github.com/modami/user-service/internal/port"
	apperror "github.com/modami/user-service/internal/apperror"
)

type SellerService struct {
	sellerRepo port.SellerProfileRepository
	userRepo   port.UserRepository
	cache      port.CacheService
}

func NewSellerService(
	sellerRepo port.SellerProfileRepository,
	userRepo port.UserRepository,
	cache port.CacheService,
) *SellerService {
	return &SellerService{
		sellerRepo: sellerRepo,
		userRepo:   userRepo,
		cache:      cache,
	}
}

func (s *SellerService) Register(ctx context.Context, userID uuid.UUID, req dto.RegisterSellerRequest) (*domain.SellerProfile, error) {
	existing, err := s.sellerRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperror.ErrAlreadySeller
	}

	now := time.Now()
	profile := &domain.SellerProfile{
		ID:              uuid.New(),
		UserID:          userID,
		ShopName:        req.ShopName,
		ShopSlug:        req.ShopSlug,
		ShopDescription: req.ShopDescription,
		BusinessType:    domain.BusinessType(req.BusinessType),
		TaxID:           req.TaxID,
		BankAccount:     req.BankAccount,
		BankName:        req.BankName,
		KYCStatus:       domain.KYCStatusNone,
		AvgRating:       0,
		TotalReviews:    0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.sellerRepo.Create(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

func (s *SellerService) UpdateProfile(ctx context.Context, userID uuid.UUID, req dto.UpdateSellerProfileRequest) (*domain.SellerProfile, error) {
	profile, err := s.sellerRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, apperror.ErrSellerNotFound
	}

	if req.ShopName != nil {
		profile.ShopName = *req.ShopName
	}
	if req.ShopDescription != nil {
		profile.ShopDescription = *req.ShopDescription
	}
	if req.ShopLogoURL != nil {
		profile.ShopLogoURL = *req.ShopLogoURL
	}
	if req.ShopBannerURL != nil {
		profile.ShopBannerURL = *req.ShopBannerURL
	}
	if req.TaxID != nil {
		profile.TaxID = *req.TaxID
	}
	if req.BankAccount != nil {
		profile.BankAccount = *req.BankAccount
	}
	if req.BankName != nil {
		profile.BankName = *req.BankName
	}
	profile.UpdatedAt = time.Now()

	if err := s.sellerRepo.Update(ctx, profile); err != nil {
		return nil, err
	}

	_ = s.cache.DeleteSellerProfile(ctx, userID)
	return profile, nil
}

func (s *SellerService) GetShopProfile(ctx context.Context, userID uuid.UUID) (*domain.SellerProfile, error) {
	cached, err := s.cache.GetSellerProfile(ctx, userID)
	if err == nil && cached != nil {
		return cached, nil
	}

	profile, err := s.sellerRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, apperror.ErrSellerNotFound
	}

	_ = s.cache.SetSellerProfile(ctx, userID, profile)
	return profile, nil
}
