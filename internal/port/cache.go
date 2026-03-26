package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/modami/user-service/internal/domain"
)

type CacheService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	SetProfile(ctx context.Context, user *domain.User) error
	DeleteProfile(ctx context.Context, userID uuid.UUID) error

	GetStatus(ctx context.Context, userID uuid.UUID) (domain.UserStatus, error)
	SetStatus(ctx context.Context, userID uuid.UUID, status domain.UserStatus) error

	GetSellerProfile(ctx context.Context, userID uuid.UUID) (*domain.SellerProfile, error)
	SetSellerProfile(ctx context.Context, userID uuid.UUID, profile *domain.SellerProfile) error
	DeleteSellerProfile(ctx context.Context, userID uuid.UUID) error

	GetFollowerCount(ctx context.Context, userID uuid.UUID) (int64, error)
	SetFollowerCount(ctx context.Context, userID uuid.UUID, count int64) error
	GetFollowingCount(ctx context.Context, userID uuid.UUID) (int64, error)
	SetFollowingCount(ctx context.Context, userID uuid.UUID, count int64) error
	IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error)
	SetIsFollowing(ctx context.Context, followerID, followingID uuid.UUID, val bool) error
	DeleteFollowKeys(ctx context.Context, followerID, followingID uuid.UUID) error

	GetRatingSummary(ctx context.Context, userID uuid.UUID) (*domain.RatingSummary, error)
	SetRatingSummary(ctx context.Context, userID uuid.UUID, summary *domain.RatingSummary) error
	DeleteRatingSummary(ctx context.Context, userID uuid.UUID) error

	GetAddresses(ctx context.Context, userID uuid.UUID) ([]*domain.Address, error)
	SetAddresses(ctx context.Context, userID uuid.UUID, addresses []*domain.Address) error
	DeleteAddresses(ctx context.Context, userID uuid.UUID) error

	GetKYCStatus(ctx context.Context, userID uuid.UUID) (domain.KYCStatus, error)
	SetKYCStatus(ctx context.Context, userID uuid.UUID, status domain.KYCStatus) error
	DeleteKYCStatus(ctx context.Context, userID uuid.UUID) error
}
