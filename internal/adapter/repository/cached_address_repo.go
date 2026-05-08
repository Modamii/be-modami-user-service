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

const addressListTTL = 30 * time.Minute

type cachedAddressRepo struct {
	inner port.AddressRepository
	cache pkgredis.CachePort
}

func NewCachedAddressRepository(inner port.AddressRepository, cache pkgredis.CachePort) port.AddressRepository {
	return &cachedAddressRepo{inner: inner, cache: cache}
}

func (r *cachedAddressRepo) addressesKey(userID uuid.UUID) string {
	return "user-svc:addresses:" + userID.String()
}

func (r *cachedAddressRepo) Create(ctx context.Context, addr *domain.Address) error {
	err := r.inner.Create(ctx, addr)
	if err == nil && r.cache != nil {
		_ = r.cache.Delete(ctx, r.addressesKey(addr.UserID))
	}
	return err
}

func (r *cachedAddressRepo) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Address, error) {
	return r.inner.GetByID(ctx, id, userID)
}

func (r *cachedAddressRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Address, error) {
	if r.cache != nil {
		if val, err := r.cache.Get(ctx, r.addressesKey(userID)); err == nil {
			var addrs []*domain.Address
			if json.Unmarshal([]byte(val), &addrs) == nil {
				return addrs, nil
			}
		}
	}
	addrs, err := r.inner.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if r.cache != nil {
		if b, err := json.Marshal(addrs); err == nil {
			_ = r.cache.Set(ctx, r.addressesKey(userID), string(b), addressListTTL)
		}
	}
	return addrs, nil
}

func (r *cachedAddressRepo) Update(ctx context.Context, addr *domain.Address) error {
	err := r.inner.Update(ctx, addr)
	if err == nil && r.cache != nil {
		_ = r.cache.Delete(ctx, r.addressesKey(addr.UserID))
	}
	return err
}

func (r *cachedAddressRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	err := r.inner.Delete(ctx, id, userID)
	if err == nil && r.cache != nil {
		_ = r.cache.Delete(ctx, r.addressesKey(userID))
	}
	return err
}

func (r *cachedAddressRepo) SetDefault(ctx context.Context, id, userID uuid.UUID) error {
	err := r.inner.SetDefault(ctx, id, userID)
	if err == nil && r.cache != nil {
		_ = r.cache.Delete(ctx, r.addressesKey(userID))
	}
	return err
}

func (r *cachedAddressRepo) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	return r.inner.CountByUserID(ctx, userID)
}
