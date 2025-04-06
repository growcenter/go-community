package v2

import (
	"github.com/labstack/echo/v4"
	"go-community/internal/config"
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/deliveries/http/middleware"
	"go-community/internal/models"
	"go-community/internal/pkg/validator"
	"go-community/internal/usecases"
	"net/http"
	"strconv"
)

type FlagHandler struct {
	usecase *usecases.Usecases
	conf    config.Configuration
}

func NewFlagHandler(api *echo.Group, u *usecases.Usecases, c config.Configuration) {
	handler := &FlagHandler{usecase: u, conf: c}

	endpoint := api.Group("/flags")
	endpoint.Use(middleware.UserMiddleware(&c, u, []string{"event-internal-view", "event-internal-edit"}))
	endpoint.GET("", handler.GetAll)
	endpoint.GET("/:key", handler.GetByKey)
	endpoint.POST("", handler.Create)
	endpoint.PUT("/:key", handler.Update)
	endpoint.PATCH("/:key/toggle/:action", handler.Toggle)
	endpoint.DELETE("/:key", handler.Delete)
}

func (fh *FlagHandler) GetAll(ctx echo.Context) error {
	events, err := fh.usecase.FeatureFlag.GetAll(ctx.Request().Context())
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.SuccessList(ctx, http.StatusOK, len(events), events)
}

func (fh *FlagHandler) GetByKey(ctx echo.Context) error {
	events, err := fh.usecase.FeatureFlag.GetByKey(ctx.Request().Context(), ctx.Param("key"))
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusOK, events)
}

func (fh *FlagHandler) Create(ctx echo.Context) error {
	var request models.FeatureFlagRequest
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, err)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	flag, err := fh.usecase.FeatureFlag.Create(ctx.Request().Context(), request)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusCreated, flag.ToResponse())
}

func (fh *FlagHandler) Update(ctx echo.Context) error {
	var request models.FeatureFlagRequest
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, err)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	flag, err := fh.usecase.FeatureFlag.Update(ctx.Request().Context(), ctx.Param("key"), request)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusCreated, flag.ToResponse())
}

func (fh *FlagHandler) Toggle(ctx echo.Context) error {
	actionBool, err := strconv.ParseBool(ctx.Param("action"))
	if err != nil {
		return response.Error(ctx, err)
	}

	err = fh.usecase.FeatureFlag.Toggle(ctx.Request().Context(), ctx.Param("key"), actionBool)
	if err != nil {
		return response.Error(ctx, err)
	}

	res := map[string]interface{}{
		"type":    models.TYPE_FEATURE_FLAG,
		"message": "Feature flag toggled successfully",
	}

	return response.Success(ctx, http.StatusAccepted, res)
}

func (fh *FlagHandler) Delete(ctx echo.Context) error {
	err := fh.usecase.FeatureFlag.Delete(ctx.Request().Context(), ctx.Param("key"))
	if err != nil {
		return response.Error(ctx, err)
	}

	res := map[string]interface{}{
		"type":    models.TYPE_FEATURE_FLAG,
		"message": "Feature flag deleted successfully",
	}

	return response.Success(ctx, http.StatusAccepted, res)
}
