package models

import (
	"time"

	"github.com/google/uuid"
)

type FormAssociation struct {
	ID         int        `gorm:"type:integer;primarykey"`
	Code       uuid.UUID  `gorm:"type:uuid;not null"`
	FormCode   uuid.UUID  `gorm:"type:uuid;not null"`
	EntityType string     `gorm:"type:varchar(50);not null"`
	EntityCode string     `gorm:"type:varchar(50);not null"`
	CreatedAt  time.Time  `gorm:"type:timestamptz;default:now()"`
	UpdatedAt  time.Time  `gorm:"type:timestamptz;default:now()"`
	DeletedAt  *time.Time `gorm:"type:timestamptz"`
}

type (
	CreateFormAssociationRequest struct {
		FormCode   uuid.UUID `json:"formCode" validate:"required"`
		EntityCode string    `json:"entityCode" validate:"required"`
		EntityType string    `json:"entityType" validate:"required"`
	}

	CreateFormAssociationResponse struct {
		FormCode   uuid.UUID `json:"formCode"`
		EntityCode string    `json:"entityCode"`
		EntityType string    `json:"entityType"`
	}
)
