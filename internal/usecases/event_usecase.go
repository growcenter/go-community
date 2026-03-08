package usecases

import (
	"context"
	"go-community/internal/common"
	"go-community/internal/constants"
	"go-community/internal/models"
	"go-community/internal/pkg/cache"
	"go-community/internal/pkg/contextc"
	"go-community/internal/pkg/errorc"
	"go-community/internal/pkg/generator"
	"go-community/internal/pkg/logger"
	"go-community/internal/pkg/stringc"
	"go-community/internal/repositories/pgsql"
	"time"
)

type EventUsecase interface {
	Create(ctx context.Context, request models.CreateEventRequest) (*models.Event, error)
	GetAll(ctx context.Context) ([]models.Event, error)
}

type eventUsecase struct {
	d *Dependencies
}

func NewEventUsecase(d *Dependencies) EventUsecase {
	return &eventUsecase{
		d: d,
	}
}

// Create creates a new event with comprehensive validation and error handling.
// Returns pointer to created Event model.
func (eu *eventUsecase) Create(ctx context.Context, request models.CreateEventRequest) (*models.Event, error) {
	// Anchor every log line for this operation to a single, queryable field.
	logger.Add(ctx, "operation", "event.create")

	creatorCommunityID, err := contextc.ExtractCommunityID(ctx)
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:      "ContextError",
			Code:      "CONTEXT_ERROR",
			Message:   err.Error(),
			Retriable: true,
		})
		return nil, errorc.Error(err)
	}

	normalizeEventRequest(&request)

	eventCode, err := generateUniqueEventCode(ctx, eu, constants.EventCodeMaxRetries)
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:      "GenerateEventCodeError",
			Code:      "GENERATE_EVENT_CODE_ERROR",
			Message:   err.Error(),
			Retriable: false,
		})
		return nil, errorc.Error(err)
	}

	slug, err := generateEventSlug(ctx, eu, request.Slug, *eventCode)
	if err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:      "GenerateSlugError",
			Code:      "GENERATE_SLUG_ERROR",
			Message:   err.Error(),
			Retriable: false,
		})
		return nil, errorc.Error(err)
	}

	// Seed the "event" group with all server-generated identifiers known at
	// this point. One call, one mutex acquisition, zero redundancy.
	logger.Add(ctx, "event", map[string]any{
		"code":   *eventCode,
		"slug":   slug,
		"status": request.Status,
	})

	if err := validateAccessControl(ctx, eu, &request.Access); err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:      "ValidationFailed",
			Code:      "VALIDATION_FAILED",
			Message:   err.Error(),
			Retriable: false,
		})
		return nil, errorc.Error(err)
	}

	if err := validateOrganizers(ctx, eu, &request.Organizer); err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:      "ValidationFailed",
			Code:      "VALIDATION_FAILED",
			Message:   err.Error(),
			Retriable: false,
		})
		return nil, errorc.Error(err)
	}

	if err := request.Location.Validate(&request.Category); err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:      "ValidationFailed",
			Code:      "LOCATION_VALIDATION_FAILED",
			Message:   err.Error(),
			Retriable: false,
		})
		return nil, errorc.Error(errorc.ErrorValidation, err.Error())
	}

	if request.Location.ClickToAction.TextNotEmpty() && !request.Location.ClickToAction.LinkNotEmpty() {
		request.Location.ClickToAction.Link = stringc.Pointer("NORMAL_FLOW")
	} else if !request.Location.ClickToAction.TextNotEmpty() && request.Location.ClickToAction.LinkNotEmpty() {
		request.Location.ClickToAction.Text = stringc.Pointer("Register Here!")
	} else if !request.Location.ClickToAction.TextNotEmpty() && !request.Location.ClickToAction.LinkNotEmpty() {
		request.Location.ClickToAction.Link = stringc.Pointer("NORMAL_FLOW")
		request.Location.ClickToAction.Text = stringc.Pointer("Register Here!")
	}

	if err := validateSchedule(&request.Schedule); err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:      "ValidationFailed",
			Code:      "SCHEDULE_VALIDATION_FAILED",
			Message:   err.Error(),
			Retriable: false,
		})
		return nil, errorc.Error(err)
	}

	if request.Recurrence.IsRecurring {
		if err := validateRecurrencePattern(&request.Recurrence); err != nil {
			logger.AddError(ctx, &logger.ErrorContext{
				Type:      "ValidationFailed",
				Code:      "RECURRENCE_VALIDATION_FAILED",
				Message:   err.Error(),
				Retriable: false,
			})
			return nil, errorc.Error(err)
		}
	}

	if err := validateNotification(&request.Notification); err != nil {
		logger.AddError(ctx, &logger.ErrorContext{
			Type:      "ValidationFailed",
			Code:      "NOTIFICATION_VALIDATION_FAILED",
			Message:   err.Error(),
			Retriable: false,
		})
		return nil, errorc.Error(err)
	}

	event := &models.Event{
		Code:               *eventCode,
		Title:              request.Title,
		Slug:               slug,
		PreDescription:     &request.PreDescription,
		PostDescription:    request.PostDescription,
		TermsAndConditions: &request.TermsAndConditions,
		Category:           request.Category,
		Status:             request.Status,

		// Images
		ImageURLs: request.Images.ImageLinks,

		// Organization
		CreatorCommunityID:    creatorCommunityID,
		OrganizerCommunityIDs: request.Organizer.OrganizerCommunityIDs,
		ContactCommunityIDs:   request.Organizer.ContactCommunityIDs,

		// Access
		AccessLevel:         *request.Access.AccessLevel,
		AllowedUserTypes:    request.Access.AllowedUserTypes,
		AllowedRoles:        request.Access.AllowedRoles,
		AllowedCampuses:     request.Access.AllowedCampuses,
		AllowedCommunityIDs: request.Access.AllowedCommunityIDs,

		// Location
		LocationType:       *request.Location.LocationType,
		PhysicalAddress:    request.Location.PhysicalAddress,
		PhysicalPlaceName:  request.Location.PhysicalPlaceName,
		VirtualLink:        request.Location.VirtualLink,
		VirtualPlatform:    request.Location.VirtualPlatform,
		LocationDetails:    request.Location.LocationDetails,
		LocationVisibility: *request.Location.LocationVisibility,
		CTAText:            request.Location.ClickToAction.Text,
		CTALink:            request.Location.ClickToAction.Link,

		// Schedule
		StartAt:  *request.Schedule.StartAt,
		EndAt:    *request.Schedule.EndAt,
		Timezone: *request.Schedule.Timezone,

		// Recurrence
		IsRecurring: request.Recurrence.IsRecurring,
		// RecurrencePattern will be set below if not nil

		// Template
		IsTemplate: request.Template.IsTemplate,
		// TemplateID and SeriesID are optional — set safely below via nil-guards

		// Notification
		NotificationChannels: request.Notification.NotificationChannels,
		// ReminderConfig will be set below if not nil
	}

	// Optional fields — only appended to the group when they're present.
	if request.Images.BannerLink != nil {
		event.BannerURL = *request.Images.BannerLink
		logger.AddToKey(ctx, "event", "has_banner", true)
	}
	if request.Template.TemplateID != nil {
		event.TemplateID = *request.Template.TemplateID
		logger.AddToKey(ctx, "event", "template_id", *request.Template.TemplateID)
	}
	if request.Template.SeriesID != nil {
		event.SeriesID = *request.Template.SeriesID
		logger.AddToKey(ctx, "event", "series_id", *request.Template.SeriesID)
	}

	// Marshal JSONB fields; surface the scalar summaries that are queryable.
	if request.Recurrence.RecurrencePattern != nil {
		if err := event.RecurrencePattern.Marshal(request.Recurrence.RecurrencePattern); err != nil {
			logger.AddError(ctx, &logger.ErrorContext{
				Type:      "MarshalError",
				Code:      "MARSHAL_RECURRENCE_FAILED",
				Message:   "failed to marshal recurrence pattern",
				Retriable: false,
			})
			return nil, errorc.Error(errorc.ErrorInternalServer, "failed to marshal recurrence pattern")
		}
		logger.AddToKey(ctx, "event", "recurrence_frequency", request.Recurrence.RecurrencePattern.Frequency)
	}
	if request.Notification.ReminderConfig != nil {
		if err := event.ReminderConfig.Marshal(request.Notification.ReminderConfig); err != nil {
			logger.AddError(ctx, &logger.ErrorContext{
				Type:      "MarshalError",
				Code:      "MARSHAL_REMINDER_FAILED",
				Message:   "failed to marshal reminder config",
				Retriable: false,
			})
			return nil, errorc.Error(errorc.ErrorInternalServer, "failed to marshal reminder config")
		}
		logger.AddToKey(ctx, "event", "reminder_enabled", request.Notification.ReminderConfig.Enabled)
	}

	if request.Category == string("announcement") && request.Sessions != nil {
		return nil, errorc.Error(errorc.ErrorValidation, "instances is not required for announcement category")
	}

	if err := eu.d.Repository.Transaction.Atomic(ctx, func(ctx context.Context, r *pgsql.PostgreRepositories) error {
		if err := r.Event.Create(ctx, event); err != nil {
			return errorc.Error(errorc.ErrorDatabase, "failed to create event: %s", err)
		}

		// event.ID is assigned by the DB — append it to the "event" group now.
		logger.AddToKey(ctx, "event", "id", event.ID)

		if request.Sessions != nil {
			if _, err := eu.d.EventSession.Create(ctx, request.Sessions, nil, event); err != nil {
				return errorc.Error(err, "failed to create event sessions: %s", err)
			}
			// sessions_created belongs to the event group — it describes how many
			// sub-resources were attached to this event in the same transaction.
			logger.AddToKey(ctx, "event", "sessions_created", len(request.Sessions))
		}

		if request.Questions != nil {
			formReq := models.CreateFormRequest{
				Name: event.Title,
				Entity: models.FormEntityRequest{
					Type: "event",
					Code: event.Code,
				},
				Questions: request.Questions,
			}

			formResp, err := eu.d.Form.Create(ctx, &formReq)
			if err != nil {
				return errorc.Error(err, "failed to create event questions: %s", err)
			}

			questionCodes := make([]string, len(formResp.Questions))
			for i, q := range formResp.Questions {
				questionCodes[i] = q.Code
			}
			// All three fields are server-generated and invisible in request/response.
			// AddToKey with a map writes them all into the "form" group in one call.
			logger.AddToKey(ctx, "form", map[string]any{
				"code":              formResp.Code,
				"questions_created": len(formResp.Questions),
				"question_codes":    questionCodes,
			})
		}

		return nil
	}); err != nil {
		return nil, errorc.Error(err, "failed to create event: %s", err)
	}

	return event, nil
}

// generateUniqueEventCode generates a unique event code with retry mechanism
func generateUniqueEventCode(ctx context.Context, eu *eventUsecase, maxRetries int) (*string, error) {
	for attempt := 0; attempt < maxRetries; attempt++ {
		code, err := generator.IdentifierCode(ctx, eu.d.Config.Event.EncodeCode, time.Now(), constants.EventCodePrefix)
		if err != nil {
			return nil, errorc.Error(err, "failed to generate event code")
		}

		// Check if code already exists
		isCodeExist, err := eu.d.Repository.Event.CheckByCode(ctx, code)
		if err != nil {
			return nil, errorc.Error(err, "failed to check event code uniqueness")
		}

		if !isCodeExist {
			logger.Add(ctx, "code_generation_attempts", attempt+1)
			return &code, nil
		}

		// Code collision — rare, but worth knowing about. Log under its own group
		// so it doesn't pollute the top-level event fields.
		logger.Add(ctx, "code_collision", map[string]any{
			"attempt": attempt + 1,
			"code":    code,
		})

		if attempt < maxRetries-1 {
			time.Sleep(time.Duration(eu.d.Config.Event.Backoff.CodeGeneration) * time.Millisecond * time.Duration(attempt+1)) // Exponential backoff
		}
	}

	return nil, errorc.Error(errorc.ErrorInternalServer, "failed to generate unique event code after %d attempts", maxRetries)
}

// generateEventSlug generates or validates the event slug
func generateEventSlug(ctx context.Context, eu *eventUsecase, requestSlug, eventCode string) (string, error) {
	var slug string

	if requestSlug != "" {
		// User provided a slug
		if len(requestSlug) < eu.d.Config.Event.MinSlug {
			// Slug too short, regenerate from provided slug
			slug = generator.Slug(requestSlug, time.Now())
		} else {
			slug = requestSlug
		}
	} else {
		// No slug provided, generate from event code
		slug = generator.Slug(eventCode, time.Now())
	}

	// Check if slug already exists
	exists, err := eu.d.Repository.Event.CheckBySlug(ctx, slug)
	if err != nil {
		return "", errorc.Error(err, "failed to check slug uniqueness")
	}

	if exists {
		return "", errorc.Error(errorc.ErrorAlreadyExist, "event with slug '%s' already exists", slug)
	}

	return slug, nil
}

// validateAccessControl validates all access control settings
func validateAccessControl(ctx context.Context, eu *eventUsecase, access *models.EventAccess) error {
	// Public events don't need access restrictions
	if *access.AccessLevel == string(constants.AccessLevelPublic) {
		access.AllowedCampuses = nil
		access.AllowedRoles = nil
		access.AllowedUserTypes = nil
		access.AllowedCommunityIDs = nil
		return nil
	}

	logger.Add(ctx,
		"campuses_count", len(access.AllowedCampuses),
		"roles_count", len(access.AllowedRoles),
		"user_types_count", len(access.AllowedUserTypes),
		"communities_count", len(access.AllowedCommunityIDs),
	)

	// Validate campuses
	if access.AllowedCampuses != nil && len(access.AllowedCampuses) > 0 {
		lowerCampuses := make([]string, len(access.AllowedCampuses))
		for i, c := range access.AllowedCampuses {
			lowerCampuses[i] = stringc.LowerAndTrimSpace(c)
		}

		if !common.CheckAllDataMapStructure(eu.d.Config.Campus, lowerCampuses) {
			return errorc.Error(errorc.ErrorDataNotFound, "one or more campuses do not exist")
		}
	}

	// Validate roles with caching
	if access.AllowedRoles != nil && len(access.AllowedRoles) > 0 {
		cacheKey := cache.RolesCacheKey(access.AllowedRoles)

		// Try to get from cache or fetch from database
		result, err := eu.d.Cache.GetOrSet(cacheKey, func() (interface{}, error) {
			count, err := eu.d.Repository.Role.CheckMultiple(ctx, access.AllowedRoles)
			if err != nil {
				return nil, errorc.Error(err, "failed to validate roles")
			}
			return count, nil
		})

		if err != nil {
			return err
		}

		count := result.(int64)
		if count != int64(len(access.AllowedRoles)) {
			return errorc.Error(errorc.ErrorDataNotFound, "one or more roles do not exist")
		}
	}

	// Validate user types with caching
	if access.AllowedUserTypes != nil && len(access.AllowedUserTypes) > 0 {
		cacheKey := cache.UserTypesCacheKey(access.AllowedUserTypes)

		// Try to get from cache or fetch from database
		result, err := eu.d.Cache.GetOrSet(cacheKey, func() (interface{}, error) {
			count, err := eu.d.Repository.UserType.CheckMultiple(ctx, access.AllowedUserTypes)
			if err != nil {
				return nil, errorc.Error(err, "failed to validate user types")
			}
			return count, nil
		})

		if err != nil {
			return err
		}

		count := result.(int64)
		if count != int64(len(access.AllowedUserTypes)) {
			return errorc.Error(errorc.ErrorDataNotFound, "one or more user types do not exist")
		}
	}

	// Validate community IDs
	if err := validateCommunityIDs(ctx, eu.d.Repository.User, access.AllowedCommunityIDs, "allowed"); err != nil {
		return errorc.Error(err)
	}

	return nil
}

func validateOrganizers(ctx context.Context, eu *eventUsecase, organizers *models.EventOrganizer) error {
	if organizers == nil {
		return nil
	}

	logger.Add(ctx,
		"organizer_count", len(organizers.OrganizerCommunityIDs),
		"contact_count", len(organizers.ContactCommunityIDs),
	)

	if organizers.ContactCommunityIDs != nil && len(organizers.ContactCommunityIDs) > 0 {
		cacheKey := cache.RolesCacheKey(organizers.ContactCommunityIDs)

		// Try to get from cache or fetch from database
		result, err := eu.d.Cache.GetOrSet(cacheKey, func() (interface{}, error) {
			count, err := eu.d.Repository.User.CheckMultiple(ctx, organizers.ContactCommunityIDs)
			if err != nil {
				return nil, errorc.Error(err, "failed to validate roles: %s", err.Error())
			}
			return count, nil
		})

		if err != nil {
			return err
		}

		count := result.(int64)
		if count != int64(len(organizers.ContactCommunityIDs)) {
			return errorc.Error(errorc.ErrorDataNotFound, "one or more contact community IDs do not exist")
		}
	}

	if organizers.OrganizerCommunityIDs != nil && len(organizers.OrganizerCommunityIDs) > 0 {
		cacheKey := cache.RolesCacheKey(organizers.OrganizerCommunityIDs)

		// Try to get from cache or fetch from database
		result, err := eu.d.Cache.GetOrSet(cacheKey, func() (interface{}, error) {
			count, err := eu.d.Repository.User.CheckMultiple(ctx, organizers.OrganizerCommunityIDs)
			if err != nil {
				return nil, errorc.Error(err, "failed to validate roles: %s", err.Error())
			}
			return count, nil
		})

		if err != nil {
			return err
		}

		count := result.(int64)
		if count != int64(len(organizers.OrganizerCommunityIDs)) {
			return errorc.Error(errorc.ErrorDataNotFound, "one or more organizer community IDs do not exist")
		}
	}

	return nil
}

// validateCommunityIDs validates that all provided community IDs exist in the database
func validateCommunityIDs(ctx context.Context, repo pgsql.UserRepository, ids []string, entityType string) error {
	if ids == nil || len(ids) == 0 {
		return nil
	}

	count, err := repo.CheckMultiple(ctx, ids)
	if err != nil {
		return errorc.Error(errorc.ErrorDatabase, "failed to validate %s community IDs for the event", entityType)
	}

	if count != int64(len(ids)) {
		return errorc.Error(errorc.ErrorDataNotFound, "one or more %s community IDs for the event do not exist", entityType)
	}

	return nil
}

// normalizeEventRequest sets default values and normalizes the request
func normalizeEventRequest(req *models.CreateEventRequest) {
	// Set default location visibility — LocationVisibility may be nil if omitted in request
	if req.Location.LocationVisibility == nil || *req.Location.LocationVisibility == "" {
		v := string(constants.LocationVisibilityAll)
		req.Location.LocationVisibility = &v
	}

	// Set default status
	if req.Status == "" {
		req.Status = string(constants.EventStatusDraft)
	}

	// If draft, set all instances to draft
	if req.Status == string(constants.EventStatusDraft) {
		for _, session := range req.Sessions {
			session.Status = string(constants.EventStatusDraft)
		}
	}

	// Set default timezone if not provided
	if req.Schedule.Timezone == nil {
		tz := common.DefaultTimeZone
		req.Schedule.Timezone = &tz
	}

	// ClickToAction defaults — must assign a new pointer, not dereference a nil one
	if req.Location.ClickToAction.Link == nil {
		req.Location.ClickToAction.Link = stringc.Pointer("NORMAL_FLOW")
	}

	if req.Location.ClickToAction.Text == nil {
		req.Location.ClickToAction.Text = stringc.Pointer("Register Here!")
	}
}

// validateSchedule validates that the event schedule is valid
func validateSchedule(schedule *models.EventSchedule) error {
	if schedule.StartAt == nil {
		return errorc.Error(errorc.ErrorValidation, "start time is required")
	}
	if schedule.EndAt == nil {
		return errorc.Error(errorc.ErrorValidation, "end time is required")
	}

	// Explicit check: EndAt must be after StartAt
	if !schedule.EndAt.After(*schedule.StartAt) {
		return errorc.Error(errorc.ErrorValidation, "end time must be after start time")
	}

	// Check that the event is not too far in the past (more than 1 day)
	oneDayAgo := time.Now().AddDate(0, 0, -1)
	if schedule.StartAt.Before(oneDayAgo) {
		return errorc.Error(errorc.ErrorValidation, "event start time cannot be more than 1 day in the past")
	}

	return nil
}

// validateRecurrencePattern validates the recurrence pattern
func validateRecurrencePattern(recurrence *models.EventRecurrence) error {
	if !recurrence.IsRecurring {
		return nil
	}

	if recurrence.RecurrencePattern == nil {
		return errorc.Error(errorc.ErrorValidation, "recurrence pattern is required when event is recurring")
	}

	// Use the existing validation function from models
	if err := models.ValidateRecurrencePattern(recurrence.RecurrencePattern); err != nil {
		return errorc.Error(errorc.ErrorValidation, "invalid recurrence pattern: %v", err)
	}

	return nil
}

// getNthWeekdayOfMonth gets the nth occurrence of a weekday in a month
func getNthWeekdayOfMonth(year int, month time.Month, nth string, weekday time.Weekday, template time.Time) time.Time {
	// Start at the first day of the month
	firstDay := time.Date(year, month, 1, template.Hour(), template.Minute(), template.Second(), 0, template.Location())

	// Find the first occurrence of the target weekday
	daysUntilWeekday := int(weekday - firstDay.Weekday())
	if daysUntilWeekday < 0 {
		daysUntilWeekday += 7
	}

	firstOccurrence := firstDay.AddDate(0, 0, daysUntilWeekday)

	// Calculate the nth occurrence
	var offset int
	switch nth {
	case "first":
		offset = 0
	case "second":
		offset = 7
	case "third":
		offset = 14
	case "fourth":
		offset = 21
	case "last":
		// Find the last occurrence by going to the 5th and checking if it's in the same month
		fifthOccurrence := firstOccurrence.AddDate(0, 0, 28)
		if fifthOccurrence.Month() == month {
			return fifthOccurrence
		}
		return firstOccurrence.AddDate(0, 0, 21)
	default:
		offset = 0
	}

	return firstOccurrence.AddDate(0, 0, offset)
}

// isExcluded checks if a date is in the exclusion list
func isExcluded(date time.Time, excludeDates []time.Time) bool {
	for _, excluded := range excludeDates {
		if date.Year() == excluded.Year() &&
			date.Month() == excluded.Month() &&
			date.Day() == excluded.Day() {
			return true
		}
	}
	return false
}

func validateNotification(notification *models.EventNotification) error {
	if notification == nil {
		return nil
	}

	if notification.ReminderConfig != nil && (notification.NotificationChannels == nil || len(notification.NotificationChannels) == 0) {
		return errorc.Error(errorc.ErrorValidation, "notification channels is required when reminder config is set")
	}

	return nil
}

func (eu *eventUsecase) GetAll(ctx context.Context) ([]models.Event, error) {
	return eu.d.Repository.Event.GetDummy(ctx)
}
