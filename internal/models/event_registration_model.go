package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventRegistration struct {
	ID           int
	Code         uuid.UUID
	InstanceCode string
	CommunityId  string
	Status       string
	Quantity     int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

type (
	CreateEventRegistrationRequest struct {
		EventCode    string            `json:"eventCode" validate:"required,min=7,max=7" example:"xxxxxxx"`
		InstanceCode string            `json:"instanceCode" validate:"required,min=15,max=15" example:"xxxxxxx-yyyyyyy"`
		Quantity     int               `json:"quantity" validate:"required,numeric"`
		Registrant   RegistrantRequest `json:"registrant" validate:"required,dive"`
		RegisterAt   time.Time         `json:"registerAt" validate:"required"`
		Method       string            `json:"method" validate:"required,oneof=personal-qr event-qr registration-qr" example:"personal-qr"`
		Attendees    []Attendees       `json:"attendees" validate:"required,dive"`
	}
	RegistrantRequest struct {
		Name        string `json:"name" validate:"required"`
		Identifier  string `json:"identifier" validate:"required,emailPhone"`
		CommunityId string `json:"communityId" validate:"required,communityId"`
	}
	Attendees struct {
		IsParent bool         `json:"isParent"`
		Form     []AnswerItem `json:"form" validate:"required,dive"`
	}
)
