package models

import (
	"database/sql"
	"time"
)

var TYPE_EVENT_SESSION = "eventSession"

type EventSession struct {
	ID              int
	EventCode       string
	Name            string
	Code            string
	Status          string
	Description     string
	Time            time.Time
	MaxSeating      int
	AvailableSeats  int
	RegisteredSeats int
	ScannedSeats    int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       sql.NullTime
}

func (es *EventSession) ToResponse() *GetEventSessionsDataResponse {
	return &GetEventSessionsDataResponse{
		Type:            TYPE_EVENT_SESSION,
		EventCode:       es.EventCode,
		Code:            es.Code,
		Name:            es.Name,
		Time:            es.Time,
		Description:     es.Description,
		MaxSeating:      es.MaxSeating,
		AvailableSeats:  es.AvailableSeats,
		RegisteredSeats: es.RegisteredSeats,
		ScannedSeats:    es.ScannedSeats,
		UnscannedSeats:  (es.RegisteredSeats - es.ScannedSeats),
		Status:          es.Status,
	}
}

type (
	GetEventSessionsDetailResponse struct {
		Type        string    `json:"type" example:"coolCategory"`
		EventCode   string    `json:"eventCode"`
		EventName   string    `json:"eventName"`
		CurrentTime time.Time `json:"currentTime"`
		IsUserValid bool      `json:"isUserValid" example:"isUserValid"`
	}
	GetEventSessionsDataResponse struct {
		Type            string    `json:"type" example:"coolCategory"`
		EventCode       string    `json:"eventCode"`
		Code            string    `json:"code"`
		Name            string    `json:"name" example:"Profesionals"`
		Time            time.Time `json:"time"`
		Description     string    `json:"description"`
		MaxSeating      int       `json:"maxSeating"`
		AvailableSeats  int       `json:"availableSeats"`
		RegisteredSeats int       `json:"registeredSeats"`
		ScannedSeats    int       `json:"scannedSeats"`
		UnscannedSeats  int       `json:"unscannedSeats"`
		Status          string    `json:"status" example:"active"`
	}
)
