package constants

type EventStatus int32

const (
	EVENT_STATUS_DRAFT EventStatus = iota
	EVENT_STATUS_INACTIVE
	EVENT_STATUS_ACTIVE
)

const (
	EventStatusActive   = "active"
	EventStatusDraft    = "draft"
	EventStatusInactive = "inactive"
)

var (
	MapEventStatus = map[EventStatus]string{
		EVENT_STATUS_ACTIVE:   EventStatusActive,
		EVENT_STATUS_INACTIVE: EventStatusInactive,
		EVENT_STATUS_DRAFT:    EventStatusDraft,
	}
)

type EventVisibility int32

const (
	EVENT_VISIBILITY_PUBLIC EventVisibility = iota
	EVENT_VISIBILITY_PRIVATE
)

const (
	EventVisibilityPublic  = "public"
	EventVisibilityPrivate = "private"
)

var (
	MapEventVisibility = map[EventVisibility]string{
		EVENT_VISIBILITY_PUBLIC:  EventVisibilityPublic,
		EVENT_VISIBILITY_PRIVATE: EventVisibilityPrivate,
	}
)

// AttendeeRole defines the role of the attendee to whom the question applies.
type AttendeeRole string

const (
	AttendeeRoleParent AttendeeRole = "PARENT"
	AttendeeRoleChild  AttendeeRole = "CHILD"
	AttendeeRoleAll    AttendeeRole = "ALL"
)

type EventCategory string

const (
	EventCategoryGeneral      EventCategory = "GENERAL"
	EventCategoryAnnouncement EventCategory = "ANNOUNCEMENT"
)
