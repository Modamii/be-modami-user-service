package port

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/modami/user-service/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByKeycloakID(ctx context.Context, keycloakID string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	Search(ctx context.Context, query string, limit int, cursor *time.Time) ([]*domain.User, error)
	UpdateTrustScore(ctx context.Context, userID uuid.UUID, score float64) error
	UpdateRole(ctx context.Context, userID uuid.UUID, role domain.UserRole) error
	UpdateStatus(ctx context.Context, userID uuid.UUID, status domain.UserStatus) error
	IncrFollowerCount(ctx context.Context, userID uuid.UUID, delta int) error
	IncrFollowingCount(ctx context.Context, userID uuid.UUID, delta int) error
}

type SellerProfileRepository interface {
	Create(ctx context.Context, profile *domain.SellerProfile) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.SellerProfile, error)
	Update(ctx context.Context, profile *domain.SellerProfile) error
	UpdateKYCStatus(ctx context.Context, userID uuid.UUID, status domain.KYCStatus, verifiedAt *time.Time) error
}

type FollowRepository interface {
	Follow(ctx context.Context, followerID, followingID uuid.UUID) error
	Unfollow(ctx context.Context, followerID, followingID uuid.UUID) error
	IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error)
	GetFollowers(ctx context.Context, userID uuid.UUID, limit int, cursor *time.Time) ([]*domain.FollowUser, error)
	GetFollowing(ctx context.Context, userID uuid.UUID, limit int, cursor *time.Time) ([]*domain.FollowUser, error)
}

type ReviewRepository interface {
	Create(ctx context.Context, review *domain.Review) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) (*domain.Review, error)
	ListByReviewee(ctx context.Context, revieweeID uuid.UUID, limit int, cursor *time.Time) ([]*domain.Review, error)
	GetRatingSummary(ctx context.Context, userID uuid.UUID) (*domain.RatingSummary, error)
	UpsertRatingSummary(ctx context.Context, userID uuid.UUID, rating int) error
}

type AddressRepository interface {
	Create(ctx context.Context, addr *domain.Address) error
	GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Address, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Address, error)
	Update(ctx context.Context, addr *domain.Address) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
	SetDefault(ctx context.Context, id, userID uuid.UUID) error
	CountByUserID(ctx context.Context, userID uuid.UUID) (int, error)
}

type KYCRepository interface {
	Create(ctx context.Context, doc *domain.KYCDocument) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.KYCDocument, error)
	UpdateStatus(ctx context.Context, userID uuid.UUID, status domain.KYCDocStatus, reason string, reviewedBy uuid.UUID) error
}

type OutboxRepository interface {
	Create(ctx context.Context, topic, key string, payload []byte) error
	GetPending(ctx context.Context, limit int) ([]*domain.OutboxEvent, error)
	MarkSent(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID) error
}

type ProcessedEventRepository interface {
	IsProcessed(ctx context.Context, eventID string) (bool, error)
	MarkProcessed(ctx context.Context, eventID, topic string) error
	Cleanup(ctx context.Context, olderThan time.Duration) error
}

type TxManager interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}
