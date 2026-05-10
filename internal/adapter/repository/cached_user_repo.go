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

const userProfileTTL = 30 * time.Minute

type cachedUserRepo struct {
	inner port.UserRepository
	cache pkgredis.CachePort
}

func NewCachedUserRepository(inner port.UserRepository, cache pkgredis.CachePort) port.UserRepository {
	return &cachedUserRepo{inner: inner, cache: cache}
}

func (r *cachedUserRepo) profileKey(userID uuid.UUID) string {
	return "user-svc:profile:" + userID.String()
}

func (r *cachedUserRepo) Create(ctx context.Context, user *domain.User) error {
	return r.inner.Create(ctx, user)
}

func (r *cachedUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if r.cache != nil {
		if val, err := r.cache.Get(ctx, r.profileKey(id)); err == nil {
			u := &domain.User{}
			if json.Unmarshal([]byte(val), u) == nil {
				return u, nil
			}
		}
	}
	user, err := r.inner.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if r.cache != nil {
		if b, err := json.Marshal(user); err == nil {
			_ = r.cache.Set(ctx, r.profileKey(id), string(b), userProfileTTL)
		}
	}
	return user, nil
}

func (r *cachedUserRepo) GetByKeycloakID(ctx context.Context, keycloakID string) (*domain.User, error) {
	return r.inner.GetByKeycloakID(ctx, keycloakID)
}

func (r *cachedUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.inner.GetByEmail(ctx, email)
}

func (r *cachedUserRepo) Update(ctx context.Context, user *domain.User) error {
	err := r.inner.Update(ctx, user)
	if err == nil && r.cache != nil {
		_ = r.cache.Delete(ctx, r.profileKey(user.ID))
	}
	return err
}

func (r *cachedUserRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	err := r.inner.SoftDelete(ctx, id)
	if err == nil && r.cache != nil {
		_ = r.cache.Delete(ctx, r.profileKey(id))
	}
	return err
}

func (r *cachedUserRepo) Search(ctx context.Context, query string, limit int, cursor *time.Time) ([]*domain.User, error) {
	return r.inner.Search(ctx, query, limit, cursor)
}

func (r *cachedUserRepo) UpdateKeycloakSyncFields(ctx context.Context, userID uuid.UUID, fields domain.KeycloakSyncFields) error {
	err := r.inner.UpdateKeycloakSyncFields(ctx, userID, fields)
	if err == nil && r.cache != nil {
		_ = r.cache.Delete(ctx, r.profileKey(userID))
	}
	return err
}

func (r *cachedUserRepo) UpdateTrustScore(ctx context.Context, userID uuid.UUID, score float64) error {
	return r.inner.UpdateTrustScore(ctx, userID, score)
}

func (r *cachedUserRepo) UpdateRole(ctx context.Context, userID uuid.UUID, role domain.UserRole) error {
	err := r.inner.UpdateRole(ctx, userID, role)
	if err == nil && r.cache != nil {
		_ = r.cache.Delete(ctx, r.profileKey(userID))
	}
	return err
}

func (r *cachedUserRepo) UpdateStatus(ctx context.Context, userID uuid.UUID, status domain.UserStatus) error {
	err := r.inner.UpdateStatus(ctx, userID, status)
	if err == nil && r.cache != nil {
		_ = r.cache.Delete(ctx, r.profileKey(userID))
	}
	return err
}

func (r *cachedUserRepo) IncrFollowerCount(ctx context.Context, userID uuid.UUID, delta int) error {
	return r.inner.IncrFollowerCount(ctx, userID, delta)
}

func (r *cachedUserRepo) IncrFollowingCount(ctx context.Context, userID uuid.UUID, delta int) error {
	return r.inner.IncrFollowingCount(ctx, userID, delta)
}
