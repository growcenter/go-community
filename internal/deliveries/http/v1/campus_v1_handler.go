package v1

import (
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/models"
	"go-community/internal/pkg/validator"
	"go-community/internal/usecases"
	"net/http"

	"github.com/labstack/echo/v4"
)

type CampusHandler struct {
	usecase *usecases.Usecases
}

func NewCampusHandler(api *echo.Group, u *usecases.Usecases) {
    handler := &CampusHandler{usecase: u}

    // Define campus routes
	endpoint := api.Group("/campus")
    endpoint.POST("", handler.Create)
	endpoint.GET("", handler.GetAllCampus)
	endpoint.GET("/:code", handler.GetByCode)  
}

func (ch *CampusHandler) Create(ctx echo.Context) error {
	// Bind the JSON Request in order to get the usecase work
	var request models.CreateCampusRequest
	if err := ctx.Bind(&request); err != nil {
        return response.Error(ctx, models.ErrorInvalidInput)
    }

	// Validate inputs based on requirement
	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	// Usage of the usecase
	new, err := ch.usecase.Campus.Create(ctx.Request().Context(), &request)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusCreated, new.ToResponse())

}

func (ch *CampusHandler) GetAllCampus(ctx echo.Context) error {
	data, err := ch.usecase.Campus.GetAll(ctx.Request().Context())
	if err != nil {
		return response.Error(ctx, err)
	}

	var res []models.CampusResponse
	for _, v := range data {
		res = append(res, *v.ToResponse())
	}

	return response.SuccessList(ctx, http.StatusOK, len(res), res)
}
func (ch *CampusHandler) GetByCode(ctx echo.Context) error {
    code := ctx.Param("code")
    campus, err := ch.usecase.Campus.GetByCode(ctx.Request().Context(), code)
    if err != nil {
        return response.Error(ctx, err)
    }

    return response.Success(ctx, http.StatusOK, campus.ToResponse())
}
