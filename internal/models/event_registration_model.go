package models

import (
	"database/sql"
	"time"
)

var TYPE_EVENT_REGISTRATION = "eventRegistration"

type EventRegistration struct {
	ID            int
	Name          string
	Identifier    string
	Address       string
	AccountNumber string
	Code          string
	EventCode     string
	SessionCode   string
	RegisteredBy  string
	UpdatedBy     string
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     sql.NullTime

	EventGeneral EventGeneral `gorm:"foreignKey:EventCode"`
	EventSession EventSession `gorm:"foreignKey:SessionCode"`
}

func (er *CreateEventRegistrationResponse) ToCreate() *CreateEventRegistrationResponse {
	return &CreateEventRegistrationResponse{
		Type:          TYPE_EVENT_REGISTRATION,
		Name:          er.Name,
		Identifier:    er.Identifier,
		Address:       er.Address,
		EventCode:     er.EventCode,
		EventName:     er.EventGeneral.Name,
		SessionCode:   er.SessionCode,
		SessionName:   er.EventSession.Name,
		IsValid:       true,
		Seats:         1,
		AccountNumber: er.AccountNumber,
		Code:          er.Code,
		Status:        er.Status,
		Others:        er.Others,
	}
}

type (
	CreateEventRegistrationRequest struct {
		Name        string                                `json:"name" validate:"required,min=1,max=50,nospecial,noStartEndSpaces" example:"Professionals"`
		Identifier  string                                `json:"identifier" validate:"required,noStartEndSpaces,emailPhoneFormat"`
		Address     string                                `json:"address" validate:"required,min=15,noStartEndSpaces"`
		EventCode   string                                `json:"eventCode" validate:"required,min=1,max=30,noStartEndSpaces" example:"Professionals"`
		SessionCode string                                `json:"sessionCode" validate:"required,min=1,max=30,noStartEndSpaces" example:"Professionals"`
		Others      []CreateOtherEventRegistrationRequest `json:"otherRegister" validate:"dive,required"`
	}
	CreateOtherEventRegistrationRequest struct {
		Name    string `json:"name" validate:"required,noStartEndSpaces"`
		Address string `json:"address" validate:"required,min=15,noStartEndSpaces"`
	}
	CreateEventRegistrationResponse struct {
		Type          string                                 `json:"type"`
		Name          string                                 `json:"name"`
		Identifier    string                                 `json:"identifier"`
		Address       string                                 `json:"address"`
		AccountNumber string                                 `json:"accountNumber,omitempty"`
		Code          string                                 `json:"code"`
		EventCode     string                                 `json:"eventCode"`
		EventName     string                                 `json:"eventName"`
		SessionCode   string                                 `json:"sessionCode"`
		SessionName   string                                 `json:"sessionName"`
		IsValid       bool                                   `json:"isValid"`
		Seats         int                                    `json:"seats"`
		Status        string                                 `json:"status"`
		Others        []CreateOtherEventRegistrationResponse `json:"otherBooking"`

		EventGeneral EventGeneral `json:"-"`
		EventSession EventSession `json:"-"`
	}
	CreateOtherEventRegistrationResponse struct {
		Type    string `json:"type"`
		Name    string `json:"name"`
		Address string `json:"address"`
		Code    string `json:"code"`
	}
)
