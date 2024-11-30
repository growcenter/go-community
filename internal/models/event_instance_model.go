package models

import (
	"database/sql"
	"time"
)

var TYPE_EVENT_INSTANCE = "eventInstance"

type EventInstance struct {
	ID              int
	Code            string
	Title           string
	Location        string
	EventCode       string
	InstanceStartAt time.Time
	InstanceEndAt   time.Time
	RegisterStartAt time.Time
	RegisterEndAt   time.Time
	Description     string
	MaxRegister     int
	TotalSeats      int
	BookedSeats     int
	ScannedSeats    int
	IsRequired      bool
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       sql.NullTime
}

type (
	CreateInstanceRequest struct {
		IsInherit       bool   `json:"isInherit"`
		Title           string `json:"title" validate:"required"`
		Location        string `json:"location" validate:"required"`
		EventCode       string `json:"eventCode" validate:"required"`
		InstanceStartAt string `json:"instanceStartAt" validate:"required"`
		InstanceEndAt   string `json:"instanceEndAt" validate:"required"`
		RegisterStartAt string `json:"registerStartAt" validate:"required"`
		RegisterEndAt   string `json:"registerEndAt" validate:"required"`
		Description     string `json:"description"`
		MaxRegister     int    `json:"maxRegister"`
		TotalSeats      int    `json:"totalSeats"`
		IsRequired      bool   `json:"isRequired" validate:"required"`
	}
	CreateInstanceResponse struct {
		Type               string    `json:"type"`
		InstanceCode       string    `json:"instanceCode"`
		Title              string    `json:"title"`
		Location           string    `json:"location"`
		Description        string    `json:"description"`
		EventCode          string    `json:"eventCode"`
		InstanceStartAt    time.Time `json:"instanceStartAt"`
		InstanceEndAt      time.Time `json:"instanceEndAt"`
		RegisterStartAt    time.Time `json:"registerStartAt"`
		RegisterEndAt      time.Time `json:"registerEndAt"`
		MaxRegister        int       `json:"maxRegister,omitempty"`
		TotalSeats         int       `json:"totalSeats,omitempty"`
		AvailabilityStatus string    `json:"availabilityStatus" example:"available"`
		IsRequired         bool      `json:"isRequired"`
	}
)
