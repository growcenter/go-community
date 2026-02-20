package models

import (
	"time"

	"github.com/google/uuid"
)

type Form struct {
	ID                 int        `gorm:"type:integer;primarykey"`
	Code               uuid.UUID  `gorm:"type:uuid;not null"`
	Name               string     `gorm:"type:varchar(255);not null"`
	Description        string     `gorm:"type:text"`
	FormType           string     `gorm:"type:varchar(50);not null"`
	IsTemplate         bool       `gorm:"type:boolean;default:false"`
	CreatorCommunityID string     `gorm:"type:varchar(50);not null"`
	Status             string     `gorm:"type:varchar(50);not null"`
	CreatedAt          time.Time  `gorm:"type:timestamptz;default:now()"`
	UpdatedAt          time.Time  `gorm:"type:timestamptz;default:now()"`
	DeletedAt          *time.Time `gorm:"type:timestamptz"`
}

type (
	CreateFormRequest struct {
		Name        string                          `json:"name" validate:"required"`
		Description string                          `json:"description"`
		Entity      FormEntityRequest               `json:"entity" validate:"required"`
		Questions   []BulkCreateFormQuestionRequest `json:"questions"`
	}
	FormEntityRequest struct {
		Type string `json:"type" validate:"required,oneof=event event_session"`
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
