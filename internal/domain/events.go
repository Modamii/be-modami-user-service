package domain

import (
	"time"

	"github.com/google/uuid"
)

// Consumed events (from auth service)
type UserRegisteredEvent struct {
	EventID    string    `json:"event_id"`
	KeycloakID string    `json:"keycloak_id"`
	Email      string    `json:"email"`
	FullName   string    `json:"full_name"`
	Phone      string    `json:"phone"`
	CreatedAt  time.Time `json:"created_at"`
}

type UserDeletedEvent struct {
	EventID    string `json:"event_id"`
	KeycloakID string `json:"keycloak_id"`
}

type UserEmailVerifiedEvent struct {
	EventID    string    `json:"event_id"`
	KeycloakID string    `json:"keycloak_id"`
	VerifiedAt time.Time `json:"verified_at"`
}

// Produced events
type UserProfileCreatedEvent struct {
	UserID     uuid.UUID  `json:"user_id"`
	KeycloakID string     `json:"keycloak_id"`
	Role       UserRole   `json:"role"`
	Status     UserStatus `json:"status"`
}

type UserUpdatedEvent struct {
	UserID        uuid.UUID              `json:"user_id"`
	ChangedFields map[string]interface{} `json:"changed_fields"`
}

type UserRoleUpgradedEvent struct {
	UserID  uuid.UUID `json:"user_id"`
	OldRole UserRole  `json:"old_role"`
	NewRole UserRole  `json:"new_role"`
}

type UserSuspendedEvent struct {
	UserID      uuid.UUID `json:"user_id"`
	Reason      string    `json:"reason"`
	SuspendedAt time.Time `json:"suspended_at"`
}

type UserFollowedEvent struct {
	FollowerID  uuid.UUID `json:"follower_id"`
	FollowingID uuid.UUID `json:"following_id"`
	Timestamp   time.Time `json:"timestamp"`
}

type UserUnfollowedEvent struct {
	FollowerID  uuid.UUID `json:"follower_id"`
	FollowingID uuid.UUID `json:"following_id"`
}

type UserReviewCreatedEvent struct {
	ReviewerID uuid.UUID `json:"reviewer_id"`
	RevieweeID uuid.UUID `json:"reviewee_id"`
	OrderID    uuid.UUID `json:"order_id"`
	Rating     int       `json:"rating"`
}
