package v1

import (
	"go-community/internal/config"
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/deliveries/http/middleware"
	"go-community/internal/models"
	"go-community/internal/pkg/validator"
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
	eventEndpoint.Use(middleware.AdminMiddleware(c))
	eventRegistrationEndpoint := eventEndpoint.Group("/registrations")
	eventRegistrationEndpoint.GET("", handler.GetRegistered)
	eventRegistrationEndpoint.PATCH("/:code", handler.Verify)
	eventRegistrationEndpoint.GET("/:eventCode/summary", handler.GetSummary)

	// No need for bearer or role
	noBearerEventUserEndpoint := api.Group("/users")
	noBearerEventUserEndpoint.PATCH("", handler.UpdateAccountRole)
	noBearerEventEndpoint := api.Group("/events")
	noBearerEventEndpoint.GET("/summary/:sessionCode", handler.GetSummaryPerSession)
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

func (eih *EventInternalHandler) Verify(ctx echo.Context) error {
	code := ctx.Param("code")

	var request models.VerifyRegistrationRequest
	request.Code = code
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, models.ErrorInvalidInput)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	accountNumber := ctx.Get("accountNumber").(string)

	updated, err := eih.usecase.EventRegistration.Verify(ctx.Request().Context(), request, accountNumber)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusOK, updated.ToUpdate())
}

func (eih *EventInternalHandler) UpdateAccountRole(ctx echo.Context) error {
	var request models.UpdateAccountRoleRequest
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, models.ErrorInvalidInput)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	updated, err := eih.usecase.EventUser.UpdateRole(ctx.Request().Context(), request)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusOK, updated.ToUpdateAccountRole())
}

func (eih *EventInternalHandler) GetSummary(ctx echo.Context) error {
	detail, data, err := eih.usecase.EventSession.GetAllByEventCode(ctx.Request().Context(), ctx.Param("eventCode"))
	if err != nil {
		return response.Error(ctx, err)
	}

	var res []models.GetEventSessionsDataResponse
	for _, v := range data {
		res = append(res, *v.ToResponse())
	}

	return response.SuccessListWithDetail(ctx, http.StatusOK, len(res), detail, res)
}

func (eih *EventInternalHandler) GetSummaryPerSession(ctx echo.Context) error {
	count, err := eih.usecase.EventRegistration.Summary(ctx.Request().Context(), ctx.Param("sessionCode"))
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusOK, count)
}
