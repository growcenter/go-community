package models

import (
	"database/sql"
	"time"
)

var TYPE_EVENT_REGISTRATION = "eventRegistration"

type EventRegistration struct {
	ID                  int
	Name                string
	Identifier          string
	Address             string
	AccountNumber       string
	Code                string
	EventCode           string
	SessionCode         string
	RegisteredBy        string
	UpdatedBy           string
	AccountNumberOrigin string
	Status              string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           sql.NullTime

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
		Seats:         len(er.Others) + 1,
		AccountNumber: er.AccountNumber,
		Code:          er.Code,
		Status:        er.Status,
		Others:        er.Others,
	}
}

type (
	CreateEventRegistrationRequest struct {
		Name        string                                `json:"name" validate:"required,min=1,max=50" example:"Professionals"`
		Identifier  string                                `json:"identifier" validate:"required,noStartEndSpaces,emailPhoneFormat"`
		Address     string                                `json:"address" validate:"required,min=15,noStartEndSpaces"`
		EventCode   string                                `json:"eventCode" validate:"required,min=1,max=30,noStartEndSpaces" example:"Professionals"`
		SessionCode string                                `json:"sessionCode" validate:"required,min=1,max=30,noStartEndSpaces" example:"Professionals"`
		Others      []CreateOtherEventRegistrationRequest `json:"otherRegister" validate:"dive,required"`
	}
	CreateOtherEventRegistrationRequest struct {
		Name    string `json:"name" validate:"required"`
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

func (er *GetRegisteredResponse) ToResponse() *GetRegisteredResponse {
	return &GetRegisteredResponse{
		Type:          TYPE_EVENT_REGISTRATION,
		Name:          er.Name,
		Identifier:    er.Identifier,
		Address:       er.Address,
		AccountNumber: er.AccountNumber,
		Code:          er.Code,
		EventCode:     er.EventCode,
		EventName:     er.EventGeneral.Name,
		SessionCode:   er.SessionCode,
		SessionName:   er.EventSession.Name,
		Status:        er.Status,
		// Others:        er.Others,
	}
}

type (
	GetRegisteredRequest struct {
		RegisteredBy string `json:"registeredBy" query:"registeredBy" validate:"required" example:"tono@amartha.com"`
	}
	GetRegisteredResponse struct {
		Type          string `json:"type"`
		Name          string `json:"name"`
		Identifier    string `json:"identifier"`
		Address       string `json:"address"`
		AccountNumber string `json:"accountNumber,omitempty"`
		Code          string `json:"code"`
		EventCode     string `json:"eventCode"`
		EventName     string `json:"eventName"`
		SessionCode   string `json:"sessionCode"`
		SessionName   string `json:"sessionName"`
		Status        string `json:"status"`

		EventGeneral EventGeneral `json:"-"`
		EventSession EventSession `json:"-"`
	}

	GetRegisteredRepository struct {
		EventRegistration EventRegistration
		EventGeneral      EventGeneral
		EventSession      EventSession
	}

	GetRegisteredRaw struct {
		ID            uint   `gorm:"column:id"`
		Name          string `gorm:"column:name"`
		Identifier    string `gorm:"column:identifier"`
		Address       string `gorm:"column:address"`
		AccountNumber string `gorm:"column:account_number"`
		Code          string `gorm:"column:code"`
		EventCode     string `gorm:"column:event_code"`
		SessionCode   string `gorm:"column:session_code"`
		RegisteredBy  string `gorm:"column:registered_by"`
		UpdatedBy     string `gorm:"column:updated_by"`
		Status        string `gorm:"column:status"`
		GeneralName   string
		SessionName   string
	}
)

func (er *GetAllRegisteredResponse) ToResponse() *GetAllRegisteredResponse {
	return &GetAllRegisteredResponse{
		Type:          TYPE_EVENT_REGISTRATION,
		Name:          er.Name,
		Identifier:    er.Identifier,
		Address:       er.Address,
		AccountNumber: er.AccountNumber,
		Code:          er.Code,
		RegisteredBy:  er.RegisteredBy,
		UpdatedBy:     er.UpdatedBy,
		EventCode:     er.EventCode,
		EventName:     er.EventGeneral.Name,
		SessionCode:   er.SessionCode,
		SessionName:   er.EventSession.Name,
		Status:        er.Status,
		// Others:        er.Others,
	}
}

type (
	GetAllPaginationParams struct {
		Page              int
		Limit             int
		Sort              string
		Search            string
		FilterSessionCode string
		FilterEventCode   string
	}

	GetAllPaginationParamsResponse struct {
		Search            string `json:"search,omitempty"`
		FilterSessionCode string `json:"sessionCode,omitempty"`
		FilterEventCode   string `json:"eventCode,omitempty"`
	}

	GetAllRegisteredResponse struct {
		Type          string `json:"type"`
		Name          string `json:"name"`
		Identifier    string `json:"identifier"`
		Address       string `json:"address"`
		AccountNumber string `json:"accountNumber,omitempty"`
		Code          string `json:"code"`
		RegisteredBy  string `json:"registeredBy"`
		UpdatedBy     string `json:"updatedBy"`
		EventCode     string `json:"eventCode"`
		EventName     string `json:"eventName"`
		SessionCode   string `json:"sessionCode"`
		SessionName   string `json:"sessionName"`
		Status        string `json:"status"`

		EventGeneral EventGeneral `json:"-"`
		EventSession EventSession `json:"-"`
	}
)

func (er *EventRegistration) ToUpdate() *VerifyRegistrationResponse {
	return &VerifyRegistrationResponse{
		Type:          TYPE_EVENT_REGISTRATION,
		Name:          er.Name,
		Identifier:    er.Identifier,
		AccountNumber: er.AccountNumber,
		Code:          er.Code,
		RegisteredBy:  er.RegisteredBy,
		UpdatedBy:     er.UpdatedBy,
		Status:        er.Status,
	}
}

type (
	VerifyRegistrationRequest struct {
		Code        string `param:"code" validate:"required,uuid"`
		Status      string `json:"status" validate:"required,oneof=active cancelled verified" example:"female"`
		SessionCode string `json:"sessionCode" validate:"required,min=1,max=30,noStartEndSpaces" example:"Professionals"`
	}
	VerifyRegistrationResponse struct {
		Type          string `json:"type"`
		Name          string `json:"name"`
		Identifier    string `json:"identifier"`
		AccountNumber string `json:"accountNumber,omitempty"`
		Code          string `json:"code"`
		RegisteredBy  string `json:"registeredBy"`
		UpdatedBy     string `json:"updatedBy"`
		Status        string `json:"status"`
	}
)

func (er *EventRegistration) ToCancel() *CancelRegistrationResponse {
	return &CancelRegistrationResponse{
		Type:          TYPE_EVENT_REGISTRATION,
		Name:          er.Name,
		Identifier:    er.Identifier,
		AccountNumber: er.AccountNumber,
		Code:          er.Code,
		RegisteredBy:  er.RegisteredBy,
		UpdatedBy:     er.UpdatedBy,
		Status:        er.Status,
		DeletedAt:     er.DeletedAt,
	}
}

type (
	CancelRegistrationRequest struct {
		Code string `param:"code" validate:"required,uuid"`
	}
	CancelRegistrationResponse struct {
		Type          string       `json:"type"`
		Name          string       `json:"name"`
		Identifier    string       `json:"identifier"`
		AccountNumber string       `json:"accountNumber,omitempty"`
		Code          string       `json:"code"`
		RegisteredBy  string       `json:"registeredBy"`
		UpdatedBy     string       `json:"updatedBy"`
		Status        string       `json:"status"`
		DeletedAt     sql.NullTime `json:"deletedAt"`
	}
)
