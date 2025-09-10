package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type FormAnswer struct {
	ID           uuid.UUID    `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CommunityID  string       `gorm:"type:varchar(255);not null"`
	FormCode     string       `gorm:"type:uuid;not null"`
	QuestionCode string       `gorm:"type:uuid;not null"`
	Answer       string       `gorm:"type:text;not null"`
	IsCorrect    sql.NullBool `gorm:"type:boolean"`
	SubmittedAt  time.Time
}

type AnswerItem struct {
	QuestionCode string `json:"questionCode" validate:"required"`
	Answer       string `json:"answer" validate:"required"`
}

type AnswerResponseItem struct {
	QuestionCode string `json:"questionCode"`
	Answer       string `json:"answer"`
	IsCorrect    *bool  `json:"isCorrect,omitempty"`
}

type CreateFormAnswerRequest struct {
	FormCode    string       `json:"formCode" validate:"required,uuid"`
	CommunityID string       `json:"communityId" validate:"required"`
	Answers     []AnswerItem `json:"answers" validate:"required,dive"`
}

type CreateFormAnswerResponse struct {
	FormCode    string               `json:"formCode"`
	CommunityID string               `json:"communityId"`
	SubmittedAt time.Time            `json:"submittedAt"`
	Answers     []AnswerResponseItem `json:"answers"`
}
