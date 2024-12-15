package v2

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

// Create godoc
// @Summary Create Roles
// @Description Create roles which would be for access
// @Tags roles
// @Accept json
// @Produce json
// @Param user body models.CreateRoleRequest true "User object that needs to be added"
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Success 201 {object} models.RoleResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorValidationResponse{errors=validator.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v2/roles [post]
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

// GetAllRoles godoc
// @Summary Get All Roles
// @Description Get All Roles
// @Tags roles
// @Accept json
// @Produce json
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Success 200 {object} models.List{data=[]models.RoleResponse} "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorValidationResponse{errors=validator.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v2/roles [get]
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
