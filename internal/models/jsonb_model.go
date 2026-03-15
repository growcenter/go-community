package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

// JSONB stores raw JSON bytes and can represent any JSON value (object, array, etc.).
// It implements driver.Valuer, sql.Scanner, json.Marshaler, and json.Unmarshaler.
type JSONB json.RawMessage

// Value returns the JSON string to be stored in the database.
// Implements driver.Valuer interface for database/sql.
// NOTE: Must return string (not []byte) — lib/pq treats []byte as bytea,
// which PostgreSQL cannot cast to json/jsonb, causing "invalid input syntax for type json".
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return string(j), nil
}

// Scan reads the raw bytes from the database.
// Implements sql.Scanner interface for database/sql.
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("JSONB: cannot scan type %T into JSONB", value)
	}

	*j = JSONB(b)
	return nil
}

// MarshalJSON implements json.Marshaler interface.
// This ensures proper JSON encoding in HTTP responses.
func (j JSONB) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return j, nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
// This ensures proper JSON decoding from HTTP requests.
func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("JSONB: UnmarshalJSON on nil pointer")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

// Unmarshal decodes the JSONB into a specific Go struct.
// Returns nil if JSONB is nil (no error for null values).
func (j JSONB) Unmarshal(v interface{}) error {
	if j == nil {
		return nil
	}
	return json.Unmarshal(j, v)
}

// Marshal encodes a Go value into JSONB.
// This is a convenience method for setting JSONB values.
func (j *JSONB) Marshal(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("JSONB: failed to marshal: %w", err)
	}
	*j = JSONB(b)
	return nil
}

// IsValid checks if the JSONB contains valid JSON.
// Returns true for nil (null is valid JSON).
func (j JSONB) IsValid() bool {
	if j == nil {
		return true
	}
	return json.Valid(j)
}

// String returns a string representation of the JSONB.
// Useful for debugging and logging.
func (j JSONB) String() string {
	if j == nil {
		return "null"
	}
	return string(j)
}

// Bytes returns the raw bytes of the JSONB.
// Returns nil if JSONB is nil.
func (j JSONB) Bytes() []byte {
	if j == nil {
		return nil
	}
	return []byte(j)
}

// IsNull checks if the JSONB is null.
func (j *JSONB) IsNull() bool {
	return j == nil || len(*j) == 0
}
