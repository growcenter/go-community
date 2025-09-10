package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONB implements the json.Marshaler and sql.Scanner interfaces
// to allow for flexible JSON data to be stored in a JSONB column.
type JSONB map[string]interface{}

// Value returns the JSON-encoded representation of the map.
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		// store as empty object instead of null
		return "{}", nil
	}
	return json.Marshal(j)
}

// Scan decodes a JSON-encoded value into the map.
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = JSONB{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &j)
}
