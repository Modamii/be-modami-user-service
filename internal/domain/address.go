package domain

import (
	"time"

	"github.com/google/uuid"
)

type Address struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	Label         string    `json:"label"`
	RecipientName string    `json:"recipient_name"`
	Phone         string    `json:"phone"`
	AddressLine1  string    `json:"address_line_1"`
	AddressLine2  string    `json:"address_line_2"`
	Ward          string    `json:"ward"`
	District      string    `json:"district"`
	Province      string    `json:"province"`
	PostalCode    string    `json:"postal_code"`
	Country       string    `json:"country"`
	IsDefault     bool      `json:"is_default"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
