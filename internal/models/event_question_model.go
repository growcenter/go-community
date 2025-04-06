package models

import (
	"database/sql"
	"github.com/google/uuid"
	"go-community/internal/constants"
	"time"
)

type Question struct {
	ID                   string                 `json:"id"`
	EventCode            string                 `json:"event_code"`
	InstanceCode         *string                `json:"instance_code"` // Optional - if null, it's a general event question
	Question             string                 `json:"question"`
	Description          string                 `json:"description"`
	Type                 constants.QuestionType `json:"type"`
	Options              []string               `json:"options,omitempty"` // For choice questions
	IsMainRequired       bool                   `json:"is_required"`
	IsRegistrantRequired bool                   `json:"is_registrant_required"`
	DisplayOrder         *int                   `json:"display_order"`
	Status               string                 `json:"status" example:"active"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
	DeletedAt            sql.NullTime           `json:"deleted_at,omitempty"`
}

type (
	CreateQuestionRequest struct {
		EventCode            string   `json:"eventCode" validate:"required" example:"event_code_123"`
		InstanceCode         *string  `json:"instanceCode" validate:"omitempty" example:"instance_code_123"` // Optional - if null, it's a general event question
		Question             string   `json:"question" validate:"required,min=1,max=255" example:"What is your favorite color?"`
		Description          string   `json:"description" validate:"omitempty,min=1,max=255" example:"Please select your favorite color from the options below."`
		Type                 string   `json:"type" validate:"required" example:"single_choice"`
		IsMainRequired       bool     `json:"isMainRequired" validate:"omitempty" example:"true"`
		IsRegistrantRequired bool     `json:"isRegistrantRequired" validate:"omitempty" example:"false"`
		Options              []string `json:"options" validate:"omitempty,dive,min=1,max=255" example:"red,blue,green"`
		DisplayOrder         *int     `json:"displayOrder" validate:"omitempty" example:"1"`
		Status               string   `json:"status" validate:"required,oneof=active inactive" example:"active"`
	}
	CreateQuestionResponse struct {
		Type                 string    `json:"type" example:"question"`
		ID                   uuid.UUID `json:"id" example:"question_id_123"`
		EventCode            string    `json:"eventCode" example:"event_code_123"`
		InstanceCode         *string   `json:"instanceCode" example:"instance_code_123"` // Optional - if null, it's a general event question
		Question             string    `json:"question" example:"What is your favorite color?"`
		Description          string    `json:"description" example:"Please select your favorite color from the options below."`
		QuestionType         string    `json:"questionType" example:"single_choice"`
		IsMainRequired       bool      `json:"isMainRequired" example:"true"`
		IsRegistrantRequired bool      `json:"isRegistrantRequired" example:"false"`
		Options              []string  `json:"options" example:"red,blue,green"`
		DisplayOrder         *int      `json:"displayOrder" example:"1"`
		Status               string    `json:"status" example:"active"`
	}
)
