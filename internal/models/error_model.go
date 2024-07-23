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
	ErrorEmailPhoneNumberEmpty = errors.New("you should enter either phone number or email")
	ErrorCannotRegisterYet     = errors.New("you cannot register yet, wait until the time allowed first")
	ErrorRegistrationDisabled  = errors.New("you cannot register anymore, since the time is already closed")

	// Google Error
	ErrorFetchGoogle = errors.New("error while retrieving user from google")

	// Auth Error
	ErrorTokenSignature = errors.New("invalid signature")
	ErrorInvalidToken   = errors.New("token is invalid")
	ErrorExpiredToken   = errors.New("token is expired. please login again to use the account")
	ErrorEmptyToken     = errors.New("token is empty")

	// Special for Validation Error
	ErrorInvalidInput = errors.New("invalid request input")
	ErrorEmailInput   = errors.New("email format is invalid, should be: xxxx@xxxx.com")
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
	case ErrorRegistrationDisabled:
		return ErrorResponse{
			Code:    http.StatusForbidden,
			Status:  "FORBIDDEN_REGISTRATION",
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
