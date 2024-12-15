package models

import (
	"database/sql"
	"github.com/google/uuid"
	"time"
)

var (
	TYPE_EVENT_REGISTRATION_RECORD = "eventRegistrationRecord"
)

type EventRegistrationRecord struct {
	ID                uuid.UUID
	Name              string
	Identifier        string
	CommunityId       string
	EventCode         string
	InstanceCode      string
	IdentifierOrigin  string
	CommunityIdOrigin string
	UpdatedBy         string
	Status            string
	RegisteredAt      time.Time
	VerifiedAt        sql.NullTime
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         sql.NullTime
}

func (erer *CreateEventRegistrationRecordResponse) ToResponse() *CreateEventRegistrationRecordResponse {
	return &CreateEventRegistrationRecordResponse{
		Type:             TYPE_EVENT_REGISTRATION_RECORD,
		ID:               erer.ID,
		Status:           erer.Status,
		Name:             erer.Name,
		Identifier:       erer.Identifier,
		CommunityID:      erer.CommunityID,
		EventCode:        erer.EventCode,
		EventTitle:       erer.EventTitle,
		InstanceCode:     erer.InstanceCode,
		InstanceTitle:    erer.InstanceTitle,
		TotalRegistrants: erer.TotalRegistrants,
		RegisterAt:       erer.RegisterAt,
		Registrants:      erer.Registrants,
	}
}

type (
	CreateEventRegistrationRecordRequest struct {
		//IsInheritUser bool                                        `json:"isInheritUser" validate:"required"`
		IsPersonalQR bool                                        `json:"isPersonalQR"`
		Name         string                                      `json:"name" validate:"omitempty,min=1,max=50,nameIdentifierCommunityIdField" example:"Professionals"`
		Identifier   string                                      `json:"identifier" validate:"omitempty,emailPhoneFormat"`
		CommunityId  string                                      `json:"communityId" validate:"omitempty,communityId"`
		EventCode    string                                      `json:"eventCode" validate:"required,min=7,max=7"`
		InstanceCode string                                      `json:"instanceCode" validate:"required,min=15,max=15"`
		RegisterAt   string                                      `json:"registerAt" validate:"required"`
		Registrants  []CreateOtherEventRegistrationRecordRequest `json:"registrants" validate:"dive,required"`
	}
	CreateOtherEventRegistrationRecordRequest struct {
		Name string `json:"name" validate:"required"`
	}
	CreateEventRegistrationRecordResponse struct {
		Type             string                                       `json:"type"`
		ID               uuid.UUID                                    `json:"registrationId"`
		Status           string                                       `json:"status"`
		Name             string                                       `json:"name"`
		Identifier       string                                       `json:"identifier,omitempty"`
		CommunityID      string                                       `json:"communityId,omitempty"`
		EventCode        string                                       `json:"eventCode"`
		EventTitle       string                                       `json:"eventTitle"`
		InstanceCode     string                                       `json:"instanceCode"`
		InstanceTitle    string                                       `json:"instanceTitle"`
		TotalRegistrants int                                          `json:"totalRegistrants"`
		RegisterAt       time.Time                                    `json:"registerAt"`
		Registrants      []CreateOtherEventRegistrationRecordResponse `json:"registrants,omitempty"`
	}
	CreateOtherEventRegistrationRecordResponse struct {
		Type   string    `json:"type"`
		ID     uuid.UUID `json:"id"`
		Status string    `json:"status"`
		Name   string    `json:"name"`
	}
)

func (erer *UpdateRegistrationStatusResponse) ToResponse() *UpdateRegistrationStatusResponse {
	return &UpdateRegistrationStatusResponse{
		Type:          TYPE_EVENT_REGISTRATION_RECORD,
		ID:            erer.ID,
		Status:        erer.Status,
		Name:          erer.Name,
		Identifier:    erer.Identifier,
		CommunityID:   erer.CommunityID,
		EventCode:     erer.EventCode,
		EventTitle:    erer.EventTitle,
		InstanceCode:  erer.InstanceCode,
		InstanceTitle: erer.InstanceTitle,
		UpdatedBy:     erer.UpdatedBy,
		VerifiedAt:    erer.VerifiedAt,
	}
}

type (
	UpdateRegistrationStatusParameter struct {
		ID string `json:"id" validate:"required,uuid"`
	}
	UpdateRegistrationStatusRequest struct {
		Status    string `json:"status" validate:"required,oneof=success cancelled" example:"success"`
		UpdatedAt string `json:"updatedAt" validate:"required"`
	}
	UpdateRegistrationStatusResponse struct {
		Type          string    `json:"type"`
		ID            uuid.UUID `json:"registrationId"`
		Status        string    `json:"status"`
		Name          string    `json:"name"`
		Identifier    string    `json:"identifier,omitempty"`
		CommunityID   string    `json:"communityId,omitempty"`
		EventCode     string    `json:"eventCode"`
		EventTitle    string    `json:"eventTitle"`
		InstanceCode  string    `json:"instanceCode"`
		InstanceTitle string    `json:"instanceTitle"`
		UpdatedBy     string    `json:"updatedBy"`
		VerifiedAt    time.Time `json:"verifiedAt"`
	}
)
