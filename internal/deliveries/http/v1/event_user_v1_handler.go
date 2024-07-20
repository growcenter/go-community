package v1

import (
	"fmt"
	"go-community/internal/config"
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/deliveries/http/middleware"
	"go-community/internal/models"
	"go-community/internal/pkg/validator"
	"go-community/internal/usecases"
	"net/http"

	"github.com/labstack/echo/v4"
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
	authEventUserEndpoint := eventUserEndpoint.Group("")
	authEventUserEndpoint.Use(middleware.JWTMiddleware(c))
	authEventUserEndpoint.GET("", handler.GetByToken)
}

func (euh *EventUserHandler) GoogleRedirect(ctx echo.Context) error {
	url, err := euh.usecase.EventUser.Redirect(ctx.Request().Context())
	if err != nil {
		return response.Error(ctx, err)
	}

	return ctx.Redirect(http.StatusTemporaryRedirect, url)
}

func (euh *EventUserHandler) GoogleCallback(ctx echo.Context) error {
	state := ctx.QueryParam("state")
	code := ctx.QueryParam("code")

	user, token, statusCode, err := euh.usecase.EventUser.Account(ctx.Request().Context(), state, code)
	if err != nil {
		return response.Error(ctx, err)
	}

	res := models.CreateEventUserResponse{Type: models.TYPE_EVENT_USER, Name: user.Name, AccountNumber: user.AccountNumber, Email: user.Email, Role: user.Role, Status: user.Status, Token: token}
	return response.Success(ctx, statusCode, res.ToCreateEventUser())
}

func (euh *EventUserHandler) ManualLogin(ctx echo.Context) error {
	var request models.LoginEventUserManualRequest
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, models.ErrorInvalidInput)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	user, token, err := euh.usecase.EventUser.ManualLogin(ctx.Request().Context(), request)
	if err != nil {
		return response.Error(ctx, err)
	}

	fmt.Println("bearer2: " + token)

	res := models.LoginEventUserManualResponse{Type: models.TYPE_EVENT_USER, PhoneNumber: user.PhoneNumber, Name: user.Name, AccountNumber: user.AccountNumber, Email: user.Email, Role: user.Role, Status: user.Status, Token: token}
	return response.Success(ctx, http.StatusCreated, res.ToLoginEventUserManual())
}

func (euh *EventUserHandler) ManualRegister(ctx echo.Context) error {
	var request models.CreateEventUserManualRequest
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, models.ErrorInvalidInput)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	user, token, err := euh.usecase.EventUser.ManualRegister(ctx.Request().Context(), request)
	if err != nil {
		return response.Error(ctx, err)
	}

	res := models.CreateEventUserManualResponse{Type: models.TYPE_EVENT_USER, Name: user.Name, Email: user.Email, PhoneNumber: user.PhoneNumber, AccountNumber: user.AccountNumber, Role: user.Role, Token: token, Status: user.Status}
	return response.Success(ctx, http.StatusCreated, res.ToCreateEventUserManual())
}

func (euh *EventUserHandler) GetByToken(ctx echo.Context) error {
	accountNumber := ctx.Get("accountNumber").(string)

	user, err := euh.usecase.EventUser.GetByAccountNumber(ctx.Request().Context(), accountNumber)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusOK, user.ToResponse())
}
