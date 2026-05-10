package service

import (
	"context"
	"encoding/json"
	"time"

	"be-modami-user-service/internal/domain"
	"be-modami-user-service/internal/port"
	apperror "be-modami-user-service/pkg/apperror"
	"be-modami-user-service/pkg/pagination"

	"github.com/google/uuid"
)

type FollowService struct {
	followRepo port.FollowRepository
	txManager  port.TxManager
	outboxRepo port.OutboxRepository
}

func NewFollowService(
	followRepo port.FollowRepository,
	txManager port.TxManager,
	outboxRepo port.OutboxRepository,
) *FollowService {
	return &FollowService{
		followRepo: followRepo,
		txManager:  txManager,
		outboxRepo: outboxRepo,
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

	payload, err := json.Marshal(&domain.UserFollowedEvent{
		FollowerID:  followerID,
		FollowingID: followingID,
		Timestamp:   time.Now(),
	})
	if err != nil {
		return err
	}

	return s.txManager.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.followRepo.Follow(ctx, followerID, followingID); err != nil {
			return err
		}
		return s.outboxRepo.Create(ctx, domain.OutboxAggregateFollow, followerID.String(), domain.OutboxEventUserFollowed, payload)
	})
}

func (s *FollowService) Unfollow(ctx context.Context, followerID, followingID uuid.UUID) error {
	if followerID == followingID {
		return apperror.ErrSelfFollow
	}

	payload, err := json.Marshal(&domain.UserUnfollowedEvent{
		FollowerID:  followerID,
		FollowingID: followingID,
	})
	if err != nil {
		return err
	}

	return s.txManager.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.followRepo.Unfollow(ctx, followerID, followingID); err != nil {
			return err
		}
		return s.outboxRepo.Create(ctx, domain.OutboxAggregateFollow, followerID.String(), domain.OutboxEventUserUnfollowed, payload)
	})
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
	return s.followRepo.IsFollowing(ctx, followerID, followingID)
}
