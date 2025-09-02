package constants

type GeneralStatus int32

const (
	STATUS_ACTIVE GeneralStatus = iota
	STATUS_INACTIVE
)

const (
	StatusActive   = "active"
	StatusInActive = "inactive"
)

var (
	MapStatus = map[GeneralStatus]string{
		STATUS_ACTIVE:   StatusActive,
		STATUS_INACTIVE: StatusInActive,
	}
)

type CoolJoinerStatus int32

const (
	COOL_JOINER_STATUS_PENDING CoolJoinerStatus = iota
	COOL_JOINER_STATUS_FOLLOWED
	COOL_JOINER_STATUS_COMPLETED
)

const (
	CoolJoinerStatusPending   = "pending"
	CoolJoinerStatusFollowed  = "followed"
	CoolJoinerStatusCompleted = "completed"
)

var (
	MapCoolJoinerStatus = map[CoolJoinerStatus]string{
		COOL_JOINER_STATUS_PENDING:   CoolJoinerStatusPending,
		COOL_JOINER_STATUS_FOLLOWED:  CoolJoinerStatusFollowed,
		COOL_JOINER_STATUS_COMPLETED: CoolJoinerStatusCompleted,
	}
)

type CoolRsvpStatus int32

const (
	COOL_RSVP_STATUS_ATTEND CoolRsvpStatus = iota
	COOL_RSVP_STATUS_CANNOT_ATTEND
	COOL_RSVP_STATUS_UNKNOWN
)

const (
	CoolRsvpStatusAttend  = "attend"
	CoolRsvpStatusCannot  = "cannotAttend"
	CoolRsvpStatusUnknown = "unknown"
)

var (
	MapCoolRsvpStatus = map[CoolRsvpStatus]string{
		COOL_RSVP_STATUS_ATTEND:        CoolRsvpStatusAttend,
		COOL_RSVP_STATUS_CANNOT_ATTEND: CoolRsvpStatusCannot,
		COOL_RSVP_STATUS_UNKNOWN:       CoolRsvpStatusUnknown,
	}
)
