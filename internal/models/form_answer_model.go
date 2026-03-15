package models

import (
	"time"

	"github.com/google/uuid"
)

type FormAnswer struct {
	ID             int        `gorm:"type:integer;primarykey"`
	Code           uuid.UUID  `gorm:"type:uuid;not null"`
	FormCode       uuid.UUID  `gorm:"type:uuid;not null"`
	QuestionCode   uuid.UUID  `gorm:"type:uuid;not null"`
	IdentifierType string     `gorm:"type:varchar(50);not null"`
	IdentifierCode string     `gorm:"type:varchar(50);not null"`
	Answer         string     `gorm:"type:text;not null"`
	IsCorrect      *bool      `gorm:"type:boolean"`
	Status         string     `gorm:"type:varchar(50);not null"`
	SubmittedAt    time.Time  `gorm:"type:timestamptz;default:now()"`
	CreatedAt      time.Time  `gorm:"type:timestamptz;default:now()"`
	UpdatedAt      time.Time  `gorm:"type:timestamptz;default:now()"`
	DeletedAt      *time.Time `gorm:"type:timestamptz"`
}

// AnswerItem represents a single question–answer pair in a submission request.
// Answer is intentionally not marked required: optional questions may be skipped.
// The usecase validates mandatory questions and per-type rules.
type AnswerItem struct {
	QuestionCode string `json:"questionCode" validate:"required,uuid4"`
	Answer       string `json:"answer"`
}

type AnswerResponse struct {
	Type           string     `json:"type"`
	Code           string     `json:"code"`
	IdentifierType string     `json:"identifierType"`
	Identifier     string     `json:"identifier"`
	FormCode       *uuid.UUID `json:"formCode,omitempty"`
	QuestionCode   uuid.UUID  `json:"questionCode"`
	Answer         string     `json:"answer"`
	IsCorrect      *bool      `json:"isCorrect,omitempty"`
	SubmittedAt    time.Time  `json:"submittedAt"`
}

type FormQuestionAnswerResponse struct {
	Question FormQuestionResponse `json:"question"`
	Answer   AnswerResponse       `json:"answer"`
}

type (
	// CreateFormAnswerRequest represents a full form submission.
	// Exactly one of FormCode or Entity must be provided.
	// FormCode identifies a specific form; Entity resolves associated forms dynamically.
	CreateFormAnswerRequest struct {
		FormCode       string                     `json:"formCode" validate:"required_without=Entity,omitempty,uuid4"`
		Entity         []FormQuestionEntityFilter `json:"entity"   validate:"required_without=FormCode,omitempty,dive"`
		Identifier     string                     `json:"identifier"     validate:"required"`
		IdentifierType string                     `json:"identifierType" validate:"required,oneof=eventAttendance communityId"`
		IsParent       bool                       `json:"isParent"`
		Answers        []AnswerItem               `json:"answers" validate:"required,min=1,dive"`
	}

	CreateFormAnswerResponse struct {
		FormCode       string                       `json:"formCode"`
		Identifier     string                       `json:"identifier"`
		IdentifierType string                       `json:"identifierType"`
		SubmittedAt    time.Time                    `json:"submittedAt"`
		Forms          []FormQuestionAnswerResponse `json:"forms"`
	}
)
