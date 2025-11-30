package v2

import (
	"fmt"
	"go-community/internal/config"
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/deliveries/http/middleware"
	"go-community/internal/models"
	"go-community/internal/pkg/authorization"
	"go-community/internal/pkg/errorgen"
	"go-community/internal/pkg/validator"
	"go-community/internal/usecases"
	"net/http"

	"strings"

	"github.com/labstack/echo/v4"
)

type EventHandler struct {
	usecase *usecases.Usecases
}

func NewEventHandler(api *echo.Group, u *usecases.Usecases, c *config.Configuration, a *authorization.Auth) {
	handler := &EventHandler{usecase: u}

	// Define campus routes
	endpoint := api.Group("/events")

	endpointUserAuth := endpoint.Group("")
	endpointUserAuth.Use(middleware.UserMiddleware(c, u, nil))
	endpointUserAuth.GET("", handler.GetAll)
	endpointUserAuth.GET("/:code", handler.GetByCode)
	endpointUserAuth.GET("/:instanceCode/questions", handler.GetQuestions)
	endpointUserAuth.POST("/registers", handler.Register)
	endpointUserAuth.GET("/registers", handler.GetAllRegistered)
	// endpointUserAuth.PATCH("/registers/:id/status", handler.UpdateStatus)
	// endpointUserAuth.GET("/attendance", handler.GetEventAttendance)

	endpointUserInternal := api.Group("/internal/events")
	endpointUserInternal.Use(middleware.UserMiddleware(c, u, []string{"event-internal-view", "event-internal-edit"}))
	endpointUserInternal.POST("", handler.Create)
	endpointUserInternal.GET("", handler.GetTitles)
	endpointUserInternal.GET("/:eventCode/summary", handler.GetSummary)
	// endpointUserInternal.GET("/registers", handler.GetAllRegisteredInternal)
	// endpointUserInternal.POST("/instances", handler.CreateInstance)
	// endpointUserInternal.GET("/registers/download", handler.DownloadInternal)
}

// Create godoc
// @Summary Create Event
// @Description Create event with the instances/sessions
// @Tags events-internal
// @Accept json
// @Produce json
// @Security BearerAuth header string true "Bearer token"
// @Param Authorization header string true "Bearer token"
// @Param event body models.CreateEventRequest true "Event creation request body"
// @Security ApiKeyAuth
// @Success 201 {object} models.CreateEventResponse "Response indicates that the request succeeded and the event has been created."
// @Failure 400 {object} models.ErrorResponse "Bad Request. Can be caused by invalid time ranges, or missing fields for private events."
// @Failure 404 {object} models.ErrorResponse "Not Found. Can be caused by non-existent roles, users, campuses, etc."
// @Failure 409 {object} models.ErrorResponse "Conflict. The event slug or a generated code already exists."
// @Failure 422 {object} models.ErrorResponse{errors=models.ErrorValidateResponse} "Validation Error. The request body is invalid."
// @Failure 500 {object} models.ErrorResponse "Internal Server Error. Something went wrong on the server."
// @Router /api/v2/internal/events [post]
func (eh *EventHandler) Create(ctx echo.Context) error {
	var request models.CreateEventRequest
	if err := ctx.Bind(&request); err != nil {
		// Check for a specific JSON unmarshal error to provide a better message.
		if httpErr, ok := err.(*echo.HTTPError); ok && httpErr.Code == http.StatusBadRequest {
			return response.ErrorV2(ctx, errorgen.Error(errorgen.InvalidInput, "Invalid request body: expected a JSON object but received a different format. Please check the client request."))
		}
		return response.ErrorV2(ctx, errorgen.Error(errorgen.InvalidInput, err.Error()))
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	event, err := eh.usecase.Event.Create(ctx.Request().Context(), request, ctx.Get("id").(string))
	if err != nil {
		return response.ErrorV2(ctx, err)
	}

	return response.SuccessV2(ctx, http.StatusCreated, event.ToResponse())

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
// @Failure 422 {object} models.ErrorResponse{errors=models.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v2/events [get]
func (eh *EventHandler) GetAll(ctx echo.Context) error {
	// events, err := eh.usecase.Event.GetAll(ctx.Request().Context(), ctx.Get("roles").([]string), ctx.Get("userTypes").([]string))
	// if err != nil {
	// 	return response.Error(ctx, err)
	// }

	parameter := models.GetAllEventsParams{
		AllowedUserTypes:    ctx.Get("userTypes").([]string),
		AllowedRoles:        ctx.Get("roles").([]string),
		AllowedCommunityIDs: ctx.Get("communityIDs").([]string),
		Status:              ctx.QueryParam("status"),
		Title:               ctx.QueryParam("title"),
		Visibility:          ctx.QueryParam("visibility"),
	}

	campusesParam := ctx.QueryParam("campuses")
	if campusesParam != "" {
		parameter.AllowedCampuses = strings.Split(campusesParam, ",")
	}

	if err := validator.Validate(parameter); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	events, err := eh.usecase.Event.GetAllV2(ctx.Request().Context(), &parameter)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.SuccessListV2(ctx, http.StatusOK, "", events)
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
// @Success 200 {object} models.ListWithDetail{details=models.GetEventByCodeResponse,data=[]models.GetInstancesByEventCodeResponse} "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorResponse{errors=models.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Failure 404 {object} models.ErrorResponse "Not Found. Can be caused by non-existent event code."
// @Router /v2/events/{code} [get]
func (eh *EventHandler) GetByCode(ctx echo.Context) error {
	parameter := models.GetEventByCodeParameter{
		Code: ctx.Param("code"),
	}

	if err := validator.Validate(parameter); err != nil {
		fmt.Println("error validation", err)
		return response.ErrorValidation(ctx, err)
	}

	tokenValue, err := models.GetValueFromToken(ctx)
	if err != nil {
		fmt.Println("error get value from token", err)
		return response.ErrorV2(ctx, err)
	}

	detail, data, err := eh.usecase.Event.GetByCode(ctx.Request().Context(), &parameter, &tokenValue)
	if err != nil {
		fmt.Println("error get event by code", err)
		return response.ErrorV2(ctx, err)
	}

	return response.SuccessListWithDetail(ctx, http.StatusOK, len(data), detail, data)
}

func (eh *EventHandler) GetQuestions(ctx echo.Context) error {
	parameter := models.GetEventQuestionParameter{
		InstanceCode: ctx.Param("instanceCode"),
	}

	if err := validator.Validate(parameter); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	questions, err := eh.usecase.Event.GetQuestions(ctx.Request().Context(), parameter)
	if err != nil {
		return response.ErrorV2(ctx, err)
	}

	return response.SuccessV2(ctx, http.StatusOK, questions)
}

// Register godoc
// @Summary Register User to Event
// @Description This endpoint handles the registration of one or more attendees to a specific event instance. It supports different registration methods like QR code scanning or direct registration. The endpoint validates event status, instance availability, user permissions for private events, and registration quotas before creating the registration records.
// @Tags events
// @Accept json
// @Produce json
// @Param registrationRequest body models.CreateEventRegistrationRequest true "The registration request payload, containing event/instance codes, registrant details, and a list of attendees with their information and form answers."
// @Security BearerAuth
// @Security ApiKeyAuth
// @Success 201 {object} models.CreateEventRegistrationResponse "Successfully created the registration. The response includes the registration code and details of all attendees."
// @Failure 400 {object} models.ErrorResponse "Bad Request. Possible causes: Mismatched event/instance codes, missing primary attendee, invalid registration method, or registration window is closed/not yet open."
// @Failure 403 {object} models.ErrorResponse "Forbidden. The user does not have the required roles or permissions to register for this private event."
// @Failure 404 {object} models.ErrorResponse "Not Found. The specified event, instance, or user does not exist."
// @Failure 409 {object} models.ErrorResponse "Conflict. The user has already registered for this event, or one of the attendees is already registered (if uniqueness is enforced)."
// @Failure 422 {object} models.ErrorResponse{errors=models.ErrorValidateResponse} "Unprocessable Entity. The request body contains validation errors, such as invalid email/phone format or missing required fields."
// @Failure 429 {object} models.ErrorResponse "Too Many Requests. The user has exceeded the registration quota for this event instance."
// @Failure 500 {object} models.ErrorResponse "Internal Server Error. An unexpected error occurred on the server."
// @Router /api/v2/events/registers [post]
func (eh *EventHandler) Register(ctx echo.Context) error {
	var request models.CreateEventRegistrationRequest
	if err := ctx.Bind(&request); err != nil {
		return response.ErrorV2(ctx, err)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	tokenValue, err := models.GetValueFromToken(ctx)
	if err != nil {
		return response.ErrorV2(ctx, err)
	}

	register, err := eh.usecase.EventRegistration.Create(ctx.Request().Context(), request, tokenValue)
	if err != nil {
		return response.ErrorV2(ctx, err)
	}

	return response.SuccessV2(ctx, http.StatusCreated, register.ToResponse())
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
// @Failure 422 {object} models.ErrorResponse{errors=models.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v2/events/registers [get]
func (eh *EventHandler) GetAllRegistered(ctx echo.Context) error {
	parameter := models.GetAllRegisteredUserParameter{
		CommunityId: ctx.Get("id").(string),
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
// @Failure 422 {object} models.ErrorResponse{errors=models.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v2/events/registers/{id}/status [patch]
// func (eh *EventHandler) UpdateStatus(ctx echo.Context) error {
// 	requestParam := models.UpdateRegistrationStatusParameter{
// 		ID: ctx.Param("id"),
// 	}

// 	if err := validator.Validate(requestParam); err != nil {
// 		return response.ErrorValidation(ctx, err)
// 	}

// 	var requestBody models.UpdateRegistrationStatusRequest
// 	if err := ctx.Bind(&requestBody); err != nil {
// 		return response.Error(ctx, err)
// 	}

// 	if err := validator.Validate(requestBody); err != nil {
// 		return response.ErrorValidation(ctx, err)
// 	}

// 	tokenValue, err := models.GetValueFromToken(ctx)
// 	if err != nil {
// 		return response.Error(ctx, err)
// 	}

// 	record, err := eh.usecase.EventRegistrationRecord.UpdateStatus(ctx.Request().Context(), &requestParam, &requestBody, &tokenValue)
// 	if err != nil {
// 		return response.Error(ctx, err)
// 	}

// 	return response.Success(ctx, http.StatusOK, record.ToResponse())
// }

// GetTitles godoc
// @Summary Get Events Titles
// @Description For Internal Purposes Only
// @Tags events
// @Accept json
// @Produce json
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Security BearerAuth
// @Success 200 {object} models.List{data=[]models.GetEventTitlesResponse} "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorResponse{errors=models.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v2/events/registers [get]
func (eh *EventHandler) GetTitles(ctx echo.Context) error {
	res, err := eh.usecase.Event.GetTitles(ctx.Request().Context())
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.SuccessList(ctx, http.StatusOK, len(res), res)
}

// GetSummary godoc
// @Summary Get Event and Sessions by Event Code
// @Description For Internal Purposes Only
// @Tags events
// @Accept json
// @Produce json
// @Param code path int true "object that needs to be added"
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Security BearerAuth
// @Success 200 {object} models.ListWithDetail{details=models.GetEventSummaryResponse,data=[]models.GetInstanceSummaryResponse} "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorResponse{errors=models.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v2/internal/events/{eventCode}/summary [get]
func (eh *EventHandler) GetSummary(ctx echo.Context) error {
	detail, data, err := eh.usecase.Event.GetSummary(ctx.Request().Context(), ctx.Param("eventCode"))
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.SuccessListWithDetail(ctx, http.StatusOK, len(data), detail, data)
}

// CreateInstance godoc
// @Summary Create Instance
// @Description Create instance from existing Event
// @Tags events
// @Accept json
// @Produce json
// @Param user body models.CreateInstanceExistingEventRequest true "User object that needs to be added"
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Security BearerAuth
// @Success 201 {object} models.CreateInstanceResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorResponse{errors=models.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v2/internal/events/instances [post]
// func (eh *EventHandler) CreateInstance(ctx echo.Context) error {
// 	var request models.CreateInstanceExistingEventRequest
// 	if err := ctx.Bind(&request); err != nil {
// 		return response.Error(ctx, err)
// 	}

// 	if err := validator.Validate(request); err != nil {
// 		return response.ErrorValidation(ctx, err)
// 	}

// 	instance, err := eh.usecase.EventInstance.Create(ctx.Request().Context(), request)
// 	if err != nil {
// 		return response.Error(ctx, err)
// 	}

// 	return response.Success(ctx, http.StatusCreated, instance.ToResponse())
// }

// func (eh *EventHandler) GetEventAttendance(ctx echo.Context) error {
// 	var request models.GetEventAttendanceParameter
// 	if err := ctx.Bind(&request); err != nil {
// 		return response.Error(ctx, err)
// 	}

// 	if err := validator.Validate(request); err != nil {
// 		return response.ErrorValidation(ctx, err)
// 	}

// 	detail, list, err := eh.usecase.EventRegistrationRecord.GetAttendance(ctx.Request().Context(), request)
// 	if err != nil {
// 		return response.Error(ctx, err)
// 	}

// 	return response.SuccessListWithDetail(ctx, http.StatusOK, len(list), detail, list)
// }

//func (eh *EventHandler) GetAllRegisteredInternal(ctx echo.Context) error {
//	var param models.GetAllRegisteredCursorParam
//	if err := ctx.Bind(&param); err != nil {
//		return response.Error(ctx, err)
//	}
//
//	if err := validator.Validate(param); err != nil {
//		return response.ErrorValidation(ctx, err)
//	}
//
//	data, info, err := eh.usecase.EventRegistrationRecord.GetAllCursor(ctx.Request().Context(), param)
//	if err != nil {
//		return response.Error(ctx, err)
//	}
//
//	return response.SuccessCursor(ctx, http.StatusOK, info, data)
//}

//func (eh *EventHandler) GetAllRegisteredInternal(ctx echo.Context) error {
//	var param models.GetAllRegisteredCursorParam
//	if err := ctx.Bind(&param); err != nil {
//		return response.Error(ctx, err)
//	}
//
//	if err := validator.Validate(param); err != nil {
//		return response.ErrorValidation(ctx, err)
//	}
//
//	data, info, err := eh.usecase.EventRegistrationRecord.GetAllCursor(ctx.Request().Context(), param)
//	if err != nil {
//		return response.Error(ctx, err)
//	}
//
//	//return response.SuccessCursor[models.GetAllRegisteredCursorResponse](ctx, data, 0, 0, nil)
//	return response.SuccessCursor(ctx, http.StatusOK, info, data)
//}

// func (eh *EventHandler) DownloadInternal(ctx echo.Context) error {
// 	var param models.GetDownloadAllRegisteredParam
// 	if err := ctx.Bind(&param); err != nil {
// 		return response.Error(ctx, err)
// 	}

// 	if err := validator.Validate(param); err != nil {
// 		return response.ErrorValidation(ctx, err)
// 	}

// 	data, contentType, fileName, err := eh.usecase.EventRegistrationRecord.Download(ctx.Request().Context(), param)
// 	if err != nil {
// 		return response.Error(ctx, err)
// 	}

// 	return response.SuccessDownload(ctx, http.StatusOK, contentType, fileName, data)
// }
