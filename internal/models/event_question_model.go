package models

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"go-community/internal/constants"
	"time"
)

var TYPE_EVENT_QUESTION = "eventQuestion"

type EventQuestion struct {
	ID                    uuid.UUID              `json:"id"`
	EventCode             string                 `json:"event_code"`
	InstanceCode          *string                `json:"instance_code"` // Optional - if null, it's a general event question
	Question              string                 `json:"question"`
	Description           string                 `json:"description"`
	Type                  constants.QuestionType `json:"type"`
	Options               []string               `json:"options,omitempty"` // For choice questions
	IsMainRequired        bool                   `json:"is_required"`
	IsRegistrantRequired  bool                   `json:"is_registrant_required"`
	DisplayOrder          *int                   `json:"display_order"`
	IsVisibleToRegistrant bool                   `json:"is_visible_to_registrant"`
	Rules                 *QuestionRules         `gorm:"type:jsonb;serializer:json"`
	Status                string                 `json:"status" example:"active"`
	CreatedAt             time.Time              `json:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at"`
	DeletedAt             sql.NullTime           `json:"deleted_at,omitempty"`
}

type QuestionRules struct {
	MinString *int `json:"min_string,omitempty"` // Minimum length for string answers
	MaxString *int `json:"max_string,omitempty"` // Maximum length for string answers
}

type (
	CreateQuestionRequest struct {
		EventCode    string                        `json:"eventCode" validate:"required" example:"event_code_123"`
		InstanceCode *string                       `json:"instanceCode" validate:"omitempty" example:"instance_code_123"` // Optional - if null, it's a general event question
		Questions    []CreateQuestionDetailRequest `json:"questions" validate:"required,dive"`
	}
	CreateQuestionDetailRequest struct {
		Question              string                 `json:"question" validate:"required,min=1,max=255" example:"What is your favorite color?"`
		Description           string                 `json:"description" validate:"omitempty,min=1,max=255" example:"Please select your favorite color from the options below."`
		Type                  constants.QuestionType `json:"type" validate:"required" example:"single_choice"`
		IsMainRequired        bool                   `json:"isMainRequired" validate:"omitempty" example:"true"`
		IsRegistrantRequired  bool                   `json:"isRegistrantRequired" validate:"omitempty" example:"false"`
		Options               []string               `json:"options" validate:"omitempty,dive,min=1,max=255" example:"red,blue,green"`
		DisplayOrder          *int                   `json:"displayOrder" validate:"omitempty" example:"1"`
		IsVisibleToRegistrant bool                   `json:"isVisibleToRegistrant" validate:"omitempty" example:"true"`
		Rules                 *QuestionRules         `json:"rules,omitempty"`
		Status                string                 `json:"status" validate:"required,oneof=active inactive" example:"active"`
	}
	CreateQuestionResponse struct {
		Type                  string                 `json:"type" example:"question"`
		ID                    uuid.UUID              `json:"id" example:"question_id_123"`
		EventCode             string                 `json:"eventCode" example:"event_code_123"`
		InstanceCode          *string                `json:"instanceCode" example:"instance_code_123"` // Optional - if null, it's a general event question
		Question              string                 `json:"question" example:"What is your favorite color?"`
		Description           string                 `json:"description" example:"Please select your favorite color from the options below."`
		QuestionType          constants.QuestionType `json:"questionType" example:"single_choice"`
		IsMainRequired        bool                   `json:"isMainRequired" example:"true"`
		IsRegistrantRequired  bool                   `json:"isRegistrantRequired" example:"false"`
		Options               []string               `json:"options" example:"red,blue,green"`
		DisplayOrder          *int                   `json:"displayOrder" example:"1"`
		IsVisibleToRegistrant bool                   `json:"isVisibleToRegistrant"`
		Rules                 *QuestionRules         `json:"rules,omitempty"`
		Status                string                 `json:"status" example:"active"`
	}

	QuestionResponseForAnswer struct {
		Type         string    `json:"type" example:"question"`
		ID           uuid.UUID `json:"id" example:"question_id_123"`
		Question     string    `json:"question" example:"What is your favorite color?"`
		Description  string    `json:"description" example:"Please select your favorite color from the options below."`
		QuestionType string    `json:"questionType" example:"single_choice"`
		Options      []string  `json:"options,omitempty" example:"red,blue,green"`
	}
)

func CreateQuestionSetup(qType constants.QuestionType, desc string, options []string, rules QuestionRules) (*constants.QuestionType, string, error) {
	var questionType constants.QuestionType
	var description string

	switch {
	case qType == constants.QuestionTypeShortText || qType == constants.QuestionTypeLongText:
		if desc != "" {
			description = desc
		}

		// Validate min and max
		if *rules.MinString < 0 {
			return nil, "", fmt.Errorf("min cannot be below 0")
		}
		if *rules.MaxString < 0 {
			return nil, "", fmt.Errorf("max cannot be below 0")
		}

		// Build description dynamically
		switch {
		case *rules.MinString != 0 && *rules.MaxString != 0:
			description = fmt.Sprintf("minimum of %d maximum of %d", *rules.MinString, *rules.MaxString)
		case *rules.MinString != 0:
			description = fmt.Sprintf("minimum of %d", *rules.MinString)
		case *rules.MaxString != 0:
			description = fmt.Sprintf("maximum of %d", *rules.MaxString)
		default:
			description = desc // No description if both are 0
		}

	case qType == constants.QuestionTypeMultiple:
		if options == nil {
			return nil, "", fmt.Errorf("%w: %s", ErrorCannotBeEmpty, "Options")
		}

	case qType == constants.QuestionTypeSingle:
		if options == nil {
			return nil, "", fmt.Errorf("%w: %s", ErrorCannotBeEmpty, "Options")
		}
	case qType == constants.QuestionTypeEmail:
		if desc == "" {
			description = "Please enter a valid email address. Example: example@mail.com"
		} else {
			description = desc
		}
	case qType == constants.QuestionTypePhone:
		if desc == "" {
			description = "Please enter a valid phone number. Example: +6281234567890"
		} else {
			description = desc
		}
	case qType == constants.QuestionTypeDate:
		if desc == "" {
			description = "Please select a date. Example: 2023-10-01"
		} else {
			description = desc
		}
	case qType == constants.QuestionTypeTime:
		if desc == "" {
			description = "Please select a time. Example: 14:00"
		} else {
			description = desc
		}
	default:
		questionType = qType
		description = desc
	}

	return &questionType, description, nil
}
