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
	COOL_JOINER_STATUS_CANCELLED
	COOL_JOINER_STATUS_COMPLETED
)

const (
	CoolJoinerStatusPending   = "pending"
	CoolJoinerStatusFollowed  = "followed"
	CoolJoinerStatusCancelled = "cancelled"
	CoolJoinerStatusCompleted = "completed"
)

var (
	MapCoolJoinerStatus = map[CoolJoinerStatus]string{
		COOL_JOINER_STATUS_PENDING:   CoolJoinerStatusPending,
		COOL_JOINER_STATUS_FOLLOWED:  CoolJoinerStatusFollowed,
		COOL_JOINER_STATUS_CANCELLED: CoolJoinerStatusCancelled,
		COOL_JOINER_STATUS_COMPLETED: CoolJoinerStatusCompleted,
	}
)
