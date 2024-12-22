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

type RegistrationStatus int32

const (
	REGISTER_STATUS_SUCCESS RegistrationStatus = iota
	REGISTER_STATUS_PENDING
	REGISTER_STATUS_FAILED
	REGISTER_STATUS_CANCELLED
	REGISTER_STATUS_PERMIT
)

const (
	RegisterStatusSuccess   = "success"
	RegisterStatusPending   = "pending"
	RegisterStatusFailed    = "failed"
	RegisterStatusCancelled = "cancelled"
	RegisterStatusPermitted = "permit"
)

var (
	MapRegisterStatus = map[RegistrationStatus]string{
		REGISTER_STATUS_SUCCESS:   RegisterStatusSuccess,
		REGISTER_STATUS_PENDING:   RegisterStatusPending,
		REGISTER_STATUS_FAILED:    RegisterStatusFailed,
		REGISTER_STATUS_CANCELLED: RegisterStatusCancelled,
		REGISTER_STATUS_PERMIT:    RegisterStatusPermitted,
	}
)
