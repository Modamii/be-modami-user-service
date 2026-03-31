package dto

type UpdateProfileRequest struct {
	FullName    *string `json:"full_name" validate:"omitempty,min=1,max=255"`
	Phone       *string `json:"phone" validate:"omitempty"`
	Bio         *string `json:"bio" validate:"omitempty,max=500"`
	Gender      *string `json:"gender" validate:"omitempty,oneof=male female other undisclosed"`
	DateOfBirth *string `json:"date_of_birth" validate:"omitempty"`
}

type UpdateAvatarRequest struct {
	AvatarURL string `json:"avatar_url" validate:"required,url"`
}

type UpdateCoverRequest struct {
	CoverURL string `json:"cover_url" validate:"required,url"`
}

type UserProfileResponse struct {
	ID             string  `json:"id"`
	Email          string  `json:"email"`
	FullName       string  `json:"full_name"`
	Phone          string  `json:"phone,omitempty"`
	AvatarURL      string  `json:"avatar_url,omitempty"`
	CoverURL       string  `json:"cover_url,omitempty"`
	Bio            string  `json:"bio,omitempty"`
	Gender         string  `json:"gender"`
	DateOfBirth    *string `json:"date_of_birth,omitempty"`
	Role           string  `json:"role"`
	Status         string  `json:"status"`
	TrustScore     float64 `json:"trust_score"`
	FollowerCount  int     `json:"follower_count"`
	FollowingCount int     `json:"following_count"`
	EmailVerified  bool    `json:"email_verified"`
	CreatedAt      string  `json:"created_at"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active inactive suspended banned"`
	Reason string `json:"reason" validate:"omitempty,max=500"`
}

type SearchUsersRequest struct {
	Query  string `form:"q" validate:"required,min=1"`
	Limit  int    `form:"limit" validate:"omitempty,min=1,max=100"`
	Cursor string `form:"cursor"`
}
