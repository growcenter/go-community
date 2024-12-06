package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
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
	IsRequired        bool
	IsOnePerAccount   bool
	IsOnePerTicket    bool
	AllowPersonalQr   bool
	AttendanceType    string
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
		IsRequired        bool   `json:"isRequired"`
		IsOnePerAccount   bool   `json:"isOnePerAccount"`
		IsOnePerTicket    bool   `json:"isOnePerTicket"`
		AllowPersonalQr   bool   `json:"allowPersonalQr"`
		AttendanceType    string `json:"attendanceType" validate:"omitempty,oneof=check-in check-out both none"`
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
		IsRequired        bool      `json:"isRequired"`
		IsOnePerAccount   bool      `json:"isOnePerAccount"`
		IsOnePerTicket    bool      `json:"isOnePerTicket"`
		AllowPersonalQr   bool      `json:"allowPersonalQr"`
		TotalSeats        int       `json:"totalSeats,omitempty"`
		AttendanceType    string    `json:"attendanceType,omitempty"`
		Status            string    `json:"status,omitempty" example:"active"`
	}
)

// InstanceData struct to hold the individual instance data
type InstanceDataDBOutput struct {
	TotalSeats  int  `json:"total_seats"`
	BookedSeats int  `json:"booked_seats"`
	IsRequired  bool `json:"is_required"`
}

// InstancesData is a custom type that implements the Scanner and Valuer interfaces
type InstancesData []InstanceDataDBOutput

// Implement the Scanner interface to unmarshal the JSONB data into the InstancesData slice
func (id *InstancesData) Scan(value interface{}) error {
	if value == nil {
		*id = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, id)
	case string:
		return json.Unmarshal([]byte(v), id)
	default:
		return fmt.Errorf("unsupported scan type for InstancesData: %T", v)
	}
}

// Implement the Valuer interface to marshal the InstancesData slice into JSONB for storage
func (id InstancesData) Value() (driver.Value, error) {
	if id == nil {
		return nil, nil
	}
	return json.Marshal(id)
}
