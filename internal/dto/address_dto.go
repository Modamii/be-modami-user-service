package dto

type CreateAddressRequest struct {
	Label         string `json:"label" validate:"required,max=50"`
	RecipientName string `json:"recipient_name" validate:"required,max=255"`
	Phone         string `json:"phone" validate:"required"`
	AddressLine1  string `json:"address_line_1" validate:"required,max=512"`
	AddressLine2  string `json:"address_line_2" validate:"omitempty,max=512"`
	Ward          string `json:"ward" validate:"omitempty,max=128"`
	District      string `json:"district" validate:"omitempty,max=128"`
	Province      string `json:"province" validate:"omitempty,max=128"`
	PostalCode    string `json:"postal_code" validate:"omitempty,max=20"`
	Country       string `json:"country" validate:"omitempty,max=64"`
	IsDefault     bool   `json:"is_default"`
}

type UpdateAddressRequest struct {
	Label         *string `json:"label" validate:"omitempty,max=50"`
	RecipientName *string `json:"recipient_name" validate:"omitempty,max=255"`
	Phone         *string `json:"phone" validate:"omitempty"`
	AddressLine1  *string `json:"address_line_1" validate:"omitempty,max=512"`
	AddressLine2  *string `json:"address_line_2" validate:"omitempty,max=512"`
	Ward          *string `json:"ward" validate:"omitempty,max=128"`
	District      *string `json:"district" validate:"omitempty,max=128"`
	Province      *string `json:"province" validate:"omitempty,max=128"`
	PostalCode    *string `json:"postal_code" validate:"omitempty,max=20"`
	Country       *string `json:"country" validate:"omitempty,max=64"`
	IsDefault     *bool   `json:"is_default"`
}

type AddressResponse struct {
	ID            string `json:"id"`
	Label         string `json:"label"`
	RecipientName string `json:"recipient_name"`
	Phone         string `json:"phone"`
	AddressLine1  string `json:"address_line_1"`
	AddressLine2  string `json:"address_line_2,omitempty"`
	Ward          string `json:"ward,omitempty"`
	District      string `json:"district,omitempty"`
	Province      string `json:"province,omitempty"`
	PostalCode    string `json:"postal_code,omitempty"`
	Country       string `json:"country,omitempty"`
	IsDefault     bool   `json:"is_default"`
	CreatedAt     string `json:"created_at"`
}
