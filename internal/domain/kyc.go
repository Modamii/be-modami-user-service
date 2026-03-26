package domain

import (
	"time"

	"github.com/google/uuid"
)

type KYCDocument struct {
	ID         uuid.UUID    `json:"id"`
	UserID     uuid.UUID    `json:"user_id"`
	DocType    KYCDocType   `json:"doc_type"`
	DocURL     string       `json:"doc_url"`
	Status     KYCDocStatus `json:"status"`
	Reason     string       `json:"reason,omitempty"`
	ReviewedBy *uuid.UUID   `json:"reviewed_by,omitempty"`
	ReviewedAt *time.Time   `json:"reviewed_at,omitempty"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

type OutboxEvent struct {
	ID        uuid.UUID  `json:"id"`
	Topic     string     `json:"topic"`
	Key       string     `json:"key"`
	Payload   []byte     `json:"payload"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	SentAt    *time.Time `json:"sent_at,omitempty"`
}
