package models

import (
	"database/sql"
	"time"
)

var TYPE_EVENT_GENERAL = "generalEvent"

type EventGeneral struct {
	ID                 int
	Name               string
	Code               string `gorm:"primaryKey"`
	CampusCode         string
	Status             string
	Description        string
	OpenRegistration   time.Time
	ClosedRegistration time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          sql.NullTime
}

func (eg *EventGeneral) ToResponse() *GetGeneralEventDataResponse {
	return &GetGeneralEventDataResponse{
		Type:               TYPE_EVENT_GENERAL,
		Code:               eg.Code,
		Name:               eg.Name,
		Description:        eg.Description,
		CampusCode:         eg.CampusCode,
		OpenRegistration:   eg.OpenRegistration,
		ClosedRegistration: eg.ClosedRegistration,
		Status:             eg.Status,
	}
}

type (
	GetGeneralEventDetailResponse struct {
		Type        string    `json:"type" example:"coolCategory"`
		CurrentTime time.Time `json:"currentTime"`
		IsUserValid bool      `json:"isUserValid" example:"isUserValid"`
	}

	GetGeneralEventDataResponse struct {
		Type               string    `json:"type" example:"coolCategory"`
		Code               string    `json:"code"`
		Name               string    `json:"name" example:"Profesionals"`
		Description        string    `json:"description"`
		CampusCode         string    `json:"campusCode"`
		OpenRegistration   time.Time `json:"openRegistration"`
		ClosedRegistration time.Time `json:"closedRegistration"`
		Status             string    `json:"status" example:"active"`
	}
)
