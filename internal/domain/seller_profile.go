package domain

import (
	"time"

	"github.com/google/uuid"
)

type SellerProfile struct {
	ID              uuid.UUID    `json:"id"`
	UserID          uuid.UUID    `json:"user_id"`
	ShopName        string       `json:"shop_name"`
	ShopSlug        string       `json:"shop_slug"`
	ShopDescription string       `json:"shop_description"`
	ShopLogoURL     string       `json:"shop_logo_url"`
	ShopBannerURL   string       `json:"shop_banner_url"`
	BusinessType    BusinessType `json:"business_type"`
	TaxID           string       `json:"tax_id"`
	BankAccount     string       `json:"bank_account"`
	BankName        string       `json:"bank_name"`
	KYCStatus       KYCStatus    `json:"kyc_status"`
	KYCVerifiedAt   *time.Time   `json:"kyc_verified_at,omitempty"`
	AvgRating       float64      `json:"avg_rating"`
	TotalReviews    int          `json:"total_reviews"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}
