package v1

import (
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/models"
	"go-community/internal/pkg/validator"
	"go-community/internal/usecases"
	"net/http"

	"github.com/labstack/echo/v4"
)

type LocationHandler struct {
	usecase *usecases.Usecases
}

func NewLocationHandler(api *echo.Group, u *usecases.Usecases) {
    handler := &LocationHandler{usecase: u}

	endpoint := api.Group("/location")
    endpoint.POST("", handler.Create)
	endpoint.GET("", handler.GetAllLocation) 
	endpoint.GET("/:campusCode", handler.GetByCampusCode)
}

func (lh *LocationHandler) Create(ctx echo.Context) error {
	var request models.CreateLocationRequest
	if err := ctx.Bind(&request); err != nil {
        return response.Error(ctx, models.ErrorInvalidInput)
    }

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	new, err := lh.usecase.Location.Create(ctx.Request().Context(), &request)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusCreated, new.ToResponse())

}

func (lh *LocationHandler) GetAllLocation(ctx echo.Context) error {
	data, err := lh.usecase.Location.GetAll(ctx.Request().Context())
	if err != nil {
		return response.Error(ctx, err)
	}

	var res []models.LocationResponse
	for _, v := range data {
		res = append(res, *v.ToResponse())
	}

	return response.SuccessList(ctx, http.StatusOK, len(res), res)
}

func (lh *LocationHandler) GetByCampusCode(ctx echo.Context) error {
	data, err := lh.usecase.Location.GetByCampusCode(ctx.Request().Context(), ctx.Param("campusCode"))
	if err != nil {
		return response.Error(ctx, err)
	}

	var res []models.LocationResponse
	for _, v := range data {
		res = append(res, *v.ToResponse())
	}

	return response.SuccessList(ctx, http.StatusOK, len(res), res)
}