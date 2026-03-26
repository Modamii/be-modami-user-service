package pagination

import (
	"encoding/base64"
	"fmt"
	"time"
)

// EncodeCursor encodes a time.Time into a base64 cursor string.
func EncodeCursor(t time.Time) string {
	val := fmt.Sprintf("%d", t.UnixNano())
	return base64.StdEncoding.EncodeToString([]byte(val))
}

// DecodeCursor decodes a base64 cursor string into a time.Time pointer.
func DecodeCursor(cursor string) (*time.Time, error) {
	if cursor == "" {
		return nil, nil
	}
	b, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor: %w", err)
	}
	var nanos int64
	_, err = fmt.Sscanf(string(b), "%d", &nanos)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor value: %w", err)
	}
	t := time.Unix(0, nanos).UTC()
	return &t, nil
}
