package v1

import (
	"go-community/internal/config"
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/deliveries/http/middleware"
	"go-community/internal/models"
	"go-community/internal/pkg/validator"
	"go-community/internal/usecases"
	"net/http"

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
}

func (erh *EventRegistrationHandler) Create(ctx echo.Context) error {
	var request models.CreateEventRegistrationRequest
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, models.ErrorInvalidInput)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	register, err := erh.usecase.EventRegistration.Create(ctx.Request().Context(), request)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusCreated, register.ToCreate())
}
