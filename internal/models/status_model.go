package models

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
