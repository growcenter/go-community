package models

import (
	"database/sql"
	"github.com/google/uuid"
	"time"
)

type Answer struct {
	ID             uuid.UUID    `json:"id"`
	QuestionID     uuid.UUID    `json:"question_id"`
	RegistrationId uuid.UUID    `json:"registration_id"`
	Value          string       `json:"value"`  // For single value answers (text, radio, select)
	Values         []string     `json:"values"` // For multiple value answers (checkboxes)
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
	DeletedAt      sql.NullTime `json:"deleted_at,omitempty"`
}
