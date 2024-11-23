package v1

import (
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/models"
	"go-community/internal/pkg/validator"
	"go-community/internal/usecases"
	"net/http"

	"github.com/labstack/echo/v4"
)

type RoleHandler struct {
	usecase *usecases.Usecases
}

func NewRoleHandler(api *echo.Group, u *usecases.Usecases) {
	handler := &RoleHandler{usecase: u}

	// Define campus routes
	endpoint := api.Group("/roles")
	endpoint.POST("", handler.Create)
	endpoint.GET("", handler.GetAllRoles)
}

func (rh *RoleHandler) Create(ctx echo.Context) error {
	// Bind the JSON Request in order to get the usecase work
	var request models.CreateRoleRequest
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, models.ErrorInvalidInput)
	}

	// Validate inputs based on requirement
	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	// Usage of the usecase
	new, err := rh.usecase.Role.Create(ctx.Request().Context(), &request)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusCreated, new.ToResponse())

}

func (rh *RoleHandler) GetAllRoles(ctx echo.Context) error {
	data, err := rh.usecase.Role.GetAll(ctx.Request().Context())
	if err != nil {
		return response.Error(ctx, err)
	}

	var res []models.RoleResponse
	for _, v := range data {
		res = append(res, *v.ToResponse())
	}

	return response.SuccessList(ctx, http.StatusOK, len(res), res)
}
