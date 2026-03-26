package dto

type CreateReviewRequest struct {
	OrderID     string `json:"order_id" validate:"required,uuid"`
	Rating      int    `json:"rating" validate:"required,min=1,max=5"`
	Comment     string `json:"comment" validate:"omitempty,max=1000"`
	IsAnonymous bool   `json:"is_anonymous"`
}

type ReviewResponse struct {
	ID          string `json:"id"`
	ReviewerID  string `json:"reviewer_id,omitempty"`
	RevieweeID  string `json:"reviewee_id"`
	OrderID     string `json:"order_id"`
	Rating      int    `json:"rating"`
	Comment     string `json:"comment,omitempty"`
	Role        string `json:"role"`
	IsAnonymous bool   `json:"is_anonymous"`
	CreatedAt   string `json:"created_at"`
}

type ReviewListResponse struct {
	Reviews []ReviewResponse `json:"reviews"`
	Cursor  string           `json:"cursor,omitempty"`
}

type RatingSummaryResponse struct {
	UserID       string  `json:"user_id"`
	AvgRating    float64 `json:"avg_rating"`
	TotalReviews int     `json:"total_reviews"`
	Count1       int     `json:"count_1"`
	Count2       int     `json:"count_2"`
	Count3       int     `json:"count_3"`
	Count4       int     `json:"count_4"`
	Count5       int     `json:"count_5"`
}
