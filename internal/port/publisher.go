package port

import (
	"context"

	"github.com/modami/user-service/internal/domain"
)

type EventPublisher interface {
	PublishUserProfileCreated(ctx context.Context, event *domain.UserProfileCreatedEvent) error
	PublishUserUpdated(ctx context.Context, event *domain.UserUpdatedEvent) error
	PublishUserRoleUpgraded(ctx context.Context, event *domain.UserRoleUpgradedEvent) error
	PublishUserSuspended(ctx context.Context, event *domain.UserSuspendedEvent) error
	PublishUserFollowed(ctx context.Context, event *domain.UserFollowedEvent) error
	PublishUserUnfollowed(ctx context.Context, event *domain.UserUnfollowedEvent) error
	PublishUserReviewCreated(ctx context.Context, event *domain.UserReviewCreatedEvent) error
}
