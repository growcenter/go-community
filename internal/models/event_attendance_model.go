package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	TYPE_EVENT_ATTENDANCE = "eventAttendance"
)

type EventAttendance struct {
	Code             uuid.UUID
	InstanceCode     string
	RegistrationCode uuid.UUID
	Role             string
	Name             string
	Email            string
	PhoneNumber      string
	LegalId          string
	ReferenceCode    *string
	Remarks          *string
	Status           string
	RegisterAt       time.Time
	VerifiedBy       *string
	VerifiedAt       *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

type EventAttendanceStatusCount struct {
	Pending   int `json:"pending"`
	Success   int `json:"success"`
	Cancelled int `json:"cancelled"`
}

type CreateEventAttendanceRequest struct {
	RegistrationCode uuid.UUID `json:"registrationCode"`
	Name             string    `json:"name" validate:"required"`
	Identifier       string    `json:"identifier" validate:"omitempty,emailPhone"`
}

type QrConfigRequest struct {
	Type string `json:"type"`
	// CommunityId string `json:"communityId" validate:"required,communityId"`
	Content JSONB `json:"content"`
}
