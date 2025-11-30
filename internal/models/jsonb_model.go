package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONB stores raw JSON bytes and can represent any JSON value (object, array, etc.).
type JSONB json.RawMessage

// Value returns the raw bytes to be stored in the database.
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return []byte(j), nil
}

// Scan reads the raw bytes from the database.
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	*j = JSONB(b)
	return nil
}

// Unmarshal allows you to decode the raw JSON into a specific Go struct.
func (j JSONB) Unmarshal(v interface{}) error {
	if j == nil {
		return nil
	}
	return json.Unmarshal(j, v)
}