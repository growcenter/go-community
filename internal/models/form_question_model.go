package models

import (
	"database/sql"
	"go-community/internal/constants"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// QuestionOptions holds the choices for multiple/single choice questions.
type QuestionOptions struct {
	Choices []string `json:"choices"`
}

// QuestionValidationRules holds the validation rules for a question.
type QuestionValidationRules struct {
	MinSelection *int    `json:"minSelection,omitempty"`
	MaxSelection *int    `json:"maxSelection,omitempty"`
	MinLength    *int    `json:"minLength,omitempty"`
	MaxLength    *int    `json:"maxLength,omitempty"`
	NotBefore    *string `json:"notBefore,omitempty"` // e.g. "today", or a specific date "2025-10-01"
	NotAfter     *string `json:"notAfter,omitempty"`
	MinValue     *int    `json:"minValue,omitempty"`
	MaxValue     *int    `json:"maxValue,omitempty"`
	Pattern      *string `json:"pattern,omitempty"`
}

// FormQuestion represents an individual question in a registration form.
type FormQuestion struct {
	ID            string
	Code          string                   `gorm:"type:uuid;not null" json:"question_id"`
	FormCode      string                   `gorm:"type:uuid;not null" json:"form_id"`
	Text          string                   `gorm:"type:text;not null" json:"question_text"`
	Type          string                   `gorm:"type:varchar(255);not null" json:"question_type"`
	MandatoryFor  pq.StringArray           `gorm:"type:text[]" json:"mandatory_for"`
	ApplyFor      pq.StringArray           `gorm:"type:text[]" json:"apply_for"`
	Options       *QuestionOptions         `gorm:"type:jsonb" json:"options"`
	Rules         *QuestionValidationRules `gorm:"type:jsonb" json:"rules"`
	CorrectAnswer sql.NullString           `gorm:"type:text" json:"correctAnswer,omitempty"`
	DisplayOrder  int                      `gorm:"not null;default:0" json:"display_order"`
	CreatedAt     time.Time                `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt     time.Time                `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt     gorm.DeletedAt           `gorm:"index" json:"deleted_at"`
}

type (
	BulkCreateFormQuestionItem struct {
		Text          string                   `json:"text" validate:"required"`
		QuestionType  constants.QuestionType   `json:"type" validate:"required,oneof=short_text long_text single_choice multiple_choice date time email phone number"`
		MandatoryFor  []string                 `json:"mandatoryFor" validate:"required,dive,oneof=parent child"`
		ApplyFor      []string                 `json:"applyFor" validate:"required,dive,oneof=parent child"`
		Options       *QuestionOptions         `json:"options"`
		Rules         *QuestionValidationRules `json:"rules"`
		CorrectAnswer *string                  `json:"correctAnswer,omitempty"`
		DisplayOrder  int                      `json:"displayOrder" validate:"omitempty,numeric"`
	}
	BulkCreateFormQuestionRequest struct {
		FormID    string                       `json:"formId" validate:"required,uuid"`
		Questions []BulkCreateFormQuestionItem `json:"questions" validate:"required,dive"`
	}

	FormQuestionResponse struct {
		Code          string                   `json:"code"`
		FormCode      string                   `json:"formCode"`
		Text          string                   `json:"text"`
		Type          string                   `json:"type"`
		MandatoryFor  []string                 `json:"mandatoryFor"`
		ApplyFor      []string                 `json:"applyFor"`
		Options       *QuestionOptions         `json:"options"`
		Rules         *QuestionValidationRules `json:"rules"`
		CorrectAnswer *string                  `json:"correctAnswer,omitempty"`
		DisplayOrder  int                      `json:"displayOrder"`
	}
)

func (fq *FormQuestion) ToResponse() *FormQuestionResponse {
	var correctAnswer *string
	if fq.CorrectAnswer.Valid {
		correctAnswer = &fq.CorrectAnswer.String
	}

	return &FormQuestionResponse{
		Code:          fq.Code,
		FormCode:      fq.FormCode,
		Text:          fq.Text,
		Type:          fq.Type,
		MandatoryFor:  fq.MandatoryFor,
		ApplyFor:      fq.ApplyFor,
		Options:       fq.Options,
		Rules:         fq.Rules,
		CorrectAnswer: correctAnswer,
		DisplayOrder:  fq.DisplayOrder,
	}
}
