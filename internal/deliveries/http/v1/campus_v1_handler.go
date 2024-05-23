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