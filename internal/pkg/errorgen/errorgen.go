package errorgen

import (
	"errors"
	"fmt"
	"go-community/internal/models"
	"net/http"
)

// Metadata (optional extension if needed later)
// type Metadata struct {
// 	RequestId string `json:"request_id"`
// 	Timestamp string `json:"timestamp"`
// }

type Response struct {
	Code     int              `json:"code"`
	Status   string           `json:"status"`
	Message  string           `json:"message"`
	Metadata *models.Metadata `json:"metadata,omitempty"` // Enable if needed
}

type HTTPError struct {
	Response Response
	Err      error
}

func (e *HTTPError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Response.Message
}

type ErrorMapping struct {
	Code   int
	Status string
}

// ==== Error Mapping Configuration ====

var errorMappings = map[error]ErrorMapping{
	ErrUserNotFound: {Code: http.StatusNotFound, Status: "DATA_NOT_FOUND"},
	ErrInvalidInput: {Code: http.StatusBadRequest, Status: "BAD_REQUEST"},
	ErrUnauthorized: {Code: http.StatusUnauthorized, Status: "UNAUTHORIZED"},
	ErrForbidden:    {Code: http.StatusForbidden, Status: "FORBIDDEN"},
	ErrEmailExists:  {Code: http.StatusConflict, Status: "CONFLICT"},
	ErrTokenExpired: {Code: http.StatusUnauthorized, Status: "TOKEN_EXPIRED"},
	DataNotFound:    {Code: http.StatusNotFound, Status: "DATA_NOT_FOUND"},
	InvalidInput:    {Code: http.StatusBadRequest, Status: "INVALID_INPUT"},
	AlreadyExist:    {Code: http.StatusConflict, Status: "ALREADY_EXISTS"},
	InvalidData:     {Code: http.StatusBadRequest, Status: "INVALID_DATA"},
}

// ==== Predefined Errors ====

var (
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrEmailExists  = errors.New("email already exists")
	ErrTokenExpired = errors.New("token expired")
	AlreadyExist    = errors.New("the resource that a client tried to create already exists")
	DataNotFound    = errors.New("a specified resource is not found")
	InvalidInput    = errors.New("invalid request input")
	InvalidData     = errors.New("invalid data")
)

// ==== Error Constructor ====

func Error(err error, message ...string) *HTTPError {
	mapping, ok := errorMappings[err]
	if !ok {
		mapping = ErrorMapping{
			Code:   http.StatusInternalServerError,
			Status: "INTERNAL_SERVER_ERROR",
		}
	}

	var msg string
	switch len(message) {
	case 0:
		msg = err.Error()
	case 1:
		msg = message[0]
	default:
		msg = fmt.Sprintf(message[0], toInterfaces(message[1:])...)
	}

	return &HTTPError{
		Response: Response{
			Code:    mapping.Code,
			Status:  mapping.Status,
			Message: msg,
		},
		Err: err,
	}
}

// ==== Fallback Response Extractor ====

func GetResponse(err error) Response {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.Response
	}
	return Response{
		Code:    http.StatusInternalServerError,
		Status:  "INTERNAL_SERVER_ERROR",
		Message: err.Error(),
	}
}

// ==== Add Custom Error Mapping ====

func AddMapping(err error, code int, status string) {
	errorMappings[err] = ErrorMapping{
		Code:   code,
		Status: status,
	}
}

// ==== Helper: Convert []string to []interface{} ====

func toInterfaces(args []string) []interface{} {
	result := make([]interface{}, len(args))
	for i, v := range args {
		result[i] = v
	}
	return result
}
