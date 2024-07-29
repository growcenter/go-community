package v1

import (
	"go-community/internal/config"
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/deliveries/http/middleware"
	"go-community/internal/models"
	"go-community/internal/usecases"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type EventInternalHandler struct {
	usecase *usecases.Usecases
}

func NewEventInternalHandler(api *echo.Group, u *usecases.Usecases, c *config.Configuration) {
	handler := &EventInternalHandler{usecase: u}

	// Define campus routes
	eventEndpoint := api.Group("/internal/events")
	eventEndpoint.Use(middleware.UserMiddleware(c))
	eventEndpoint.GET("", handler.GetRegistered)
}

func (eih *EventInternalHandler) GetRegistered(ctx echo.Context) error {
	page, _ := strconv.Atoi(ctx.QueryParam("page"))
	limit, _ := strconv.Atoi(ctx.QueryParam("limit"))
	sort := ctx.QueryParam("sort")
	search := ctx.QueryParam("search")
	filterSessionCode := ctx.QueryParam("sessionCode")
	filterEventCode := ctx.QueryParam("eventCode")

	// Set default values if necessary
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	params := models.GetAllPaginationParams{
		Page:              page,
		Limit:             limit,
		Sort:              sort,
		Search:            search,
		FilterSessionCode: filterSessionCode,
		FilterEventCode:   filterEventCode,
	}

	registers, count, err := eih.usecase.EventRegistration.GetAll(ctx.Request().Context(), params)
	if err != nil {
		return response.Error(ctx, err)
	}

	// Calculate pagination info
	pages := (count + int64(limit) - 1) / int64(limit)

	info := models.PaginationInfo{
		CurrentPage: page,
		TotalPages:  int(pages),
		TotalData:   int(count),
		Limit:       limit,
		Parameter: models.GetAllPaginationParamsResponse{
			Search:            search,
			FilterSessionCode: filterSessionCode,
			FilterEventCode:   filterEventCode,
		},
	}

	return response.SuccessPagination(ctx, http.StatusOK, info, registers)
}
