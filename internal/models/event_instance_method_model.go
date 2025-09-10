package models

type RegisterMethod int32

const (
	REGISTER_METHOD_PERSONAL RegisterMethod = iota
	REGISTER_METHOD_EVENT
	REGISTER_METHOD_REGISTRATION
)

const (
	RegisterMethodPersonal = "personal-qr"
	RegisterMethodEvent    = "event-qr"
	RegisterMethodRegister = "register-qr"
	RegisterMethodBoth     = "both-qr"
	RegisterMethodNone     = "none"
)

var (
	MapRegisterMethod = map[RegisterMethod]string{
		REGISTER_METHOD_PERSONAL:     RegisterMethodPersonal,
		REGISTER_METHOD_EVENT:        RegisterMethodEvent,
		REGISTER_METHOD_REGISTRATION: RegisterMethodRegister,
	}
)
