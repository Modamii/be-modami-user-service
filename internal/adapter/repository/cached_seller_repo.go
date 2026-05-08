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

const sellerProfileTTL = 30 * time.Minute

type cachedSellerRepo struct {
	inner port.SellerProfileRepository
	cache pkgredis.CachePort
}

func NewCachedSellerRepository(inner port.SellerProfileRepository, cache pkgredis.CachePort) port.SellerProfileRepository {
	return &cachedSellerRepo{inner: inner, cache: cache}
}

func (r *cachedSellerRepo) sellerKey(userID uuid.UUID) string {
	return "user-svc:seller:" + userID.String()
}

func (r *cachedSellerRepo) Create(ctx context.Context, profile *domain.SellerProfile) error {
	return r.inner.Create(ctx, profile)
}

func (r *cachedSellerRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.SellerProfile, error) {
	if r.cache != nil {
		if val, err := r.cache.Get(ctx, r.sellerKey(userID)); err == nil {
			p := &domain.SellerProfile{}
			if json.Unmarshal([]byte(val), p) == nil {
				return p, nil
			}
		}
	}
	profile, err := r.inner.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile != nil && r.cache != nil {
		if b, err := json.Marshal(profile); err == nil {
			_ = r.cache.Set(ctx, r.sellerKey(userID), string(b), sellerProfileTTL)
		}
	}
	return profile, nil
}

func (r *cachedSellerRepo) Update(ctx context.Context, profile *domain.SellerProfile) error {
	err := r.inner.Update(ctx, profile)
	if err == nil && r.cache != nil {
		_ = r.cache.Delete(ctx, r.sellerKey(profile.UserID))
	}
	return err
}

func (r *cachedSellerRepo) UpdateKYCStatus(ctx context.Context, userID uuid.UUID, status domain.KYCStatus, verifiedAt *time.Time) error {
	err := r.inner.UpdateKYCStatus(ctx, userID, status, verifiedAt)
	if err == nil && r.cache != nil {
		_ = r.cache.Delete(ctx, r.sellerKey(userID))
	}
	return err
}
