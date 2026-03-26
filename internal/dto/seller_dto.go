package dto

type RegisterSellerRequest struct {
	ShopName        string `json:"shop_name" validate:"required,min=3,max=255"`
	ShopSlug        string `json:"shop_slug" validate:"required,min=3,max=255,alphanum"`
	ShopDescription string `json:"shop_description" validate:"omitempty,max=1000"`
	BusinessType    string `json:"business_type" validate:"required,oneof=individual business"`
	TaxID           string `json:"tax_id" validate:"omitempty"`
	BankAccount     string `json:"bank_account" validate:"omitempty"`
	BankName        string `json:"bank_name" validate:"omitempty,max=128"`
}

type UpdateSellerProfileRequest struct {
	ShopName        *string `json:"shop_name" validate:"omitempty,min=3,max=255"`
	ShopDescription *string `json:"shop_description" validate:"omitempty,max=1000"`
	ShopLogoURL     *string `json:"shop_logo_url" validate:"omitempty,url"`
	ShopBannerURL   *string `json:"shop_banner_url" validate:"omitempty,url"`
	TaxID           *string `json:"tax_id" validate:"omitempty"`
	BankAccount     *string `json:"bank_account" validate:"omitempty"`
	BankName        *string `json:"bank_name" validate:"omitempty,max=128"`
}

type SellerProfileResponse struct {
	UserID          string  `json:"user_id"`
	ShopName        string  `json:"shop_name"`
	ShopSlug        string  `json:"shop_slug"`
	ShopDescription string  `json:"shop_description,omitempty"`
	ShopLogoURL     string  `json:"shop_logo_url,omitempty"`
	ShopBannerURL   string  `json:"shop_banner_url,omitempty"`
	BusinessType    string  `json:"business_type"`
	KYCStatus       string  `json:"kyc_status"`
	AvgRating       float64 `json:"avg_rating"`
	TotalReviews    int     `json:"total_reviews"`
	CreatedAt       string  `json:"created_at"`
}

type SubmitKYCRequest struct {
	Documents []KYCDocumentInput `json:"documents" validate:"required,min=1,dive"`
}

type KYCDocumentInput struct {
	DocType string `json:"doc_type" validate:"required,oneof=id_card_front id_card_back business_license selfie_with_id"`
	DocURL  string `json:"doc_url" validate:"required,url"`
}

type RejectKYCRequest struct {
	Reason string `json:"reason" validate:"required,min=1,max=500"`
}

type KYCStatusResponse struct {
	Status    string `json:"status"`
	UpdatedAt string `json:"updated_at,omitempty"`
}
