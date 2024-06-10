package v1

import (
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/models"
	"go-community/internal/pkg/validator"
	"go-community/internal/usecases"
	"net/http"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	usecase *usecases.Usecases
}

func NewUserHandler(api *echo.Group, u *usecases.Usecases) {
	handler := &UserHandler{usecase: u}

	endpoint := api.Group("/user")
	endpoint.POST("/cool", handler.CreateCool)
	// endpoint.GET("/:accountNumber")
}

func (uh *UserHandler) CreateCool(ctx echo.Context) error {
	var request models.CreateUserCoolRequest
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, models.ErrorInvalidInput)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	new, err := uh.usecase.User.CreateCool(ctx.Request().Context(), &request)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusCreated, new.ToCreateUserCool())
}
