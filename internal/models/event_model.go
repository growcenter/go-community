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

func (e *GetEventByCodeResponse) ToResponse() GetEventByCodeResponse {
	return GetEventByCodeResponse{
		Type:               TYPE_EVENT,
		Code:               e.Code,
		Title:              e.Title,
		Location:           e.Location,
		Description:        e.Description,
		CampusCode:         e.CampusCode,
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
	GetEventByCodeDBOutput struct {
		EventCode            string
		EventTitle           string
		EventLocation        string
		EventDescription     string
		EventCampusCode      pq.StringArray `gorm:"type:text[]"`
		EventAllowedRoles    pq.StringArray `gorm:"type:text[]"`
		EventIsRecurring     bool
		EventRecurrence      string
		EventStartAt         time.Time
		EventEndAt           time.Time
		EventRegisterStartAt time.Time
		EventRegisterEndAt   time.Time
		EventStatus          string
	}
	GetInstanceByEventCodeDBOutput struct {
		Code                string    `json:"instance_code"`
		Title               string    `json:"instance_title"`
		Location            string    `json:"instance_location"`
		Description         string    `json:"instance_description"`
		RegisterStartAt     time.Time `json:"instance_register_start_at"`
		RegisterEndAt       time.Time `json:"instance_register_end_at"`
		InstanceStartAt     time.Time `json:"instance_start_at"`
		InstanceEndAt       time.Time `json:"instance_end_at"`
		MaxRegister         int       `json:"instance_max_register"`
		TotalSeats          int       `json:"instance_total_seats"`
		BookedSeats         int       `json:"instance_booked_seats"`
		ScannedSeats        int       `json:"instance_scanned_seats"`
		IsRequired          bool      `json:"instance_is_required"`
		Status              string    `json:"instance_status"`
		TotalRemainingSeats int       `json:"total_remaining_seats"`
	}
	GetEventByCodeParameter struct {
		Code string `json:"string" validate:"required,min=2"`
	}
	GetEventByCodeResponse struct {
		Type               string                            `json:"type" example:"Event"`
		Code               string                            `json:"code" example:"2024-HOMEBASE"`
		Title              string                            `json:"title" example:"Homebase"`
		Location           string                            `json:"location" example:"PIOT 6 Lt. 6"`
		Description        string                            `json:"description" example:"Homebase"`
		CampusCode         []string                          `json:"campusCode"`
		IsRecurring        bool                              `json:"isRecurring" example:"true"`
		Recurrence         string                            `json:"recurrence,omitempty" example:"monthly"`
		EventStartAt       time.Time                         `json:"eventStartAt,omitempty" example:""`
		EventEndAt         time.Time                         `json:"eventEndAt,omitempty" example:""`
		RegisterStartAt    time.Time                         `json:"registerStartAt,omitempty" example:""`
		RegisterEndAt      time.Time                         `json:"registerEndAt,omitempty" example:""`
		AvailabilityStatus string                            `json:"availabilityStatus,omitempty" example:"available"`
		Instances          []GetInstancesByEventCodeResponse `json:"instances"`
	}
	GetInstancesByEventCodeResponse struct {
		Type                string    `json:"type" example:"eventInstance"`
		Code                string    `json:"code" example:"2024-HOMEBASE"`
		Title               string    `json:"title" example:"Homebase"`
		Description         string    `json:"description" example:"Homebase"`
		Location            string    `json:"location" example:"PIOT 6 Lt. 6"`
		InstanceIsRequired  bool      `json:"isRequired" example:"true"`
		InstanceStartAt     time.Time `json:"instanceStartAt" example:""`
		InstanceEndAt       time.Time `json:"instanceEndAt" example:""`
		RegisterStartAt     time.Time `json:"registerStartAt" example:""`
		RegisterEndAt       time.Time `json:"registerEndAt" example:""`
		MaxRegister         int       `json:"maxRegister" example:"0"`
		TotalSeats          int       `json:"totalSeats" example:"0"`
		BookedSeats         int       `json:"bookedSeats" example:"0"`
		TotalRemainingSeats int       `json:"totalRemainingSeats" example:"0"`
		AvailabilityStatus  string    `json:"availabilityStatus,omitempty" example:"available"`
	}
)
