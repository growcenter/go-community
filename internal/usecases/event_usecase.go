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
	"go-community/internal/repositories/pgsql"
	"strings"
	"time"
)

type EventUsecase interface {
	Create(ctx context.Context, request models.CreateEventRequest) (*models.Event, error)
}

type eventUsecase struct {
	d *Dependencies
}

func NewEventUsecase(d *Dependencies) EventUsecase {
	return &eventUsecase{
		d: d,
	}
}

// Create creates a new event with comprehensive validation and error handling
// Returns pointer to created Event model
func (eu *eventUsecase) Create(ctx context.Context, request models.CreateEventRequest) (*models.Event, error) {
	// Enrich wide event with operation context
	logger.EnrichContextWith(ctx, map[string]any{
		"operation":      "event.create",
		"event_category": request.Category,
		"event_title":    request.Title,
		"location_type":  request.Location.LocationType,
		"access_level":   request.Access.AccessLevel,
	})

	logger.EnrichContextMap(ctx,
		map[string]any{
			"category":      request.Category,
			"location_type": request.Location.LocationType,
		},
	)

	// Step 1: Extract creator community ID from context safely
	creatorID, err := contextc.ExtractCommunityID(ctx)
	if err != nil {
		logger.SetErrorContext(ctx, &logger.ErrorContext{
			Type:      "ContextError",
			Code:      "CONTEXT_ERROR",
			Message:   err.Error(),
			Retriable: true,
		})
		return nil, errorc.Error(err)
	}
	logger.EnrichContext(ctx, "creator_community_id", creatorID)

	// Step 2: Normalize request (set defaults, cleanup)
	normalizeEventRequest(&request)

	// Step 3: Early validation for announcement category
	if err := validateAnnouncementCategory(request.Category, request.Location.LocationType); err != nil {
		logger.SetErrorContext(ctx, &logger.ErrorContext{
			Type:      "ValidationFailed",
			Code:      "VALIDATION_FAILED",
			Message:   err.Error(),
			Retriable: false,
		})
		return nil, errorc.Error(err)
	}

	// Step 4: Generate unique event code with retry mechanism
	eventCode, err := generateUniqueEventCode(ctx, eu, constants.EventCodeMaxRetries)
	if err != nil {
		logger.SetErrorContext(ctx, &logger.ErrorContext{
			Type:      "GenerateEventCodeError",
			Code:      "GENERATE_EVENT_CODE_ERROR",
			Message:   err.Error(),
			Retriable: false,
		})
		return nil, errorc.Error(err)
	}
	logger.EnrichContext(ctx, "event_code", eventCode)

	// Step 5: Generate or validate slug
	slug, err := generateEventSlug(ctx, eu, request.Slug, eventCode)
	if err != nil {
		logger.SetErrorContext(ctx, &logger.ErrorContext{
			Type:      "GenerateSlugError",
			Code:      "GENERATE_SLUG_ERROR",
			Message:   err.Error(),
			Retriable: false,
		})
		return nil, errorc.Error(err)
	}
	logger.EnrichContext(ctx, "event_slug", slug)

	// Step 6: Validate access control settings
	if err := validateAccessControl(ctx, eu, &request.Access); err != nil {
		logger.SetErrorContext(ctx, &logger.ErrorContext{
			Type:      "ValidationFailed",
			Code:      "VALIDATION_FAILED",
			Message:   err.Error(),
			Retriable: false,
		})
		return nil, errorc.Error(err)
	}

	// Step 7: Validate organizer community IDs
	if err := validateCommunityIDs(ctx, eu.d.Repository.User, request.OrganizerCommunityIDs, "organizer"); err != nil {
		logger.SetErrorContext(ctx, &logger.ErrorContext{
			Type:      "ValidationFailed",
			Code:      "VALIDATION_FAILED",
			Message:   err.Error(),
			Retriable: false,
		})
		return nil, errorc.Error(err)
	}

	// Step 8: Validate contact community IDs
	if err := validateCommunityIDs(ctx, eu.d.Repository.User, request.ContactCommunityIDs, "contact"); err != nil {
		logger.SetErrorContext(ctx, &logger.ErrorContext{
			Type:      "ValidationFailed",
			Code:      "VALIDATION_FAILED",
			Message:   err.Error(),
			Retriable: false,
		})
		return nil, errorc.Error(err)
	}

	// Step 9: Build event model from request
	event := models.NewEventFromRequest(request, eventCode, slug, creatorID)

	// Step 10: Create event in database within transaction
	if err := eu.d.Repository.Transaction.Atomic(ctx, func(ctx context.Context, r *pgsql.PostgreRepositories) error {
		if err := r.Event.Create(ctx, event); err != nil {
			return errorc.Error(errorc.ErrorDatabase, "failed to create event: %v", err)
		}

		// Create event instances if provided in request
		// GORM automatically uses the transaction context
		if request.Instances != nil {
			_, err := eu.d.EventInstance.Create(ctx, event, *request.Instances)
			if err != nil {
				return errorc.Error(err, "failed to create event instances")
			}
		}

		return nil
	}); err != nil {
		logger.SetErrorContext(ctx, &logger.ErrorContext{
			Type:      "DatabaseError",
			Code:      "DATABASE_ATOMIC_ERROR",
			Message:   err.Error(),
			Retriable: true,
		})
		return nil, errorc.Error(err)
	}

	// Final enrichment with success metrics
	logger.EnrichContextWith(ctx, map[string]any{
		"event_created":    true,
		"organizers_count": len(request.OrganizerCommunityIDs),
		"contacts_count":   len(request.ContactCommunityIDs),
		"has_recurrence":   request.Schedule.Recurrence != "",
	})

	return event, nil
}

// generateUniqueEventCode generates a unique event code with retry mechanism
func generateUniqueEventCode(ctx context.Context, eu *eventUsecase, maxRetries int) (string, error) {
	for attempt := 0; attempt < maxRetries; attempt++ {
		code, err := generator.IdentifierCode(ctx, eu.d.Config.Event.EncodeCode, time.Now(), constants.EventCodePrefix)
		if err != nil {
			return "", errorc.Error(err, "failed to generate event code")
		}

		// Check if code already exists
		exists, err := eu.d.Repository.Event.CheckEventByCodeOrSlug(ctx, code, "")
		if err != nil {
			return "", errorc.Error(err, "failed to check event code uniqueness")
		}

		if !exists {
			logger.EnrichContext(ctx, "code_generation_attempts", attempt+1)
			return code, nil
		}

		// Code collision detected, retry
		logger.EnrichContextWith(ctx, map[string]any{
			"code_collision":    true,
			"collision_attempt": attempt + 1,
			"colliding_code":    code,
		})

		if attempt < maxRetries-1 {
			time.Sleep(time.Millisecond * time.Duration(10*(attempt+1))) // Exponential backoff
		}
	}

	logger.EnrichContext(ctx, "code_generation_failed", true)
	return "", errorc.Error(errorc.ErrorInternalServer, "failed to generate unique event code after %d attempts", maxRetries)
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
	exists, err := eu.d.Repository.Event.CheckEventByCodeOrSlug(ctx, "", slug)
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
	if access.AccessLevel == string(constants.AccessLevelPublic) {
		access.AllowedCampuses = nil
		access.AllowedRoles = nil
		access.AllowedUserTypes = nil
		access.AllowedCommunityIDs = nil
		return nil
	}

	// Log private event access control validation
	logger.EnrichContextWith(ctx, map[string]any{
		"validating_access_control": true,
		"campuses_count":            len(access.AllowedCampuses),
		"roles_count":               len(access.AllowedRoles),
		"user_types_count":          len(access.AllowedUserTypes),
		"communities_count":         len(access.AllowedCommunityIDs),
	})

	// Validate campuses
	if access.AllowedCampuses != nil && len(access.AllowedCampuses) > 0 {
		lowerCampuses := make([]string, len(access.AllowedCampuses))
		for i, c := range access.AllowedCampuses {
			lowerCampuses[i] = strings.ToLower(c)
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
		return err
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
		return errorc.Error(err, "failed to validate %s community IDs", entityType)
	}

	if count != int64(len(ids)) {
		return errorc.Error(errorc.ErrorDataNotFound, "one or more %s community IDs do not exist", entityType)
	}

	return nil
}

// validateAnnouncementCategory validates that announcement events must be online
func validateAnnouncementCategory(category, locationType string) error {
	if category == string(constants.CategoryAnnouncement) && locationType != string(constants.LocationTypeOnline) {
		return errorc.Error(errorc.ErrorInvalidInput, "announcement events must have online location type")
	}
	return nil
}

// normalizeEventRequest sets default values and normalizes the request
func normalizeEventRequest(req *models.CreateEventRequest) {
	// Set default location visibility
	if req.Location.LocationVisibility == "" {
		req.Location.LocationVisibility = string(constants.LocationVisibilityAll)
	}

	// Set default status
	if req.Status == "" {
		req.Status = string(constants.EventStatusDraft)
	}

	// If draft, set all instances to draft
	if req.Status == string(constants.EventStatusDraft) {
		for i := range req.Instances.InstanceRequest {
			req.Instances.InstanceRequest[i].Status = string(constants.EventStatusDraft)
		}
	}
}
