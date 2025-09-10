package constants

// AttendeeRole defines the role of the attendee to whom the question applies.
type AttendeeRole string

const (
	AttendeeRoleParent AttendeeRole = "PARENT"
	AttendeeRoleChild  AttendeeRole = "CHILD"
	AttendeeRoleAll    AttendeeRole = "ALL"
)
