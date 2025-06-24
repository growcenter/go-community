package v2

import (
	"go-community/internal/config"
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/deliveries/http/middleware"
	"go-community/internal/models"
	"go-community/internal/pkg/validator"
	"go-community/internal/usecases"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type CoolAttendanceHandler struct {
	usecase *usecases.Usecases
	conf    *config.Configuration
}

func NewCoolAttendanceHandler(api *echo.Group, u *usecases.Usecases, c *config.Configuration) {
	handler := &CoolAttendanceHandler{usecase: u, conf: c}

	endpoint := api.Group("/cools")
	// endpointInternal := api.Group("/internal/cools")

	endpointCoreAuth := endpoint.Group("")
	endpointCoreAuth.Use(middleware.UserMiddleware(c, u, []string{"cool-attendance-create"}))
	endpointCoreAuth.GET("/:coolCode/attendances", handler.GetSummaryByCoolCode)
	endpointCoreAuth.POST("/meetings", handler.CreateMeeting)
	endpointCoreAuth.POST("/meetings/attendances", handler.CreateAttendance)
	endpointCoreAuth.GET("/meetings/:id/attendances", handler.GetMeetingAttendances)

	endpointMemberAuth := endpoint.Group("")
	endpointMemberAuth.Use(middleware.UserMiddleware(c, u, nil))
	endpointMemberAuth.GET("/meetings", handler.GetMeetings)
}

// Create godoc
// @Summary Create Meeting
// @Description Create Cool Meetings
// @Tags cools-meetings
// @Accept json
// @Produce json
// @Param user body models.CreateMeetingRequest true "Create Meeting Request JSON"
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Success 201 {object} models.Response{data=models.CreateMeetingResponse,metadata=models.Metadata} "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorResponse{errors=models.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v2/cools/meetings [post]
func (cah *CoolAttendanceHandler) CreateMeeting(ctx echo.Context) error {
	var request models.CreateMeetingRequest
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, models.ErrorInvalidInput)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	new, err := cah.usecase.CoolMeeting.Create(ctx.Request().Context(), request)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.SuccessV2(ctx, http.StatusCreated, "Meeting successfully created.", new.ToResponse())
}

func (cah *CoolAttendanceHandler) CreateAttendance(ctx echo.Context) error {
	var request models.CreateAttendanceRequest
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, models.ErrorInvalidInput)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	new, err := cah.usecase.CoolAttendance.Create(ctx.Request().Context(), request)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.SuccessV2(ctx, http.StatusCreated, "Attendance successfully created.", new.ToResponse())
}

func (cah *CoolAttendanceHandler) GetMeetings(ctx echo.Context) error {
	var param models.GetPreviousUpcomingMeetingsParameter
	if err := ctx.Bind(&param); err != nil {
		return response.Error(ctx, models.ErrorInvalidInput)
	}

	if err := validator.Validate(param); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	// Declare meetings variable outside the switch statement
	var meetings interface{}
	var err error

	switch param.Type {
	case "upcoming":
		meetings, err = cah.usecase.CoolMeeting.GetUpcomingMeetings(ctx.Request().Context(), param)
		if err != nil {
			return response.Error(ctx, err)
		}
	case "previous":
		tokenValue, err := models.GetValueFromToken(ctx)
		if err != nil {
			return response.Error(ctx, err)
		}

		meetings, err = cah.usecase.CoolMeeting.GetPreviousMeetings(ctx.Request().Context(), tokenValue, param)
		if err != nil {
			return response.Error(ctx, err)
		}
	default:
		return response.Error(ctx, models.ErrorInvalidInput)
	}

	return response.SuccessListV2(ctx, http.StatusOK, "", meetings)
}

func (cah *CoolAttendanceHandler) GetMeetingAttendances(ctx echo.Context) error {
	request := models.GetAllAttendanceByMeetingIdRequest{
		MeetingId: uuid.MustParse(ctx.Param("id")),
	}

	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, err)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	attendances, err := cah.usecase.CoolAttendance.GetByMeetingId(ctx.Request().Context(), request.MeetingId)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.SuccessV2(ctx, http.StatusOK, "", attendances)
}

func (cah *CoolAttendanceHandler) GetSummaryByCoolCode(ctx echo.Context) error {
	request := models.GetSummaryAttendanceByCoolCodeRequest{
		CoolCode:  ctx.Param("coolCode"),
		StartDate: ctx.QueryParam("startDate"),
		EndDate:   ctx.QueryParam("endDate"),
	}

	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, err)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	summary, err := cah.usecase.CoolAttendance.GetSummaryByCoolCode(ctx.Request().Context(), request)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.SuccessListV2(ctx, http.StatusOK, "", summary)
}
