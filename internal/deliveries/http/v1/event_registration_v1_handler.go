package v1

import (
	"github.com/labstack/echo/v4"
	"go-community/internal/config"
	"go-community/internal/constants"
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/usecases"
	"net/http"
)

type EventRegistrationHandler struct {
	usecase *usecases.Usecases
}

func NewEventRegistrationHandler(api *echo.Group, u *usecases.Usecases, c *config.Configuration) {
	handler := &EventRegistrationHandler{usecase: u}

	// Define campus routes
	eventEndpoint := api.Group("/events/registration")
	eventEndpoint.POST("", handler.Create)
	eventEndpoint.POST("/homebase", handler.CreateHomebase)
	eventEndpoint.GET("", handler.GetRegistered)
	eventEndpoint.DELETE("/:code", handler.Cancel)

}

func (erh *EventRegistrationHandler) Create(ctx echo.Context) error {
	return response.Success(ctx, http.StatusGone, constants.DeprecatedResponse)
}

func (erh *EventRegistrationHandler) CreateHomebase(ctx echo.Context) error {
	return response.Success(ctx, http.StatusGone, constants.DeprecatedResponse)
}

func (erh *EventRegistrationHandler) GetRegistered(ctx echo.Context) error {
	return response.Success(ctx, http.StatusGone, constants.DeprecatedResponse)
}

func (erh *EventRegistrationHandler) Cancel(ctx echo.Context) error {
	return response.Success(ctx, http.StatusGone, constants.DeprecatedResponse)
}
