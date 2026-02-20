package models

import (
	"database/sql"
	"time"
)

var TYPE_EVENT_INSTANCE = "eventInstance"

// EventInstance represents a specific occurrence/session of an event
// Each instance has its own schedule, capacity, and registration settings
type EventInstance struct {
	// Primary Keys
	ID   int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Code string `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`

	// Event Relationship
	EventCode string `gorm:"type:varchar(50);index;not null" json:"event_code"`

	// Basic Information
	Title       string `gorm:"type:varchar(255);not null" json:"title"`
	Description string `gorm:"type:text" json:"description"`

	// Instance Type
	InstanceType string `gorm:"type:varchar(30);not null;default:'registration'" json:"instance_type"` // registration, announcement, volunteer-attendance

	// Registration Identifier Configuration
	ParentIdentifierFields JSONB `gorm:"type:jsonb" json:"parent_identifier_fields" swaggertype:"object"` // Fields required from parent/guardian
	ChildIdentifierFields  JSONB `gorm:"type:jsonb" json:"child_identifier_fields" swaggertype:"object"`  // Fields required from child/participant

	// Registration Enforcement Rules
	EnforceCommunityId      bool     `gorm:"default:false" json:"enforce_community_id"`      // One registration per community_id
	EnforceSelfRegistration bool     `gorm:"default:false" json:"enforce_self_registration"` // User can only register themselves
	EnforceUniqueness       bool     `gorm:"default:false" json:"enforce_uniqueness"`        // One registration per user_id
	Methods                 []string `gorm:"type:text[]" json:"methods"`                     // Allowed registration methods: personal-qr, event-qr, registration-qr

	// Location Information
	LocationType       string  `gorm:"type:varchar(20);not null" json:"location_type"`       // online, offline, hybrid
	PhysicalAddress    *string `gorm:"type:text" json:"physical_address"`                    // Physical venue address
	VirtualLink        *string `gorm:"type:text" json:"virtual_link"`                        // Online meeting link
	MeetingCTAText     *string `gorm:"type:varchar(100)" json:"meeting_cta_text"`            // Call-to-action button text
	LocationDetails    *string `gorm:"type:text" json:"location_details"`                    // Additional location information
	LocationVisibility string  `gorm:"type:varchar(30);not null" json:"location_visibility"` // pre-registration, post-registration, all

	// Schedule Information
	StartAt         time.Time `gorm:"not null" json:"start_at"`          // Instance start time
	EndAt           time.Time `gorm:"not null" json:"end_at"`            // Instance end time
	RegisterStartAt time.Time `gorm:"not null" json:"register_start_at"` // Registration opens
	RegisterEndAt   time.Time `gorm:"not null" json:"register_end_at"`   // Registration closes
	VerifyStartAt   time.Time `gorm:"not null" json:"verify_start_at"`   // Verification/check-in opens
	VerifyEndAt     time.Time `gorm:"not null" json:"verify_end_at"`     // Verification/check-in closes
	Timezone        string    `gorm:"type:varchar(50);not null" json:"timezone"`

	// Age-Based Registration
	MinAge            *int `gorm:"type:int" json:"min_age"`
	MaxAge            *int `gorm:"type:int" json:"max_age"`
	RequireParentInfo bool `gorm:"default:false" json:"require_parent_info"`

	// Family Registration
	IsFamilyRegistration bool `gorm:"default:false" json:"is_family_registration"`
	MaxFamilyMembers     *int `gorm:"type:int" json:"max_family_members"`

	// Registration Configuration
	Flow                    string `gorm:"type:varchar(20);not null" json:"flow"`                   // direct, staged, free
	QuotaPerUser            int    `gorm:"not null;default:1" json:"quota_per_user"`                // Max registrations per user
	Capacity                int    `gorm:"not null" json:"capacity"`                                // Total capacity
	PostRegistrationDetails JSONB  `gorm:"type:jsonb" json:"post_registration_details"`             // Details shown after registration
	Status                  string `gorm:"type:varchar(20);not null;default:'draft'" json:"status"` // draft, active, cancelled, completed

	// Timestamps
	CreatedAt time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt sql.NullTime `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for GORM
func (EventInstance) TableName() string {
	return "event_instances"
}

// Helper Methods - Simple getters, NO validation, NO business logic
// IsActive returns true if the instance is in active status
func (ei *EventInstance) IsActive() bool {
	return ei.Status == "active"
}

// IsDraft returns true if the instance is in draft status
func (ei *EventInstance) IsDraft() bool {
	return ei.Status == "draft"
}

// IsOnline returns true if the instance is online-only
func (ei *EventInstance) IsOnline() bool {
	return ei.LocationType == "online"
}

// IsOffline returns true if the instance is offline-only
func (ei *EventInstance) IsOffline() bool {
	return ei.LocationType == "offline"
}

// IsHybrid returns true if the instance supports both online and offline
func (ei *EventInstance) IsHybrid() bool {
	return ei.LocationType == "hybrid"
}

type (
	CreateInstanceRequest struct {
		EventConfiguration EventConfiguration `json:"eventConfiguration" validate:"omitempty,dive"`
		InstanceRequest    []InstanceRequest  `json:"instanceRequest" validate:"required"`
	}
	EventConfiguration struct {
		Code          string `json:"code"`
		IsFollowEvent bool   `json:"isFollowEvent"`
	}
	InstanceRequest struct {
		Title                  string                     `json:"title" validate:"required"`
		Description            string                     `json:"description"`
		InstanceType           string                     `json:"instanceType" validate:"required,oneof=registration announcement volunteer-attendance"`
		RegistrationConfig     InstanceRegistrationConfig `json:"registrationConfig" validate:"required"`
		Schedule               InstanceSchedule           `json:"schedule" validate:"required"`
		RegistrationIdentifier InstanceIdentifierConfig   `json:"registrationIdentifier" validate:"omitempty,dive"`
		Location               EventLocation              `json:"location" validate:"required"`
		IsUpdateEventTime      bool                       `json:"isUpdateEventTime"`
		Status                 string                     `json:"status" validate:"required"`
	}
	InstanceRegistrationConfig struct {
		Capacity                int      `json:"capacity" validate:"required,numeric,min=1"`
		QuotaPerUser            int      `json:"quotaPerUser" validate:"required,numeric,min=1"`
		EnforceCommunityId      bool     `json:"enforceCommunityId"`
		EnforceUniqueness       bool     `json:"enforceUniqueness"`
		EnforceSelfRegistration bool     `json:"enforceSelfRegistration"`
		Methods                 []string `json:"methods" validate:"omitempty,dive,oneof=personal-qr event-qr registration-qr" example:"personal-qr"`
		Flow                    string   `json:"flow" validate:"required,oneof=direct staged free" example:"direct"`

		// Age-based registration (optional, flexible)
		MinAge            *int `json:"minAge" validate:"omitempty,min=0,max=150"`
		MaxAge            *int `json:"maxAge" validate:"omitempty,min=0,max=150,gtefield=MinAge"`
		RequireParentInfo bool `json:"requireParentInfo"` // Can be auto-set based on config

		// Family registration configuration
		IsFamilyRegistration bool `json:"isFamilyRegistration"`
		MaxFamilyMembers     *int `json:"maxFamilyMembers" validate:"omitempty,min=1,max=20"`
	}
	InstanceIdentifierConfig struct {
		ParentIdentifierFields []ResgitrationIdentifierField `json:"parentIdentifierFields" validate:"omitempty,dive"`
		ChildIdentifierFields  []ResgitrationIdentifierField `json:"childIdentifierFields" validate:"omitempty,dive"`
	}
	// IdentifierField defines the structure for a single identifier, including its type and whether it is mandatory.
	ResgitrationIdentifierField struct {
		Type        string `json:"type" validate:"oneof=email phone"`
		IsMandatory bool   `json:"isMandatory"`
	}
	InstanceSchedule struct {
		StartAt         time.Time `json:"startAt" validate:"required"`
		EndAt           time.Time `json:"endAt" validate:"required"`
		RegisterStartAt time.Time `json:"registerStartAt" validate:"required"`
		RegisterEndAt   time.Time `json:"registerEndAt" validate:"required"`
		VerifyStartAt   time.Time `json:"verifyStartAt" validate:"required"`
		VerifyEndAt     time.Time `json:"verifyEndAt" validate:"required"`
		Timezone        string    `json:"timezone" validate:"required"`
	}
	CreateInstanceResponse struct {
		Type               string                             `json:"type"`
		InstanceCode       string                             `json:"instanceCode"`
		EventCode          string                             `json:"eventCode"`
		Title              string                             `json:"title"`
		Description        string                             `json:"description"`
		IdentifierConfig   InstanceIdentifierConfigResponse   `json:"identifierConfig"`
		TimeConfig         InstanceTimeConfigResponse         `json:"timeConfig"`
		Location           EventLocation                      `json:"location"`
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
		Capacity                int      `json:"capacity" example:"100"`
		QuotaPerUser            int      `json:"quotaPerUser" example:"1"`
		EnforceCommunityId      bool     `json:"enforceCommunityId" example:"false"`
		EnforceUniqueness       bool     `json:"enforceUniqueness" example:"false"`
		EnforceSelfRegistration bool     `json:"enforceSelfRegistration" example:"false"`
		Methods                 []string `json:"methods" example:"personal-qr"`
		Flow                    string   `json:"flow" example:"direct"`
	}
	InstanceIdentifierConfigResponse struct {
		ParentIdentifierFields []InstanceIdentifierConfig `json:"parentIdentifierFields,omitempty"`
		ChildIdentifierFields  []InstanceIdentifierConfig `json:"childIdentifierFields,omitempty"`
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

// ToCreateResponse transforms an EventInstance model into a CreateInstanceResponse DTO
func (ei *EventInstance) ToCreateResponse() CreateInstanceResponse {
	return CreateInstanceResponse{
		Type:         TYPE_EVENT_INSTANCE,
		InstanceCode: ei.Code,
		EventCode:    ei.EventCode,
		Title:        ei.Title,
		Description:  ei.Description,
		IdentifierConfig: InstanceIdentifierConfigResponse{
			ParentIdentifierFields: []InstanceIdentifierConfig{},
			ChildIdentifierFields:  []InstanceIdentifierConfig{},
		},
		TimeConfig: InstanceTimeConfigResponse{
			StartAt:         ei.StartAt.Format(time.RFC3339),
			EndAt:           ei.EndAt.Format(time.RFC3339),
			RegisterStartAt: ei.RegisterStartAt.Format(time.RFC3339),
			RegisterEndAt:   ei.RegisterEndAt.Format(time.RFC3339),
			VerifyStartAt:   ei.VerifyStartAt.Format(time.RFC3339),
			VerifyEndAt:     ei.VerifyEndAt.Format(time.RFC3339),
			Timezone:        ei.Timezone,
		},
		Location: EventLocation{
			LocationType:    &ei.LocationType,
			PhysicalAddress: ei.PhysicalAddress,
			VirtualLink:     ei.VirtualLink,
			ClickToAction: ClickToAction{
				Text: ei.MeetingCTAText,
			},
			LocationDetails:    ei.LocationDetails,
			LocationVisibility: &ei.LocationVisibility,
		},
		RegistrationConfig: InstanceRegistrationConfigResponse{
			Capacity:                ei.Capacity,
			QuotaPerUser:            ei.QuotaPerUser,
			EnforceCommunityId:      ei.EnforceCommunityId,
			EnforceUniqueness:       ei.EnforceUniqueness,
			EnforceSelfRegistration: ei.EnforceSelfRegistration,
			Methods:                 ei.Methods,
			Flow:                    ei.Flow,
		},
		Status: ei.Status,
	}
}

// ToResponse transforms a slice of EventInstance to slice of CreateInstanceResponse
func ToInstanceResponses(instances *[]EventInstance) []CreateInstanceResponse {
	if instances == nil {
		return []CreateInstanceResponse{}
	}

	responses := make([]CreateInstanceResponse, len(*instances))
	for i, instance := range *instances {
		responses[i] = instance.ToCreateResponse()
	}
	return responses
}

// NewInstanceFromRequest constructs an EventInstance model from InstanceRequest
// This is the recommended way to create instances - keeps construction logic in the model layer
func NewInstanceFromRequest(
	req InstanceRequest,
	event *Event,
	instanceCode string,
	parentIdentifierFields, childIdentifierFields []byte,
) *EventInstance {
	return &EventInstance{
		Code:                    instanceCode,
		EventCode:               event.Code,
		Title:                   req.Title,
		Description:             req.Description,
		InstanceType:            req.InstanceType,
		ParentIdentifierFields:  parentIdentifierFields,
		ChildIdentifierFields:   childIdentifierFields,
		EnforceCommunityId:      req.RegistrationConfig.EnforceCommunityId,
		EnforceSelfRegistration: req.RegistrationConfig.EnforceSelfRegistration,
		EnforceUniqueness:       req.RegistrationConfig.EnforceUniqueness,
		Methods:                 req.RegistrationConfig.Methods,
		LocationType:            *req.Location.LocationType,
		PhysicalAddress:         req.Location.PhysicalAddress,
		VirtualLink:             req.Location.VirtualLink,
		MeetingCTAText:          req.Location.ClickToAction.Text,
		LocationDetails:         req.Location.LocationDetails,
		LocationVisibility:      *req.Location.LocationVisibility,
		StartAt:                 req.Schedule.StartAt,
		EndAt:                   req.Schedule.EndAt,
		RegisterStartAt:         req.Schedule.RegisterStartAt,
		RegisterEndAt:           req.Schedule.RegisterEndAt,
		VerifyStartAt:           req.Schedule.VerifyStartAt,
		VerifyEndAt:             req.Schedule.VerifyEndAt,
		Timezone:                req.Schedule.Timezone,
		MinAge:                  req.RegistrationConfig.MinAge,
		MaxAge:                  req.RegistrationConfig.MaxAge,
		RequireParentInfo:       req.RegistrationConfig.RequireParentInfo,
		IsFamilyRegistration:    req.RegistrationConfig.IsFamilyRegistration,
		MaxFamilyMembers:        req.RegistrationConfig.MaxFamilyMembers,
		Flow:                    req.RegistrationConfig.Flow,
		QuotaPerUser:            req.RegistrationConfig.QuotaPerUser,
		Capacity:                req.RegistrationConfig.Capacity,
		Status:                  req.Status,
	}
}
