package response

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	Status			string
	Code			string
	Message			string
}

type ErrorMapping struct {
	Status		 	int
	Code         	error
	Response		ErrorResponse
}

func (e *ErrorMapping) Error() string {
	return fmt.Sprintf("error code %d", e.Status)
}

func (e *ErrorMapping) Builder() *ErrorMapping {
	return e
}

const (
	ERROR_STATUS = "error"
	ERROR_DATA_NOT_UNIQUE = "Data inputted is not unique"
	ERROR_NOT_FOUND = "Data not found"
	ERROR_BAD_REQUEST = "Bad request"
	ERROR_INTERNAL_SERVER_ERROR = "Internal Server Error"
	ERROR_UNAUTHORIZED = "Unauthorized"
	ERROR_UNPROCESSABLE_ENTITY = "ERROR"
)

var (
	ErrorDataNotUnique = ErrorMapping{
		Status: http.StatusConflict,
		Response: ErrorResponse{
			Status: ERROR_STATUS,
			Code: ERROR_DATA_NOT_UNIQUE,
			Message: ERROR_DATA_NOT_UNIQUE,
		},
	}
	ErrorInternalServer = ErrorMapping{
		Status: http.StatusInternalServerError,
		Response: ErrorResponse{
			Status: ERROR_STATUS,
			Code: ERROR_INTERNAL_SERVER_ERROR,
			Message: ERROR_INTERNAL_SERVER_ERROR,
		},
	}
)

func ErrorBuilder(err ErrorMapping, msg error) error {
	err.Code = msg
	return &err
}

func ErrorBuilderCustom(code int, err string, message string) error {
	return &ErrorMapping{
		Status: code,
		Response: ErrorResponse{
			Status: ERROR_STATUS,
			Code: err,
			Message: message,
		},
	}
}

func Success(ctx echo.Context, status int, data interface{}) error {
	return ctx.JSON(status, data)
}

func Error(ctx echo.Context, err error) error {
	requestBody, er := io.ReadAll(ctx.Request().Body)
	if er != nil {
		return er
	}

	requestHeader, er := json.Marshal(ctx.Request().Header)
	if er != nil {
		return er
	}

	response, ok := err.(*ErrorMapping)
	if ok {
		fmt.Println(requestBody)
		fmt.Println(requestHeader)

		return ctx.JSON(response.Builder().Status, response.Builder().Response)
	} else {
		return ctx.JSON(ErrorInternalServer.Status, ErrorInternalServer.Response)
	}
}