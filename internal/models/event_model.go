package models

import (
	"errors"
	"fmt"
	"go-community/internal/constants"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

const (
	TYPE_EVENT = "event"
)

// Event represents a church event in the system
type Event struct {
	// Primary Key
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Code string `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`

	// Core Event Information
	Title              string  `gorm:"type:varchar(255);not null;index" json:"title"`
	Slug               string  `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"` // Slugified title for URLs
	PreDescription     *string `gorm:"type:text" json:"pre_description"`                   // Details shown before registration
	PostDescription    JSONB   `gorm:"type:jsonb" json:"post_description"`                 // Details shown after registration
	TermsAndConditions *string `gorm:"type:text" json:"terms_and_conditions"`              // Terms and conditions
	Category           string  `gorm:"type:varchar(30);not null" json:"category"`          // announcement, registerable

	// Media
	ImageURLs pq.StringArray `gorm:"type:text[]" json:"image_url"`
	BannerURL string         `gorm:"type:text" json:"banner_url"`

	// Organization
	CreatorCommunityID    string         `gorm:"type:varchar(50);not null;index" json:"creator_community_id"`
	OrganizerCommunityIDs pq.StringArray `gorm:"type:text[]" json:"organizer_community_ids"`
	ContactCommunityIDs   pq.StringArray `gorm:"type:text[]" json:"contact_community_ids"`

	// Visibility & Access
	AccessLevel         string         `gorm:"type:varchar(20);not null;index" json:"access_level"` // public, private
	AllowedUserTypes    pq.StringArray `gorm:"type:text[]" json:"allowed_user_types"`
	AllowedRoles        pq.StringArray `gorm:"type:text[]" json:"allowed_roles"`
	AllowedCampuses     pq.StringArray `gorm:"type:text[]" json:"allowed_campuses"`
	AllowedCommunityIDs pq.StringArray `gorm:"type:text[]" json:"allowed_community_ids"`

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

	// Scheduling
	StartAt  time.Time `gorm:"type:timestamptz;not null;index;index:idx_events_status_start,priority:2" json:"start_at"`
	EndAt    time.Time `gorm:"type:timestamptz;not null;index" json:"end_at"`
	Timezone string    `gorm:"type:varchar(50);not null" json:"timezone"` // IANA timezone (e.g., "Asia/Jakarta")

	// Recurrence (For Sunday Service type)
	IsRecurring       bool  `gorm:"type:boolean;default:false" json:"is_recurring"`
	RecurrencePattern JSONB `gorm:"type:jsonb" json:"recurrence_pattern" swaggertype:"object"`

	// Template/Series
	IsTemplate bool   `gorm:"type:boolean;default:false" json:"is_template"`
	TemplateID string `gorm:"type:varchar(255)" json:"template_id"`
	SeriesID   string `gorm:"type:varchar(255)" json:"series_id"`

	// Registration Configuration
	SessionPerUser int `gorm:"type:integer;default:0" json:"session_per_user"` // 0 = unlimited

	// Notification
	NotificationChannels pq.StringArray `gorm:"type:text[]" json:"notification_channels"`
	ReminderConfig       JSONB          `gorm:"type:jsonb" json:"reminder_config"`

	// Status & Metadata
	Status    string         `gorm:"type:varchar(20);not null;index;index:idx_events_status_start,priority:1;default:'draft'" json:"status"`
	CreatedAt time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for GORM
func (Event) TableName() string {
	return "events"
}

// ============================================================================
// Event Helper Methods
// ============================================================================
//
// These methods provide convenient access to event properties and business logic.
// They mirror the DTO helper methods but work with the flattened Event structure.
//
// Categories:
//   - Access Control: IsPublic, IsPrivate, HasRestrictions
//   - Location: IsOnline, IsOffline, IsHybrid, HasPhysicalLocation, HasVirtualLocation
//   - Schedule: IsUpcoming, IsOngoing, IsPast, Duration
//   - Status: IsDraft, IsActive, IsInactive
//   - Recurrence: HasRecurrence, HasRecurrencePattern
//   - Template: IsTemplate, IsFromTemplate, IsPartOfSeries
//   - Media: HasImages, HasBanner, ImageCount
//   - Notification: HasNotifications, HasReminders
// ============================================================================

// ----------------------------------------------------------------------------
// Access Control Methods
// ----------------------------------------------------------------------------

// IsPublic returns true if the event is publicly accessible
func (e *Event) IsPublic() bool {
	return e.AccessLevel == string(constants.AccessLevelPublic)
}

// IsPrivate returns true if the event is private
func (e *Event) IsPrivate() bool {
	return e.AccessLevel == string(constants.AccessLevelPrivate)
}

// HasRestrictions returns true if any access restrictions are defined
func (e *Event) HasRestrictions() bool {
	return len(e.AllowedUserTypes) > 0 ||
		len(e.AllowedRoles) > 0 ||
		len(e.AllowedCampuses) > 0 ||
		len(e.AllowedCommunityIDs) > 0
}

// ----------------------------------------------------------------------------
// Location Methods
// ----------------------------------------------------------------------------

// IsOnline returns true if the event is online only
func (e *Event) IsOnline() bool {
	return e.LocationType == string(constants.LocationTypeOnline)
}

// IsOffline returns true if the event is offline only
func (e *Event) IsOffline() bool {
	return e.LocationType == string(constants.LocationTypeOffline)
}

// IsHybrid returns true if the event is hybrid
func (e *Event) IsHybrid() bool {
	return e.LocationType == string(constants.LocationTypeHybrid)
}

// HasPhysicalLocation returns true if event has a physical location
func (e *Event) HasPhysicalLocation() bool {
	return e.IsOffline() || e.IsHybrid()
}

// HasVirtualLocation returns true if event has a virtual location
func (e *Event) HasVirtualLocation() bool {
	return e.IsOnline() || e.IsHybrid()
}

// HasLocationDetails returns true if additional location details are provided
func (e *Event) HasLocationDetails() bool {
	return e.LocationDetails != nil && *e.LocationDetails != ""
}

func (e *Event) HaveLocation() bool {
	return e.IsOnline() || e.IsOffline() || e.IsHybrid() || e.LocationType != ""
}

// ----------------------------------------------------------------------------
// Schedule Methods
// ----------------------------------------------------------------------------

// Duration returns the duration of the event
func (e *Event) Duration() time.Duration {
	return e.EndAt.Sub(e.StartAt)
}

// IsUpcoming returns true if event hasn't started yet
func (e *Event) IsUpcoming() bool {
	return time.Now().Before(e.StartAt)
}

// IsOngoing returns true if event is currently happening
func (e *Event) IsOngoing() bool {
	now := time.Now()
	return now.After(e.StartAt) && now.Before(e.EndAt)
}

// IsPast returns true if event has ended
func (e *Event) IsPast() bool {
	return time.Now().After(e.EndAt)
}

// DaysUntilStart returns the number of days until the event starts
// Returns negative if event has already started
func (e *Event) DaysUntilStart() int {
	duration := time.Until(e.StartAt)
	return int(duration.Hours() / 24)
}

// ----------------------------------------------------------------------------
// Status Methods
// ----------------------------------------------------------------------------

// IsDraft returns true if the event is in draft status
func (e *Event) IsDraft() bool {
	return e.Status == string(constants.EventStatusDraft)
}

// IsActive returns true if the event is active
func (e *Event) IsActive() bool {
	return e.Status == string(constants.EventStatusActive)
}

// IsInactive returns true if the event is inactive
func (e *Event) IsInactive() bool {
	return e.Status == string(constants.EventStatusInactive)
}

// ----------------------------------------------------------------------------
// Recurrence Methods
// ----------------------------------------------------------------------------

// HasRecurrence returns true if the event is recurring
func (e *Event) HasRecurrence() bool {
	return e.IsRecurring
}

// HasRecurrencePattern returns true if recurrence pattern is defined
func (e *Event) HasRecurrencePattern() bool {
	return e.IsRecurring && !e.RecurrencePattern.IsNull()
}

// ----------------------------------------------------------------------------
// Template & Series Methods
// ----------------------------------------------------------------------------

// IsTemplateEvent returns true if this event is a template
func (e *Event) IsTemplateEvent() bool {
	return e.IsTemplate
}

// IsFromTemplate returns true if event was created from a template
func (e *Event) IsFromTemplate() bool {
	return e.TemplateID != ""
}

// IsPartOfSeries returns true if event is part of a series
func (e *Event) IsPartOfSeries() bool {
	return e.SeriesID != ""
}

// ----------------------------------------------------------------------------
// Media Methods
// ----------------------------------------------------------------------------

// HasImages returns true if event has any images
func (e *Event) HasImages() bool {
	return len(e.ImageURLs) > 0 || e.BannerURL != ""
}

// HasBanner returns true if event has a banner image
func (e *Event) HasBanner() bool {
	return e.BannerURL != ""
}

// ImageCount returns the total number of images including banner
func (e *Event) ImageCount() int {
	count := len(e.ImageURLs)
	if e.BannerURL != "" {
		count++
	}
	return count
}

// ----------------------------------------------------------------------------
// Organization Methods
// ----------------------------------------------------------------------------

// HasOrganizers returns true if event has organizers
func (e *Event) HasOrganizers() bool {
	return len(e.OrganizerCommunityIDs) > 0
}

// HasContacts returns true if event has contact persons
func (e *Event) HasContacts() bool {
	return len(e.ContactCommunityIDs) > 0
}

// OrganizerCount returns the number of organizers
func (e *Event) OrganizerCount() int {
	return len(e.OrganizerCommunityIDs)
}

// ContactCount returns the number of contacts
func (e *Event) ContactCount() int {
	return len(e.ContactCommunityIDs)
}

// ----------------------------------------------------------------------------
// Notification Methods
// ----------------------------------------------------------------------------

// HasNotifications returns true if any notification channels are configured
func (e *Event) HasNotifications() bool {
	return len(e.NotificationChannels) > 0
}

// HasReminders returns true if reminder configuration is present and enabled
func (e *Event) HasReminders() bool {
	if e.ReminderConfig.IsNull() {
		return false
	}
	var rc ReminderConfig
	if err := e.ReminderConfig.Unmarshal(&rc); err != nil {
		return false
	}
	return rc.Enabled && len(rc.Intervals) > 0
}

// NotificationChannelCount returns the number of notification channels
func (e *Event) NotificationChannelCount() int {
	return len(e.NotificationChannels)
}

// ============================================================================
// Shared Request Components
// ============================================================================
//
// Design Decision: These structs are shared between CreateEventRequest and
// UpdateEventRequest to maximize reusability and maintain consistency.
//
// Benefits:
//   - Single source of truth for validation rules
//   - Consistent API contracts across operations
//   - Reduced code duplication
//   - Easier maintenance and updates
//
// Usage:
//   - CreateEventRequest: Uses these for initial event creation
//   - UpdateEventRequest: Reuses these for partial updates
//   - Response DTOs: Can transform from Event model to these
// ============================================================================

// EventImages contains all image-related fields for an event
//
// Validation:
//   - ImageLinks: Optional, max 10 images, each must be valid URL ≤ 2048 chars
//   - BannerLink: Optional, must be valid URL ≤ 2048 chars
//   - bannerRequiresImages: BannerLink can only have a value if ImageLinks is not empty
type EventImages struct {
	ImageLinks []string `json:"imageLinks" validate:"omitempty,max=10,dive,url,max=2048" example:"https://example.com/image1.jpg,https://example.com/image2.jpg"`
	BannerLink *string  `json:"bannerLink" validate:"bannerRequiresImages,omitempty,url,max=2048" example:"https://example.com/banner.jpg"`
}

// HasImages returns true if at least one image is present
func (ei *EventImages) HasImages() bool {
	return len(ei.ImageLinks) > 0 || (ei.BannerLink != nil && *ei.BannerLink != "")
}

// ImageCount returns the total number of images including banner
func (ei *EventImages) ImageCount() int {
	count := len(ei.ImageLinks)
	if ei.BannerLink != nil && *ei.BannerLink != "" {
		count++
	}
	return count
}

// EventOrganizer contains organizer and contact information
//
// Validation:
//   - OrganizerCommunityIDs: Required, min 1, max 50, must be valid UUIDs
//   - ContactCommunityIDs: Optional, max 50, must be valid UUIDs
//
// Business Rules:
//   - At least one organizer is required for event accountability
//   - Contacts are optional and used for event inquiries
type EventOrganizer struct {
	OrganizerCommunityIDs []string `json:"organizerCommunityIds" validate:"required,min=1,max=50,dive,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
	ContactCommunityIDs   []string `json:"contactCommunityIds" validate:"omitempty,max=50,dive,uuid4" example:"550e8400-e29b-41d4-a716-446655440001"`
}

// HasContacts returns true if contact community IDs are provided
func (eo *EventOrganizer) HasContacts() bool {
	return len(eo.ContactCommunityIDs) > 0
}

// TotalPeople returns the total number of unique organizers and contacts
func (eo *EventOrganizer) TotalPeople() int {
	unique := make(map[string]bool)
	for _, id := range eo.OrganizerCommunityIDs {
		unique[id] = true
	}
	for _, id := range eo.ContactCommunityIDs {
		unique[id] = true
	}
	return len(unique)
}

// EventAccess defines access control and visibility settings
//
// Validation:
//   - AccessLevel: Required, must be "public" or "private"
//   - AllowedUserTypes: Optional, validated against system user types
//   - AllowedRoles: Optional, validated against system roles
//   - AllowedCampuses: Optional, validated against configured campuses
//   - AllowedCommunityIDs: Optional, must be valid UUIDs
//
// Business Rules:
//   - Public events: All restriction fields should be empty/nil
//   - Private events: At least one restriction should be specified
type EventAccess struct {
	AccessLevel         *string  `json:"accessLevel" validate:"required,oneof=public private"`
	AllowedUserTypes    []string `json:"allowedUserTypes" validate:"omitempty,max=20,dive,min=1,max=50"`
	AllowedRoles        []string `json:"allowedRoles" validate:"omitempty,max=20,dive,min=1,max=50"`
	AllowedCampuses     []string `json:"allowedCampuses" validate:"omitempty,max=50,dive,campusCodes"`
	AllowedCommunityIDs []string `json:"allowedCommunityIds" validate:"omitempty,max=100,dive,uuid4"`
}

// IsPublic returns true if access level is public
func (ea *EventAccess) IsPublic() bool {
	return ea.AccessLevel != nil && *ea.AccessLevel == string(constants.AccessLevelPublic)
}

// IsPrivate returns true if access level is private
func (ea *EventAccess) IsPrivate() bool {
	return ea.AccessLevel != nil && *ea.AccessLevel == string(constants.AccessLevelPrivate)
}

// HasRestrictions returns true if any access restrictions are defined
func (ea *EventAccess) HasRestrictions() bool {
	return len(ea.AllowedUserTypes) > 0 ||
		len(ea.AllowedRoles) > 0 ||
		len(ea.AllowedCampuses) > 0 ||
		len(ea.AllowedCommunityIDs) > 0
}

// EventLocation contains all location-related information
//
// Validation:
//   - LocationType: Required, must be "online", "offline", or "hybrid"
//   - LocationVisibility: Required, controls when location is shown
//   - Physical fields: Required for offline/hybrid events (via required_if)
//   - Virtual fields: Required for online/hybrid events (via required_if)
//
// Business Rules:
//   - Offline: Must have PhysicalPlaceName and PhysicalAddress
//   - Online: Must have VirtualLink and VirtualPlatform
//   - Hybrid: Must have both physical and virtual fields
//
// Note: required_if validation works by checking the sibling field value.
// Syntax: required_if=FieldName value1,required_if=FieldName value2
type EventLocation struct {
	LocationType       *string       `json:"locationType" validate:"required,oneof=online offline hybrid"`
	LocationVisibility *string       `json:"locationVisibility" validate:"required,oneof=pre-registration post-registration all"`
	PhysicalPlaceName  *string       `json:"physicalPlaceName" validate:"omitempty,required_if=LocationType offline,required_if=LocationType hybrid,min=1,max=255"`
	PhysicalAddress    *string       `json:"physicalAddress" validate:"omitempty,required_if=LocationType offline,required_if=LocationType hybrid,min=1,max=500"`
	VirtualLink        *string       `json:"virtualLink" validate:"omitempty,required_if=LocationType online,required_if=LocationType hybrid,url,max=2048"`
	VirtualPlatform    *string       `json:"virtualPlatform" validate:"omitempty,required_if=LocationType online,required_if=LocationType hybrid,min=1,max=50"`
	ClickToAction      ClickToAction `json:"clickToAction" validate:"omitempty,dive"`
	LocationDetails    *string       `json:"locationDetails" validate:"omitempty,max=1000"`
}

// IsOnline returns true if location type is online
func (el *EventLocation) IsOnline() bool {
	return el.LocationType != nil && *el.LocationType == string(constants.LocationTypeOnline)
}

// IsOffline returns true if location type is offline
func (el *EventLocation) IsOffline() bool {
	return el.LocationType != nil && *el.LocationType == string(constants.LocationTypeOffline)
}

// IsHybrid returns true if location type is hybrid
func (el *EventLocation) IsHybrid() bool {
	return el.LocationType != nil && *el.LocationType == string(constants.LocationTypeHybrid)
}

// HasPhysicalLocation returns true if event has physical location
func (el *EventLocation) HasPhysicalLocation() bool {
	return el.IsOffline() || el.IsHybrid()
}

// HasVirtualLocation returns true if event has virtual location
func (el *EventLocation) HasVirtualLocation() bool {
	return el.IsOnline() || el.IsHybrid()
}

// Validate checks if the location data is valid for the given category
func (el *EventLocation) Validate(category *string) error {
	if el.LocationType == nil {
		return errors.New("location type is required")
	}

	locationType := *el.LocationType

	if category != nil && *category == "announcement" {
		if locationType != string(constants.LocationTypeOnline) {
			return errors.New("location type must be online for announcement")
		}
	}

	switch locationType {
	case string(constants.LocationTypeOnline):
		// Online events must have virtual link and platform
		if el.VirtualLink == nil || *el.VirtualLink == "" {
			return errors.New("virtual link is required for online events")
		}
		if el.VirtualPlatform == nil || *el.VirtualPlatform == "" {
			return errors.New("virtual platform is required for online events")
		}

	case string(constants.LocationTypeOffline):
		// Offline events must have physical address
		if el.PhysicalAddress == nil || *el.PhysicalAddress == "" {
			return errors.New("physical address is required for offline events")
		}

	case string(constants.LocationTypeHybrid):
		// Hybrid events must have both virtual and physical locations
		if el.VirtualLink == nil || *el.VirtualLink == "" {
			return errors.New("virtual link is required for hybrid events")
		}
		if el.VirtualPlatform == nil || *el.VirtualPlatform == "" {
			return errors.New("virtual platform is required for hybrid events")
		}
		if el.PhysicalAddress == nil || *el.PhysicalAddress == "" {
			return errors.New("physical address is required for hybrid events")
		}

	default:
		return fmt.Errorf("invalid location type: %s", locationType)
	}

	return nil
}

// ClickToAction defines a call-to-action button configuration
//
// Validation:
//   - Text: Optional, min 1, max 100 characters
//   - Link: Optional, must be valid URL or special value "NORMAL_FLOW"
type ClickToAction struct {
	Text *string `json:"text" validate:"omitempty,min=1,max=100" example:"Register Now!"`
	Link *string `json:"link" validate:"omitempty,max=2048" example:"https://register.example.com"`
}

// IsNormalFlow returns true if CTA uses the normal registration flow
func (cta *ClickToAction) IsNormalFlow() bool {
	return cta.Link != nil && *cta.Link == "NORMAL_FLOW"
}

// HasCustomLink returns true if CTA has a custom external link
func (cta *ClickToAction) HasCustomLink() bool {
	return cta.Link != nil && *cta.Link != "" && *cta.Link != "NORMAL_FLOW"
}

func (cta *ClickToAction) TextNotEmpty() bool {
	return cta.Text != nil && *cta.Text != ""
}

func (cta *ClickToAction) LinkNotEmpty() bool {
	return cta.Link != nil && *cta.Link != ""
}

// EventSchedule contains scheduling information
//
// Validation:
//   - StartAt: Required, must be a valid timestamp
//   - EndAt: Required, must be after StartAt (via gtfield)
//   - Timezone: Optional, defaults to system timezone, must be valid IANA timezone
//
// Note: gtfield validation compares the EndAt field with StartAt field.
// With pointer fields, it compares the dereferenced values if both are non-nil.
// Additional validation in usecase layer is recommended for nil-safety.
type EventSchedule struct {
	StartAt  *time.Time `json:"startAt" validate:"required" example:"2026-12-25T09:00:00Z"`
	EndAt    *time.Time `json:"endAt" validate:"required,gtfield=StartAt" example:"2026-12-25T11:00:00Z"`
	Timezone *string    `json:"timezone" validate:"omitempty,timezone" example:"Asia/Jakarta"`
}

// Duration returns the duration of the event
func (es *EventSchedule) Duration() time.Duration {
	if es.StartAt == nil || es.EndAt == nil {
		return 0
	}
	return es.EndAt.Sub(*es.StartAt)
}

// IsValid performs basic schedule validation
func (es *EventSchedule) IsValid() bool {
	if es.StartAt == nil || es.EndAt == nil {
		return false
	}
	return es.EndAt.After(*es.StartAt)
}

// EventRecurrence defines recurrence pattern for recurring events
//
// Validation:
//   - IsRecurring: Boolean flag
//   - RecurrencePattern: Required if IsRecurring is true
//   - RecurrenceEndDate: Optional, defines when recurrence stops
//
// Business Rules:
//   - If IsRecurring is true, RecurrencePattern must be provided
//   - If IsRecurring is false, RecurrencePattern should be nil
type EventRecurrence struct {
	IsRecurring       bool               `json:"isRecurring" example:"true"`
	RecurrencePattern *RecurrencePattern `json:"recurrencePattern" validate:"omitempty,dive"`
	RecurrenceEndDate *time.Time         `json:"recurrenceEndDate" validate:"omitempty" example:"2027-12-31T23:59:59Z"`
}

// HasPattern returns true if recurrence pattern is defined
func (er *EventRecurrence) HasPattern() bool {
	return er.RecurrencePattern != nil
}

// EventTemplate contains template and series information
//
// Validation:
//   - IsTemplate: Boolean flag indicating if this event is a template
//   - TemplateID: Optional, references the template this event was created from
//   - SeriesID: Optional, groups related events together
//
// Business Rules:
//   - An event can be both a template and created from a template
//   - SeriesID links events in a series (e.g., "Summer Camp 2026" series)
type EventTemplate struct {
	IsTemplate bool    `json:"isTemplate" example:"false"`
	TemplateID *string `json:"templateId" validate:"omitempty,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
	SeriesID   *string `json:"seriesId" validate:"omitempty,uuid4" example:"550e8400-e29b-41d4-a716-446655440001"`
}

// IsFromTemplate returns true if event was created from a template
func (et *EventTemplate) IsFromTemplate() bool {
	return et.TemplateID != nil && *et.TemplateID != ""
}

// IsPartOfSeries returns true if event is part of a series
func (et *EventTemplate) IsPartOfSeries() bool {
	return et.SeriesID != nil && *et.SeriesID != ""
}

// EventNotification contains notification settings
//
// Validation:
//   - NotificationChannels: Optional, list of channels (e.g., "email", "whatsapp")
//   - ReminderConfig: Optional, defines reminder schedule
//
// Supported Channels:
//   - email: Email notifications
//   - whatsapp: WhatsApp notifications
//   - sms: SMS notifications (if configured)
type EventNotification struct {
	NotificationChannels []string        `json:"notificationChannels" validate:"omitempty,max=5,dive,oneof=email whatsapp sms push" example:"email,whatsapp"`
	ReminderConfig       *ReminderConfig `json:"reminderConfig" validate:"omitempty,dive"`
}

// HasNotifications returns true if any notification channels are configured
func (en *EventNotification) HasNotifications() bool {
	return len(en.NotificationChannels) > 0
}

// HasReminders returns true if reminder configuration is present and enabled
func (en *EventNotification) HasReminders() bool {
	return en.ReminderConfig != nil && en.ReminderConfig.Enabled
}

// ReminderConfig defines the structure for event reminders
//
// Validation:
//   - Enabled: Boolean flag
//   - Intervals: Required if Enabled is true, max 10 intervals
//
// Supported Intervals:
//   - "1h", "2h", "6h", "12h": Hours before event
//   - "24h", "48h": Days before event (as hours)
//   - "1w": One week before event
//
// Example:
//
//	{
//	  "enabled": true,
//	  "intervals": ["24h", "1h"]  // Remind 1 day before and 1 hour before
//	}
type ReminderConfig struct {
	Enabled   bool     `json:"enabled" example:"true"`
	Intervals []string `json:"intervals" validate:"required_if=Enabled true,min=1,max=10,dive,oneof=1h 2h 6h 12h 24h 48h 1w" example:"24h,1h"`
}

// HasIntervals returns true if reminder intervals are configured
func (rc *ReminderConfig) HasIntervals() bool {
	return len(rc.Intervals) > 0
}

// Create Event Request
type (
	CreateEventRequest struct {
		Title              string                          `json:"title" example:"Event Title"`
		Slug               string                          `json:"slug" example:"event-title"`
		Topics             []string                        `json:"topics" example:"topic1,topic2,topic3"`
		PreDescription     string                          `json:"preDescription" example:"Event Pre Description"`
		PostDescription    JSONB                           `json:"postDescription" swaggertype:"object"`
		TermsAndConditions string                          `json:"termsAndConditions" example:"Event Terms and Conditions"`
		Category           string                          `json:"category" validate:"required,oneof=registration internal-attendance announcement external-attendance"`
		Status             string                          `json:"status" validate:"required,oneof=draft active inactive"`
		Images             EventImages                     `json:"images" validate:"omitempty,dive"`
		Organizer          EventOrganizer                  `json:"organizer" validate:"omitempty,dive"`
		Access             EventAccess                     `json:"access" validate:"omitempty,dive"`
		Location           EventLocation                   `json:"location" validate:"omitempty,dive"`
		Schedule           EventSchedule                   `json:"schedule" validate:"omitempty,dive"`
		Recurrence         EventRecurrence                 `json:"recurrence" validate:"omitempty,dive"`
		Template           EventTemplate                   `json:"template" validate:"omitempty,dive"`
		Notification       EventNotification               `json:"notification" validate:"omitempty,dive"`
		Sessions           []CreateEventSessionRequest     `json:"sessions" validate:"omitempty,dive"`
		Questions          []BulkCreateFormQuestionRequest `json:"questions" validate:"omitempty,dive"`
	}
	CreateEventResponse struct {
		Type               string                       `json:"type" example:"event"` // Use TypeEvent constant in code
		Code               string                       `json:"code" example:"event-code"`
		Title              string                       `json:"title" example:"Event Title"`
		Slug               string                       `json:"slug" example:"event-title"`
		Category           string                       `json:"category" example:"registration"`
		PreDescription     string                       `json:"preDescription" example:"Event Pre Description"`
		PostDescription    JSONB                        `json:"postDescription" swaggertype:"object"`
		TermsAndConditions string                       `json:"termsAndConditions" example:"Event Terms and Conditions"`
		Images             EventImages                  `json:"images"`
		Organizer          EventOrganizer               `json:"organizer"`
		Access             EventAccess                  `json:"access"`
		Location           EventLocation                `json:"location"`
		Schedule           EventSchedule                `json:"schedule"`
		Recurrence         EventRecurrence              `json:"recurrence"`
		Template           EventTemplate                `json:"template"`
		Notification       EventNotification            `json:"notification"`
		Status             string                       `json:"status" example:"draft"`
		Sessions           []CreateEventSessionResponse `json:"sessions,omitempty"`
		Questions          []FormQuestionResponse       `json:"questions,omitempty"`
	}

	// CreateEventResponseOption defined as a function that modifies CreateEventResponse
	CreateEventResponseOption func(*CreateEventResponse)
)

// WithEventSessions adds sessions to the response
func WithEventSessions(sessions []CreateEventSessionResponse) CreateEventResponseOption {
	return func(r *CreateEventResponse) {
		r.Sessions = sessions
	}
}

// WithFormQuestions adds questions to the response
func EventWithFormQuestions(questions []FormQuestionResponse) CreateEventResponseOption {
	return func(r *CreateEventResponse) {
		r.Questions = questions
	}
}

func (e *Event) ToCreateResponse(opts ...CreateEventResponseOption) *CreateEventResponse {
	var rp *RecurrencePattern
	if !e.RecurrencePattern.IsNull() {
		rp = &RecurrencePattern{}
		_ = e.RecurrencePattern.Unmarshal(rp)
	}

	var rc *ReminderConfig
	if !e.ReminderConfig.IsNull() {
		rc = &ReminderConfig{}
		_ = e.ReminderConfig.Unmarshal(rc)
	}

	response := &CreateEventResponse{
		Type:               TYPE_EVENT,
		Code:               e.Code,
		Title:              e.Title,
		Slug:               e.Slug,
		PreDescription:     *e.PreDescription,
		PostDescription:    e.PostDescription,
		TermsAndConditions: *e.TermsAndConditions,
		Category:           e.Category,
		Status:             e.Status,
		Images: EventImages{
			ImageLinks: e.ImageURLs,
			BannerLink: &e.BannerURL,
		},
		Organizer: EventOrganizer{
			OrganizerCommunityIDs: e.OrganizerCommunityIDs,
			ContactCommunityIDs:   e.ContactCommunityIDs,
		},
		Access: EventAccess{
			AccessLevel:         &e.AccessLevel,
			AllowedUserTypes:    e.AllowedUserTypes,
			AllowedRoles:        e.AllowedRoles,
			AllowedCampuses:     e.AllowedCampuses,
			AllowedCommunityIDs: e.AllowedCommunityIDs,
		},
		Location: EventLocation{
			LocationType:    &e.LocationType,
			PhysicalAddress: e.PhysicalAddress,
			VirtualLink:     e.VirtualLink,
			VirtualPlatform: e.VirtualPlatform,
			ClickToAction: ClickToAction{
				Text: e.CTAText,
				Link: e.CTALink,
			},
			LocationDetails:    e.LocationDetails,
			LocationVisibility: &e.LocationVisibility,
		},
		Schedule: EventSchedule{
			StartAt:  &e.StartAt,
			EndAt:    &e.EndAt,
			Timezone: &e.Timezone,
		},
		Recurrence: EventRecurrence{
			IsRecurring:       e.IsRecurring,
			RecurrencePattern: rp,
		},
		Template: EventTemplate{
			IsTemplate: e.IsTemplate,
			TemplateID: &e.TemplateID,
			SeriesID:   &e.SeriesID,
		},
		Notification: EventNotification{
			NotificationChannels: e.NotificationChannels,
			ReminderConfig:       rc,
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(response)
	}

	return response
}

// ToResponse is an alias for ToCreateResponse for handler compatibility
func (e *Event) ToResponse(opts ...CreateEventResponseOption) *CreateEventResponse {
	return e.ToCreateResponse(opts...)
}

// UpdateEventRequest represents a partial update request for an event
// All fields are pointers to distinguish between:
// - nil: field not provided, will not be updated
// - pointer to empty value: field will be set to empty/null
// - pointer to value: field will be updated to that value
type UpdateEventRequest struct {
	// Core Information
	Title              *string            `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Slug               *string            `json:"slug,omitempty" validate:"omitempty,min=1,max=255"`
	Topics             *[]string          `json:"topics,omitempty"`
	PreDescription     *string            `json:"preDescription" example:"Event Pre Description"`
	PostDescription    JSONB              `json:"postDescription" example:"Event Post Description"`
	TermsAndConditions *string            `json:"termsAndConditions" example:"Event Terms and Conditions"`
	Category           *string            `json:"category" validate:"required,oneof=registration attendance announcement volunteer hybrid"`
	Status             *string            `json:"status" validate:"required,oneof=draft active inactive"`
	Images             *EventImages       `json:"images" validate:"omitempty,dive"`
	Organizer          *EventOrganizer    `json:"organizer" validate:"omitempty,dive"`
	Access             *EventAccess       `json:"access" validate:"omitempty,dive"`
	Location           *EventLocation     `json:"location" validate:"omitempty,dive"`
	Schedule           *EventSchedule     `json:"schedule" validate:"omitempty,dive"`
	Recurrence         *EventRecurrence   `json:"recurrence" validate:"omitempty,dive"`
	Template           *EventTemplate     `json:"template" validate:"omitempty,dive"`
	Notification       *EventNotification `json:"notification" validate:"omitempty,dive"`
}

// ToUpdateMap converts the UpdateEventRequest to a map for GORM Updates
// Only includes non-nil fields, and handles null values for pointer fields
// Returns the map and the count of fields to update
func (u *UpdateEventRequest) ToUpdateMap() (map[string]interface{}, int) {
	updateMap := make(map[string]interface{})
	fieldCount := 0

	// Core Information
	if u.Title != nil {
		updateMap["title"] = *u.Title
		fieldCount++
	}
	if u.Slug != nil {
		updateMap["slug"] = *u.Slug
		fieldCount++
	}

	// Content & Media
	if u.Topics != nil {
		updateMap["topics"] = *u.Topics
		fieldCount++
	}
	if u.Category != nil {
		updateMap["category"] = *u.Category
		fieldCount++
	}
	if u.PreDescription != nil {
		if *u.PreDescription == "" {
			updateMap["pre_description"] = nil
		} else {
			updateMap["pre_description"] = *u.PreDescription
		}
		fieldCount++
	}
	if u.TermsAndConditions != nil {
		if *u.TermsAndConditions == "" {
			updateMap["terms_and_conditions"] = nil
		} else {
			updateMap["terms_and_conditions"] = *u.TermsAndConditions
		}
		fieldCount++
	}
	if u.Images != nil {
		if u.Images.ImageLinks != nil {
			updateMap["image_urls"] = u.Images.ImageLinks
			fieldCount++
		}
		if u.Images.BannerLink != nil {
			updateMap["banner_url"] = *u.Images.BannerLink
			fieldCount++
		}
	}

	// Organizers & Contacts
	if u.Organizer != nil {
		if u.Organizer.OrganizerCommunityIDs != nil {
			updateMap["organizer_community_ids"] = u.Organizer.OrganizerCommunityIDs
			fieldCount++
		}
		if u.Organizer.ContactCommunityIDs != nil {
			updateMap["contact_community_ids"] = u.Organizer.ContactCommunityIDs
			fieldCount++
		}
	}

	// Location Information
	if u.Location != nil {
		if u.Location.LocationType != nil {
			updateMap["location_type"] = *u.Location.LocationType
			fieldCount++
		}
		if u.Location.PhysicalAddress != nil {
			updateMap["physical_address"] = *u.Location.PhysicalAddress
			fieldCount++
		}
		if u.Location.VirtualLink != nil {
			updateMap["virtual_link"] = *u.Location.VirtualLink
			fieldCount++
		}
		// ClickToAction is always present as a value in EventLocation, check its pointer fields
		if u.Location.ClickToAction.Text != nil {
			updateMap["cta_text"] = *u.Location.ClickToAction.Text
			fieldCount++
		}
		if u.Location.ClickToAction.Link != nil {
			updateMap["cta_link"] = *u.Location.ClickToAction.Link
			fieldCount++
		}
		if u.Location.LocationDetails != nil {
			updateMap["location_details"] = *u.Location.LocationDetails
			fieldCount++
		}
		if u.Location.LocationVisibility != nil {
			updateMap["location_visibility"] = *u.Location.LocationVisibility
			fieldCount++
		}
	}

	// Access Control
	if u.Access != nil {
		if u.Access.AccessLevel != nil {
			updateMap["access_level"] = *u.Access.AccessLevel
			fieldCount++
		}
		if u.Access.AllowedUserTypes != nil {
			updateMap["allowed_user_types"] = u.Access.AllowedUserTypes
			fieldCount++
		}
		if u.Access.AllowedRoles != nil {
			updateMap["allowed_roles"] = u.Access.AllowedRoles
			fieldCount++
		}
		if u.Access.AllowedCampuses != nil {
			updateMap["allowed_campuses"] = u.Access.AllowedCampuses
			fieldCount++
		}
		if u.Access.AllowedCommunityIDs != nil {
			updateMap["allowed_community_ids"] = u.Access.AllowedCommunityIDs
			fieldCount++
		}
	}

	// Schedule
	if u.Schedule != nil {
		if u.Schedule.StartAt != nil {
			updateMap["start_at"] = *u.Schedule.StartAt
			fieldCount++
		}
		if u.Schedule.EndAt != nil {
			updateMap["end_at"] = *u.Schedule.EndAt
			fieldCount++
		}
		if u.Schedule.Timezone != nil {
			updateMap["timezone"] = *u.Schedule.Timezone
			fieldCount++
		}
	}

	// Recurrence
	if u.Recurrence != nil {
		updateMap["is_recurring"] = u.Recurrence.IsRecurring
		fieldCount++
		if u.Recurrence.RecurrencePattern != nil {
			updateMap["recurrence_pattern"] = u.Recurrence.RecurrencePattern
			fieldCount++
		}
		if u.Recurrence.RecurrenceEndDate != nil {
			updateMap["recurrence_end_date"] = u.Recurrence.RecurrenceEndDate
			fieldCount++
		}
	}

	// Template
	if u.Template != nil {
		updateMap["is_template"] = u.Template.IsTemplate
		fieldCount++
		if u.Template.TemplateID != nil {
			updateMap["template_id"] = *u.Template.TemplateID
			fieldCount++
		}
		if u.Template.SeriesID != nil {
			updateMap["series_id"] = *u.Template.SeriesID
			fieldCount++
		}
	}

	// Status
	if u.Status != nil {
		updateMap["status"] = *u.Status
		fieldCount++
	}

	return updateMap, fieldCount
}
