package domain

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID          uuid.UUID  `json:"id"`
	ReviewerID  uuid.UUID  `json:"reviewer_id"`
	RevieweeID  uuid.UUID  `json:"reviewee_id"`
	OrderID     uuid.UUID  `json:"order_id"`
	Rating      int        `json:"rating"`
	Comment     string     `json:"comment"`
	Role        ReviewRole `json:"role"`
	IsAnonymous bool       `json:"is_anonymous"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type RatingSummary struct {
	UserID       uuid.UUID `json:"user_id"`
	AvgRating    float64   `json:"avg_rating"`
	TotalReviews int       `json:"total_reviews"`
	Count1       int       `json:"count_1"`
	Count2       int       `json:"count_2"`
	Count3       int       `json:"count_3"`
	Count4       int       `json:"count_4"`
	Count5       int       `json:"count_5"`
}
