package constants

// EventLocationType represents the type of event location
type EventLocationType string

const (
	LocationTypeOnline  EventLocationType = "online"
	LocationTypeOffline EventLocationType = "offline"
	LocationTypeHybrid  EventLocationType = "hybrid"
)

// IsValid checks if the location type is valid
func (e EventLocationType) IsValid() bool {
	switch e {
	case LocationTypeOnline, LocationTypeOffline, LocationTypeHybrid:
		return true
	}
	return false
}

// String returns the string representation
func (e EventLocationType) String() string {
	return string(e)
}

// AccessLevel represents event access level
type AccessLevel string

const (
	AccessLevelPublic  AccessLevel = "public"
	AccessLevelPrivate AccessLevel = "private"
)

// IsValid checks if the access level is valid
func (a AccessLevel) IsValid() bool {
	switch a {
	case AccessLevelPublic, AccessLevelPrivate:
		return true
	}
	return false
}

// String returns the string representation
func (a AccessLevel) String() string {
	return string(a)
}

// EventCategory represents event category
type EventCategory string

const (
	CategoryAnnouncement EventCategory = "announcement"
	CategoryRegisterable EventCategory = "registerable"
)

// IsValid checks if the category is valid
func (c EventCategory) IsValid() bool {
	switch c {
	case CategoryAnnouncement, CategoryRegisterable:
		return true
	}
	return false
}

// String returns the string representation
func (c EventCategory) String() string {
	return string(c)
}

// LocationVisibility represents when location details are shown
type LocationVisibility string

const (
	LocationVisibilityAll              LocationVisibility = "all"
	LocationVisibilityPreRegistration  LocationVisibility = "pre-registration"
	LocationVisibilityPostRegistration LocationVisibility = "post-registration"
)

// IsValid checks if the location visibility is valid
func (l LocationVisibility) IsValid() bool {
	switch l {
	case LocationVisibilityAll, LocationVisibilityPreRegistration, LocationVisibilityPostRegistration:
		return true
	}
	return false
}

// String returns the string representation
func (l LocationVisibility) String() string {
	return string(l)
}

// ConfirmationMethod represents how confirmation is sent
type ConfirmationMethod string

const (
	ConfirmationMethodWhatsApp ConfirmationMethod = "whatsapp"
	ConfirmationMethodEmail    ConfirmationMethod = "email"
	ConfirmationMethodBoth     ConfirmationMethod = "both"
)

// IsValid checks if the confirmation method is valid
func (c ConfirmationMethod) IsValid() bool {
	switch c {
	case ConfirmationMethodWhatsApp, ConfirmationMethodEmail, ConfirmationMethodBoth:
		return true
	}
	return false
}

// String returns the string representation
func (c ConfirmationMethod) String() string {
	return string(c)
}

// EventStatus represents the current status of an event
type EventStatus string

const (
	EventStatusDraft     EventStatus = "draft"
	EventStatusPublished EventStatus = "published"
	EventStatusCancelled EventStatus = "cancelled"
	EventStatusCompleted EventStatus = "completed"
)

// IsValid checks if the event status is valid
func (e EventStatus) IsValid() bool {
	switch e {
	case EventStatusDraft, EventStatusPublished, EventStatusCancelled, EventStatusCompleted:
		return true
	}
	return false
}

// String returns the string representation
func (e EventStatus) String() string {
	return string(e)
}

// RecurrenceType represents simple recurrence patterns
type RecurrenceType string

const (
	RecurrenceNone    RecurrenceType = ""
	RecurrenceDaily   RecurrenceType = "daily"
	RecurrenceWeekly  RecurrenceType = "weekly"
	RecurrenceMonthly RecurrenceType = "monthly"
	RecurrenceYearly  RecurrenceType = "yearly"
)

// IsValid checks if the recurrence type is valid
func (r RecurrenceType) IsValid() bool {
	switch r {
	case RecurrenceNone, RecurrenceDaily, RecurrenceWeekly, RecurrenceMonthly, RecurrenceYearly:
		return true
	}
	return false
}

// String returns the string representation
func (r RecurrenceType) String() string {
	return string(r)
}

// Event code configuration
const (
	EventCodePrefix     = "EVT"
	EventCodeMaxRetries = 5
)

// Event type constant
const TypeEvent = "event"
