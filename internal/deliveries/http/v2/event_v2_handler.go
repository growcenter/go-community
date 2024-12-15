package v2

import (
	"github.com/labstack/echo/v4"
	"go-community/internal/config"
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/deliveries/http/middleware"
	"go-community/internal/models"
	"go-community/internal/pkg/authorization"
	"go-community/internal/pkg/validator"
	"go-community/internal/usecases"
	"net/http"
)

type EventHandler struct {
	usecase *usecases.Usecases
}

func NewEventHandler(api *echo.Group, u *usecases.Usecases, c *config.Configuration, a *authorization.Auth) {
	handler := &EventHandler{usecase: u}

	// Define campus routes
	endpoint := api.Group("/events")
	endpoint.POST("", handler.Create)
	endpointUserAuth := endpoint.Group("")
	endpointUserAuth.Use(middleware.UserV2Middleware(c))
	endpointUserAuth.GET("", handler.GetAll)
	endpointUserAuth.GET("/:code", handler.GetByCode)
	endpointUserAuth.POST("/registers", handler.Register)
	endpointUserAuth.GET("/registers", handler.GetAllRegistered)
	endpointUserAuth.PATCH("/registers/:id/status", handler.UpdateStatus)
}

// Create godoc
// @Summary Create Event
// @Description Create event with the instances/sessions
// @Tags events
// @Accept json
// @Produce json
// @Param user body models.CreateEventRequest true "User object that needs to be added"
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Security BearerAuth
// @Success 201 {object} models.CreateEventResponse{instances=models.CreateInstanceResponse} "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorValidationResponse{errors=validator.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v2/events [post]
func (eh *EventHandler) Create(ctx echo.Context) error {
	var request models.CreateEventRequest
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, err)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	event, err := eh.usecase.Event.Create(ctx.Request().Context(), request)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusCreated, event.ToResponse())
}

// GetAll godoc
// @Summary Get All Events
// @Description Get All Events based on User Roles
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Success 200 {object} models.List{data=[]models.GetAllEventsResponse} "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorValidationResponse{errors=validator.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v2/events [get]
func (eh *EventHandler) GetAll(ctx echo.Context) error {
	events, err := eh.usecase.Event.GetAll(ctx.Request().Context(), ctx.Get("roles").([]string), ctx.Get("userTypes").([]string))
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.SuccessList(ctx, http.StatusOK, len(*events), events)
}

// GetByCode godoc
// @Summary Get Event by Event Code
// @Description Get Event and Instances by Event Code
// @Tags events
// @Accept json
// @Produce json
// @Param code path int true "object that needs to be added"
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Security BearerAuth
// @Success 200 {object} models.GetEventByCodeResponse{instances=[]models.GetInstancesByEventCodeResponse} "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorValidationResponse{errors=validator.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v2/events/{code} [get]
func (eh *EventHandler) GetByCode(ctx echo.Context) error {
	parameter := models.GetEventByCodeParameter{
		Code: ctx.Param("code"),
	}

	if err := validator.Validate(parameter); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	events, err := eh.usecase.Event.GetByCode(ctx.Request().Context(), parameter.Code, ctx.Get("roles").([]string), ctx.Get("userTypes").([]string))
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusOK, events.ToResponse())
}

// Register godoc
// @Summary Register User to Event
// @Description Register user to particular event and instances
// @Tags events
// @Accept json
// @Produce json
// @Param user body models.CreateEventRegistrationRecordRequest true "User object that needs to be added"
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Security BearerAuth
// @Success 201 {object} models.CreateEventRegistrationRecordResponse{registrants=models.CreateOtherEventRegistrationRecordRequest} "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorValidationResponse{errors=validator.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v2/events/registers [post]
func (eh *EventHandler) Register(ctx echo.Context) error {
	var request models.CreateEventRegistrationRecordRequest
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, err)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	tokenValue, err := models.GetValueFromToken(ctx)
	if err != nil {
		return response.Error(ctx, err)
	}

	register, err := eh.usecase.EventRegistrationRecord.Create(ctx.Request().Context(), &request, &tokenValue)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusCreated, register.ToResponse())
}

// GetAllRegistered godoc
// @Summary Get All User's Registered Event
// @Description Get All User's Registered Event
// @Tags events
// @Accept json
// @Produce json
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Security BearerAuth
// @Success 200 {object} models.GetAllRegisteredUserResponse{instances=[]models.InstancesForRegisteredRecordsResponse} "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorValidationResponse{errors=validator.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v2/events/registers [get]
func (eh *EventHandler) GetAllRegistered(ctx echo.Context) error {
	parameter := models.GetAllRegisteredUserParameter{
		CommunityId: ctx.Get("communityId").(string),
	}

	if err := validator.Validate(parameter); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	res, err := eh.usecase.Event.GetRegistered(ctx.Request().Context(), parameter.CommunityId)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.SuccessList(ctx, http.StatusOK, len(res), res)
}

// UpdateStatus godoc
// @Summary Update Registration Status
// @Description Update user registration id to success or failed
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "registration id"
// @Param user body models.UpdateRegistrationStatusRequest true "User object that needs to be added"
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Success 201 {object} models.UpdateRegistrationStatusResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorValidationResponse{errors=validator.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v2/events/registers/{id}/status [patch]
func (eh *EventHandler) UpdateStatus(ctx echo.Context) error {
	requestParam := models.UpdateRegistrationStatusParameter{
		ID: ctx.Param("id"),
	}

	if err := validator.Validate(requestParam); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	var requestBody models.UpdateRegistrationStatusRequest
	if err := ctx.Bind(&requestBody); err != nil {
		return response.Error(ctx, err)
	}

	if err := validator.Validate(requestBody); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	tokenValue, err := models.GetValueFromToken(ctx)
	if err != nil {
		return response.Error(ctx, err)
	}

	record, err := eh.usecase.EventRegistrationRecord.UpdateStatus(ctx.Request().Context(), &requestParam, &requestBody, &tokenValue)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusOK, record.ToResponse())
}
