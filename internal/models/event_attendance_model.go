package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Attendee struct {
	ID               int
	Code             uuid.UUID
	RegistrationCode uuid.UUID
	Role             string
	Name             string
	QRCodeConfig     *string
	Status           string
	VerifiedAt       *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

type CreateEventAttendanceRequest struct {
	RegistrationCode uuid.UUID `json:"registrationCode"`
	Name             string    `json:"name" validate:"required"`
	Identifier       string    `json:"identifier" validate:"omitempty,emailPhone"`
}
