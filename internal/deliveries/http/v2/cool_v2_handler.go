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
	"strings"

	"github.com/labstack/echo/v4"
)

type CoolHandler struct {
	usecase *usecases.Usecases
	conf    *config.Configuration
}

func NewCoolHandler(api *echo.Group, u *usecases.Usecases, c *config.Configuration) {
	handler := &CoolHandler{usecase: u, conf: c}

	endpoint := api.Group("/cools")
	endpoint.Use(middleware.OptionalMiddleware(c, u, nil))
	endpoint.GET("", handler.GetAll)

	// Define campus routes
	endpointOld := api.Group("/cool")
	endpointOld.POST("/category", handler.CreateCategory)
	endpointOld.GET("/category", handler.GetAllCategory)

	endpointAuth := endpoint.Group("")
	endpointAuth.Use(middleware.UserMiddleware(c, u, nil))
	endpointAuth.POST("/join", handler.CreateNewJoiner)
	endpointAuth.GET("/me", handler.GetCoolPersonal)
	endpointAuth.GET("/:code/members", handler.GetCoolMemberByCode)

	endpointInternalAuth := api.Group("/internal/cools")
	endpointInternalAuth.Use(middleware.UserMiddleware(c, u, []string{"cool-internal-view", "cool-internal-edit"}))
	endpointInternalAuth.GET("/join", handler.GetAllNewJoiner)
	endpointInternalAuth.PATCH("/join/:idNewJoiner/:status", handler.UpdateNewJoiner)
	endpointInternalAuth.POST("", handler.CreateCool)

	endpointCoreAuth := endpoint.Group("") // For cool admin leader, core team and facilitator
	endpointCoreAuth.Use(middleware.UserMiddleware(c, u, []string{"cool-member-manage"}))
	endpointCoreAuth.POST("/:code/members", handler.AddMemberByCode)
	endpointCoreAuth.GET("/:code", handler.GetCoolByCode)
	endpointCoreAuth.DELETE("/:code/members/:communityId", handler.DeleteMemberByCode)
	endpointCoreAuth.PATCH("/:code/members/:communityId", handler.UpdateMemberByCode)
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

func (clh *CoolHandler) CreateCool(ctx echo.Context) error {
	var request models.CreateCoolRequest
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, models.ErrorInvalidInput)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	new, err := clh.usecase.Cool.Create(ctx.Request().Context(), request)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.SuccessV2(ctx, http.StatusCreated, "", new.ToResponse())
}

func (clh *CoolHandler) GetAll(ctx echo.Context) error {
	header := ctx.Request().Header.Get("X-Cool-List")
	if header == "" {
		header = "option"
	}

	var userTypes []string
	var communityId string
	if ctx.Get("userTypes") != nil && ctx.Get("id") != nil {
		userTypes = ctx.Get("userTypes").([]string)
		communityId = ctx.Get("id").(string)
	}

	data, err := clh.usecase.Cool.GetAll(ctx.Request().Context(), header, userTypes, communityId)
	if err != nil {
		return response.ErrorV2(ctx, err)
	}

	return response.SuccessListV2(ctx, http.StatusOK, "", data)

	return nil // unreachable
}

func (clh *CoolHandler) GetCoolPersonal(ctx echo.Context) error {
	cool, err := clh.usecase.Cool.GetByCommunityId(ctx.Request().Context(), ctx.Get("id").(string))
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.SuccessV2(ctx, http.StatusOK, "", cool)
}

func (clh *CoolHandler) GetCoolMemberByCode(ctx echo.Context) error {
	parameter := models.GetCoolMemberByCoolCodeParameter{
		Code: ctx.Param("code"),
		Type: []string{},
	}

	if typeParam := ctx.QueryParam("type"); typeParam != "" {
		parameter.Type = strings.Split(typeParam, ",")
	}

	if err := validator.Validate(parameter); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	cools, err := clh.usecase.Cool.GetMemberByCode(ctx.Request().Context(), parameter)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.SuccessListV2(ctx, http.StatusOK, "", cools)
}

func (clh *CoolHandler) AddMemberByCode(ctx echo.Context) error {
	var request []models.AddCoolMemberRequest
	if err := ctx.Bind(&request); err != nil {
		return response.ErrorV2(ctx, models.ErrorInvalidInput)
	}

	for _, r := range request {
		if err := validator.Validate(r); err != nil {
			return response.ErrorValidation(ctx, err)
		}
	}

	members, err := clh.usecase.Cool.AddMemberByCode(ctx.Request().Context(), ctx.Get("userTypes").([]string), ctx.Param("code"), request)
	if err != nil {
		return response.ErrorV2(ctx, err)
	}

	return response.SuccessV2(ctx, http.StatusCreated, "", members)
}

func (clh *CoolHandler) DeleteMemberByCode(ctx echo.Context) error {
	request := models.DeleteCoolMemberRequest{
		CoolCode:    ctx.Param("code"),
		CommunityId: ctx.Param("communityId"),
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	if err := clh.usecase.Cool.DeleteMemberByCode(ctx.Request().Context(), ctx.Get("userTypes").([]string), request); err != nil {
		return response.ErrorV2(ctx, err)
	}

	return response.SuccessV2(ctx, http.StatusOK, "User's current COOL information has been deleted", nil)
}

func (clh *CoolHandler) UpdateMemberByCode(ctx echo.Context) error {
	parameter := models.UpdateRoleMemberParameter{
		CoolCode:    ctx.Param("code"),
		CommunityId: ctx.Param("communityId"),
	}

	var request models.UpdateRoleMemberRequest
	if err := ctx.Bind(&request); err != nil {
		return response.ErrorV2(ctx, models.ErrorInvalidInput)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	coolData, err := clh.usecase.Cool.UpdateMember(ctx.Request().Context(), parameter, request, ctx.Get("userTypes").([]string))
	if err != nil {
		return response.ErrorV2(ctx, err)
	}

	return response.SuccessV2(ctx, http.StatusOK, "User's current COOL information has been updated", coolData)
}

func (clh *CoolHandler) GetCoolByCode(ctx echo.Context) error {
	cool, err := clh.usecase.Cool.GetByCode(ctx.Request().Context(), ctx.Param("code"))
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.SuccessV2(ctx, http.StatusOK, "", cool)
}
