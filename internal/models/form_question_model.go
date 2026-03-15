package models

import (
	"go-community/internal/constants"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// QuestionOptions holds the choices for multiple/single choice questions.
type QuestionOptions struct {
	Choices []string `json:"choices" validate:"omitempty,min=1,dive,min=1"`
}

// FormQuestion is the DB model for a form question.
type FormQuestion struct {
	ID            int            `gorm:"type:integer;primarykey"`
	Code          uuid.UUID      `gorm:"type:uuid;not null"`
	FormCode      uuid.UUID      `gorm:"type:uuid;not null"`
	Text          string         `gorm:"type:text;not null"`
	Category      string         `gorm:"type:varchar(50);not null"`
	RequiredFor   pq.StringArray `gorm:"type:text[];default:ARRAY[]::TEXT[]"`
	VisibleFor    pq.StringArray `gorm:"type:text[];default:ARRAY[]::TEXT[]"`
	Options       JSONB          `gorm:"type:jsonb" swaggertype:"object"`
	Rules         JSONB          `gorm:"type:jsonb" swaggertype:"object"`
	CorrectAnswer string         `gorm:"type:text"`
	DisplayOrder  int            `gorm:"type:integer;not null"`
	CreatedAt     time.Time      `gorm:"type:timestamptz;default:now()"`
	UpdatedAt     time.Time      `gorm:"type:timestamptz;default:now()"`
	DeletedAt     *time.Time     `gorm:"type:timestamptz"`
}

// EntityFilter is used to specify an entity for filtering form questions.
type FormQuestionEntityFilter struct {
	Code string
	Type string
}

// QuestionValidationRules holds the validation rules for a question.
type QuestionValidationRules struct {
	MinSelection *int    `json:"minSelection,omitempty"`
	MaxSelection *int    `json:"maxSelection,omitempty"`
	MinLength    *int    `json:"minLength,omitempty"`
	MaxLength    *int    `json:"maxLength,omitempty"`
	NotBefore    *string `json:"notBefore,omitempty"` // "today" or a date like "2025-10-01"
	NotAfter     *string `json:"notAfter,omitempty"`
	MinValue     *int    `json:"minValue,omitempty"`
	MaxValue     *int    `json:"maxValue,omitempty"`
	Pattern      *string `json:"pattern,omitempty"` // Regex applied to text/phone/email answers
}

type (
	BulkCreateFormQuestionRequest struct {
		Text          string                   `json:"text" validate:"required"`
		QuestionType  constants.QuestionType   `json:"type" validate:"required,questionType"`
		RequiredFor   []string                 `json:"requiredFor" validate:"required,dive,oneof=parent child"`
		VisibleFor    []string                 `json:"visibleFor" validate:"required,dive,oneof=parent child"`
		Options       *QuestionOptions         `json:"options" validate:"omitempty"`
		Rules         *QuestionValidationRules `json:"rules"`
		CorrectAnswer *string                  `json:"correctAnswer,omitempty"`
		DisplayOrder  int                      `json:"displayOrder" validate:"omitempty,min=0"`
	}

	FormQuestionResponse struct {
		Type          string                   `json:"type"`
		Code          string                   `json:"code"`
		FormCode      string                   `json:"formCode"`
		Text          string                   `json:"text"`
		QuestionType  string                   `json:"questionType"`
		RequiredFor   []string                 `json:"requiredFor"`
		VisibleFor    []string                 `json:"visibleFor"`
		Options       *QuestionOptions         `json:"options"`
		Rules         *QuestionValidationRules `json:"rules"`
		CorrectAnswer *string                  `json:"correctAnswer,omitempty"`
		DisplayOrder  int                      `json:"displayOrder"`
	}
)

func (fq *FormQuestion) ToResponse() *FormQuestionResponse {
	var options *QuestionOptions
	if !fq.Options.IsNull() {
		_ = fq.Options.Unmarshal(&options)
	}

	var rules *QuestionValidationRules
	if !fq.Rules.IsNull() {
		_ = fq.Rules.Unmarshal(&rules)
	}

	var correctAnswer *string
	if fq.CorrectAnswer != "" {
		correctAnswer = &fq.CorrectAnswer
	}

	return &FormQuestionResponse{
		Type:          "form_question",
		Code:          fq.Code.String(),
		FormCode:      fq.FormCode.String(),
		Text:          fq.Text,
		QuestionType:  fq.Category,
		RequiredFor:   fq.RequiredFor,
		VisibleFor:    fq.VisibleFor,
		Options:       options,
		Rules:         rules,
		CorrectAnswer: correctAnswer,
		DisplayOrder:  fq.DisplayOrder,
	}
}
