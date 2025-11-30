package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	TYPE_EVENT_REGISTRATION = "eventRegistration"
)

type EventRegistration struct {
	Code         uuid.UUID
	EventCode    string
	InstanceCode string
	Name         string
	Email        string
	PhoneNumber  string
	CommunityId  string
	Quantity     int
	Method       string
	Status       string
	RegisterAt   time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

type (
	CreateEventRegistrationRequest struct {
		EventCode    string            `json:"eventCode" validate:"required,min=7,max=7" example:"xxxxxxx"`
		InstanceCode string            `json:"instanceCode" validate:"required,min=15,max=15" example:"xxxxxxx-yyyyyyy"`
		Registrant   RegistrantRequest `json:"registrant" validate:"required"`
		RegisterAt   time.Time         `json:"registerAt" validate:"required"`
		Method       string            `json:"method" validate:"required,oneof=personal-qr event-qr registration-qr" example:"personal-qr"`
		Attendees    []AttendeeRequest `json:"attendees" validate:"required,dive"`
	}
	RegistrantRequest struct {
		Name        string `json:"name" validate:"required"`
		Email       string `json:"email" validate:"omitempty,email"`
		PhoneNumber string `json:"phoneNumber" validate:"omitempty,phoneFormat0862"`
		CommunityId string `json:"communityId" validate:"required,communityId"`
	}
	AttendeeRequest struct {
		IsParent    bool         `json:"isParent"`
		Name        string       `json:"name" validate:"required"`
		Email       string       `json:"email" validate:"omitempty,email"`
		PhoneNumber string       `json:"phoneNumber" validate:"omitempty,phoneFormat0862"`
		LegalId     string       `json:"legalId" validate:"omitempty,min=16,max=16"`
		CommunityId string       `json:"communityId" validate:"omitempty,communityId"`
		Remarks     string       `json:"remarks" validate:"omitempty"`
		Form        []AnswerItem `json:"form" validate:""`
	}
	CreateAttendeeResponse struct {
		Type        string                       `json:"type"`
		Code        string                       `json:"code"`
		Role        string                       `json:"role"`
		Name        string                       `json:"name"`
		Email       string                       `json:"email"`
		PhoneNumber string                       `json:"phoneNumber"`
		LegalId     string                       `json:"legalId"`
		CommunityId string                       `json:"communityId"`
		Remarks     string                       `json:"remarks"`
		Status      string                       `json:"status"`
		Forms       []FormQuestionAnswerResponse `json:"forms"`
	}

	CreateEventRegistrationResponse struct {
		Type         string                   `json:"type"`
		Code         string                   `json:"code"`
		EventCode    string                   `json:"eventCode"`
		InstanceCode string                   `json:"instanceCode"`
		Method       string                   `json:"method"`
		Quantity     int                      `json:"quantity"`
		RegisterAt   time.Time                `json:"registerAt"`
		Registrant   RegistrantRequest        `json:"registrant"`
		Attendees    []CreateAttendeeResponse `json:"attendees"`
	}

	CreateRegistrantResponse struct {
		Type        string `json:"type"`
		Name        string `json:"name"`
		Email       string `json:"email"`
		PhoneNumber string `json:"phoneNumber"`
		CommunityId string `json:"communityId"`
	}
)

func (err *CreateEventRegistrationResponse) ToResponse() *CreateEventRegistrationResponse {
	return &CreateEventRegistrationResponse{
		Type:         TYPE_EVENT_REGISTRATION,
		Code:         err.Code,
		EventCode:    err.EventCode,
		InstanceCode: err.InstanceCode,
		Method:       err.Method,
		Quantity:     err.Quantity,
		RegisterAt:   err.RegisterAt,
		Registrant:   err.Registrant,
		Attendees:    err.Attendees,
	}
}
