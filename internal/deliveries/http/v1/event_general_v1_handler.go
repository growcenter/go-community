package v1

import (
	"github.com/labstack/echo/v4"
	"go-community/internal/config"
	"go-community/internal/constants"
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/usecases"
	"net/http"
)

type EventGeneralHandler struct {
	usecase *usecases.Usecases
}

func NewEventGeneralHandler(api *echo.Group, u *usecases.Usecases, c *config.Configuration) {
	handler := &EventGeneralHandler{usecase: u}

	// Define campus routes
	eventGeneralEndpoint := api.Group("/events")
	eventGeneralEndpoint.GET("", handler.GetAll)
}

func (egh *EventGeneralHandler) GetAll(ctx echo.Context) error {
	return response.Success(ctx, http.StatusGone, constants.DeprecatedResponse)
}
