package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/modami/user-service/internal/domain"
	"github.com/modami/user-service/internal/port"
	"github.com/modami/user-service/pkg/apperror"
	"github.com/modami/user-service/pkg/pagination"
)

type FollowService struct {
	followRepo port.FollowRepository
	cache      port.CacheService
	publisher  port.EventPublisher
}

func NewFollowService(
	followRepo port.FollowRepository,
	cache port.CacheService,
	publisher port.EventPublisher,
) *FollowService {
	return &FollowService{
		followRepo: followRepo,
		cache:      cache,
		publisher:  publisher,
	}
}

func (s *FollowService) Follow(ctx context.Context, followerID, followingID uuid.UUID) error {
	if followerID == followingID {
		return apperror.ErrSelfFollow
	}

	already, err := s.followRepo.IsFollowing(ctx, followerID, followingID)
	if err != nil {
		return err
	}
	if already {
		return apperror.ErrAlreadyFollowing
	}

	// Follow repo handles tx internally (insert follow + update counts)
	if err := s.followRepo.Follow(ctx, followerID, followingID); err != nil {
		return err
	}

	_ = s.cache.DeleteFollowKeys(ctx, followerID, followingID)

	_ = s.publisher.PublishUserFollowed(ctx, &domain.UserFollowedEvent{
		FollowerID:  followerID,
		FollowingID: followingID,
		Timestamp:   time.Now(),
	})

	return nil
}

func (s *FollowService) Unfollow(ctx context.Context, followerID, followingID uuid.UUID) error {
	if followerID == followingID {
		return apperror.ErrSelfFollow
	}

	if err := s.followRepo.Unfollow(ctx, followerID, followingID); err != nil {
		return err
	}

	_ = s.cache.DeleteFollowKeys(ctx, followerID, followingID)

	_ = s.publisher.PublishUserUnfollowed(ctx, &domain.UserUnfollowedEvent{
		FollowerID:  followerID,
		FollowingID: followingID,
	})

	return nil
}

func (s *FollowService) GetFollowers(ctx context.Context, userID uuid.UUID, limit int, cursorStr string) ([]*domain.FollowUser, string, error) {
	if limit <= 0 {
		limit = 20
	}
	cursor, err := pagination.DecodeCursor(cursorStr)
	if err != nil {
		return nil, "", err
	}

	users, err := s.followRepo.GetFollowers(ctx, userID, limit+1, cursor)
	if err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(users) > limit {
		nextCursor = pagination.EncodeCursor(users[limit-1].FollowedAt)
		users = users[:limit]
	}
	return users, nextCursor, nil
}

func (s *FollowService) GetFollowing(ctx context.Context, userID uuid.UUID, limit int, cursorStr string) ([]*domain.FollowUser, string, error) {
	if limit <= 0 {
		limit = 20
	}
	cursor, err := pagination.DecodeCursor(cursorStr)
	if err != nil {
		return nil, "", err
	}

	users, err := s.followRepo.GetFollowing(ctx, userID, limit+1, cursor)
	if err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(users) > limit {
		nextCursor = pagination.EncodeCursor(users[limit-1].FollowedAt)
		users = users[:limit]
	}
	return users, nextCursor, nil
}

func (s *FollowService) CheckFollowStatus(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	// Try cache first
	val, err := s.cache.IsFollowing(ctx, followerID, followingID)
	if err == nil {
		return val, nil
	}

	isFollowing, err := s.followRepo.IsFollowing(ctx, followerID, followingID)
	if err != nil {
		return false, err
	}

	_ = s.cache.SetIsFollowing(ctx, followerID, followingID, isFollowing)
	return isFollowing, nil
}
