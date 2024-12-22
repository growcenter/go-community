package cursor

import (
	"encoding/base64"
	"fmt"
	"time"
)

// EncryptCursor encodes the timestamp as a Base64 string
func EncryptCursor(timestamp string) (string, error) {
	// Encode the timestamp as Base64
	encoded := base64.StdEncoding.EncodeToString([]byte(timestamp))
	return encoded, nil
}

// DecryptCursor decodes the Base64 string back to a timestamp string
func DecryptCursor(cursor string) (time.Time, error) {
	// Decode from Base64
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to decode cursor: %w", err)
	}

	// Parse the decoded string as a time
	t, err := time.Parse(time.RFC3339, string(decoded))
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp from cursor: %w", err)
	}

	return t, nil
}
