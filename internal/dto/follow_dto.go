package dto

type FollowStatusResponse struct {
	IsFollowing bool `json:"is_following"`
}

type FollowListResponse struct {
	Users  []FollowUserItem `json:"users"`
	Cursor string           `json:"cursor,omitempty"`
}

type FollowUserItem struct {
	ID         string `json:"id"`
	FullName   string `json:"full_name"`
	AvatarURL  string `json:"avatar_url,omitempty"`
	FollowedAt string `json:"followed_at"`
}
