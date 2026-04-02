package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"be-modami-user-service/internal/domain"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
)

const (
	profileTTL       = 30 * time.Minute
	statusTTL        = 10 * time.Minute
	sellerTTL        = 30 * time.Minute
	countTTL         = 15 * time.Minute
	isFollowingTTL   = 10 * time.Minute
	ratingSummaryTTL = 30 * time.Minute
	addressesTTL     = 30 * time.Minute
	kycStatusTTL     = 15 * time.Minute
)

type redisCache struct {
	client  *redis.Client
	breaker *gobreaker.CircuitBreaker
}

func NewRedisCache(client *redis.Client) *redisCache {
	st := gobreaker.Settings{
		Name:        "redis-cache",
		MaxRequests: 5,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 5
		},
	}
	cb := gobreaker.NewCircuitBreaker(st)
	return &redisCache{client: client, breaker: cb}
}

func (c *redisCache) key(parts ...string) string {
	k := "user-svc"
	for _, p := range parts {
		k += ":" + p
	}
	return k
}

func (c *redisCache) get(ctx context.Context, key string) (string, error) {
	val, err := c.breaker.Execute(func() (interface{}, error) {
		return c.client.Get(ctx, key).Result()
	})
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", redis.Nil
		}
		return "", err
	}
	return val.(string), nil
}

func (c *redisCache) set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	_, err := c.breaker.Execute(func() (interface{}, error) {
		return nil, c.client.Set(ctx, key, value, ttl).Err()
	})
	return err
}

func (c *redisCache) del(ctx context.Context, keys ...string) error {
	_, err := c.breaker.Execute(func() (interface{}, error) {
		return nil, c.client.Del(ctx, keys...).Err()
	})
	return err
}

func (c *redisCache) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	k := c.key("profile", userID.String())
	val, err := c.get(ctx, k)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	u := &domain.User{}
	if err := json.Unmarshal([]byte(val), u); err != nil {
		return nil, err
	}
	return u, nil
}

func (c *redisCache) SetProfile(ctx context.Context, user *domain.User) error {
	b, err := json.Marshal(user)
	if err != nil {
		return err
	}
	k := c.key("profile", user.ID.String())
	return c.set(ctx, k, string(b), profileTTL)
}

func (c *redisCache) DeleteProfile(ctx context.Context, userID uuid.UUID) error {
	return c.del(ctx, c.key("profile", userID.String()))
}

func (c *redisCache) GetStatus(ctx context.Context, userID uuid.UUID) (domain.UserStatus, error) {
	k := c.key("status", userID.String())
	val, err := c.get(ctx, k)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", err
	}
	return domain.UserStatus(val), nil
}

func (c *redisCache) SetStatus(ctx context.Context, userID uuid.UUID, status domain.UserStatus) error {
	k := c.key("status", userID.String())
	return c.set(ctx, k, string(status), statusTTL)
}

func (c *redisCache) GetSellerProfile(ctx context.Context, userID uuid.UUID) (*domain.SellerProfile, error) {
	k := c.key("seller", userID.String())
	val, err := c.get(ctx, k)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	p := &domain.SellerProfile{}
	if err := json.Unmarshal([]byte(val), p); err != nil {
		return nil, err
	}
	return p, nil
}

func (c *redisCache) SetSellerProfile(ctx context.Context, userID uuid.UUID, profile *domain.SellerProfile) error {
	b, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	k := c.key("seller", userID.String())
	return c.set(ctx, k, string(b), sellerTTL)
}

func (c *redisCache) DeleteSellerProfile(ctx context.Context, userID uuid.UUID) error {
	return c.del(ctx, c.key("seller", userID.String()))
}

func (c *redisCache) GetFollowerCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	k := c.key("follower_count", userID.String())
	val, err := c.get(ctx, k)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return -1, nil
		}
		return 0, err
	}
	n, err := strconv.ParseInt(val, 10, 64)
	return n, err
}

func (c *redisCache) SetFollowerCount(ctx context.Context, userID uuid.UUID, count int64) error {
	k := c.key("follower_count", userID.String())
	return c.set(ctx, k, fmt.Sprintf("%d", count), countTTL)
}

func (c *redisCache) GetFollowingCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	k := c.key("following_count", userID.String())
	val, err := c.get(ctx, k)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return -1, nil
		}
		return 0, err
	}
	n, err := strconv.ParseInt(val, 10, 64)
	return n, err
}

func (c *redisCache) SetFollowingCount(ctx context.Context, userID uuid.UUID, count int64) error {
	k := c.key("following_count", userID.String())
	return c.set(ctx, k, fmt.Sprintf("%d", count), countTTL)
}

func (c *redisCache) IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	k := c.key("is_following", followerID.String(), followingID.String())
	val, err := c.get(ctx, k)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, redis.Nil
		}
		return false, err
	}
	return val == "1", nil
}

func (c *redisCache) SetIsFollowing(ctx context.Context, followerID, followingID uuid.UUID, val bool) error {
	k := c.key("is_following", followerID.String(), followingID.String())
	v := "0"
	if val {
		v = "1"
	}
	return c.set(ctx, k, v, isFollowingTTL)
}

func (c *redisCache) DeleteFollowKeys(ctx context.Context, followerID, followingID uuid.UUID) error {
	keys := []string{
		c.key("is_following", followerID.String(), followingID.String()),
		c.key("follower_count", followingID.String()),
		c.key("following_count", followerID.String()),
	}
	return c.del(ctx, keys...)
}

func (c *redisCache) GetRatingSummary(ctx context.Context, userID uuid.UUID) (*domain.RatingSummary, error) {
	k := c.key("rating_summary", userID.String())
	val, err := c.get(ctx, k)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	rs := &domain.RatingSummary{}
	if err := json.Unmarshal([]byte(val), rs); err != nil {
		return nil, err
	}
	return rs, nil
}

func (c *redisCache) SetRatingSummary(ctx context.Context, userID uuid.UUID, summary *domain.RatingSummary) error {
	b, err := json.Marshal(summary)
	if err != nil {
		return err
	}
	k := c.key("rating_summary", userID.String())
	return c.set(ctx, k, string(b), ratingSummaryTTL)
}

func (c *redisCache) DeleteRatingSummary(ctx context.Context, userID uuid.UUID) error {
	return c.del(ctx, c.key("rating_summary", userID.String()))
}

func (c *redisCache) GetAddresses(ctx context.Context, userID uuid.UUID) ([]*domain.Address, error) {
	k := c.key("addresses", userID.String())
	val, err := c.get(ctx, k)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	var addrs []*domain.Address
	if err := json.Unmarshal([]byte(val), &addrs); err != nil {
		return nil, err
	}
	return addrs, nil
}

func (c *redisCache) SetAddresses(ctx context.Context, userID uuid.UUID, addresses []*domain.Address) error {
	b, err := json.Marshal(addresses)
	if err != nil {
		return err
	}
	k := c.key("addresses", userID.String())
	return c.set(ctx, k, string(b), addressesTTL)
}

func (c *redisCache) DeleteAddresses(ctx context.Context, userID uuid.UUID) error {
	return c.del(ctx, c.key("addresses", userID.String()))
}

func (c *redisCache) GetKYCStatus(ctx context.Context, userID uuid.UUID) (domain.KYCStatus, error) {
	k := c.key("kyc_status", userID.String())
	val, err := c.get(ctx, k)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", err
	}
	return domain.KYCStatus(val), nil
}

func (c *redisCache) SetKYCStatus(ctx context.Context, userID uuid.UUID, status domain.KYCStatus) error {
	k := c.key("kyc_status", userID.String())
	return c.set(ctx, k, string(status), kycStatusTTL)
}

func (c *redisCache) DeleteKYCStatus(ctx context.Context, userID uuid.UUID) error {
	return c.del(ctx, c.key("kyc_status", userID.String()))
}
