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

type (
	CreateAnswerRequest struct {
		QuestionId uuid.UUID `json:"questionId" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
		Value      string    `json:"value" validate:"omitempty,min=1,max=255" example:"Red"`
		Values     []string  `json:"values" validate:"omitempty,dive,min=1,max=255" example:"Red,Blue,Green"`
	}
	CreateAnswerResponse struct {
		Type     string                    `json:"type" example:"answer"`
		ID       uuid.UUID                 `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
		Question QuestionResponseForAnswer `json:"question"`
		Value    string                    `json:"value,omitempty" example:"Red"`
		Values   string                    `json:"values,omitempty" example:"Red,Blue,Green"`
	}
)
