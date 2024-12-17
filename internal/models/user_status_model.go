package models

type UserStatus int32

const (
	USER_STATUS_ACTIVE UserStatus = iota
	USER_STATUS_INACTIVE
)

const (
	UserStatusActive   = "active"
	UserStatusInActive = "inactive"
)

var (
	MapUserStatus = map[UserStatus]string{
		USER_STATUS_ACTIVE:   UserStatusActive,
		USER_STATUS_INACTIVE: UserStatusInActive,
	}
)
