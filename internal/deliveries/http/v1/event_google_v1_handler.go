package v1

import (
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/models"
	"go-community/internal/usecases"

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
	state := ctx.QueryParam("state")
	code := ctx.QueryParam("code")

	user, token, statusCode, err := egh.usecase.EventUser.Account(ctx.Request().Context(), state, code)
	if err != nil {
		return response.Error(ctx, err)
	}

	res := models.CreateEventUserResponse{Type: models.TYPE_EVENT_USER, Name: user.Name, AccountNumber: user.AccountNumber, Email: user.Email, Role: user.Role, Status: user.Status, Token: token}
	return response.Success(ctx, statusCode, res.ToCreateEventUser())
}
