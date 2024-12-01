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

func (e *GetAllEventsResponse) ToResponse() GetAllEventsResponse {
	return GetAllEventsResponse{
		Type:               TYPE_EVENT,
		Code:               e.Code,
		Title:              e.Title,
		Location:           e.Location,
		CampusCode:         e.CampusCode,
		IsRecurring:        e.IsRecurring,
		Recurrence:         e.Recurrence,
		EventStartAt:       e.EventStartAt,
		EventEndAt:         e.EventEndAt,
		RegisterStartAt:    e.RegisterStartAt,
		RegisterEndAt:      e.RegisterEndAt,
		AvailabilityStatus: e.AvailabilityStatus,
	}
}

type (
	GetAllEventsDBOutput struct {
		EventCode            string         `json:"event_code"`
		EventTitle           string         `json:"event_title"`
		EventLocation        string         `json:"event_location"`
		EventCampusCode      pq.StringArray `json:"event_campus_code" gorm:"type:text[]"`
		EventIsRecurring     bool           `json:"event_is_recurring"`
		EventRecurrence      string         `json:"event_recurrence"`
		EventStartAt         time.Time
		EventEndAt           time.Time
		EventRegisterStartAt time.Time `json:"event_register_start_at"`
		EventRegisterEndAt   time.Time `json:"event_register_end_at"`
		TotalRemainingSeats  int       `json:"total_remaining_seats"`
		InstanceIsRequired   bool      `json:"instance_is_required"`
	}
	GetAllEventsResponse struct {
		Type               string    `json:"type" example:"Event"`
		Code               string    `json:"code" example:"2024-HOMEBASE"`
		Title              string    `json:"title" example:"Homebase"`
		Location           string    `json:"location" example:"PIOT 6 Lt. 6"`
		CampusCode         []string  `json:"campusCode"`
		IsRecurring        bool      `json:"isRecurring" example:"true"`
		Recurrence         string    `json:"recurrence,omitempty" example:"monthly"`
		EventStartAt       time.Time `json:"eventStartAt,omitempty" example:""`
		EventEndAt         time.Time `json:"eventEndAt,omitempty" example:""`
		RegisterStartAt    time.Time `json:"registerStartAt,omitempty" example:""`
		RegisterEndAt      time.Time `json:"registerEndAt,omitempty" example:""`
		AvailabilityStatus string    `json:"availabilityStatus,omitempty" example:"available"`
	}
)

type (
	GetEventByCodeDBOutput struct {
		EventCode               string         `json:"event_code"`
		EventTitle              string         `json:"event_title"`
		EventLocation           string         `json:"event_location"`
		EventDescription        string         `json:"event_description"`
		EventCampusCode         pq.StringArray `json:"event_campus_code" gorm:"type:text[]"`
		EventIsRecurring        bool           `json:"event_is_recurring"`
		EventRecurrence         string         `json:"event_recurrence"`
		EventStartAt            time.Time
		EventEndAt              time.Time
		EventRegisterStartAt    time.Time `json:"event_register_start_at"`
		EventRegisterEndAt      time.Time `json:"event_register_end_at"`
		EventStatus             string
		InstanceCode            string    `json:"instance_code"`
		InstanceTitle           string    `json:"instance_title"`
		InstanceLocation        string    `json:"instance_location"`
		InstanceStartAt         time.Time `json:"instance_start_at"`
		InstanceEndAt           time.Time `json:"instance_end_at"`
		InstanceRegisterStartAt time.Time `json:"instance_register_start_at"`
		InstanceRegisterEndAt   time.Time `json:"instance_register_end_at"`
		InstanceDescription     string    `json:"instance_description"`
		InstanceMaxRegister     int       `json:"instance_max_register"`
		InstanceTotalSeats      int       `json:"instance_total_seats"`
		InstanceBookedSeats     int       `json:"instance_booked_seats"`
		InstanceScannedSeats    int       `json:"instance_scanned_seats"`
		InstanceStatus          string    `json:"instance_status"`
		TotalRemainingSeats     int       `json:"total_remaining_seats"`
		InstanceIsRequired      bool      `json:"instance_is_required"`
	}
)
