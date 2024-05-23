package v1

import (
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/models"
	"go-community/internal/pkg/validator"
	"go-community/internal/usecases"
	"net/http"

	"github.com/labstack/echo/v4"
)

type CoolHandler struct {
	usecase *usecases.Usecases
}

func NewCoolHandler(api *echo.Group, u *usecases.Usecases) {
    handler := &CoolHandler{usecase: u}

    // Define campus routes
	endpoint := api.Group("/campus")
    endpoint.POST("/category", handler.CreateCategory)
}

func (clh *CoolHandler) CreateCategory(ctx echo.Context) error {
	var request models.CreateCoolCategoryRequest
	if err := ctx.Bind(&request); err != nil {
        return response.Error(ctx, models.ErrorInvalidInput)
    }

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	new, err := clh.usecase.CoolCategory.Create(ctx.Request().Context(), &request)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusCreated, new.ToResponse())
}