package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Form struct {
	ID          int
	Code        uuid.UUID
	Name        string
	Description string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   sql.NullTime
}

type (
	CreateFormRequest struct {
		Name        string                       `json:"name" validate:"required"`
		Description string                       `json:"description"`
		Entity      FormEntityRequest            `json:"entity" validate:"required"`
		Questions   []BulkCreateFormQuestionItem `json:"questions"`
	}
	FormEntityRequest struct {
		Type string `json:"type" validate:"required"`
		Code string `json:"code" validate:"required"`
	}
	CreateFormResponse struct {
		Type               string                 `json:"type"`
		Code               string                 `json:"code"`
		Name               string                 `json:"name"`
		Description        string                 `json:"description"`
		FormEntityResponse FormEntityResponse     `json:"entity"`
		Status             string                 `json:"status"`
		Questions          []FormQuestionResponse `json:"questions"`
	}
	FormEntityResponse struct {
		Type string `json:"type"`
		Code string `json:"code"`
	}
)
