package models

import (
	"database/sql"
	"time"
)

var TYPE_EVENT_INSTANCE = "eventInstance"

type EventInstance struct {
	ID                int
	Code              string
	EventCode         string
	Title             string
	Description       string
	InstanceStartAt   time.Time
	InstanceEndAt     time.Time
	RegisterStartAt   time.Time
	RegisterEndAt     time.Time
	LocationType      string
	LocationName      string
	MaxPerTransaction int
	IsOnePerAccount   bool
	IsOnePerTicket    bool
	RegisterFlow      string
	CheckType         string
	TotalSeats        int
	BookedSeats       int
	ScannedSeats      int
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         sql.NullTime
}

type (
	CreateInstanceRequest struct {
		Title             string `json:"title" validate:"required"`
		Description       string `json:"description"`
		InstanceStartAt   string `json:"instanceStartAt" validate:"required"`
		InstanceEndAt     string `json:"instanceEndAt" validate:"required"`
		RegisterStartAt   string `json:"registerStartAt" validate:"required"`
		RegisterEndAt     string `json:"registerEndAt" validate:"required"`
		LocationType      string `json:"locationType" validate:"required,oneof=online onsite hybrid"`
		LocationName      string `json:"locationName" validate:"required"`
		MaxPerTransaction int    `json:"maxPerTransaction"`
		IsOnePerAccount   bool   `json:"isOnePerAccount"`
		IsOnePerTicket    bool   `json:"isOnePerTicket"`
		RegisterFlow      string `json:"registerFlow" validate:"oneof=personal-qr event-qr both-qr none"`
		CheckType         string `json:"checkType" validate:"omitempty,oneof=check-in check-out both none"`
		TotalSeats        int    `json:"totalSeats"`
	}
	CreateInstanceResponse struct {
		Type              string    `json:"type"`
		InstanceCode      string    `json:"instanceCode"`
		EventCode         string    `json:"eventCode"`
		Title             string    `json:"title"`
		Description       string    `json:"description"`
		InstanceStartAt   time.Time `json:"instanceStartAt"`
		InstanceEndAt     time.Time `json:"instanceEndAt"`
		RegisterStartAt   time.Time `json:"registerStartAt"`
		RegisterEndAt     time.Time `json:"registerEndAt"`
		LocationType      string    `json:"locationType"`
		LocationName      string    `json:"locationName"`
		MaxPerTransaction int       `json:"maxPerTransaction,omitempty"`
		IsOnePerAccount   bool      `json:"isOnePerAccount"`
		IsOnePerTicket    bool      `json:"isOnePerTicket"`
		RegisterFlow      string    `json:"registerFlow"`
		TotalSeats        int       `json:"totalSeats,omitempty"`
		CheckType         string    `json:"checkType,omitempty"`
		Status            string    `json:"status,omitempty" example:"active"`
	}
)

type GetInstanceByCodeDBOutput struct {
	InstanceCode              string    `json:"instance_code"`
	InstanceEventCode         string    `json:"instance_event_code"`
	InstanceTitle             string    `json:"instance_title"`
	InstanceDescription       string    `json:"instance_description"`
	InstanceStartAt           time.Time `json:"instance_start_at"`
	InstanceEndAt             time.Time `json:"instance_end_at"`
	InstanceRegisterStartAt   time.Time `json:"instance_register_start_at"`
	InstanceRegisterEndAt     time.Time `json:"instance_register_end_at"`
	InstanceLocationType      string    `json:"instance_location"`
	InstanceLocationName      string    `json:"instance_location_name"`
	InstanceMaxPerTransaction int       `json:"instance_max_register"`
	InstanceIsOnePerAccount   bool      `json:"instance_is_one_per_account"`
	InstanceIsOnePerTicket    bool      `json:"instance_is_one_per_ticket"`
	InstanceRegisterFlow      string    `json:"instance_register_flow"`
	InstanceCheckType         string    `json:"instance_check_type"`
	InstanceTotalSeats        int       `json:"instance_total_seats"`
	InstanceBookedSeats       int       `json:"instance_booked_seats"`
	InstanceScannedSeats      int       `json:"instance_scanned_seats"`
	InstanceStatus            string    `json:"instance_status"`
	TotalRemainingSeats       int       `json:"total_remaining_seats"`
}

type GetSeatsAndNamesByInstanceCodeDBOutput struct {
	TotalSeats          int    `json:"total_seats"`
	BookedSeats         int    `json:"booked_seats"`
	ScannedSeats        int    `json:"scanned_seats"`
	EventInstanceTitle  string `json:"event_instance_title"`
	EventTitle          string `json:"event_title"`
	TotalRemainingSeats int    `json:"total_remaining_seats"`
}
