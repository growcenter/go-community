package cursor

import (
	"encoding/base64"
	"fmt"
	"strconv"
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
		return time.Time{}, fmt.Errorf("failed to decode cursor properly: %w", err)
	}

	// Parse the decoded string as a time
	t, err := time.Parse(time.RFC3339, string(decoded))
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp from cursor: %w", err)
	}

	return t, nil
}

func DecryptCursorFromInteger(cursor string) (int64, error) {
	// Decode from Base64
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return 0, fmt.Errorf("failed to decode cursor: %w", err)
	}

	// Convert the decoded byte slice back to an integer
	idStr := string(decoded)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse cursor as integer: %v", err)
	}

	return id, nil
}

func DecryptCursorFromString(cursor string) (string, error) {
	// Decode from Base64
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return "", fmt.Errorf("failed to decode cursor: %w", err)
	}

	return string(decoded), nil
}
