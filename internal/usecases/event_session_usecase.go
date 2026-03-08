package usecases

import (
	"context"
	"time"

	"go-community/internal/common"
	"go-community/internal/constants"
	"go-community/internal/models"
	"go-community/internal/pkg/errorc"
	"go-community/internal/pkg/generator"
	"go-community/internal/pkg/logger"
	"go-community/internal/pkg/stringc"
)

type EventSessionUsecase interface {
	// Create validates and persists sessions, using the default repository set.
	// When called inside an Atomic callback the transactional DB is embedded in ctx
	// by Atomic and picked up transparently by each repository — no special variant needed.
	Create(ctx context.Context, requests []models.CreateEventSessionRequest, eventCode *string, event *models.Event) (*models.EventSession, error)

	GetByEventCode(ctx context.Context, eventCode string) ([]models.EventSession, error)
}

type eventSessionUsecase struct {
	d *Dependencies
}

func NewEventSessionUsecase(d *Dependencies) EventSessionUsecase {
	return &eventSessionUsecase{d: d}
}

// Create validates and persists sessions using the usecase's own repository set.
// Suitable for direct (non-atomic) calls from the API layer.
func (es *eventSessionUsecase) Create(ctx context.Context, requests []models.CreateEventSessionRequest, eventCode *string, event *models.Event) (*models.EventSession, error) {
	logger.Add(ctx, "session", map[string]any{
		"operation": "create",
	})

	event, err := es.resolveAndValidateEvent(ctx, event, eventCode)
	if err != nil {
		return nil, err
	}

	var created *models.EventSession

	for i := range requests {
		req := &requests[i]

		// ── 1. Location inheritance: Event → Session ──────────────────────────
		es.inheritLocationFromEvent(req, event)

		// ── 2. Location inheritance: Parent Session → Child Session ───────────
		if req.ParentSessionCode != "" {
			parentSession, err := es.d.Repository.EventSession.GetByCode(ctx, req.ParentSessionCode)
			if err != nil {
				return nil, errorc.Error(err)
			}
			if parentSession == nil {
				return nil, errorc.Error(errorc.ErrorDataNotFound, "parent session '%s' not found", req.ParentSessionCode)
			}
			es.inheritLocationFromSession(req, parentSession)
		}

		// ── 3. Location validation ─────────────────────────────────────────────
		if err := req.Location.Validate(nil); err != nil {
			return nil, errorc.Error(err)
		}

		// ── 4. Geolocation config validation ────────────────────────────────────
		if req.Geolocation != nil && req.Geolocation.Enabled {
			if err := req.Geolocation.Validate(); err != nil {
				return nil, errorc.Error(err)
			}
		}

		// ── 5. Normalize defaults ────────────────────────────────────────────────
		normalizeSessionRequest(req, event)

		// ── 6. Schedule validation ────────────────────────────────────────────────
		if err := validateSchedule(&req.Schedule); err != nil {
			logger.AddError(ctx, &logger.ErrorContext{
				Type: "ValidationFailed", Code: "SESSION_SCHEDULE_INVALID",
				Message: err.Error(), Retriable: false,
			})
			return nil, errorc.Error(err)
		}

		// ── 7. Time windows (registration, check-in, check-out) ───────────────
		if err := validateSessionTimeConfiguration(&req.Times); err != nil {
			logger.AddError(ctx, &logger.ErrorContext{
				Type: "ValidationFailed", Code: "SESSION_TIME_CONFIG_INVALID",
				Message: err.Error(), Retriable: false,
			})
			return nil, errorc.Error(err)
		}

		// ── 8. Capacity ───────────────────────────────────────────────────────────
		if err := validateSessionCapacity(&req.SessionCapacity, &req.SessionRules); err != nil {
			logger.AddError(ctx, &logger.ErrorContext{
				Type: "ValidationFailed", Code: "SESSION_CAPACITY_INVALID",
				Message: err.Error(), Retriable: false,
			})
			return nil, errorc.Error(err)
		}

		// ── 9. Rules ──────────────────────────────────────────────────────────────
		if err := validateSessionRules(req, event); err != nil {
			logger.AddError(ctx, &logger.ErrorContext{
				Type: "ValidationFailed", Code: "SESSION_RULES_INVALID",
				Message: err.Error(), Retriable: false,
			})
			return nil, errorc.Error(err)
		}

		// ── 11. Status gate ───────────────────────────────────────────────────────
		if event.Status != constants.EventStatusActive.String() && req.Status == constants.EventStatusActive.String() {
			return nil, errorc.Error(errorc.ErrorInvalidInput, "session cannot be active when the parent event is not active")
		}

		// ── 12. Generate unique session code ──────────────────────────────────────
		sessionCode, err := generateUniqueSessionCode(ctx, es, constants.SessionCodeMaxRetries)
		if err != nil {
			logger.AddError(ctx, &logger.ErrorContext{
				Type: "GenerateSessionCodeError", Code: "GENERATE_SESSION_CODE_ERROR",
				Message: err.Error(), Retriable: false,
			})
			return nil, errorc.Error(err)
		}

		// Seed the "session" group with identifiers known at this point.
		// One call, one mutex acquisition, zero redundancy.
		logger.Add(ctx, "session", map[string]any{
			"code":         *sessionCode,
			"event_code":   event.Code,
			"status":       req.Status,
			"session_type": req.SessionType,
		})

		// ── 13. Map request → model ───────────────────────────────────────────────
		session := buildSessionModel(req, event.Code, *sessionCode)

		if req.Geolocation != nil {
			if err := session.Geolocation.Marshal(req.Geolocation); err != nil {
				return nil, errorc.Error(errorc.ErrorInternalServer, "failed to marshal geolocation config")
			}
		}

		// ── 14. Persist ─────────────────────────────────────────────────────────
		// ctx carries the active Atomic transaction (if any); the repository picks
		// it up transparently via BaseRepository.db(ctx).
		if err := es.d.Repository.EventSession.Create(ctx, session); err != nil {
			logger.AddError(ctx, &logger.ErrorContext{
				Type: "DatabaseError", Code: "SESSION_CREATE_FAILED",
				Message: err.Error(), Retriable: true,
			})
			return nil, errorc.Error(err)
		}

		if req.Questions != nil {
			form := models.CreateFormRequest{
				Name: session.Title,
				Entity: models.FormEntityRequest{
					Type: "event_session",
					Code: session.Code,
				},
				Questions: req.Questions,
			}

			formResp, err := es.d.Form.Create(ctx, &form)
			if err != nil {
				logger.AddError(ctx, &logger.ErrorContext{
					Type:      "DatabaseError",
					Code:      "QUESTIONS_CREATE_FAILED",
					Message:   err.Error(),
					Retriable: true,
				})
				return nil, errorc.Error(err, "failed to create event questions: %s", err)
			}

			questionCodes := make([]string, len(formResp.Questions))
			for i, q := range formResp.Questions {
				questionCodes[i] = q.Code
			}

			// Group form creation details neatly under 'form'
			logger.AddToKey(ctx, "form", map[string]any{
				"code":              formResp.Code,
				"questions_created": len(formResp.Questions),
				"question_codes":    questionCodes,
			})
		}

		// ── 15. Optionally update parent event schedule ───────────────────────────
		if err := es.updateEvent(ctx, req, event); err != nil {
			logger.AddError(ctx, &logger.ErrorContext{
				Type: "DatabaseError", Code: "EVENT_UPDATE_FAILED",
				Message: err.Error(), Retriable: true,
			})
			return nil, errorc.Error(err)
		}

		created = session
	}

	logger.Add(ctx, "session", map[string]any{
		"sessions_created": len(requests),
	})

	return created, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// resolveAndValidateEvent ensures we have a valid event to work with.
func (es *eventSessionUsecase) resolveAndValidateEvent(ctx context.Context, event *models.Event, eventCode *string) (*models.Event, error) {
	if event == nil && eventCode == nil {
		return nil, errorc.Error(errorc.ErrorMissingFields)
	}

	if event != nil {
		if eventCode != nil {
			if stringc.UpperAndTrimSpace(event.Code) != stringc.UpperAndTrimSpace(*eventCode) {
				return nil, errorc.Error(errorc.ErrorInvalidInput, "event code mismatch")
			}
		}
		return event, nil
	}

	fetched, err := es.d.Repository.Event.GetByCode(ctx, *eventCode)
	if err != nil {
		return nil, err
	}
	if fetched == nil {
		return nil, errorc.Error(errorc.ErrorDataNotFound, "event '%s' not found", *eventCode)
	}
	return fetched, nil
}

// generateUniqueSessionCode generates a collision-free session code with exponential backoff.
func generateUniqueSessionCode(ctx context.Context, es *eventSessionUsecase, maxRetries int) (*string, error) {
	for attempt := 0; attempt < maxRetries; attempt++ {
		code, err := generator.IdentifierCode(ctx, es.d.Config.Event.EncodeCode, time.Now(), constants.SessionCodePrefix)
		if err != nil {
			return nil, errorc.Error(err, "failed to generate session code")
		}

		exists, err := es.d.Repository.EventSession.CheckByCode(ctx, code)
		if err != nil {
			return nil, errorc.Error(err, "failed to check session code uniqueness")
		}
		if !exists {
			logger.Add(ctx, "code_generation_attempts", attempt+1)
			return &code, nil
		}

		// Code collision — rare, but worth knowing about. Log under its own group
		// so it doesn't pollute top-level fields.
		logger.Add(ctx, "code_collision", map[string]any{
			"attempt": attempt + 1,
			"code":    code,
		})

		if attempt < maxRetries-1 {
			time.Sleep(time.Duration(es.d.Config.Event.Backoff.CodeGeneration) * time.Millisecond * time.Duration(attempt+1))
		}
	}

	return nil, errorc.Error(errorc.ErrorInternalServer, "failed to generate unique session code after %d attempts", maxRetries)
}

// normalizeSessionRequest applies sensible defaults to a session request before validation.
// Mirrors normalizeEventRequest in event_usecase.go.
func normalizeSessionRequest(request *models.CreateEventSessionRequest, event *models.Event) {
	// Status default: inherit event's status so sessions are created in the same state
	if request.Status == "" {
		request.Status = event.Status
	}

	if request.Status == "active" && event.Status == "draft" {
		request.Status = event.Status
	}

	// Timezone: fall back to event timezone, then system default
	if request.Schedule.Timezone == nil || *request.Schedule.Timezone == "" {
		tz := event.Timezone
		if tz == "" {
			tz = common.DefaultTimeZone
		}
		request.Schedule.Timezone = &tz
	}

	// Location visibility default
	if request.Location.LocationVisibility == nil || *request.Location.LocationVisibility == "" {
		v := string(constants.LocationVisibilityAll)
		request.Location.LocationVisibility = &v
	}

	// CTA defaults — mirror event usecase pattern
	if !request.Location.ClickToAction.TextNotEmpty() {
		request.Location.ClickToAction.Text = stringc.Pointer("Register Here!")
	}
	if !request.Location.ClickToAction.LinkNotEmpty() {
		request.Location.ClickToAction.Link = stringc.Pointer("NORMAL_FLOW")
	}

	// Check-in default per spec §4.2: enabled by default only when times are provided.
	// If the caller provides check-in times, honour their Enabled flag as-is.
	// If no check-in times are present, leave Enabled as false to avoid a
	// validation failure (validateCheckInConfig requires StartAt/EndAt when Enabled=true).
	if !request.Times.CheckIn.Enabled && !request.Times.CheckIn.Required {
		if request.Times.CheckIn.StartAt != nil && request.Times.CheckIn.EndAt != nil {
			// Times were provided but Enabled was not explicitly set — default to true.
			request.Times.CheckIn.Enabled = true
		}
		// If no times provided, leave Enabled=false so the session is created
		// without a check-in window (the spec default "enabled" only applies when
		// the operator actually configures a window).
	}

	// MaxRegistrationsPerUser default: 1
	if request.SessionRules.MaxRegistrationsPerUser == 0 {
		request.SessionRules.MaxRegistrationsPerUser = 1
	}
}

// validateSessionCapacity validates capacity settings against spec §3.4.
func validateSessionCapacity(capacity *models.SessionCapacity, rules *models.SessionRules) error {
	if capacity.Capacity < 0 {
		return errorc.Error(errorc.ErrorValidation, "capacity must be 0 (unlimited) or a positive number")
	}

	if capacity.WaitlistEnabled {
		if capacity.Capacity == 0 {
			return errorc.Error(errorc.ErrorValidation, "waitlist cannot be enabled for unlimited-capacity sessions")
		}
		if capacity.WaitlistCapacity < 0 {
			return errorc.Error(errorc.ErrorValidation, "waitlist capacity must be a positive number")
		}

		// Spec §3.4: waitlist is not meaningful when QR-based walk-in is the only method
		if stringc.OneOf(rules.RegistrationMethods, "personal-qr", "event-qr") {
			return errorc.Error(errorc.ErrorValidation, "waitlist cannot be enabled for 'personal-qr' or 'event-qr' registration methods")
		}
	} else {
		// Clear waitlist capacity when waitlist is disabled to avoid confusion
		capacity.WaitlistCapacity = 0
	}

	return nil
}

// validateSessionRules validates and coerces registration rules per spec §3.4.
func validateSessionRules(req *models.CreateEventSessionRequest, event *models.Event) error {
	rules := &req.SessionRules

	// Coerce MaxRegistrationsPerUser for self-only or personal-qr flows
	// (a user can only register themselves — 1 slot)
	if rules.RegistrationMode == "self_only" || stringc.OneOf(rules.RegistrationMethods, "personal-qr") {
		rules.MaxRegistrationsPerUser = 1
	}

	// For recurring events in self_only mode, default max to 1
	if rules.MaxRegistrationsPerUser == 0 && event.IsRecurring && rules.RegistrationMode == "self_only" {
		rules.MaxRegistrationsPerUser = 1
	}

	// Age range consistency
	if rules.MinAge > 0 && rules.MaxAge > 0 {
		if rules.MaxAge <= rules.MinAge {
			return errorc.Error(errorc.ErrorValidation, "max_age (%d) must be greater than min_age (%d)", rules.MaxAge, rules.MinAge)
		}
	}

	return nil
}

// buildSessionModel maps a validated CreateEventSessionRequest to a models.EventSession.
func buildSessionModel(req *models.CreateEventSessionRequest, eventCode, sessionCode string) *models.EventSession {
	session := &models.EventSession{
		Code:      sessionCode,
		EventCode: eventCode,
		ParentSessionCode: func() *string {
			if req.ParentSessionCode == "" {
				return nil
			}
			return stringc.Pointer(req.ParentSessionCode)
		}(),

		Title:       req.Title,
		Description: req.Description,
		SessionType: req.SessionType,
		Status:      req.Status,

		// Location — LocationType and LocationVisibility are guaranteed non-nil after
		// inheritance + normalization, so direct dereference is safe here.
		LocationType:       *req.Location.LocationType,
		PhysicalPlaceName:  req.Location.PhysicalPlaceName,
		PhysicalAddress:    req.Location.PhysicalAddress,
		VirtualLink:        req.Location.VirtualLink,
		VirtualPlatform:    req.Location.VirtualPlatform,
		LocationDetails:    req.Location.LocationDetails,
		LocationVisibility: *req.Location.LocationVisibility,
		CTAText:            req.Location.ClickToAction.Text,
		CTALink:            req.Location.ClickToAction.Link,

		// Schedule
		StartAt:  *req.Schedule.StartAt,
		EndAt:    *req.Schedule.EndAt,
		Timezone: *req.Schedule.Timezone,

		// Registration window
		RegistrationStartAt: req.Times.Registration.StartAt,
		RegistrationEndAt:   req.Times.Registration.EndAt,

		// Capacity
		Capacity:         req.SessionCapacity.Capacity,
		WaitlistEnabled:  req.SessionCapacity.WaitlistEnabled,
		WaitlistCapacity: req.SessionCapacity.WaitlistCapacity,

		// Rules
		RequireApproval:         req.SessionRules.RequireApproval,
		RegistrationMethods:     req.SessionRules.RegistrationMethods,
		RegistrationMode:        req.SessionRules.RegistrationMode,
		MaxRegistrationsPerUser: req.SessionRules.MaxRegistrationsPerUser,
		OneSessionPerEvent:      req.SessionRules.OneSessionPerEvent,
		MinAge:                  req.SessionRules.MinAge,
		MaxAge:                  req.SessionRules.MaxAge,
		Prerequisites:           req.SessionRules.Prerequisites,

		// Check-in
		CheckInEnabled:       req.Times.CheckIn.Enabled,
		CheckInRequired:      req.Times.CheckIn.Required,
		CheckInStartAt:       req.Times.CheckIn.StartAt,
		CheckInEndAt:         req.Times.CheckIn.EndAt,
		CheckInAllowLate:     req.Times.CheckIn.AllowLate,
		CheckInLateThreshold: req.Times.CheckIn.LateThreshold,

		// Check-out
		CheckOutEnabled:       req.Times.CheckOut.Enabled,
		CheckOutRequired:      req.Times.CheckOut.Required,
		CheckOutStartAt:       req.Times.CheckOut.StartAt,
		CheckOutEndAt:         req.Times.CheckOut.EndAt,
		CheckOutAllowLate:     req.Times.CheckOut.AllowLate,
		CheckOutLateThreshold: req.Times.CheckOut.LateThreshold,
	}

	return session
}

// ─────────────────────────────────────────────────────────────────────────────
// Location inheritance helpers
// ─────────────────────────────────────────────────────────────────────────────

// inheritLocationFromEvent applies parent event's location to session if not explicitly provided.
// First level of inheritance: Event → Session → Child Session.
func (es *eventSessionUsecase) inheritLocationFromEvent(req *models.CreateEventSessionRequest, event *models.Event) {
	if req.Location.LocationType == nil || *req.Location.LocationType == "" {
		req.Location.LocationType = &event.LocationType
	}
	if req.Location.PhysicalPlaceName == nil && event.PhysicalPlaceName != nil {
		req.Location.PhysicalPlaceName = event.PhysicalPlaceName
	}
	if req.Location.PhysicalAddress == nil && event.PhysicalAddress != nil {
		req.Location.PhysicalAddress = event.PhysicalAddress
	}
	if req.Location.VirtualLink == nil && event.VirtualLink != nil {
		req.Location.VirtualLink = event.VirtualLink
	}
	if req.Location.VirtualPlatform == nil && event.VirtualPlatform != nil {
		req.Location.VirtualPlatform = event.VirtualPlatform
	}
	if req.Location.LocationDetails == nil && event.LocationDetails != nil {
		req.Location.LocationDetails = event.LocationDetails
	}
	if req.Location.LocationVisibility == nil || *req.Location.LocationVisibility == "" {
		req.Location.LocationVisibility = &event.LocationVisibility
	}
	if req.Location.ClickToAction.Text == nil && event.CTAText != nil {
		req.Location.ClickToAction.Text = event.CTAText
	}
	if req.Location.ClickToAction.Link == nil && event.CTALink != nil {
		req.Location.ClickToAction.Link = event.CTALink
	}
}

// inheritLocationFromSession applies parent session's location to child session if not explicitly provided.
// Second level of inheritance: Session → Child Session (Track → Breakout).
func (es *eventSessionUsecase) inheritLocationFromSession(req *models.CreateEventSessionRequest, parent *models.EventSession) {
	if req.Location.LocationType == nil || *req.Location.LocationType == "" {
		req.Location.LocationType = &parent.LocationType
	}
	if req.Location.PhysicalPlaceName == nil && parent.PhysicalPlaceName != nil {
		req.Location.PhysicalPlaceName = parent.PhysicalPlaceName
	}
	if req.Location.PhysicalAddress == nil && parent.PhysicalAddress != nil {
		req.Location.PhysicalAddress = parent.PhysicalAddress
	}
	if req.Location.VirtualLink == nil && parent.VirtualLink != nil {
		req.Location.VirtualLink = parent.VirtualLink
	}
	if req.Location.VirtualPlatform == nil && parent.VirtualPlatform != nil {
		req.Location.VirtualPlatform = parent.VirtualPlatform
	}
	if req.Location.LocationDetails == nil && parent.LocationDetails != nil {
		req.Location.LocationDetails = parent.LocationDetails
	}
	if req.Location.LocationVisibility == nil || *req.Location.LocationVisibility == "" {
		req.Location.LocationVisibility = &parent.LocationVisibility
	}
	if req.Location.ClickToAction.Text == nil && parent.CTAText != nil {
		req.Location.ClickToAction.Text = parent.CTAText
	}
	if req.Location.ClickToAction.Link == nil && parent.CTALink != nil {
		req.Location.ClickToAction.Link = parent.CTALink
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Time configuration validators
// ─────────────────────────────────────────────────────────────────────────────

func validateSessionTimeConfiguration(times *models.SessionTimeConfiguration) error {
	// Registration window: both ends must be present and coherent
	reg := &times.Registration
	if reg.StartAt != nil || reg.EndAt != nil {
		if reg.StartAt == nil {
			return errorc.Error(errorc.ErrorValidation, "registration start time is required when end time is set")
		}
		if reg.EndAt == nil {
			return errorc.Error(errorc.ErrorValidation, "registration end time is required when start time is set")
		}
		if !reg.EndAt.After(*reg.StartAt) {
			return errorc.Error(errorc.ErrorValidation, "registration end time must be after start time")
		}
		oneDayAgo := time.Now().AddDate(0, 0, -1)
		if reg.StartAt.Before(oneDayAgo) {
			return errorc.Error(errorc.ErrorValidation, "registration start time cannot be more than 1 day in the past")
		}
	}

	if err := validateCheckInConfig(&times.CheckIn); err != nil {
		return err
	}
	if err := validateCheckOutConfig(&times.CheckOut); err != nil {
		return err
	}

	return nil
}

func validateCheckInConfig(config *models.SessionCheckInConfig) error {
	if config.Required && !config.Enabled {
		return errorc.Error(errorc.ErrorValidation, "check-in cannot be required when it is not enabled")
	}

	if !config.Enabled {
		return nil
	}

	if config.StartAt == nil {
		return errorc.Error(errorc.ErrorValidation, "check-in start time is required when check-in is enabled")
	}
	if config.EndAt == nil {
		return errorc.Error(errorc.ErrorValidation, "check-in end time is required when check-in is enabled")
	}
	if !config.EndAt.After(*config.StartAt) {
		return errorc.Error(errorc.ErrorValidation, "check-in end time must be after check-in start time")
	}

	oneDayAgo := time.Now().AddDate(0, 0, -1)
	if config.StartAt.Before(oneDayAgo) {
		return errorc.Error(errorc.ErrorValidation, "check-in start time cannot be more than 1 day in the past")
	}

	if config.LatePolicy != "" {
		if !config.AllowLate {
			return errorc.Error(errorc.ErrorValidation, "late policy requires allow_late to be true")
		}
		switch config.LatePolicy {
		case "reject":
			if config.LateThreshold > 0 {
				return errorc.Error(errorc.ErrorValidation, "late threshold must not be set for 'reject' policy")
			}
		case "warn":
			if !config.TrackLate {
				return errorc.Error(errorc.ErrorValidation, "'warn' policy requires track_late to be true")
			}
			if config.LateThreshold <= 0 {
				return errorc.Error(errorc.ErrorValidation, "late threshold is required for 'warn' policy")
			}
		case "allow":
			if config.TrackLate {
				return errorc.Error(errorc.ErrorValidation, "'allow' policy must not have track_late enabled")
			}
		default:
			return errorc.Error(errorc.ErrorValidation, "invalid late policy: %s", config.LatePolicy)
		}
	}

	if config.TrackLate && config.LateThreshold <= 0 {
		return errorc.Error(errorc.ErrorValidation, "late threshold is required when track_late is enabled")
	}

	return nil
}

func validateCheckOutConfig(config *models.SessionCheckOutConfig) error {
	if config.Required && !config.Enabled {
		return errorc.Error(errorc.ErrorValidation, "check-out cannot be required when it is not enabled")
	}

	if !config.Enabled {
		return nil
	}

	if config.StartAt == nil {
		return errorc.Error(errorc.ErrorValidation, "check-out start time is required when check-out is enabled")
	}
	if config.EndAt == nil {
		return errorc.Error(errorc.ErrorValidation, "check-out end time is required when check-out is enabled")
	}
	if !config.EndAt.After(*config.StartAt) {
		return errorc.Error(errorc.ErrorValidation, "check-out end time must be after check-out start time")
	}

	oneDayAgo := time.Now().AddDate(0, 0, -1)
	if config.StartAt.Before(oneDayAgo) {
		return errorc.Error(errorc.ErrorValidation, "check-out start time cannot be more than 1 day in the past")
	}

	if config.LatePolicy != "" {
		if !config.AllowLate {
			return errorc.Error(errorc.ErrorValidation, "late policy requires allow_late to be true")
		}
		switch config.LatePolicy {
		case "reject":
			if config.LateThreshold > 0 {
				return errorc.Error(errorc.ErrorValidation, "late threshold must not be set for 'reject' policy")
			}
		case "warn":
			if !config.TrackLate {
				return errorc.Error(errorc.ErrorValidation, "'warn' policy requires track_late to be true")
			}
			if config.LateThreshold <= 0 {
				return errorc.Error(errorc.ErrorValidation, "late threshold is required for 'warn' policy")
			}
		case "allow":
			if config.TrackLate {
				return errorc.Error(errorc.ErrorValidation, "'allow' policy must not have track_late enabled")
			}
		default:
			return errorc.Error(errorc.ErrorValidation, "invalid late policy: %s", config.LatePolicy)
		}
	}

	if config.TrackLate && config.LateThreshold <= 0 {
		return errorc.Error(errorc.ErrorValidation, "late threshold is required when track_late is enabled")
	}

	return nil
}

// updateEvent optionally propagates a session's schedule back to the parent event.
// ctx carries the active Atomic transaction (if any); the repository resolves it
// transparently via BaseRepository.db(ctx) so both writes share the same connection.
func (es *eventSessionUsecase) updateEvent(ctx context.Context, request *models.CreateEventSessionRequest, event *models.Event) error {
	if !request.IsUpdateEvent {
		return nil
	}

	logger.AddToKey(ctx, "event", "updated_by_session", true)

	// Guard against nil schedule pointers — these should have been validated by
	// validateSchedule earlier in the loop, but we defend here explicitly because
	// dereferencing a nil *time.Time causes a panic that crashes the transaction.
	if request.Schedule.StartAt == nil || request.Schedule.EndAt == nil || request.Schedule.Timezone == nil {
		return errorc.Error(errorc.ErrorValidation, "schedule.startAt, schedule.endAt, and schedule.timezone are required when isUpdateEvent is true")
	}

	event.StartAt = *request.Schedule.StartAt
	event.EndAt = *request.Schedule.EndAt
	event.Timezone = *request.Schedule.Timezone
	event.Status = request.Status

	return es.d.Repository.Event.UpdatePartial(ctx, event.Code, &models.UpdateEventRequest{
		Schedule: &models.EventSchedule{
			StartAt:  request.Schedule.StartAt,
			EndAt:    request.Schedule.EndAt,
			Timezone: request.Schedule.Timezone,
		},
		Status: &request.Status,
	})
}

func (es *eventSessionUsecase) GetByEventCode(ctx context.Context, eventCode string) ([]models.EventSession, error) {
	sessions, err := es.d.Repository.EventSession.GetByEventCode(ctx, eventCode)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}
