package repository

import (
	"context"
	"time"

	"be-modami-user-service/internal/domain"
	"be-modami-user-service/internal/port"

	"github.com/google/uuid"
	pkgredis "gitlab.com/lifegoeson-libs/pkg-gokit/redis"
)

const isFollowingTTL = 10 * time.Minute

type cachedFollowRepo struct {
	inner port.FollowRepository
	cache pkgredis.CachePort
}

func NewCachedFollowRepository(inner port.FollowRepository, cache pkgredis.CachePort) port.FollowRepository {
	return &cachedFollowRepo{inner: inner, cache: cache}
}

func (r *cachedFollowRepo) isFollowingKey(followerID, followingID uuid.UUID) string {
	return "user-svc:is_following:" + followerID.String() + ":" + followingID.String()
}

func (r *cachedFollowRepo) followerCountKey(userID uuid.UUID) string {
	return "user-svc:follower_count:" + userID.String()
}

func (r *cachedFollowRepo) followingCountKey(userID uuid.UUID) string {
	return "user-svc:following_count:" + userID.String()
}

func (r *cachedFollowRepo) invalidateFollowKeys(ctx context.Context, followerID, followingID uuid.UUID) {
	_ = r.cache.Delete(ctx,
		r.isFollowingKey(followerID, followingID),
		r.followerCountKey(followingID),
		r.followingCountKey(followerID),
	)
}

func (r *cachedFollowRepo) Follow(ctx context.Context, followerID, followingID uuid.UUID) error {
	err := r.inner.Follow(ctx, followerID, followingID)
	if err == nil && r.cache != nil {
		r.invalidateFollowKeys(ctx, followerID, followingID)
	}
	return err
}

func (r *cachedFollowRepo) Unfollow(ctx context.Context, followerID, followingID uuid.UUID) error {
	err := r.inner.Unfollow(ctx, followerID, followingID)
	if err == nil && r.cache != nil {
		r.invalidateFollowKeys(ctx, followerID, followingID)
	}
	return err
}

func (r *cachedFollowRepo) IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	if r.cache != nil {
		if val, err := r.cache.Get(ctx, r.isFollowingKey(followerID, followingID)); err == nil {
			return val == "1", nil
		}
	}
	result, err := r.inner.IsFollowing(ctx, followerID, followingID)
	if err != nil {
		return false, err
	}
	if r.cache != nil {
		v := "0"
		if result {
			v = "1"
		}
		_ = r.cache.Set(ctx, r.isFollowingKey(followerID, followingID), v, isFollowingTTL)
	}
	return result, nil
}

func (r *cachedFollowRepo) GetFollowers(ctx context.Context, userID uuid.UUID, limit int, cursor *time.Time) ([]*domain.FollowUser, error) {
	return r.inner.GetFollowers(ctx, userID, limit, cursor)
}

func (r *cachedFollowRepo) GetFollowing(ctx context.Context, userID uuid.UUID, limit int, cursor *time.Time) ([]*domain.FollowUser, error) {
	return r.inner.GetFollowing(ctx, userID, limit, cursor)
}
