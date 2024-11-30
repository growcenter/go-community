package models

import (
	"database/sql"
	"github.com/lib/pq"
	"time"
)

var (
	TYPE_EVENT = "event"
)

type Event struct {
	ID              int
	Code            string
	Title           string
	Location        string
	Description     string
	CampusCode      pq.StringArray `gorm:"type:text[]"`
	AllowedUsers    pq.StringArray `gorm:"type:text[]"`
	AllowedRoles    pq.StringArray `gorm:"type:text[]"`
	IsRecurring     bool
	Recurrence      string
	EventStartAt    time.Time
	EventEndAt      time.Time
	RegisterStartAt time.Time
	RegisterEndAt   time.Time
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       sql.NullTime
}

func (e *CreateEventResponse) ToResponse() *CreateEventResponse {
	return &CreateEventResponse{
		Type:               TYPE_EVENT,
		Code:               e.Code,
		Title:              e.Title,
		Location:           e.Location,
		Description:        e.Description,
		CampusCode:         e.CampusCode,
		AllowedUsers:       e.AllowedUsers,
		AllowedRoles:       e.AllowedRoles,
		IsRecurring:        e.IsRecurring,
		Recurrence:         e.Recurrence,
		EventStartAt:       e.EventStartAt,
		EventEndAt:         e.EventEndAt,
		RegisterStartAt:    e.RegisterStartAt,
		RegisterEndAt:      e.RegisterEndAt,
		AvailabilityStatus: e.AvailabilityStatus,
		Instances:          e.Instances,
	}
}

type (
	CreateEventRequest struct {
		Code            string                  `json:"code" validate:"required,min=1,max=30"`
		Title           string                  `json:"name" validate:"required"`
		Location        string                  `json:"location" validate:"required"`
		Description     string                  `json:"description"`
		CampusCode      []string                `json:"campusCode" validate:"required,dive,min=3"`
		AllowedUsers    []string                `json:"allowedUsers"`
		AllowedRoles    []string                `json:"allowedRoles" validate:"required"`
		IsRecurring     bool                    `json:"isRecurring" validate:"required"`
		Recurrence      string                  `json:"recurrence"`
		EventStartAt    string                  `json:"eventStartAt"`
		EventEndAt      string                  `json:"eventEndAt"`
		RegisterStartAt string                  `json:"registerStartAt"`
		RegisterEndAt   string                  `json:"registerEndAt"`
		Instances       []CreateInstanceRequest `json:"instances" validate:"dive,required"`
	}
	CreateEventResponse struct {
		Type               string                   `json:"type" example:"Event"`
		Code               string                   `json:"code" example:"2024-HOMEBASE"`
		Title              string                   `json:"title" example:"Homebase"`
		Location           string                   `json:"location" example:"PIOT 6 Lt. 6"`
		Description        string                   `json:"description" example:"This event blabla"`
		CampusCode         []string                 `json:"campusCode"`
		AllowedUsers       []string                 `json:"allowedUsers,omitempty"`
		AllowedRoles       []string                 `json:"allowedRoles"`
		IsRecurring        bool                     `json:"isRecurring" example:"true"`
		Recurrence         string                   `json:"recurrence,omitempty" example:"monthly"`
		EventStartAt       time.Time                `json:"eventStartAt,omitempty" example:""`
		EventEndAt         time.Time                `json:"eventEndAt,omitempty" example:""`
		RegisterStartAt    time.Time                `json:"registerStartAt,omitempty" example:""`
		RegisterEndAt      time.Time                `json:"registerEndAt,omitempty" example:""`
		AvailabilityStatus string                   `json:"availabilityStatus,omitempty" example:"available"`
		Instances          []CreateInstanceResponse `json:"instances" validate:"dive,required"`
	}
)
