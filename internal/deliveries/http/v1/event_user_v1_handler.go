package v1

import (
	"github.com/labstack/echo/v4"
	"go-community/internal/config"
	"go-community/internal/constants"
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/usecases"
	"net/http"
)

type EventUserHandler struct {
	usecase *usecases.Usecases
}

func NewEventUserHandler(api *echo.Group, u *usecases.Usecases, c *config.Configuration) {
	handler := &EventUserHandler{usecase: u}

	// Define campus routes
	eventGoogleEndpoint := api.Group("/event/google")
	eventGoogleEndpoint.GET("/login", handler.GoogleRedirect)
	eventGoogleEndpoint.GET("/callback", handler.GoogleCallback)

	eventUserEndpoint := api.Group("/event/user")
	eventUserEndpoint.POST("/login", handler.ManualLogin)
	eventUserEndpoint.POST("/register", handler.ManualRegister)
	eventUserEndpoint.POST("/register-worker", handler.ManualRegisterWorker)
	eventUserEndpoint.PATCH("/forgot", handler.ForgotPassword)
	authEventUserEndpoint := eventUserEndpoint.Group("")
	authEventUserEndpoint.GET("", handler.GetByToken)
	authEventUserEndpoint.PATCH("/logout", handler.Logout)
}

func (euh *EventUserHandler) GoogleRedirect(ctx echo.Context) error {
	return response.Success(ctx, http.StatusGone, constants.DeprecatedResponse)
}

func (euh *EventUserHandler) GoogleCallback(ctx echo.Context) error {
	return response.Success(ctx, http.StatusGone, constants.DeprecatedResponse)
}

func (euh *EventUserHandler) ManualLogin(ctx echo.Context) error {
	return response.Success(ctx, http.StatusGone, constants.DeprecatedResponse)
}

func (euh *EventUserHandler) ManualRegister(ctx echo.Context) error {
	return response.Success(ctx, http.StatusGone, constants.DeprecatedResponse)
}

func (euh *EventUserHandler) ManualRegisterWorker(ctx echo.Context) error {
	return response.Success(ctx, http.StatusGone, constants.DeprecatedResponse)
}

func (euh *EventUserHandler) GetByToken(ctx echo.Context) error {
	return response.Success(ctx, http.StatusGone, constants.DeprecatedResponse)
}

func (euh *EventUserHandler) Logout(ctx echo.Context) error {
	return response.Success(ctx, http.StatusGone, constants.DeprecatedResponse)
}

func (euh *EventUserHandler) ForgotPassword(ctx echo.Context) error {
	return response.Success(ctx, http.StatusGone, constants.DeprecatedResponse)
}
