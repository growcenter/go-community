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

type EventSessionHandler struct {
	usecase *usecases.Usecases
}

func NewEventSessionHandler(api *echo.Group, u *usecases.Usecases, c *config.Configuration) {
	handler := &EventSessionHandler{usecase: u}

	// Define campus routes
	eventGeneralEndpoint := api.Group("/events")
	eventGeneralEndpoint.Use(middleware.JWTMiddleware(c))
	eventGeneralEndpoint.GET("/:eventCode/sessions", handler.GetAll)
}

func (esh *EventSessionHandler) GetAll(ctx echo.Context) error {
	detail, data, err := esh.usecase.EventSession.GetAllByEventCode(ctx.Request().Context(), ctx.Param("eventCode"))
	if err != nil {
		return response.Error(ctx, err)
	}

	var res []models.GetEventSessionsDataResponse
	for _, v := range data {
		res = append(res, *v.ToResponse())
	}

	return response.SuccessListWithDetail(ctx, http.StatusOK, len(res), detail, res)
}
