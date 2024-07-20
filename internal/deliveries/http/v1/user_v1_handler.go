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
	endpoint.GET("/check", handler.Check)
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

func (uh *UserHandler) CreateUser(ctx echo.Context) error {
	var request models.CreateUserRequest
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, models.ErrorInvalidInput)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	new, err := uh.usecase.User.CreateUser(ctx.Request().Context(), &request)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusCreated, new.ToCreateUser())
}

func (uh *UserHandler) Check(ctx echo.Context) error {
	request := ctx.QueryParam("email")

	isExist, data, err := uh.usecase.User.CheckByEmail(ctx.Request().Context(), request)
	if err != nil {
		return response.Error(ctx, err)
	}

	res := models.CheckUserEmailResponse{IsExist: isExist, UserType: data.UserType, Email: request}
	return response.Success(ctx, http.StatusOK, res.ToCheck())
}
