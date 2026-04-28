package pagination

import (
	"encoding/base64"
	"fmt"
	"time"
)

// EncodeCursor encodes a time.Time into a base64 URL-safe cursor string.
func EncodeCursor(t time.Time) string {
	return base64.URLEncoding.EncodeToString([]byte(t.UTC().Format(time.RFC3339Nano)))
}

// DecodeCursor decodes a base64 cursor string back to a *time.Time.
// Returns nil, nil when the cursor is empty (first page).
func DecodeCursor(cursor string) (*time.Time, error) {
	if cursor == "" {
		return nil, nil
	}
	b, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor: %w", err)
	}
	t, err := time.Parse(time.RFC3339Nano, string(b))
	if err != nil {
		return nil, fmt.Errorf("invalid cursor: %w", err)
	}
	return &t, nil
}
