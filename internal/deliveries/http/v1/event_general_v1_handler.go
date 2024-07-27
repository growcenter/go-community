package v1

import (
	"go-community/internal/config"
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/deliveries/http/middleware"
	"go-community/internal/models"
	"go-community/internal/usecases"
	"net/http"

	"github.com/labstack/echo/v4"
)

type EventGeneralHandler struct {
	usecase *usecases.Usecases
}

func NewEventGeneralHandler(api *echo.Group, u *usecases.Usecases, c *config.Configuration) {
	handler := &EventGeneralHandler{usecase: u}

	// Define campus routes
	eventGeneralEndpoint := api.Group("/events")
	eventGeneralEndpoint.Use(middleware.UserMiddleware(c))
	eventGeneralEndpoint.GET("", handler.GetAll)
}

func (egh *EventGeneralHandler) GetAll(ctx echo.Context) error {
	detail, data, err := egh.usecase.EventGeneral.GetAll(ctx.Request().Context())
	if err != nil {
		return response.Error(ctx, err)
	}

	var res []models.GetGeneralEventDataResponse
	for _, v := range data {
		res = append(res, *v.ToResponse())
	}

	return response.SuccessListWithDetail(ctx, http.StatusOK, len(res), detail, res)
}
