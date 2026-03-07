package models

import (
	"time"

	"github.com/lib/pq"
)

var (
	TYPE_EVENT_SESSION = "event_session"
)

type EventSession struct {
	ID                int    `gorm:"type:integer;primarykey"`
	Code              string `gorm:"type:varchar(50);not null"`
	EventCode         string `gorm:"type:varchar(50);not null"`
	ParentSessionCode string `gorm:"type:varchar(50);not null"`

	Title       string `gorm:"type:text;not null"`
	Description string `gorm:"type:text"`
	SessionType string `gorm:"type:varchar(50);not null"`

	// Location Information
	LocationType       string  `gorm:"type:varchar(20);not null" json:"location_type"`       // online, offline, hybrid
	PhysicalPlaceName  *string `gorm:"type:text" json:"physical_place_name"`                 // Place name
	PhysicalAddress    *string `gorm:"type:text" json:"physical_address"`                    // Physical address
	VirtualLink        *string `gorm:"type:text" json:"virtual_link"`                        // Meeting link
	VirtualPlatform    *string `gorm:"type:text" json:"virtual_platform"`                    // Meeting platform: YouTube, Zoom, etc
	LocationDetails    *string `gorm:"type:text" json:"location_details"`                    // Additional location info
	LocationVisibility string  `gorm:"type:varchar(30);not null" json:"location_visibility"` // pre-registration, post-registration, all
	CTAText            *string `gorm:"type:varchar(100)" json:"cta_text"`                    // CTA button text
	CTALink            *string `gorm:"type:text" json:"cta_link"`                            // CTA button link

	Geolocation JSONB `gorm:"type:jsonb" json:"geolocation" swaggertype:"object"`

	// Scheduling
	StartAt  time.Time `gorm:"type:timestamptz;not null;index;index:idx_events_status_start,priority:2" json:"start_at"`
	EndAt    time.Time `gorm:"type:timestamptz;not null;index" json:"end_at"`
	Timezone string    `gorm:"type:varchar(50);not null" json:"timezone"` // IANA timezone (e.g., "Asia/Jakarta")

	// Registration Window
	RegistrationStartAt *time.Time `gorm:"type:timestamptz" json:"registration_start_at"`
	RegistrationEndAt   *time.Time `gorm:"type:timestamptz" json:"registration_end_at"`

	// Capacity
	Capacity         int  `gorm:"type:integer" json:"capacity"`
	WaitlistEnabled  bool `gorm:"type:boolean;default:false" json:"waitlist_enabled"`
	WaitlistCapacity int  `gorm:"type:integer" json:"waitlist_capacity"`

	// Registration Rules
	RequireApproval     bool           `gorm:"type:boolean;default:false" json:"require_approval"`
	RegistrationMethods pq.StringArray `gorm:"type:text[]" json:"registration_methods" swaggertype:"array,string"`

	// Group/Family Registration Config
	RegistrationMode        string `gorm:"type:varchar(30);default:'self_and_others'" json:"registration_mode"` // self_only, self_and_registered, self_and_others
	MaxRegistrationsPerUser int    `gorm:"type:integer;default:1" json:"max_registrations_per_user"`            // How many people a user can register (including self)
	OneSessionPerEvent      bool   `gorm:"type:boolean;default:false" json:"one_session_per_event"`             // If true, user can only register for ONE session in this event (example: if event has 4 sessions, user can only register for 1 session)

	// // Form Requirements for Additional Registrants
	// AdditionalRegistrantFormMode string `gorm:"type:varchar(30);default:'name_only'" json:"additional_registrant_form_mode"` // same_as_primary, name_only, custom
	// AdditionalRegistrantFormID   *int   `gorm:"type:bigint" json:"additional_registrant_form_id"`

	// Check-in Configuration (Detailed)
	CheckInEnabled       bool       `gorm:"type:boolean;default:true" json:"check_in_enabled"`   // Is check-in feature enabled?
	CheckInRequired      bool       `gorm:"type:boolean;default:false" json:"check_in_required"` // Is check-in mandatory?
	CheckInStartAt       *time.Time `gorm:"type:timestamptz" json:"check_in_start_at"`
	CheckInEndAt         *time.Time `gorm:"type:timestamptz" json:"check_in_end_at"`
	CheckInAllowLate     bool       `gorm:"type:boolean;default:true" json:"check_in_allow_late"`   // Allow check-in after window?
	CheckInLateThreshold int        `gorm:"type:integer;default:10" json:"check_in_late_threshold"` // Minutes to mark as "late"

	// Check-out Configuration (Detailed)
	CheckOutEnabled       bool       `gorm:"type:boolean;default:false" json:"check_out_enabled"`  // Is check-out feature enabled?
	CheckOutRequired      bool       `gorm:"type:boolean;default:false" json:"check_out_required"` // Is check-out mandatory?
	CheckOutStartAt       *time.Time `gorm:"type:timestamptz" json:"check_out_start_at"`
	CheckOutEndAt         *time.Time `gorm:"type:timestamptz" json:"check_out_end_at"`
	CheckOutAllowLate     bool       `gorm:"type:boolean;default:true" json:"check_out_allow_late"`   // Allow check-out after window?
	CheckOutLateThreshold int        `gorm:"type:integer;default:30" json:"check_out_late_threshold"` // Minutes to mark as "late checkout"

	// Age/Eligibility (for kids, youth, etc.)
	MinAge        int    `gorm:"type:integer" json:"min_age"`
	MaxAge        int    `gorm:"type:integer" json:"max_age"`
	Prerequisites string `gorm:"type:text" json:"prerequisites"`

	// Custom Questions (for primary registrant)
	// RegistrationFormID *int `gorm:"type:bigint" json:"registration_form_id"`

	// Identifier Configuration (JSONB for flexibility)
	// Separate configs for primary registrant and additional registrants
	IdentifierConfig JSONB `gorm:"type:jsonb" json:"identifier_config" swaggertype:"object"`

	// Form Field Overrides (session-specific field visibility/requirements)
	// FormFieldOverrides JSONB `gorm:"type:jsonb" json:"form_field_overrides" swaggertype:"object"`

	// Status
	Status    string     `gorm:"type:varchar(20);not null;default:'draft'" json:"status"` // draft, published, cancelled, archived
	CreatedAt time.Time  `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt *time.Time `gorm:"type:timestamptz" json:"deleted_at"`
}

// TableName specifies the table name for GORM
func (EventSession) TableName() string {
	return "event_sessions"
}

// Common Requests
type (
	QRValidationRule struct {
		QRType        string `json:"qrType" validate:"required,oneof=personal-qr event-qr session-qr registration-qr"`
		Required      bool   `json:"required"`
		AllowOverride bool   `json:"allowOverride"`
	}

	SessionTimeConfiguration struct {
		Registration EventSchedule         `json:"registration" validate:"omitempty"`
		CheckIn      SessionCheckInConfig  `json:"checkIn" validate:"omitempty"`
		CheckOut     SessionCheckOutConfig `json:"checkOut" validate:"omitempty"`
	}

	SessionCapacity struct {
		Capacity         int  `json:"capacity" validate:"required"`
		WaitlistEnabled  bool `json:"waitlistEnabled"`
		WaitlistCapacity int  `json:"waitlistCapacity"`
	}

	SessionRules struct {
		RequireApproval         bool     `json:"requireApproval"`
		RegistrationMethods     []string `json:"registrationMethods" validate:"omitempty,dive,oneof=personal-qr event-qr session-qr registration-qr"`
		RegistrationMode        string   `json:"registrationMode" validate:"required,oneof=self_only self_and_registered self_and_others"`
		MaxRegistrationsPerUser int      `json:"maxRegistrationsPerUser"`
		OneSessionPerEvent      bool     `json:"oneSessionPerEvent"`
		MinAge                  int      `json:"minAge"`
		MaxAge                  int      `json:"maxAge"`
		Prerequisites           string   `json:"prerequisites"`
	}

	SessionCheckInConfig struct {
		Enabled       bool       `json:"enabled"`                                                                                         // Is check-in feature enabled?
		Required      bool       `json:"required"`                                                                                        // Is check-in mandatory?
		StartAt       *time.Time `gorm:"type:timestamptz" json:"startAt"`                                                                 // Check-in start time
		EndAt         *time.Time `gorm:"type:timestamptz" json:"endAt"`                                                                   // Check-in end time
		AllowLate     bool       `json:"allowLate"`                                                                                       // Allow check-in after window?
		TrackLate     bool       `json:"trackLate"`                                                                                       // Whether to track/flag late check-ins
		LatePolicy    string     `json:"latePolicy" validate:"omitempty,oneof=reject warn allow"`                                         // How to handle late check-ins: reject (hard cutoff), warn (accept but flag), allow (accept silently)
		LateThreshold int        `json:"lateThreshold" validate:"required_if=LatePolicy warn,required_if=TrackLate true,omitempty,min=0"` // Minutes to mark as "late"
	}

	SessionCheckOutConfig struct {
		Enabled       bool       `json:"enabled"`                                                                                         // Is check-out feature enabled?
		Required      bool       `json:"required"`                                                                                        // Is check-out mandatory?
		StartAt       *time.Time `gorm:"type:timestamptz" json:"startAt"`                                                                 // Check-out start time
		EndAt         *time.Time `gorm:"type:timestamptz" json:"endAt"`                                                                   // Check-out end time
		AllowLate     bool       `json:"allowLate"`                                                                                       // Allow check-out after window?
		TrackLate     bool       `json:"trackLate"`                                                                                       // Whether to track/flag late check-outs
		LatePolicy    string     `json:"latePolicy" validate:"omitempty,oneof=reject warn allow"`                                         // How to handle late check-outs: reject (hard cutoff), warn (accept but flag), allow (accept silently)
		LateThreshold int        `json:"lateThreshold" validate:"required_if=LatePolicy warn,required_if=TrackLate true,omitempty,min=0"` // Minutes to mark as "late checkout"
	}

	SessionIdentifierConfig struct {
		Primary    PrimaryIdentifierConfig    `json:"primary" validate:"omitempty"`
		Additional AdditionalIdentifierConfig `json:"additional" validate:"omitempty"`
	}

	PrimaryIdentifierConfig struct {
		Name        FieldConfig `json:"name" validate:"omitempty"`
		Email       FieldConfig `json:"email" validate:"omitempty"`
		Phone       FieldConfig `json:"phone" validate:"omitempty"`
		CommunityID FieldConfig `json:"communityId" validate:"omitempty"`
	}

	AdditionalIdentifierConfig struct {
		Name        FieldConfig `json:"name" validate:"omitempty"`
		Email       FieldConfig `json:"email" validate:"omitempty"`
		Phone       FieldConfig `json:"phone" validate:"omitempty"`
		CommunityID FieldConfig `json:"communityId" validate:"omitempty"`
	}

	FieldConfig struct {
		Visible  bool `json:"visible"`
		Required bool `json:"required"`
	}
)

type (
	CreateEventSessionRequest struct {
		Title             string                          `json:"title" validate:"required"`
		ParentSessionCode string                          `json:"parentSessionCode" validate:"omitempty"`
		Description       string                          `json:"description" validate:"omitempty"`
		SessionType       string                          `json:"sessionType" validate:"required,oneof=service class track breakout workshop general kids youth teen adult"`
		Status            string                          `json:"status" validate:"required,oneof=draft active inactive"`
		IsUpdateEvent     bool                            `json:"isUpdateEvent"`
		Location          EventLocation                   `json:"location" validate:"omitempty"`
		Geolocation       *GeolocationConfiguration       `json:"geolocation" validate:"omitempty"`
		Schedule          EventSchedule                   `json:"schedule" validate:"omitempty"`
		Times             SessionTimeConfiguration        `json:"times" validate:"omitempty"`
		SessionCapacity   SessionCapacity                 `json:"sessionCapacity" validate:"required"`
		SessionRules      SessionRules                    `json:"sessionRules" validate:"required"`
		Questions         []BulkCreateFormQuestionRequest `json:"questions" validate:"omitempty,dive"`
	}

	CreateEventSessionResponse struct {
		Type              string                   `json:"type"`
		Code              string                   `json:"code" example:"event-code"`
		EventCode         string                   `json:"eventCode" example:"event-code"`
		ParentSessionCode string                   `json:"parentSessionCode" example:"event-code"`
		Title             string                   `json:"title" example:"Event Title"`
		Description       string                   `json:"description" example:"Event Description"`
		SessionType       string                   `json:"sessionType" example:"service"`
		Status            string                   `json:"status" example:"draft"`
		Location          EventLocation            `json:"location"`
		Geolocation       GeolocationConfiguration `json:"geolocation"`
		Schedule          EventSchedule            `json:"schedule"`
		Times             SessionTimeConfiguration `json:"times"`
		SessionCapacity   SessionCapacity          `json:"sessionCapacity"`
		SessionRules      SessionRules             `json:"sessionRules"`
		CheckInConfig     SessionCheckInConfig     `json:"checkInConfig"`
		CheckOutConfig    SessionCheckOutConfig    `json:"checkOutConfig"`
		Questions         []FormQuestionResponse   `json:"questions" validate:"omitempty,dive"`
	}

	// CreateEventResponseOption defined as a function that modifies CreateEventResponse
	CreateEventSessionResponseOption func(*CreateEventSessionResponse)
)

// WithFormQuestions adds questions to the response
func EventSessionWithFormQuestions(questions []FormQuestionResponse) CreateEventSessionResponseOption {
	return func(r *CreateEventSessionResponse) {
		r.Questions = questions
	}
}

func (es *EventSession) ToResponse(options ...CreateEventSessionResponseOption) CreateEventSessionResponse {
	var geo GeolocationConfiguration
	if !es.Geolocation.IsNull() {
		_ = es.Geolocation.Unmarshal(&geo)
	}

	var ic *SessionIdentifierConfig
	if !es.IdentifierConfig.IsNull() {
		ic = &SessionIdentifierConfig{}
		_ = es.IdentifierConfig.Unmarshal(ic)
	}
	_ = ic // reserved for future use

	response := CreateEventSessionResponse{
		Type:              TYPE_EVENT_SESSION,
		Code:              es.Code,
		EventCode:         es.EventCode,
		ParentSessionCode: es.ParentSessionCode,
		Title:             es.Title,
		Description:       es.Description,
		SessionType:       es.SessionType,
		Status:            es.Status,
		Location: EventLocation{
			LocationType:    &es.LocationType,
			PhysicalAddress: es.PhysicalAddress,
			VirtualLink:     es.VirtualLink,
			VirtualPlatform: es.VirtualPlatform,
			ClickToAction: ClickToAction{
				Text: es.CTAText,
				Link: es.CTALink,
			},
			LocationDetails:    es.LocationDetails,
			LocationVisibility: &es.LocationVisibility,
		},
		Geolocation: geo,
		Schedule: EventSchedule{
			StartAt:  &es.StartAt,
			EndAt:    &es.EndAt,
			Timezone: &es.Timezone,
		},
		Times: SessionTimeConfiguration{
			Registration: EventSchedule{
				StartAt:  es.RegistrationStartAt,
				EndAt:    es.RegistrationEndAt,
				Timezone: &es.Timezone,
			},
			CheckIn: SessionCheckInConfig{
				Enabled:       es.CheckInEnabled,
				Required:      es.CheckInRequired,
				StartAt:       es.CheckInStartAt,
				EndAt:         es.CheckInEndAt,
				AllowLate:     es.CheckInAllowLate,
				LateThreshold: es.CheckInLateThreshold,
			},
			CheckOut: SessionCheckOutConfig{
				Enabled:       es.CheckOutEnabled,
				Required:      es.CheckOutRequired,
				StartAt:       es.CheckOutStartAt,
				EndAt:         es.CheckOutEndAt,
				AllowLate:     es.CheckOutAllowLate,
				LateThreshold: es.CheckOutLateThreshold,
			},
		},
		SessionCapacity: SessionCapacity{
			Capacity:         es.Capacity,
			WaitlistEnabled:  es.WaitlistEnabled,
			WaitlistCapacity: es.WaitlistCapacity,
		},
		SessionRules: SessionRules{
			RequireApproval:         es.RequireApproval,
			RegistrationMode:        es.RegistrationMode,
			MaxRegistrationsPerUser: es.MaxRegistrationsPerUser,
			OneSessionPerEvent:      es.OneSessionPerEvent,
			MinAge:                  es.MinAge,
			MaxAge:                  es.MaxAge,
			Prerequisites:           es.Prerequisites,
		},
	}

	// Apply options
	for _, option := range options {
		option(&response)
	}

	return response
}
