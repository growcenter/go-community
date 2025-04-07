package v1

import (
	"github.com/labstack/echo/v4"
	"go-community/internal/config"
	"go-community/internal/constants"
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/usecases"
	"net/http"
)

type EventInternalHandler struct {
	usecase *usecases.Usecases
}

func NewEventInternalHandler(api *echo.Group, u *usecases.Usecases, c *config.Configuration) {
	handler := &EventInternalHandler{usecase: u}

	// Define campus routes
	eventEndpoint := api.Group("/internal/events")
	eventRegistrationEndpoint := eventEndpoint.Group("/registrations")
	eventRegistrationEndpoint.GET("", handler.GetRegistered)
	eventRegistrationEndpoint.PATCH("/:code", handler.Verify)
	eventRegistrationEndpoint.GET("/:eventCode/summary", handler.GetSummary)

	// No need for bearer or role
	noBearerEventUserEndpoint := api.Group("/users")
	noBearerEventUserEndpoint.PATCH("", handler.UpdateAccountRole)
	noBearerEventEndpoint := api.Group("/events")
	noBearerEventEndpoint.GET("/summary/:sessionCode", handler.GetSummaryPerSession)
}

func (eih *EventInternalHandler) GetRegistered(ctx echo.Context) error {
	return response.Success(ctx, http.StatusGone, constants.DeprecatedResponse)
}

func (eih *EventInternalHandler) Verify(ctx echo.Context) error {
	return response.Success(ctx, http.StatusGone, constants.DeprecatedResponse)
}

func (eih *EventInternalHandler) UpdateAccountRole(ctx echo.Context) error {
	return response.Success(ctx, http.StatusGone, constants.DeprecatedResponse)
}

func (eih *EventInternalHandler) GetSummary(ctx echo.Context) error {
	return response.Success(ctx, http.StatusGone, constants.DeprecatedResponse)
}

func (eih *EventInternalHandler) GetSummaryPerSession(ctx echo.Context) error {
	return response.Success(ctx, http.StatusGone, constants.DeprecatedResponse)
}
