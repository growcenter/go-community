package v2

import (
	"errors"
	"go-community/internal/config"
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/deliveries/http/middleware"
	"go-community/internal/models"
	"go-community/internal/pkg/authorization"
	"go-community/internal/pkg/validator"
	"go-community/internal/usecases"
	"net/http"

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
	endpointUserAuth.GET("/:eventCode/sessions", handler.GetAllSessionsByEventCode)

	endpointUserInternal := api.Group("/internal/events")
	endpointUserInternal.Use(middleware.UserMiddleware(c, u, []string{"event-internal-view"}))

	endpointCreateEvent := endpointUserInternal.Group("")
	endpointCreateEvent.Use(middleware.UserMiddleware(c, u, []string{"event-internal-create"}))
	endpointCreateEvent.POST("", handler.Create)
}

// Create godoc
// @Summary Create a new Event
// @Description Creates a new event along with its associated sessions and form questions.
// @Description The event can be of various categories (e.g., registration, announcement, internal-attendance).
// @Description It supports comprehensive configurations for images, organizers, access control, location (online/offline/hybrid), scheduling, recurrences, and notifications.
// @Description
// @Description **Important Notes:**
// @Description - `sessions` are required unless the category is `announcement`.
// @Description - `location` must match the event's location type (`online`, `offline`, or `hybrid`).
// @Description - `access` controls who can view or register for the event.
// @Tags events
// @Accept json
// @Produce json
// @Param request body models.CreateEventRequest true "Event creation payload containing core details, settings, sessions, and optional questions"
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Security BearerAuth
// @Success 201 {object} models.CreateEventResponse "Successfully created event with its sessions and form questions"
// @Failure 400 {object} models.ErrorResponse "Bad Request - Invalid payload format or missing required fields"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - Invalid or missing authentication token"
// @Failure 403 {object} models.ErrorResponse "Forbidden - User does not have the required permissions"
// @Failure 422 {object} models.ErrorResponse{errors=models.ErrorValidateResponse} "Unprocessable Entity - Validation errors (e.g., invalid location type, missing required linked fields)"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error"
// @Router /v2/internal/events [post]
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

// CreateSession godoc
// @Summary Create a new Event Session
// @Description Creates a new event session for a specific event. This can be a standalone session, or a child session (e.g., a breakout room or a specific track) if `parentSessionCode` is provided.
// @Description
// @Description **Inheritance & Defaults:**
// @Description - If location details are omitted, the session inherits the location from the parent event (or parent session).
// @Description - If status is omitted, it defaults to the event's status.
// @Description - Timezone falls back to the event's timezone.
// @Description
// @Description **Registration & Capacity:**
// @Description - `sessionCapacity.capacity`: Set to 0 for unlimited. Waitlist can only be enabled if capacity > 0 and registration method is not QR-based walk-in.
// @Description - `sessionRules.registrationMode`: Supported modes are `self_only`, `self_and_registered`, and `self_and_others`.
// @Description - `sessionRules.registrationMethods`: Options include `personal-qr`, `event-qr`, `session-qr`, `registration-qr`.
// @Description - Supports age limits (`minAge`, `maxAge`) and prerequisites.
// @Description
// @Description **Check-in & Check-out:**
// @Description - `checkIn` is enabled by default. You can configure `allowLate` and `lateThreshold` with specific late policies (`reject`, `warn`, `allow`).
// @Description - Validations ensure that check-in/check-out times align with the session schedule.
// @Tags events
// @Accept json
// @Produce json
// @Param eventCode path string true "Event Code"
// @Param request body models.CreateEventSessionRequest true "Event session creation payload containing core details, settings, and optional questions"
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Security BearerAuth
// @Success 201 {object} models.Response{data=models.CreateEventSessionResponse} "Successfully created event session"
// @Failure 400 {object} models.ErrorResponse "Bad Request - Invalid payload format or missing required fields"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - Invalid or missing authentication token"
// @Failure 403 {object} models.ErrorResponse "Forbidden - User does not have the required permissions"
// @Failure 422 {object} models.ErrorResponse{errors=models.ErrorValidateResponse} "Unprocessable Entity - Validation errors"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error"
// @Router /v2/internal/events/{eventCode}/sessions [post]
func (eh *EventHandler) CreateSession(ctx echo.Context) error {
	var request models.CreateEventSessionRequest
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, err)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	eventCode := ctx.Param("eventCode")
	if eventCode == "" {
		return response.Error(ctx, errors.New("event code is required"))
	}

	var eventCodePtr *string
	if eventCode != "" {
		eventCodePtr = &eventCode
	}

	eventSession, err := eh.usecase.EventSession.Create(ctx.Request().Context(), []models.CreateEventSessionRequest{request}, eventCodePtr, nil)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.SuccessListV2(ctx, http.StatusCreated, "Event Session is created successfully.", eventSession.ToResponse())
}

// GetAllSessionsByEventCode godoc
// @Summary Get All Sessions By Event Code
// @Description Fetch all event sessions associated with a specific event code
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param eventCode path string true "Event Code"
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Success 200 {object} models.List{data=[]models.EventSession} "Successfully fetched event sessions"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorResponse{errors=models.ErrorValidateResponse} "Validation error."
// @Router /v2/events/{eventCode}/sessions [get]
func (eh *EventHandler) GetAllSessionsByEventCode(ctx echo.Context) error {
	events, err := eh.usecase.EventSession.GetByEventCode(ctx.Request().Context(), ctx.Param("eventCode"))
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.SuccessList(ctx, http.StatusOK, len(events), events)
}

func (eh *EventHandler) GetAll(ctx echo.Context) error {
	events, err := eh.usecase.Event.GetAll(ctx.Request().Context())
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.SuccessList(ctx, http.StatusOK, len(events), events)
}
