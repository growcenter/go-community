package v2

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
	endpoint := api.Group("/cool")
	endpoint.POST("/category", handler.CreateCategory)
	endpoint.GET("/category", handler.GetAllCategory)
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

func (clh *CoolHandler) GetAllCategory(ctx echo.Context) error {
	data, err := clh.usecase.CoolCategory.GetAll(ctx.Request().Context())
	if err != nil {
		return response.Error(ctx, err)
	}

	var res []models.CoolCategoryResponse
	for _, v := range data {
		res = append(res, *v.ToResponse())
	}

	return response.SuccessList(ctx, http.StatusOK, len(res), res)
}
