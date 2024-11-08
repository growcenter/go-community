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
	ErrorDidNotFillKKJNumber = errors.New("please input the kkj number if you input jemaat id")
)

type (
	ErrorResponse struct {
		Code    int    `json:"code"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	ErrorValidationResponse struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Errors  interface{} `json:"errors"`
	}
)

var ValidationErrorMapping = map[string]error{
	"User_Name_required": ErrorInvalidInput,
	"User_Email_email":   ErrorEmailInput,
}

func ErrorMapping(err error) ErrorResponse {
	switch err {
	case ErrorUserNotFound:
		return ErrorResponse{
			Code:    http.StatusNotFound,
			Status:  "DATA_NOT_FOUND",
			Message: err.Error(),
		}
	case ErrorDataNotFound:
		return ErrorResponse{
			Code:    http.StatusNotFound,
			Status:  "DATA_NOT_FOUND",
			Message: err.Error(),
		}
	case ErrorInvalidInput:
		return ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "INVALID_INPUT",
			Message: err.Error(),
		}
	case ErrorNoRows:
		return ErrorResponse{
			Code:    http.StatusInternalServerError,
			Status:  "DATABASE_ERROR",
			Message: err.Error(),
		}
	case ErrorAlreadyExist:
		return ErrorResponse{
			Code:    http.StatusConflict,
			Status:  "ALREADY_EXISTS",
			Message: err.Error(),
		}
	case ErrorInvalidInput:
		return ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "INVALID_ARGUMENT",
			Message: err.Error(),
		}
	case ErrorEmailInput:
		return ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "INVALID_ARGUMENT",
			Message: err.Error(),
		}
	case ErrorAgeRange:
		return ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "INVALID_ARGUMENT",
			Message: err.Error(),
		}
	case ErrorFetchGoogle:
		return ErrorResponse{
			Code:    http.StatusInternalServerError,
			Status:  "INTERNAL_SERVER_ERROR",
			Message: err.Error(),
		}
	case ErrorTokenSignature:
		return ErrorResponse{
			Code:    http.StatusUnauthorized,
			Status:  "INVALID_TOKEN_SIGNATURE",
			Message: err.Error(),
		}
	case ErrorInvalidToken:
		return ErrorResponse{
			Code:    http.StatusUnauthorized,
			Status:  "INVALID_TOKEN",
			Message: err.Error(),
		}
	case ErrorExpiredToken:
		return ErrorResponse{
			Code:    http.StatusUnauthorized,
			Status:  "EXPIRED_TOKEN",
			Message: err.Error(),
		}
	case ErrorEmptyToken:
		return ErrorResponse{
			Code:    http.StatusUnauthorized,
			Status:  "MISSING_TOKEN",
			Message: err.Error(),
		}
	case ErrorEmailPhoneNumberEmpty:
		return ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "MISSING_FIELD",
			Message: err.Error(),
		}
	case ErrorInvalidPassword:
		return ErrorResponse{
			Code:    http.StatusUnauthorized,
			Status:  "INVALID_CREDENTIALS",
			Message: err.Error(),
		}
	case ErrorCannotRegisterYet:
		return ErrorResponse{
			Code:    http.StatusForbidden,
			Status:  "FORBIDDEN_REGISTRATION",
			Message: err.Error(),
		}
	case ErrorRegistrationTimeDisabled:
		return ErrorResponse{
			Code:    http.StatusForbidden,
			Status:  "FORBIDDEN_REGISTRATION",
			Message: err.Error(),
		}
	case ErrorInvalidAPIKey:
		return ErrorResponse{
			Code:    http.StatusUnauthorized,
			Status:  "INVALID_KEY",
			Message: err.Error(),
		}
	case ErrorEmptyAPIKey:
		return ErrorResponse{
			Code:    http.StatusUnauthorized,
			Status:  "MISSING_KEY",
			Message: err.Error(),
		}
	case ErrorEventNotValid:
		return ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "INVALID_INPUT",
			Message: err.Error(),
		}
	case ErrorExceedMaxSeating:
		return ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "FORBIDDEN_REGISTRATION",
			Message: err.Error(),
		}
	case ErrorRegisterQuotaNotAvailable:
		return ErrorResponse{
			Code:    http.StatusForbidden,
			Status:  "FORBIDDEN_REGISTRATION",
			Message: err.Error(),
		}
	case ErrorRegistrationCodeInvalid:
		return ErrorResponse{
			Code:    http.StatusForbidden,
			Status:  "FORBIDDEN_REGISTRATION",
			Message: err.Error(),
		}
	case ErrorRegistrationAlreadyCancel:
		return ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "ALREADY_CHANGED",
			Message: err.Error(),
		}
	case ErrorRegistrationAlreadyVerified:
		return ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "ALREADY_CHANGED",
			Message: err.Error(),
		}
	case ErrorForbiddenRole:
		return ErrorResponse{
			Code:    http.StatusForbidden,
			Status:  "FORBIDDEN_ROLE",
			Message: err.Error(),
		}
	case ErrorRegistrationWrongTime:
		return ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "INVALID_ARGUMENT",
			Message: err.Error(),
		}
	case ErrorNoRegistrationNeeded:
		return ErrorResponse{
			Code:    http.StatusForbidden,
			Status:  "FORBIDDEN_REGISTRATION",
			Message: err.Error(),
		}
	case ErrorLoggedOut:
		return ErrorResponse{
			Code:    http.StatusUnauthorized,
			Status:  "LOGGED_OUT",
			Message: err.Error(),
		}
	case ErrorDidNotFillKKJNumber:
		return ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "MISSING_FIELD",
			Message: err.Error(),
		}

	default:
		return ErrorResponse{
			Code:    http.StatusInternalServerError,
			Status:  "INTERNAL_SERVER_ERROR",
			Message: err.Error(),
		}
	}
}
