package models

import (
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
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// Core Event Information
	Code  string `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	Title string `gorm:"type:varchar(255);not null;index" json:"title"`
	Slug  string `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"` // Slugified title for URLs

	// Content & Media
	Topics                  pq.StringArray `gorm:"type:text[]" json:"topics"`
	Category                string         `gorm:"type:varchar(50);not null" json:"category"`
	Description             *string        `gorm:"type:text" json:"description"`
	TermsAndConditions      *string        `gorm:"type:text" json:"terms_and_conditions"`
	ImageLinks              pq.StringArray `gorm:"type:text[]" json:"imageLinks"`
	PostRegistrationDetails JSONB          `gorm:"type:jsonb" json:"post_registration_details"` // Details shown after registration

	// Creator & Organizers
	CreatorCommunityID    string         `gorm:"type:varchar(50);not null;index" json:"creator_community_id"`
	OrganizerCommunityIDs pq.StringArray `gorm:"type:text[]" json:"organizer_community_ids"`
	ContactCommunityIDs   pq.StringArray `gorm:"type:text[]" json:"contact_community_ids"`

	// Location Information
	LocationType       string  `gorm:"type:varchar(20);not null" json:"location_type"`       // online, offline, hybrid
	PhysicalAddress    *string `gorm:"type:text" json:"physical_address"`                    // Physical address
	VirtualLink        *string `gorm:"type:text" json:"virtual_link"`                        // Meeting link
	MeetingCTAText     *string `gorm:"type:varchar(100)" json:"meeting_cta_text"`            // CTA button text
	LocationDetails    *string `gorm:"type:text" json:"location_details"`                    // Additional location info
	LocationVisibility string  `gorm:"type:varchar(30);not null" json:"location_visibility"` // pre-registration, post-registration, all

	// Access Control
	AccessLevel         string         `gorm:"type:varchar(20);not null;index;index:idx_events_access_status" json:"access_level"` // public, private
	AllowedUserTypes    pq.StringArray `gorm:"type:text[]" json:"allowed_user_types"`
	AllowedRoles        pq.StringArray `gorm:"type:text[]" json:"allowed_roles"`
	AllowedCampuses     pq.StringArray `gorm:"type:text[]" json:"allowed_campuses"`
	AllowedCommunityIDs pq.StringArray `gorm:"type:text[]" json:"allowed_community_ids"`

	// Scheduling
	Recurrence *string   `gorm:"type:varchar(255)" json:"recurrence"` // daily, weekly, monthly, yearly, or Quartz cron
	StartAt    time.Time `gorm:"type:timestamptz;not null;index;index:idx_events_status_start,priority:2" json:"start_at"`
	EndAt      time.Time `gorm:"type:timestamptz;not null;index" json:"end_at"`
	Timezone   string    `gorm:"type:varchar(50);not null" json:"timezone"` // IANA timezone (e.g., "Asia/Jakarta")

	// Registration Configurations
	ConfirmationMethod *string `gorm:"type:varchar(20)" json:"confirmation_method"` // whatsapp, email, both
	ValidationMethod   *string `gorm:"type:varchar(20)" json:"validation_method"`   // location

	// Status & Metadata
	Status    string         `gorm:"type:varchar(20);not null;index;index:idx_events_status_start,priority:1;default:'draft'" json:"status"`
	CreatedAt time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships (loaded via joins, not stored)
	Instances []EventInstance `gorm:"foreignKey:EventCode;references:Code" json:"instances,omitempty"`
}

// TableName specifies the table name for GORM
func (Event) TableName() string {
	return "events"
}

// Helper Methods - Simple getters, NO validation, NO business logic

// IsPublic returns true if the event is publicly accessible
func (e *Event) IsPublic() bool {
	return e.AccessLevel == string(constants.AccessLevelPublic)
}

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

// IsDraft returns true if the event is in draft status
func (e *Event) IsDraft() bool {
	return e.Status == string(constants.EventStatusDraft)
}

// IsRecurring returns true if the event has recurrence
func (e *Event) IsRecurring() bool {
	return e.Recurrence != nil && *e.Recurrence != string(constants.RecurrenceNone)
}

// Common Requests

type (
	EventLocation struct {
		LocationType       string `json:"locationType" validate:"required,oneof=online offline hybrid"`
		PhysicalAddress    string `json:"physicalAddress" validate:"required_if=LocationType offline|required_if=LocationType hybrid"`
		VirtualLink        string `json:"virtualLink" validate:"required_if=LocationType online|required_if=LocationType hybrid"`
		MeetingCTAText     string `json:"meetingCtaText"`
		LocationDetails    string `json:"locationDetails"`
		LocationVisibility string `json:"locationVisibility" validate:"required,oneof=pre-registration post-registration all"`
	}

	EventAccess struct {
		AccessLevel         string   `json:"accessLevel" validate:"required,oneof=public private"`
		AllowedUserTypes    []string `json:"allowedUserTypes" validate:"omitempty"`
		AllowedRoles        []string `json:"allowedRoles" validate:"omitempty"`
		AllowedCampuses     []string `json:"allowedCampuses" validate:"omitempty"`
		AllowedCommunityIDs []string `json:"allowedCommunityIds" validate:"omitempty"`
	}

	EventSchedule struct {
		Recurrence string    `json:"recurrence" validate:"omitempty,oneof=daily weekly monthly yearly"`
		StartAt    time.Time `json:"startAt" validate:"required"`
		EndAt      time.Time `json:"endAt" validate:"required,gtfield=StartAt"`
		Timezone   string    `json:"timezone" validate:"required"`
	}

	RegistrationConfig struct {
		PostRegistrationDetails JSONB  `json:"postRegistrationDetails" example:"Event Post Registration Details"`
		ConfirmationMethod      string `json:"confirmationMethod" validate:"required,oneof=whatsapp email both"`
		ValidationMethod        string `json:"validationMethod" validate:"omitempty,oneof=location"`
	}
)

// Create Event Request
type (
	CreateEventRequest struct {
		Title                 string                 `json:"title" validate:"required" example:"Event Title"`
		Slug                  string                 `json:"slug" example:"event-title"`
		Topics                []string               `json:"topics" example:"topic1,topic2,topic3"`
		Category              string                 `json:"category" validate:"required,oneof=announcement registerable"`
		Description           string                 `json:"description" example:"Event Description"`
		TermsAndConditions    string                 `json:"termsAndConditions" example:"Event Terms and Conditions"`
		ImageLinks            []string               `json:"imageLinks" validate:"omitempty,dive,url" example:"image1,image2,image3"`
		OrganizerCommunityIDs []string               `json:"organizerCommunityIds" validate:"required" example:"community1,community2,community3"`
		ContactCommunityIDs   []string               `json:"contactCommunityIds" validate:"required" example:"community1,community2,community3"`
		RegistrationConfig    RegistrationConfig     `json:"registrationConfig" validate:"required"`
		Location              EventLocation          `json:"location" validate:"required"`
		Access                EventAccess            `json:"access" validate:"required"`
		Schedule              EventSchedule          `json:"schedule" validate:"required"`
		Status                string                 `json:"status" example:"draft"`
		Instances             *CreateInstanceRequest `json:"instances" validate:"omitempty"`
	}
	CreateEventResponse struct {
		Type                  string             `json:"type" example:"event"` // Use TypeEvent constant in code
		Code                  string             `json:"code" example:"event-code"`
		Title                 string             `json:"title" example:"Event Title"`
		Slug                  string             `json:"slug" example:"event-title"`
		Description           string             `json:"description" example:"Event Description"`
		TermsAndConditions    string             `json:"termsAndConditions" example:"Event Terms and Conditions"`
		ImageLinks            []string           `json:"imageLinks" example:"image1,image2,image3"`
		OrganizerCommunityIDs []string           `json:"organizerCommunityIds" example:"community1,community2,community3"`
		ContactCommunityIDs   []string           `json:"contactCommunityIds" example:"community1,community2,community3"`
		RegistrationConfig    RegistrationConfig `json:"registrationConfig"`
		Location              EventLocation      `json:"location"`
		Access                EventAccess        `json:"access"`
		Schedule              EventSchedule      `json:"schedule"`
		Status                string             `json:"status" example:"draft"`
	}
)

func (e *Event) ToCreateResponse() *CreateEventResponse {
	return &CreateEventResponse{
		Type:                  TYPE_EVENT,
		Code:                  e.Code,
		Title:                 e.Title,
		Slug:                  e.Slug,
		Description:           *e.Description,
		TermsAndConditions:    *e.TermsAndConditions,
		ImageLinks:            e.ImageLinks,
		OrganizerCommunityIDs: e.OrganizerCommunityIDs,
		ContactCommunityIDs:   e.ContactCommunityIDs,
		RegistrationConfig: RegistrationConfig{
			PostRegistrationDetails: e.PostRegistrationDetails,
			ConfirmationMethod:      *e.ConfirmationMethod,
			ValidationMethod:        *e.ValidationMethod,
		},
		Location: EventLocation{
			LocationType:       e.LocationType,
			PhysicalAddress:    *e.PhysicalAddress,
			VirtualLink:        *e.VirtualLink,
			MeetingCTAText:     *e.MeetingCTAText,
			LocationDetails:    *e.LocationDetails,
			LocationVisibility: e.LocationVisibility,
		},
		Access: EventAccess{
			AccessLevel:         e.AccessLevel,
			AllowedUserTypes:    e.AllowedUserTypes,
			AllowedRoles:        e.AllowedRoles,
			AllowedCampuses:     e.AllowedCampuses,
			AllowedCommunityIDs: e.AllowedCommunityIDs,
		},
		Schedule: EventSchedule{
			Recurrence: *e.Recurrence,
			StartAt:    e.StartAt,
			EndAt:      e.EndAt,
			Timezone:   e.Timezone,
		},
		Status: e.Status,
	}
}

// ToResponse is an alias for ToCreateResponse for handler compatibility
func (e *Event) ToResponse() *CreateEventResponse {
	return e.ToCreateResponse()
}

// NewEventFromRequest constructs an Event model from CreateEventRequest
func NewEventFromRequest(req CreateEventRequest, code, slug, creatorID string) *Event {
	return &Event{
		Code:                    code,
		Title:                   req.Title,
		Slug:                    slug,
		Topics:                  req.Topics,
		Category:                req.Category,
		Description:             &req.Description,
		TermsAndConditions:      &req.TermsAndConditions,
		ImageLinks:              req.ImageLinks,
		PostRegistrationDetails: req.RegistrationConfig.PostRegistrationDetails,
		CreatorCommunityID:      creatorID,
		OrganizerCommunityIDs:   req.OrganizerCommunityIDs,
		ContactCommunityIDs:     req.ContactCommunityIDs,
		LocationType:            req.Location.LocationType,
		PhysicalAddress:         &req.Location.PhysicalAddress,
		VirtualLink:             &req.Location.VirtualLink,
		MeetingCTAText:          &req.Location.MeetingCTAText,
		LocationDetails:         &req.Location.LocationDetails,
		LocationVisibility:      req.Location.LocationVisibility,
		AccessLevel:             req.Access.AccessLevel,
		AllowedUserTypes:        req.Access.AllowedUserTypes,
		AllowedRoles:            req.Access.AllowedRoles,
		AllowedCampuses:         req.Access.AllowedCampuses,
		AllowedCommunityIDs:     req.Access.AllowedCommunityIDs,
		Recurrence:              &req.Schedule.Recurrence,
		StartAt:                 req.Schedule.StartAt,
		EndAt:                   req.Schedule.EndAt,
		Timezone:                req.Schedule.Timezone,
		ConfirmationMethod:      &req.RegistrationConfig.ConfirmationMethod,
		Status:                  req.Status,
	}
}

// UpdateEventRequest represents a partial update request for an event
// All fields are pointers to distinguish between:
// - nil: field not provided, will not be updated
// - pointer to empty value: field will be set to empty/null
// - pointer to value: field will be updated to that value
type UpdateEventRequest struct {
	// Core Information
	Title *string `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Slug  *string `json:"slug,omitempty" validate:"omitempty,min=1,max=255"`

	// Content & Media
	Topics             *[]string `json:"topics,omitempty"`
	Category           *string   `json:"category,omitempty" validate:"omitempty,oneof=announcement registerable"`
	Description        *string   `json:"description,omitempty"`
	TermsAndConditions *string   `json:"termsAndConditions,omitempty"`
	ImageLinks         *[]string `json:"imageLinks,omitempty" validate:"omitempty,dive,url"`

	// Organizers & Contacts
	OrganizerCommunityIDs *[]string `json:"organizerCommunityIds,omitempty"`
	ContactCommunityIDs   *[]string `json:"contactCommunityIds,omitempty"`

	// Location Information
	LocationType       *string `json:"locationType,omitempty" validate:"omitempty,oneof=online offline hybrid"`
	PhysicalAddress    *string `json:"physicalAddress,omitempty"`
	VirtualLink        *string `json:"virtualLink,omitempty"`
	MeetingCTAText     *string `json:"meetingCtaText,omitempty"`
	LocationDetails    *string `json:"locationDetails,omitempty"`
	LocationVisibility *string `json:"locationVisibility,omitempty" validate:"omitempty,oneof=pre-registration post-registration all"`

	// Access Control
	AccessLevel         *string   `json:"accessLevel,omitempty" validate:"omitempty,oneof=public private"`
	AllowedUserTypes    *[]string `json:"allowedUserTypes,omitempty"`
	AllowedRoles        *[]string `json:"allowedRoles,omitempty"`
	AllowedCampuses     *[]string `json:"allowedCampuses,omitempty"`
	AllowedCommunityIDs *[]string `json:"allowedCommunityIds,omitempty"`

	// Scheduling
	Recurrence *string    `json:"recurrence,omitempty" validate:"omitempty,oneof=daily weekly monthly yearly"`
	StartAt    *time.Time `json:"startAt,omitempty"`
	EndAt      *time.Time `json:"endAt,omitempty"`
	Timezone   *string    `json:"timezone,omitempty"`

	// Registration Configuration
	PostRegistrationDetails *JSONB  `json:"postRegistrationDetails,omitempty"`
	ConfirmationMethod      *string `json:"confirmationMethod,omitempty" validate:"omitempty,oneof=whatsapp email both"`
	ValidationMethod        *string `json:"validationMethod,omitempty" validate:"omitempty,oneof=location"`

	// Status
	Status *string `json:"status,omitempty" validate:"omitempty,oneof=draft published cancelled completed"`
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
	if u.Description != nil {
		if *u.Description == "" {
			updateMap["description"] = nil
		} else {
			updateMap["description"] = *u.Description
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
	if u.ImageLinks != nil {
		updateMap["image_links"] = *u.ImageLinks
		fieldCount++
	}

	// Organizers & Contacts
	if u.OrganizerCommunityIDs != nil {
		updateMap["organizer_community_ids"] = *u.OrganizerCommunityIDs
		fieldCount++
	}
	if u.ContactCommunityIDs != nil {
		updateMap["contact_community_ids"] = *u.ContactCommunityIDs
		fieldCount++
	}

	// Location Information
	if u.LocationType != nil {
		updateMap["location_type"] = *u.LocationType
		fieldCount++
	}
	if u.PhysicalAddress != nil {
		if *u.PhysicalAddress == "" {
			updateMap["physical_address"] = nil
		} else {
			updateMap["physical_address"] = *u.PhysicalAddress
		}
		fieldCount++
	}
	if u.VirtualLink != nil {
		if *u.VirtualLink == "" {
			updateMap["virtual_link"] = nil
		} else {
			updateMap["virtual_link"] = *u.VirtualLink
		}
		fieldCount++
	}
	if u.MeetingCTAText != nil {
		if *u.MeetingCTAText == "" {
			updateMap["meeting_cta_text"] = nil
		} else {
			updateMap["meeting_cta_text"] = *u.MeetingCTAText
		}
		fieldCount++
	}
	if u.LocationDetails != nil {
		if *u.LocationDetails == "" {
			updateMap["location_details"] = nil
		} else {
			updateMap["location_details"] = *u.LocationDetails
		}
		fieldCount++
	}
	if u.LocationVisibility != nil {
		updateMap["location_visibility"] = *u.LocationVisibility
		fieldCount++
	}

	// Access Control
	if u.AccessLevel != nil {
		updateMap["access_level"] = *u.AccessLevel
		fieldCount++
	}
	if u.AllowedUserTypes != nil {
		updateMap["allowed_user_types"] = *u.AllowedUserTypes
		fieldCount++
	}
	if u.AllowedRoles != nil {
		updateMap["allowed_roles"] = *u.AllowedRoles
		fieldCount++
	}
	if u.AllowedCampuses != nil {
		updateMap["allowed_campuses"] = *u.AllowedCampuses
		fieldCount++
	}
	if u.AllowedCommunityIDs != nil {
		updateMap["allowed_community_ids"] = *u.AllowedCommunityIDs
		fieldCount++
	}

	// Scheduling
	if u.Recurrence != nil {
		if *u.Recurrence == "" {
			updateMap["recurrence"] = nil
		} else {
			updateMap["recurrence"] = *u.Recurrence
		}
		fieldCount++
	}
	if u.StartAt != nil {
		updateMap["start_at"] = *u.StartAt
		fieldCount++
	}
	if u.EndAt != nil {
		updateMap["end_at"] = *u.EndAt
		fieldCount++
	}
	if u.Timezone != nil {
		updateMap["timezone"] = *u.Timezone
		fieldCount++
	}

	// Registration Configuration
	if u.PostRegistrationDetails != nil {
		updateMap["post_registration_details"] = *u.PostRegistrationDetails
		fieldCount++
	}
	if u.ConfirmationMethod != nil {
		if *u.ConfirmationMethod == "" {
			updateMap["confirmation_method"] = nil
		} else {
			updateMap["confirmation_method"] = *u.ConfirmationMethod
		}
		fieldCount++
	}
	if u.ValidationMethod != nil {
		if *u.ValidationMethod == "" {
			updateMap["validation_method"] = nil
		} else {
			updateMap["validation_method"] = *u.ValidationMethod
		}
		fieldCount++
	}

	// Status
	if u.Status != nil {
		updateMap["status"] = *u.Status
		fieldCount++
	}

	return updateMap, fieldCount
}
