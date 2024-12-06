package models

import (
	"database/sql"
	"github.com/google/uuid"
	"time"
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

type (
	CreateEventRegistrationRecordRequest struct {
		IsUsingQR     bool                                        `json:"isUsingQR" validate:"required"`
		IsInheritUser bool                                        `json:"isInheritUser" validate:"required"`
		Name          string                                      `json:"name" validate:"omitempty,min=1,max=50,nameIdentifierCommunityIdField" example:"Professionals"`
		Identifier    string                                      `json:"identifier" validate:"omitempty,emailPhoneFormat"`
		CommunityId   string                                      `json:"communityId" validate:"omitempty,communityId"`
		EventCode     string                                      `json:"eventCode" validate:"required,min=1,max=30"`
		InstanceCode  string                                      `json:"instanceCode" validate:"required"`
		RegisterAt    string                                      `json:"registerAt" validate:"required"`
		Registrants   []CreateOtherEventRegistrationRecordRequest `json:"registrants" validate:"dive,required"`
	}
	CreateOtherEventRegistrationRecordRequest struct {
		Name string `json:"name" validate:"required"`
	}
	CreateEventRegistrationRecordResponse struct {
		Type             string                                 `json:"type"`
		ID               uuid.UUID                              `json:"id"`
		Status           string                                 `json:"status"`
		Name             string                                 `json:"name"`
		Identifier       string                                 `json:"identifier,omitempty"`
		CommunityID      string                                 `json:"communityId,omitempty"`
		EventCode        string                                 `json:"eventCode"`
		EventName        string                                 `json:"eventName"`
		SessionCode      string                                 `json:"sessionCode"`
		SessionName      string                                 `json:"sessionName"`
		TotalRegistrants int                                    `json:"totalRegistrants"`
		Registrants      []CreateOtherEventRegistrationResponse `json:"registrants"`
	}
	CreateOtherEventRegistrationRecordResponse struct {
		Type   string    `json:"type"`
		ID     uuid.UUID `json:"id"`
		Status string    `json:"status"`
		Name   string    `json:"name"`
	}
)
