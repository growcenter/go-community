package v2

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

type CoolHandler struct {
	usecase *usecases.Usecases
	conf    *config.Configuration
}

func NewCoolHandler(api *echo.Group, u *usecases.Usecases, c *config.Configuration) {
	handler := &CoolHandler{usecase: u, conf: c}

	endpoint := api.Group("/cools")

	// Define campus routes
	endpointOld := api.Group("/cool")
	endpointOld.POST("/category", handler.CreateCategory)
	endpointOld.GET("/category", handler.GetAllCategory)

	endpointAuth := endpoint.Group("")
	endpointAuth.Use(middleware.UserMiddleware(c, u, nil))
	endpointAuth.POST("/join", handler.CreateNewJoiner)

	endpointInternalAuth := api.Group("/internal/cools")
	endpointInternalAuth.Use(middleware.UserMiddleware(c, u, []string{"event-internal-view", "event-internal-edit"}))
	endpointInternalAuth.GET("/join", handler.GetAllNewJoiner)
	endpointInternalAuth.PATCH("/join/:idNewJoiner/:status", handler.UpdateNewJoiner)
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

func (clh *CoolHandler) CreateNewJoiner(ctx echo.Context) error {
	var request models.CreateCoolNewJoinerRequest
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, err)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	cool, err := clh.usecase.CoolNewJoiner.Create(ctx.Request().Context(), &request)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.SuccessV2(ctx, http.StatusCreated, "", cool)
}

func (clh *CoolHandler) GetAllNewJoiner(ctx echo.Context) error {
	var param models.GetAllCoolNewJoinerCursorParam
	if err := ctx.Bind(&param); err != nil {
		return response.Error(ctx, err)
	}

	if err := validator.Validate(param); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	data, info, err := clh.usecase.CoolNewJoiner.GetAll(ctx.Request().Context(), param)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.SuccessPaginationV2(ctx, http.StatusOK, "", *info, data)
}

func (clh *CoolHandler) UpdateNewJoiner(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("idNewJoiner")) // convert string to int
	if err != nil {
		return response.Error(ctx, err)
	}

	request := models.UpdateCoolNewJoinerRequest{
		Status:    ctx.Param("status"),
		Id:        id,
		UpdatedBy: ctx.Get("id").(string),
	}

	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, err)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	cool, err := clh.usecase.CoolNewJoiner.UpdateStatus(ctx.Request().Context(), &request)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.SuccessV2(ctx, http.StatusOK, "", cool)
}
