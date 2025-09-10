package models

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

var TYPE_EVENT_INSTANCE = "eventInstance"

type EventInstance struct {
	ID                       int
	Code                     string
	EventCode                string
	Title                    string
	Description              string
	ValidateParentIdentifier bool
	ParentIdentifierInput    pq.StringArray `gorm:"type:text[]"`
	ValidateChildIdentifier  bool
	ChildIdentifierInput     pq.StringArray `gorm:"type:text[]"`
	EnforceCommunityId       bool
	EnforceUniqueness        bool
	Methods                  pq.StringArray `gorm:"type:text[]"`
	Flow                     string
	StartAt                  time.Time
	EndAt                    time.Time
	RegisterStartAt          time.Time
	RegisterEndAt            time.Time
	VerifyStartAt            time.Time
	VerifyEndAt              time.Time
	Timezone                 string
	LocationType             string
	LocationOfflineVenue     string
	LocationOnlineLink       string
	QuotaPerUser             int
	Capacity                 int
	PostDetails              JSONB `gorm:"type:jsonb;default:'{}'"`
	Status                   string
	CreatedAt                time.Time
	UpdatedAt                time.Time
	DeletedAt                sql.NullTime
}

type (
	CreateInstanceRequest struct {
		EventCode          string                     `json:"eventCode" validate:"required"`
		IsFollowEvent      bool                       `json:"isFollowEvent"`
		IsPublish          bool                       `json:"isPublish"`
		Title              string                     `json:"title" validate:"required"`
		Description        string                     `json:"description"`
		RegistrationConfig InstanceRegistrationConfig `json:"registrationConfig" validate:"required,dive"`
		TimeConfig         InstanceTimeConfig         `json:"timeConfig" validate:"required,dive"`
		IdentifierConfig   InstanceIdentifierConfig   `json:"identifierConfig" validate:"required,dive"`
		Location           EventLocationRequest       `json:"location" validate:"required,dive"`
		IsUpdateEventTime  bool                       `json:"isUpdateEventTime"`
		Questions          []CreateFormRequest        `json:"questions" validate:"omitempty,dive"`
	}
	InstanceRegistrationConfig struct {
		Capacity           int      `json:"capacity" validate:"required,numeric,min=1"`
		QuotaPerUser       int      `json:"quotaPerUser" validate:"required,numeric,min=1"`
		EnforceCommunityId bool     `json:"enforceCommunityId"`
		EnforceUniqueness  bool     `json:"enforceUniqueness"`
		Methods            []string `json:"methods" validate:"omitempty,dive,oneof=personal-qr event-qr registration-qr" example:"personal-qr"`
		Flow               string   `json:"flow" validate:"required,oneof=direct staged free" example:"direct"`
	}
	InstanceIdentifierConfig struct {
		ValidateParentIdentifier bool     `json:"validateParentIdentifier"`
		ParentIdentifierInput    []string `json:"parentIdentifier" validate:"required_if=ValidateParentIdentifier true,dive,oneof=email phone" example:"email"`
		ValidateChildIdentifier  bool     `json:"validateChildIdentifier"`
		ChildIdentifierInput     []string `json:"childIdentifier" validate:"required_if=ValidateChildIdentifier true,dive,oneof=email phone" example:"email"`
	}
	InstanceTimeConfig struct {
		StartAt         string `json:"startAt" validate:"required"`
		EndAt           string `json:"endAt" validate:"required"`
		RegisterStartAt string `json:"registerStartAt" validate:"required"`
		RegisterEndAt   string `json:"registerEndAt" validate:"required"`
		VerifyStartAt   string `json:"verifyStartAt" validate:"required"`
		VerifyEndAt     string `json:"verifyEndAt" validate:"required"`
		Timezone        string `json:"timezone" validate:"required"`
	}
	CreateInstanceResponse struct {
		Type               string                             `json:"type"`
		InstanceCode       string                             `json:"instanceCode"`
		EventCode          string                             `json:"eventCode"`
		Title              string                             `json:"title"`
		Description        string                             `json:"description"`
		IdentifierConfig   InstanceIdentifierConfigResponse   `json:"identifierConfig"`
		TimeConfig         InstanceTimeConfigResponse         `json:"timeConfig"`
		Location           EventLocationResponse              `json:"location"`
		RegistrationConfig InstanceRegistrationConfigResponse `json:"registrationConfig"`
		Status             string                             `json:"status,omitempty" example:"active"`
	}
	InstanceTimeConfigResponse struct {
		StartAt         string `json:"startAt" example:"2024-12-10T09:02:42Z"`
		EndAt           string `json:"endAt" example:"2024-12-10T09:02:42Z"`
		RegisterStartAt string `json:"registerStartAt" example:"2024-12-10T09:02:42Z"`
		RegisterEndAt   string `json:"registerEndAt" example:"2024-12-10T09:02:42Z"`
		VerifyStartAt   string `json:"verifyStartAt" example:"2024-12-10T09:02:42Z"`
		VerifyEndAt     string `json:"verifyEndAt" example:"2024-12-10T09:02:42Z"`
		Timezone        string `json:"timezone" example:"Asia/Jakarta"`
	}
	InstanceRegistrationConfigResponse struct {
		IdentifierInput    []string `json:"identifier" example:"email"`
		Capacity           int      `json:"capacity" example:"100"`
		QuotaPerUser       int      `json:"quotaPerUser" example:"1"`
		EnforceCommunityId bool     `json:"enforceCommunityId" example:"false"`
		EnforceUniqueness  bool     `json:"enforceUniqueness" example:"false"`
		Methods            []string `json:"methods" example:"personal-qr"`
		Flow               string   `json:"flow" example:"direct"`
	}
	InstanceIdentifierConfigResponse struct {
		ValidateParentIdentifier bool     `json:"validateParentIdentifier"`
		ParentIdentifierInput    []string `json:"parentIdentifier" example:"email"`
		ValidateChildIdentifier  bool     `json:"validateChildIdentifier"`
		ChildIdentifierInput     []string `json:"childIdentifier" example:"email"`
	}
)

func (ir *CreateInstanceResponse) ToResponse() *CreateInstanceResponse {
	return &CreateInstanceResponse{
		Type:               ir.Type,
		InstanceCode:       ir.InstanceCode,
		EventCode:          ir.EventCode,
		Title:              ir.Title,
		Description:        ir.Description,
		TimeConfig:         ir.TimeConfig,
		Location:           ir.Location,
		RegistrationConfig: ir.RegistrationConfig,
		Status:             ir.Status,
	}
}

type GetInstanceByCodeDBOutput struct {
	InstanceCode              string    `json:"instance_code"`
	InstanceEventCode         string    `json:"instance_event_code"`
	InstanceTitle             string    `json:"instance_title"`
	InstanceDescription       string    `json:"instance_description"`
	InstanceStartAt           time.Time `json:"instance_start_at"`
	InstanceEndAt             time.Time `json:"instance_end_at"`
	InstanceRegisterStartAt   time.Time `json:"instance_register_start_at"`
	InstanceRegisterEndAt     time.Time `json:"instance_register_end_at"`
	InstanceAllowVerifyAt     time.Time `json:"instance_allow_verify_at"`
	InstanceDisallowVerifyAt  time.Time `json:"instance_disallow_verify_at"`
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
	TotalSeats               int       `json:"total_seats"`
	BookedSeats              int       `json:"booked_seats"`
	ScannedSeats             int       `json:"scanned_seats"`
	EventInstanceTitle       string    `json:"event_instance_title"`
	EventTitle               string    `json:"event_title"`
	TotalRemainingSeats      int       `json:"total_remaining_seats"`
	InstanceAllowVerifyAt    time.Time `json:"instance_allow_verify_at"`
	InstanceDisallowVerifyAt time.Time `json:"instance_disallow_verify_at"`
}

type (
	CreateInstanceExistingEventRequest struct {
		Title             string `json:"title" validate:"required"`
		Description       string `json:"description"`
		EventCode         string `json:"eventCode" validate:"required"`
		InstanceStartAt   string `json:"instanceStartAt" validate:"required"`
		InstanceEndAt     string `json:"instanceEndAt" validate:"required"`
		RegisterStartAt   string `json:"registerStartAt" validate:"required"`
		RegisterEndAt     string `json:"registerEndAt" validate:"required"`
		AllowVerifyAt     string `json:"allowVerifyAt" validate:"required"`
		DisallowVerifyAt  string `json:"disallowVerifyAt" validate:"required"`
		LocationType      string `json:"locationType" validate:"required,oneof=online onsite hybrid"`
		LocationName      string `json:"locationName" validate:"required"`
		MaxPerTransaction int    `json:"maxPerTransaction"`
		IsOnePerAccount   bool   `json:"isOnePerAccount"`
		IsOnePerTicket    bool   `json:"isOnePerTicket"`
		RegisterFlow      string `json:"registerFlow" validate:"oneof=personal-qr event-qr both-qr none"`
		CheckType         string `json:"checkType" validate:"omitempty,oneof=check-in check-out both none"`
		TotalSeats        int    `json:"totalSeats"`
		IsUpdateEventTime bool   `json:"isUpdateEventTime"`
	}
)
