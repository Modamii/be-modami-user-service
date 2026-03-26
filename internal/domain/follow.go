package domain

import (
	"time"

	"github.com/google/uuid"
)

type Follow struct {
	FollowerID  uuid.UUID `json:"follower_id"`
	FollowingID uuid.UUID `json:"following_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type FollowUser struct {
	ID         uuid.UUID `json:"id"`
	FullName   string    `json:"full_name"`
	AvatarURL  string    `json:"avatar_url"`
	FollowedAt time.Time `json:"followed_at"`
}
