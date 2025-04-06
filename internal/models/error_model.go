package models

import (
	"database/sql"
	"errors"
	"net/http"
)

var (
	// Errors
	ErrorUserNotFound    = errors.New("user is not registered yet")
	ErrorInvalidPassword = errors.New("invalid password")
	ErrorDataNotFound    = errors.New("a specified resource is not found")
	ErrorAlreadyExist    = errors.New("the resource that a client tried to create already exists")
	ErrorUnauthorized    = errors.New("request not authenticated due to missing, invalid, or expired token")
	ErrorNoRows          = sql.ErrNoRows

	// Specific for COOL Category
	ErrorAgeRange = errors.New("ageStart should be less than ageEnd")

	// Event
	ErrorEmailPhoneNumberEmpty       = errors.New("you should enter either phone number or email")
	ErrorCannotRegisterYet           = errors.New("you cannot register yet, wait until the time allowed first")
	ErrorRegistrationTimeDisabled    = errors.New("you cannot register anymore, since the time is already closed")
	ErrorEventNotValid               = errors.New("event code is not valid")
	ErrorExceedMaxSeating            = errors.New("you cannot register more than 4 people")
	ErrorRegisterQuotaNotAvailable   = errors.New("you cannot register anymore, since there are no available seats anymore")
	ErrorRegistrationCodeInvalid     = errors.New("you cannot register using an invalid registration code")
	ErrorRegistrationWrongTime       = errors.New("your registration is not valid for this session, please check again")
	ErrorRegistrationAlreadyCancel   = errors.New("you already cancelled the registration")
	ErrorRegistrationAlreadyVerified = errors.New("your registration code is already verified")
	ErrorNoRegistrationNeeded        = errors.New("you do not need to register for this session")
	ErrorViolateAllowedForPrivate    = errors.New("allowedFor is private but either allowedUsers, allowedRoles, allowedCampuses are empty")
	ErrorMaxPerTrxIsZero             = errors.New("since registration is required, cannot set maxPerTrx to 0")
	ErrorAttendanceTypeWhenRequired  = errors.New("since registration is required, attendance type cannot be empty")
	ErrorEventNotAvailable           = errors.New("event is not available")
	ErrorEventCanOnlyRegisterOnce    = errors.New("your account already registered for this event")
	ErrorAlreadyRegistered           = errors.New("your main or other register data already registered for this event")
	ErrorIdentifierCommunityIdEmpty  = errors.New("at least should filled either identifier or community id")
	ErrorQRForMoreThanOneRegister    = errors.New("your personal QR cannot be used for more than one registration")
	ErrorCannotUsePersonalQR         = errors.New("you cannot register this event by your personal qr. To register, please register manually")
	ErrorAlreadyVerified             = errors.New("your registration is already verified")
	ErrorAlreadyCancelled            = errors.New("your registration is already cancelled")
	ErrorForbiddenStatus             = errors.New("you are not allowed to use this status on this event")
	ErrorReasonEmpty                 = errors.New("reason cannot be empty when you entered for permission")

	// Google Error
	ErrorFetchGoogle = errors.New("error while retrieving user from google")

	// User Auth Error
	ErrorTokenSignature = errors.New("invalid signature")
	ErrorInvalidToken   = errors.New("token is invalid")
	ErrorExpiredToken   = errors.New("token is expired. please login again to use the account")
	ErrorEmptyToken     = errors.New("token is empty")
	ErrorForbiddenRole  = errors.New("you are not allowed to access this feature")
	ErrorLoggedOut      = errors.New("you are already logged out")

	// API Auth Error
	ErrorInvalidAPIKey = errors.New("api key is invalid")
	ErrorEmptyAPIKey   = errors.New("no api key is found")

	// Special for Validation Error
	ErrorInvalidInput = errors.New("invalid request input")
	ErrorEmailInput   = errors.New("email format is invalid, should be: xxxx@xxxx.com")

	// Idempotency Error
	ErrorEmptyRequestID     = errors.New("request id is empty")
	ErrorProcessedRequestID = errors.New("request id has already been processed")

	// Rate Limiter Error
	ErrorRateLimiterExceeds = errors.New("too much input, please try again later")

	// User Error
	ErrorDidNotFillKKJNumber   = errors.New("please input the kkj number if you input jemaat id")
	ErrorMismatchFields        = errors.New("please input the same input on both fields")
	ErrorMissingDepartmentCool = errors.New("please input department code or cool id")

	// Update User Error
	ErrorDifferentCommunityId   = errors.New("you cannot take action with this different account from your account")
	ErrorConflictRelationDelete = errors.New("its not allowed to place the same user relation in update and delete")

	// Time error
	ErrorStartDateLater = errors.New("start time cannot be later than end time")

	// Pagination Error
	ErrorLimitMustBeGreaterThanZero = errors.New("the limit must be greater than zero")

	// Download Error
	ErrorCSVOrXLSX = errors.New("should be either csv or xlsx")
)

func ErrorMapping(err error) Response {
	switch err {
	case ErrorUserNotFound:
		return Response{
			Code:    http.StatusNotFound,
			Status:  "DATA_NOT_FOUND",
			Message: err.Error(),
		}
	case ErrorDataNotFound:
		return Response{
			Code:    http.StatusNotFound,
			Status:  "DATA_NOT_FOUND",
			Message: err.Error(),
		}
	case ErrorInvalidInput:
		return Response{
			Code:    http.StatusBadRequest,
			Status:  "INVALID_INPUT",
			Message: err.Error(),
		}
	case ErrorNoRows:
		return Response{
			Code:    http.StatusInternalServerError,
			Status:  "DATABASE_ERROR",
			Message: err.Error(),
		}
	case ErrorAlreadyExist:
		return Response{
			Code:    http.StatusConflict,
			Status:  "ALREADY_EXISTS",
			Message: err.Error(),
		}
	case ErrorInvalidInput:
		return Response{
			Code:    http.StatusBadRequest,
			Status:  "INVALID_ARGUMENT",
			Message: err.Error(),
		}
	case ErrorEmailInput:
		return Response{
			Code:    http.StatusBadRequest,
			Status:  "INVALID_ARGUMENT",
			Message: err.Error(),
		}
	case ErrorAgeRange:
		return Response{
			Code:    http.StatusBadRequest,
			Status:  "INVALID_ARGUMENT",
			Message: err.Error(),
		}
	case ErrorFetchGoogle:
		return Response{
			Code:    http.StatusInternalServerError,
			Status:  "INTERNAL_SERVER_ERROR",
			Message: err.Error(),
		}
	case ErrorTokenSignature:
		return Response{
			Code:    http.StatusUnauthorized,
			Status:  "INVALID_TOKEN_SIGNATURE",
			Message: err.Error(),
		}
	case ErrorInvalidToken:
		return Response{
			Code:    http.StatusUnauthorized,
			Status:  "INVALID_TOKEN",
			Message: err.Error(),
		}
	case ErrorExpiredToken:
		return Response{
			Code:    http.StatusUnauthorized,
			Status:  "EXPIRED_TOKEN",
			Message: err.Error(),
		}
	case ErrorEmptyToken:
		return Response{
			Code:    http.StatusUnauthorized,
			Status:  "MISSING_TOKEN",
			Message: err.Error(),
		}
	case ErrorEmailPhoneNumberEmpty:
		return Response{
			Code:    http.StatusBadRequest,
			Status:  "MISSING_FIELD",
			Message: err.Error(),
		}
	case ErrorInvalidPassword:
		return Response{
			Code:    http.StatusUnauthorized,
			Status:  "INVALID_CREDENTIALS",
			Message: err.Error(),
		}
	case ErrorCannotRegisterYet:
		return Response{
			Code:    http.StatusForbidden,
			Status:  "FORBIDDEN_REGISTRATION",
			Message: err.Error(),
		}
	case ErrorRegistrationTimeDisabled:
		return Response{
			Code:    http.StatusForbidden,
			Status:  "FORBIDDEN_REGISTRATION",
			Message: err.Error(),
		}
	case ErrorInvalidAPIKey:
		return Response{
			Code:    http.StatusUnauthorized,
			Status:  "INVALID_KEY",
			Message: err.Error(),
		}
	case ErrorEmptyAPIKey:
		return Response{
			Code:    http.StatusUnauthorized,
			Status:  "MISSING_KEY",
			Message: err.Error(),
		}
	case ErrorEventNotValid:
		return Response{
			Code:    http.StatusBadRequest,
			Status:  "INVALID_INPUT",
			Message: err.Error(),
		}
	case ErrorExceedMaxSeating:
		return Response{
			Code:    http.StatusBadRequest,
			Status:  "FORBIDDEN_REGISTRATION",
			Message: err.Error(),
		}
	case ErrorRegisterQuotaNotAvailable:
		return Response{
			Code:    http.StatusForbidden,
			Status:  "FORBIDDEN_REGISTRATION",
			Message: err.Error(),
		}
	case ErrorRegistrationCodeInvalid:
		return Response{
			Code:    http.StatusForbidden,
			Status:  "FORBIDDEN_REGISTRATION",
			Message: err.Error(),
		}
	case ErrorRegistrationAlreadyCancel:
		return Response{
			Code:    http.StatusBadRequest,
			Status:  "ALREADY_CHANGED",
			Message: err.Error(),
		}
	case ErrorRegistrationAlreadyVerified:
		return Response{
			Code:    http.StatusBadRequest,
			Status:  "ALREADY_CHANGED",
			Message: err.Error(),
		}
	case ErrorForbiddenRole:
		return Response{
			Code:    http.StatusForbidden,
			Status:  "FORBIDDEN_ROLE",
			Message: err.Error(),
		}
	case ErrorRegistrationWrongTime:
		return Response{
			Code:    http.StatusBadRequest,
			Status:  "INVALID_ARGUMENT",
			Message: err.Error(),
		}
	case ErrorNoRegistrationNeeded:
		return Response{
			Code:    http.StatusForbidden,
			Status:  "FORBIDDEN_REGISTRATION",
			Message: err.Error(),
		}
	case ErrorLoggedOut:
		return Response{
			Code:    http.StatusUnauthorized,
			Status:  "LOGGED_OUT",
			Message: err.Error(),
		}
	case ErrorDidNotFillKKJNumber:
		return Response{
			Code:    http.StatusBadRequest,
			Status:  "MISSING_FIELD",
			Message: err.Error(),
		}
	case ErrorMismatchFields:
		return Response{
			Code:    http.StatusBadRequest,
			Status:  "MISMATCH_FIELDS",
			Message: err.Error(),
		}
	case ErrorStartDateLater:
		return Response{
			Code:    http.StatusBadRequest,
			Status:  "INVALID_VALUES",
			Message: err.Error(),
		}
	case ErrorMissingDepartmentCool:
		return Response{
			Code:    http.StatusBadRequest,
			Status:  "MISSING_FIELDS",
			Message: err.Error(),
		}
	case ErrorViolateAllowedForPrivate:
		return Response{
			Code:    http.StatusBadRequest,
			Status:  "MISSING_FIELDS",
			Message: err.Error(),
		}
	case ErrorMaxPerTrxIsZero:
		return Response{
			Code:    http.StatusBadRequest,
			Status:  "INVALID_VALUES",
			Message: err.Error(),
		}
	case ErrorAttendanceTypeWhenRequired:
		return Response{
			Code:    http.StatusBadRequest,
			Status:  "MISSING_FIELDS",
			Message: err.Error(),
		}
	case ErrorEventNotAvailable:
		return Response{
			Code:    http.StatusBadRequest,
			Status:  "FORBIDDEN_REGISTRATION",
			Message: err.Error(),
		}
	case ErrorEventCanOnlyRegisterOnce:
		return Response{
			Code:    http.StatusConflict,
			Status:  "ALREADY_REGISTERED",
			Message: err.Error(),
		}
	case ErrorAlreadyRegistered:
		return Response{
			Code:    http.StatusConflict,
			Status:  "ALREADY_REGISTERED",
			Message: err.Error(),
		}
	case ErrorCannotUsePersonalQR:
		return Response{
			Code:    http.StatusForbidden,
			Status:  "FORBIDDEN_REGISTRATION",
			Message: err.Error(),
		}
	case ErrorAlreadyVerified:
		return Response{
			Code:    http.StatusConflict,
			Status:  "ALREADY_UPDATED",
			Message: err.Error(),
		}
	case ErrorAlreadyCancelled:
		return Response{
			Code:    http.StatusConflict,
			Status:  "ALREADY_UPDATED",
			Message: err.Error(),
		}
	case ErrorForbiddenStatus:
		return Response{
			Code:    http.StatusBadRequest,
			Status:  "FORBIDDEN_STATUS",
			Message: err.Error(),
		}
	case ErrorReasonEmpty:
		return Response{
			Code:    http.StatusUnprocessableEntity,
			Status:  "MISSING_FIELDS",
			Message: err.Error(),
		}
	case ErrorDifferentCommunityId:
		return Response{
			Code:    http.StatusForbidden,
			Status:  "FORBIDDEN_ACTION",
			Message: err.Error(),
		}
	case ErrorConflictRelationDelete:
		return Response{
			Code:    http.StatusConflict,
			Status:  "CONFLICT_USERS",
			Message: err.Error(),
		}
	case ErrorLimitMustBeGreaterThanZero:
		return Response{
			Code:    http.StatusBadRequest,
			Status:  "INVALID_VALUES",
			Message: err.Error(),
		}
	case ErrorCSVOrXLSX:
		return Response{
			Code:    http.StatusUnprocessableEntity,
			Status:  "INVALID_FORMAT",
			Message: err.Error(),
		}

	default:
		return Response{
			Code:    http.StatusInternalServerError,
			Status:  "INTERNAL_SERVER_ERROR",
			Message: err.Error(),
		}
	}
}
