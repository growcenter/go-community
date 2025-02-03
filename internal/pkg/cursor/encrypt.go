package cursor

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go-community/internal/models"
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

func EncryptCursorFromStruct(cursor interface{}) string {
	bytes, err := json.Marshal(cursor)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(bytes)
}

func DecryptCursorToStruct(cursor string, target interface{}) (interface{}, error) {
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cursor: %w", err)
	}

	err = json.Unmarshal(decoded, target)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cursor: %w", err)
	}

	return target, nil
}

func DecryptCursorForGetRegisteredRecord(s string) (*models.GetAllRegisteredRecordCursor, error) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cursor: %w", err)
	}

	var cursor models.GetAllRegisteredRecordCursor
	err = json.Unmarshal(decoded, &cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cursor: %w", err)
	}

	return &cursor, nil
}

func DecryptCursorForGetAllUser(s string) (*models.GetAllUserCursor, error) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cursor: %w", err)
	}

	var cursor models.GetAllUserCursor
	err = json.Unmarshal(decoded, &cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cursor: %w", err)
	}

	return &cursor, nil
}
