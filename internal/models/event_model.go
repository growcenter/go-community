package models

import (
	"database/sql"
	"github.com/lib/pq"
	"go-community/internal/common"
	"time"
)

var (
	TYPE_EVENT = "event"
)

type Event struct {
	ID                 int
	Code               string
	Title              string
	Topics             pq.StringArray `gorm:"type:text[]"`
	Description        string
	TermsAndConditions string
	AllowedFor         string
	AllowedUsers       pq.StringArray `gorm:"type:text[]"`
	AllowedRoles       pq.StringArray `gorm:"type:text[]"`
	AllowedCampuses    pq.StringArray `gorm:"type:text[]"`
	IsRecurring        bool
	Recurrence         string
	EventStartAt       time.Time
	EventEndAt         time.Time
	RegisterStartAt    time.Time
	RegisterEndAt      time.Time
	LocationType       string
	LocationName       string
	Status             string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          sql.NullTime
}

func (e *CreateEventResponse) ToResponse() *CreateEventResponse {
	return &CreateEventResponse{
		Type:               TYPE_EVENT,
		Code:               e.Code,
		Title:              e.Title,
		Topics:             e.Topics,
		Description:        e.Description,
		TermsAndConditions: e.TermsAndConditions,
		AllowedFor:         e.AllowedFor,
		AllowedUsers:       e.AllowedUsers,
		AllowedRoles:       e.AllowedRoles,
		AllowedCampuses:    e.AllowedCampuses,
		IsRecurring:        e.IsRecurring,
		Recurrence:         e.Recurrence,
		EventStartAt:       e.EventStartAt,
		EventEndAt:         e.EventEndAt,
		RegisterStartAt:    e.RegisterStartAt,
		RegisterEndAt:      e.RegisterEndAt,
		LocationType:       e.LocationType,
		LocationName:       e.LocationName,
		Status:             e.Status,
		Instances:          e.Instances,
	}
}

type (
	CreateEventRequest struct {
		Title              string                  `json:"name" validate:"required"`
		Topics             []string                `json:"topics"`
		Description        string                  `json:"description"`
		TermsAndConditions string                  `json:"termsAndConditions"`
		AllowedFor         string                  `json:"allowedFor" validate:"required,oneof=public private"`
		AllowedUsers       []string                `json:"allowedUsers" validate:"required"`
		AllowedRoles       []string                `json:"allowedRoles" validate:"required"`
		AllowedCampuses    []string                `json:"allowedCampuses" validate:"required,dive,min=3"`
		IsRecurring        bool                    `json:"isRecurring"`
		Recurrence         string                  `json:"recurrence"`
		EventStartAt       string                  `json:"eventStartAt"`
		EventEndAt         string                  `json:"eventEndAt"`
		RegisterStartAt    string                  `json:"registerStartAt"`
		RegisterEndAt      string                  `json:"registerEndAt"`
		LocationType       string                  `json:"locationType" validate:"required,oneof=online onsite hybrid"`
		LocationName       string                  `json:"locationName" validate:"required"`
		Instances          []CreateInstanceRequest `json:"instances" validate:"dive,required"`
	}
	CreateEventResponse struct {
		Type               string                   `json:"type" example:"event"`
		Code               string                   `json:"code" example:"bhfe382"`
		Title              string                   `json:"title" example:"Homebase"`
		Topics             []string                 `json:"topics"`
		Description        string                   `json:"description" example:"This event blabla"`
		TermsAndConditions string                   `json:"termsAndConditions" example:"This event blabla"`
		AllowedFor         string                   `json:"allowedFor" example:"public"`
		AllowedUsers       []string                 `json:"allowedUsers,omitempty"`
		AllowedRoles       []string                 `json:"allowedRoles,omitempty"`
		AllowedCampuses    []string                 `json:"allowedCampuses,omitempty"`
		IsRecurring        bool                     `json:"isRecurring" example:"true"`
		Recurrence         string                   `json:"recurrence,omitempty" example:"monthly"`
		EventStartAt       time.Time                `json:"eventStartAt,omitempty" example:""`
		EventEndAt         time.Time                `json:"eventEndAt,omitempty" example:""`
		RegisterStartAt    time.Time                `json:"registerStartAt,omitempty" example:""`
		RegisterEndAt      time.Time                `json:"registerEndAt,omitempty" example:""`
		LocationType       string                   `json:"locationType" example:"offline"`
		LocationName       string                   `json:"locationName" example:"PIOT 6 Lt. 6"`
		Status             string                   `json:"status,omitempty" example:"available"`
		Instances          []CreateInstanceResponse `json:"instances" validate:"dive,required"`
	}
)

func (e *GetAllEventsResponse) ToResponse() GetAllEventsResponse {
	return GetAllEventsResponse{
		Type:               TYPE_EVENT,
		Code:               e.Code,
		Title:              e.Title,
		Topics:             e.Topics,
		AllowedFor:         e.AllowedFor,
		AllowedUsers:       e.AllowedUsers,
		AllowedRoles:       e.AllowedRoles,
		AllowedCampuses:    e.AllowedCampuses,
		IsRecurring:        e.IsRecurring,
		Recurrence:         e.Recurrence,
		EventStartAt:       e.EventStartAt,
		EventEndAt:         e.EventEndAt,
		RegisterStartAt:    e.RegisterStartAt,
		RegisterEndAt:      e.RegisterEndAt,
		LocationType:       e.LocationType,
		AvailabilityStatus: e.AvailabilityStatus,
	}
}

type (
	GetAllEventsDBOutput struct {
		EventCode            string         `json:"event_code"`
		EventTitle           string         `json:"event_title"`
		EventTopics          pq.StringArray `gorm:"type:text[]"`
		EventLocationType    string
		EventAllowedFor      string
		EventAllowedRoles    pq.StringArray `gorm:"type:text[]"`
		EventAllowedUsers    pq.StringArray `gorm:"type:text[]"`
		EventAllowedCampuses pq.StringArray `gorm:"type:text[]"`
		EventIsRecurring     bool           `json:"event_is_recurring"`
		EventRecurrence      string         `json:"event_recurrence"`
		EventStartAt         time.Time
		EventEndAt           time.Time
		EventRegisterStartAt time.Time `json:"event_register_start_at"`
		EventRegisterEndAt   time.Time `json:"event_register_end_at"`
		InstanceTotalSeats   int
		TotalRemainingSeats  int            `json:"total_remaining_seats"`
		InstanceIsRequired   pq.BoolArray   `gorm:"type:boolean[]"`
		EventStatus          string         `json:"event_status"`
		InstancesData        pq.StringArray `gorm:"type:text[]"`
	}

	GetAllEventsResponse struct {
		Type                string    `json:"type" example:"Event"`
		Code                string    `json:"code" example:"2024-HOMEBASE"`
		Title               string    `json:"title" example:"Homebase"`
		Topics              []string  `json:"topics"`
		LocationType        string    `json:"locationType" example:"offline"`
		AllowedFor          string    `json:"allowedFor" example:"public"`
		AllowedUsers        []string  `json:"allowedUsers,omitempty"`
		AllowedRoles        []string  `json:"allowedRoles,omitempty"`
		AllowedCampuses     []string  `json:"allowedCampuses,omitempty"`
		IsRecurring         bool      `json:"isRecurring" example:"true"`
		Recurrence          string    `json:"recurrence,omitempty" example:"monthly"`
		EventStartAt        time.Time `json:"eventStartAt,omitempty" example:""`
		EventEndAt          time.Time `json:"eventEndAt,omitempty" example:""`
		RegisterStartAt     time.Time `json:"registerStartAt,omitempty" example:""`
		RegisterEndAt       time.Time `json:"registerEndAt,omitempty" example:""`
		TotalRemainingSeats int       `json:"totalRemainingSeats" example:"2"`
		AvailabilityStatus  string    `json:"availabilityStatus,omitempty" example:"available"`
	}
)

func (e *GetEventByCodeResponse) ToResponse() GetEventByCodeResponse {
	return GetEventByCodeResponse{
		Type:               TYPE_EVENT,
		Code:               e.Code,
		Title:              e.Title,
		Topics:             e.Topics,
		Description:        e.Description,
		TermsAndConditions: e.TermsAndConditions,
		AllowedFor:         e.AllowedFor,
		AllowedUsers:       e.AllowedUsers,
		AllowedRoles:       e.AllowedRoles,
		AllowedCampuses:    e.AllowedCampuses,
		IsRecurring:        e.IsRecurring,
		Recurrence:         e.Recurrence,
		EventStartAt:       e.EventStartAt,
		EventEndAt:         e.EventEndAt,
		RegisterStartAt:    e.RegisterStartAt,
		RegisterEndAt:      e.RegisterEndAt,
		LocationType:       e.LocationType,
		LocationName:       e.LocationName,
		AvailabilityStatus: e.AvailabilityStatus,
		Instances:          e.Instances,
	}
}

type (
	GetEventByCodeDBOutput struct {
		EventCode               string
		EventTitle              string
		EventTopics             pq.StringArray `gorm:"type:text[]"`
		EventDescription        string
		EventTermsAndConditions string
		EventAllowedFor         string
		EventAllowedRoles       pq.StringArray `gorm:"type:text[]"`
		EventAllowedUsers       pq.StringArray `gorm:"type:text[]"`
		EventAllowedCampuses    pq.StringArray `gorm:"type:text[]"`
		EventIsRecurring        bool
		EventRecurrence         string
		EventStartAt            time.Time
		EventEndAt              time.Time
		EventRegisterStartAt    time.Time
		EventRegisterEndAt      time.Time
		EventLocationType       string
		EventLocationName       string
		EventStatus             string
		InstanceTotalSeats      int
		TotalRemainingSeats     int            `json:"total_remaining_seats"`
		InstanceIsRequired      pq.BoolArray   `gorm:"type:boolean[]"`
		InstancesData           pq.StringArray `gorm:"type:text[]"`
	}
	GetInstanceByEventCodeDBOutput struct {
		InstanceCode              string    `json:"instance_code"`
		InstanceTitle             string    `json:"instance_title"`
		InstanceDescription       string    `json:"instance_description"`
		InstanceStartAt           time.Time `json:"instance_start_at"`
		InstanceEndAt             time.Time `json:"instance_end_at"`
		InstanceRegisterStartAt   time.Time `json:"instance_register_start_at"`
		InstanceRegisterEndAt     time.Time `json:"instance_register_end_at"`
		InstanceLocationType      string    `json:"instance_location"`
		InstanceLocationName      string    `json:"instance_location_name"`
		InstanceMaxPerTransaction int       `json:"instance_max_register"`
		InstanceIsRequired        bool      `json:"instance_is_required"`
		InstanceIsOnePerAccount   bool      `json:"instance_is_one_per_account"`
		InstanceIsOnePerTicket    bool      `json:"instance_is_one_per_ticket"`
		InstanceAllowPersonalQr   bool      `json:"allow_personal_qr"`
		InstanceAttendanceType    string    `json:"attendance_type"`
		InstanceTotalSeats        int       `json:"instance_total_seats"`
		InstanceBookedSeats       int       `json:"instance_booked_seats"`
		InstanceScannedSeats      int       `json:"instance_scanned_seats"`
		InstanceStatus            string    `json:"instance_status"`
		TotalRemainingSeats       int       `json:"total_remaining_seats"`
		EventAllowedFor           string    `json:"event_allowed_for"`
	}
	GetEventByCodeParameter struct {
		Code string `json:"string" validate:"required,min=2"`
	}
	GetEventByCodeResponse struct {
		Type               string                            `json:"type" example:"event"`
		Code               string                            `json:"code" example:"bhfe382"`
		Title              string                            `json:"title" example:"Homebase"`
		Topics             []string                          `json:"topics"`
		Description        string                            `json:"description" example:"This event blabla"`
		TermsAndConditions string                            `json:"termsAndConditions" example:"This event blabla"`
		AllowedFor         string                            `json:"allowedFor" example:"public"`
		AllowedUsers       []string                          `json:"allowedUsers,omitempty"`
		AllowedRoles       []string                          `json:"allowedRoles,omitempty"`
		AllowedCampuses    []string                          `json:"allowedCampuses,omitempty"`
		IsRecurring        bool                              `json:"isRecurring" example:"true"`
		Recurrence         string                            `json:"recurrence,omitempty" example:"monthly"`
		EventStartAt       time.Time                         `json:"eventStartAt,omitempty" example:""`
		EventEndAt         time.Time                         `json:"eventEndAt,omitempty" example:""`
		RegisterStartAt    time.Time                         `json:"registerStartAt,omitempty" example:""`
		RegisterEndAt      time.Time                         `json:"registerEndAt,omitempty" example:""`
		LocationType       string                            `json:"locationType" example:"offline"`
		LocationName       string                            `json:"locationName" example:"PIOT 6 Lt. 6"`
		AvailabilityStatus string                            `json:"availabilityStatus,omitempty" example:"available"`
		Instances          []GetInstancesByEventCodeResponse `json:"instances"`
	}
	GetInstancesByEventCodeResponse struct {
		Type                string    `json:"type" example:"eventInstance"`
		Code                string    `json:"code" example:"2024-HOMEBASE"`
		Title               string    `json:"title" example:"Homebase"`
		Description         string    `json:"description" example:"Homebase"`
		InstanceStartAt     time.Time `json:"instanceStartAt" example:""`
		InstanceEndAt       time.Time `json:"instanceEndAt" example:""`
		RegisterStartAt     time.Time `json:"registerStartAt" example:""`
		RegisterEndAt       time.Time `json:"registerEndAt" example:""`
		LocationType        string    `json:"locationType" example:"offline"`
		LocationName        string    `json:"LocationName" example:"PIOT 6 Lt. 6"`
		MaxPerTransaction   int       `json:"maxPerTransaction,omitempty"`
		IsRequired          bool      `json:"isRequired"`
		IsOnePerAccount     bool      `json:"isOnePerAccount"`
		IsOnePerTicket      bool      `json:"isOnePerTicket"`
		AllowPersonalQr     bool      `json:"allowPersonalQr"`
		AttendanceType      string    `json:"attendanceType"`
		TotalSeats          int       `json:"totalSeats" example:"0"`
		BookedSeats         int       `json:"bookedSeats" example:"0"`
		TotalRemainingSeats int       `json:"totalRemainingSeats" example:"0"`
		AvailabilityStatus  string    `json:"availabilityStatus,omitempty" example:"available"`
	}
)

type EventAvailabilityStatus int32

const (
	AVAILABILITY_STATUS_AVAILABLE EventAvailabilityStatus = iota
	AVAILABILITY_STATUS_UNAVAILABLE
	AVAILABILITY_STATUS_FULL
	AVAILABILITY_STATUS_SOON
)

const (
	AvailibilityStatusAvailable   = "available"
	AvailibilityStatusUnavailable = "unavailable"
	AvailibilityStatusFull        = "full"
	AvailibilityStatusSoon        = "soon"
)

var (
	MapAvailabilityStatus = map[EventAvailabilityStatus]string{
		AVAILABILITY_STATUS_AVAILABLE:   AvailibilityStatusAvailable,
		AVAILABILITY_STATUS_UNAVAILABLE: AvailibilityStatusUnavailable,
		AVAILABILITY_STATUS_FULL:        AvailibilityStatusFull,
		AVAILABILITY_STATUS_SOON:        AvailibilityStatusSoon,
	}
)

func DefineAvailabilityStatus(event interface{}) (string, error) {
	var totalRemainingSeats int
	var countInstanceIsRequired int
	var totalSeats int
	var eventAllowedFor string
	var eventRegisterStartAt, eventRegisterEndAt time.Time
	var instanceIsRequired []bool

	// Type assertion to extract fields from the concrete type
	switch e := event.(type) {
	case GetAllEventsDBOutput:
		totalRemainingSeats = e.TotalRemainingSeats
		totalSeats = e.InstanceTotalSeats
		eventAllowedFor = e.EventAllowedFor
		eventRegisterStartAt = e.EventRegisterStartAt
		eventRegisterEndAt = e.EventRegisterEndAt

	case *GetEventByCodeDBOutput:
		totalRemainingSeats = e.TotalRemainingSeats
		totalSeats = e.InstanceTotalSeats
		instanceIsRequired = common.GetBooleanArrayFromStringArray(e.InstancesData)
		countInstanceIsRequired = common.CountTrue(instanceIsRequired)
		eventAllowedFor = e.EventAllowedFor
		eventRegisterStartAt = e.EventRegisterStartAt
		eventRegisterEndAt = e.EventRegisterEndAt
	case GetInstanceByEventCodeDBOutput:
		totalRemainingSeats = e.TotalRemainingSeats
		totalSeats = e.InstanceTotalSeats
		countInstanceIsRequired = common.BoolToInt(e.InstanceIsRequired)
		eventRegisterStartAt = e.InstanceRegisterStartAt
		eventRegisterEndAt = e.InstanceRegisterEndAt
		eventAllowedFor = e.EventAllowedFor
		instanceIsRequired = []bool{e.InstanceIsRequired}
	default:
		// Return a default or error if the type is not recognized
		return "", ErrorInvalidInput
	}

	switch {
	case totalRemainingSeats <= 0 && countInstanceIsRequired < len(instanceIsRequired):
		return MapAvailabilityStatus[AVAILABILITY_STATUS_AVAILABLE], nil
	case totalRemainingSeats <= 0 && countInstanceIsRequired == len(instanceIsRequired) && eventAllowedFor != "private" && totalSeats > 0:
		return MapAvailabilityStatus[AVAILABILITY_STATUS_FULL], nil
	case common.Now().Before(eventRegisterStartAt.In(common.GetLocation())):
		return MapAvailabilityStatus[AVAILABILITY_STATUS_SOON], nil
	case common.Now().After(eventRegisterEndAt.In(common.GetLocation())):
		return MapAvailabilityStatus[AVAILABILITY_STATUS_UNAVAILABLE], nil
	default:
		return MapAvailabilityStatus[AVAILABILITY_STATUS_AVAILABLE], nil
	}
}
