package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

var (
	TYPE_FORM_ANSWER = "formAnswer"
)

type FormAnswer struct {
	ID             uuid.UUID    `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Identifier     string       `gorm:"type:varchar(255);not null"`
	IdentifierType string       `gorm:"type:varchar(255);not null"`
	FormCode       *uuid.UUID   `gorm:"type:uuid"`
	QuestionCode   uuid.UUID    `gorm:"type:uuid;not null"`
	Answer         string       `gorm:"type:text;not null"`
	IsCorrect      sql.NullBool `gorm:"type:boolean"`
	SubmittedAt    time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      sql.NullTime
}

type AnswerItem struct {
	QuestionCode string `json:"questionCode" validate:"required,uuid"`
	Answer       string `json:"answer" validate:"required"`
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
	Question QuestionsResponse `json:"question"`
	Answer   AnswerResponse    `json:"answer"`
}

type (
	CreateFormAnswerRequest struct {
		FormCode       string                     `json:"formCode" validate:"omitempty,uuid,required_without=Entity"`
		Entity         []FormQuestionEntityFilter `json:"entity" validate:"dive,required_without=FormCode"`
		Identifier     string                     `json:"identifier" validate:"required"`
		IdentifierType string                     `json:"identifierType" validate:"required,oneof=eventAttendance communityId"`
		IsParent       bool                       `json:"isParent"`
		Answers        []AnswerItem               `json:"answers" validate:"required,dive"`
	}

	CreateFormAnswerResponse struct {
		FormCode       string                       `json:"formCode"`
		Identifier     string                       `json:"identifier"`
		IdentifierType string                       `json:"identifierType"`
		SubmittedAt    time.Time                    `json:"submittedAt"`
		Forms          []FormQuestionAnswerResponse `json:"forms"`
	}
)
