package v1

import (
	"go-community/internal/config"
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/deliveries/http/middleware"
	"go-community/internal/models"
	"go-community/internal/pkg/validator"
	"go-community/internal/usecases"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type EventRegistrationHandler struct {
	usecase *usecases.Usecases
}

func NewEventRegistrationHandler(api *echo.Group, u *usecases.Usecases, c *config.Configuration) {
	handler := &EventRegistrationHandler{usecase: u}

	// Define campus routes
	eventEndpoint := api.Group("/events/registration")
	eventEndpoint.Use(middleware.UserMiddleware(c))
	eventEndpoint.POST("", handler.Create)
	eventEndpoint.GET("", handler.GetRegistered)
	eventEndpoint.DELETE("/:code", handler.Cancel)

}

func (erh *EventRegistrationHandler) Create(ctx echo.Context) error {
	var request models.CreateEventRegistrationRequest
	accountNumber := ctx.Get("accountNumber").(string)
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, models.ErrorInvalidInput)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	register, err := erh.usecase.EventRegistration.Create(ctx.Request().Context(), request, accountNumber)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusCreated, register.ToCreate())
}

func (erh *EventRegistrationHandler) GetRegistered(ctx echo.Context) error {
	var request models.GetRegisteredRequest
	accountNumber := ctx.Get("accountNumber").(string)
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, models.ErrorInvalidInput)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	registers, err := erh.usecase.EventRegistration.GetRegistered(ctx.Request().Context(), strings.ToLower(request.RegisteredBy), accountNumber)
	if err != nil {
		return response.Error(ctx, err)
	}

	// var res []models.GetRegisteredResponse
	// for _, v := range registers {
	// 	res = append(res, *v.ToResponse())
	// }

	return response.SuccessList(ctx, http.StatusOK, len(registers), registers)
}

func (erh *EventRegistrationHandler) Cancel(ctx echo.Context) error {
	var request models.CancelRegistrationRequest
	request.Code = ctx.Param("code")
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, models.ErrorInvalidInput)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	cancel, err := erh.usecase.EventRegistration.Cancel(ctx.Request().Context(), request)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusOK, cancel.ToCancel())
}
