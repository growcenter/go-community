package v1

import (
	"go-community/internal/constants"
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/usecases"
	"net/http"

	"github.com/labstack/echo/v4"
)

type EventGoogleHandler struct {
	usecase *usecases.Usecases
}

func NewEventGoogleHandler(api *echo.Group, u *usecases.Usecases) {
	handler := &EventUserHandler{usecase: u}

	// Define campus routes
	eventGoogleEndpoint := api.Group("/event/google")
	eventGoogleEndpoint.GET("/callback", handler.GoogleCallback)
}

func (egh *EventGoogleHandler) GoogleCallback(ctx echo.Context) error {
	return response.Success(ctx, http.StatusGone, constants.DeprecatedResponse)
}
