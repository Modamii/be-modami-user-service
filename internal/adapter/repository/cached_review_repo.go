package repository

import (
	"context"
	"encoding/json"
	"time"

	"be-modami-user-service/internal/domain"
	"be-modami-user-service/internal/port"

	"github.com/google/uuid"
	pkgredis "gitlab.com/lifegoeson-libs/pkg-gokit/redis"
)

const ratingSummaryTTL = 30 * time.Minute

type cachedReviewRepo struct {
	inner port.ReviewRepository
	cache pkgredis.CachePort
}

func NewCachedReviewRepository(inner port.ReviewRepository, cache pkgredis.CachePort) port.ReviewRepository {
	return &cachedReviewRepo{inner: inner, cache: cache}
}

func (r *cachedReviewRepo) ratingSummaryKey(userID uuid.UUID) string {
	return "user-svc:rating_summary:" + userID.String()
}

func (r *cachedReviewRepo) Create(ctx context.Context, review *domain.Review) error {
	return r.inner.Create(ctx, review)
}

func (r *cachedReviewRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error) {
	return r.inner.GetByID(ctx, id)
}

func (r *cachedReviewRepo) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*domain.Review, error) {
	return r.inner.GetByOrderID(ctx, orderID)
}

func (r *cachedReviewRepo) ListByReviewee(ctx context.Context, revieweeID uuid.UUID, limit int, cursor *time.Time) ([]*domain.Review, error) {
	return r.inner.ListByReviewee(ctx, revieweeID, limit, cursor)
}

func (r *cachedReviewRepo) GetRatingSummary(ctx context.Context, userID uuid.UUID) (*domain.RatingSummary, error) {
	if r.cache != nil {
		if val, err := r.cache.Get(ctx, r.ratingSummaryKey(userID)); err == nil {
			rs := &domain.RatingSummary{}
			if json.Unmarshal([]byte(val), rs) == nil {
				return rs, nil
			}
		}
	}
	summary, err := r.inner.GetRatingSummary(ctx, userID)
	if err != nil {
		return nil, err
	}
	if r.cache != nil {
		if b, err := json.Marshal(summary); err == nil {
			_ = r.cache.Set(ctx, r.ratingSummaryKey(userID), string(b), ratingSummaryTTL)
		}
	}
	return summary, nil
}

func (r *cachedReviewRepo) UpsertRatingSummary(ctx context.Context, userID uuid.UUID, rating int) error {
	err := r.inner.UpsertRatingSummary(ctx, userID, rating)
	if err == nil && r.cache != nil {
		_ = r.cache.Delete(ctx, r.ratingSummaryKey(userID))
	}
	return err
}
