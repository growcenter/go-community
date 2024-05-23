package response

import (
	"go-community/internal/models"
	"net/http"

	"github.com/hashicorp/go-multierror"
	"github.com/labstack/echo/v4"
)

func Error(ctx echo.Context, err error) error {
    errorResponse := models.ErrorMapping(err)
    return ctx.JSON(errorResponse.Code, errorResponse)
}

func ErrorValidation(ctx echo.Context, errors interface{}) error {
	res := models.ErrorValidationResponse{
		Code:  http.StatusUnprocessableEntity,
		Message: "Validation failed for one or more fields.",
	}
	if data, ok := errors.(*multierror.Error); ok {
		res.Errors = data.Errors
	}

	return ctx.JSON(res.Code, res)
}

func Success(ctx echo.Context, code int,  response interface{}) error {
    return ctx.JSON(code, response)
}
