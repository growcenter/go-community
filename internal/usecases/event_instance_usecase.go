package usecases

import (
	"context"
	"encoding/json"
	"go-community/internal/constants"
	"go-community/internal/models"
	"go-community/internal/pkg/errorc"
	"go-community/internal/pkg/generator"
	"go-community/internal/pkg/logger"
	"time"
)

type EventInstanceUsecase interface {
	Create(ctx context.Context, event *models.Event, requests models.CreateInstanceRequest) (*[]models.EventInstance, error)
}

type eventInstanceUsecase struct {
	d *Dependencies
}

func NewEventInstanceUsecase(d *Dependencies) EventInstanceUsecase {
	return &eventInstanceUsecase{
		d: d,
	}
}

// Create creates event instances from requests
// Works both standalone and within a transaction (GORM handles this automatically)
// Returns pointer to created instances slice
func (eiu *eventInstanceUsecase) Create(ctx context.Context, event *models.Event, requests models.CreateInstanceRequest) (*[]models.EventInstance, error) {
	// Fetch parent event once if not provided
	if event == nil {
		if len(requests.InstanceRequest) == 0 {
			return nil, errorc.Error(errorc.ErrorEmptyInput)
		}

		// Fetch event once
		eventExist, err := eiu.d.Repository.Event.GetByCode(ctx, requests.EventConfiguration.Code)
		if err != nil {
			// Original code logs error context here, replicating that behavior
			logger.SetErrorContext(ctx, &logger.ErrorContext{
				Type:      "GetEventByCodeError",
				Code:      "GET_EVENT_BY_CODE_ERROR",
				Message:   err.Error(),
				Retriable: false,
			})
			return nil, errorc.Error(err, "failed to fetch event %s", requests.EventConfiguration.Code)
		}

		if eventExist.ID == 0 {
			logger.SetErrorContext(ctx, &logger.ErrorContext{
				Type:      "EventNotFound",
				Code:      "EVENT_NOT_FOUND",
				Message:   "Event not found",
				Retriable: false,
			})
			return nil, errorc.Error(errorc.ErrorDataNotFound)
		}

		event = eventExist
	}

	if len(requests.InstanceRequest) == 0 {
		return nil, errorc.Error(errorc.ErrorEmptyInput)
	}

	var instances []models.EventInstance

	for _, request := range requests.InstanceRequest {
		// Step 2: Generate instance code with retry for uniqueness
		var instanceCode string
		for attempt := 0; attempt < constants.EventCodeMaxRetries; attempt++ {
			code, err := generator.InstanceCode(ctx, event.Code, eiu.d.Config.Event.EncodeCode)
			if err != nil {
				logger.SetErrorContext(ctx, &logger.ErrorContext{
					Type:      "GenerateInstanceCodeError",
					Code:      "GENERATE_INSTANCE_CODE_ERROR",
					Message:   err.Error(),
					Retriable: false,
				})
				return nil, errorc.Error(err, "failed to generate instance code")
			}

			// Check uniqueness
			exists, err := eiu.d.Repository.EventInstance.CheckByCode(ctx, code)
			if err != nil {
				return nil, errorc.Error(err, "failed to check instance code uniqueness")
			}

			if !exists {
				instanceCode = code
				logger.EnrichContext(ctx, "instance_code_generation_attempts", attempt+1)
				break
			}

			// Collision detected (very rare with random), retry
			logger.EnrichContextWith(ctx, map[string]any{
				"instance_code_collision": true,
				"collision_attempt":       attempt + 1,
				"colliding_code":          code,
			})
		}

		if instanceCode == "" {
			logger.SetErrorContext(ctx, &logger.ErrorContext{
				Type:      "GenerateInstanceCodeError",
				Code:      "GENERATE_INSTANCE_CODE_ERROR",
				Message:   "failed to generate unique instance code after retries",
				Retriable: false,
			})
			return nil, errorc.Error(errorc.ErrorInternalServer,
				"failed to generate unique instance code after %d attempts",
				constants.EventCodeMaxRetries)
		}

		// Step 3: Validate schedule constraints
		if err := validateInstanceSchedule(ctx, request.Schedule); err != nil {
			return nil, err
		}

		// Step 4: Validate instance type and configure accordingly
		if err := validateInstanceType(ctx, &request); err != nil {
			return nil, err
		}

		// Step 5: Validate and auto-configure age-based settings
		if err := eiu.validateAgeConfiguration(ctx, &request.RegistrationConfig); err != nil {
			return nil, err
		}

		// Step 6: Validate registration configuration
		if len(request.RegistrationConfig.Methods) != 0 {
			if request.RegistrationConfig.Capacity == 0 {
				return nil, errorc.Error(errorc.ErrorInvalidInput, "capacity cannot be zero")
			}

			// Set default quota for recurring events with uniqueness enforcement
			if event.IsRecurring() && request.RegistrationConfig.EnforceUniqueness {
				if request.RegistrationConfig.QuotaPerUser == 0 {
					request.RegistrationConfig.QuotaPerUser = 1
				}
			}

			if request.RegistrationConfig.Flow == "" {
				return nil, errorc.Error(errorc.ErrorMissingFields, "registration flow cannot be empty")
			}
		} else {
			request.RegistrationConfig.EnforceCommunityId = false
			request.RegistrationConfig.EnforceUniqueness = false
			request.RegistrationConfig.EnforceSelfRegistration = false
			request.RegistrationConfig.Capacity = 0
			request.RegistrationConfig.Flow = "free"
			request.RegistrationConfig.QuotaPerUser = 0
		}

		// Step 4: Inherit from parent event if requested
		if request.IsUpdateEventTime {
			request.Title = event.Title

			// Safe dereferencing of pointer fields
			if event.Description != nil {
				request.Description = *event.Description
			}

			request.Schedule.StartAt = event.StartAt
			request.Schedule.EndAt = event.EndAt
			request.Location.LocationType = event.LocationType

			if event.PhysicalAddress != nil {
				request.Location.PhysicalAddress = *event.PhysicalAddress
			}
			if event.VirtualLink != nil {
				request.Location.VirtualLink = *event.VirtualLink
			}
		}

		// Set default location visibility
		if request.Location.LocationVisibility == "" {
			request.Location.LocationVisibility = string(constants.LocationVisibilityAll)
		}

		// Set default status
		if request.Status == "" {
			request.Status = string(constants.EventStatusDraft)
		}

		// Step 5: Marshal identifier fields
		parentIdentifierFields, err := json.Marshal(request.RegistrationIdentifier.ParentIdentifierFields)
		if err != nil {
			return nil, errorc.Error(err, "failed to marshal parent identifier fields")
		}
		childIdentifierFields, err := json.Marshal(request.RegistrationIdentifier.ChildIdentifierFields)
		if err != nil {
			return nil, errorc.Error(err, "failed to marshal child identifier fields")
		}

		// Step 6: Build instance model using constructor
		instance := models.NewInstanceFromRequest(
			request,
			event,
			instanceCode,
			parentIdentifierFields,
			childIdentifierFields,
		)

		// Step 7: Append to instances slice
		instances = append(instances, *instance)

		// Step 8: Update parent event times if requested
		if request.IsUpdateEventTime {
			eventUpdateRequest := models.UpdateEventRequest{
				StartAt: &instance.StartAt,
				EndAt:   &instance.EndAt,
			}

			logger.EnrichContextWith(ctx, map[string]any{
				"is_update_event_time": request.IsUpdateEventTime,
				"event_code":           event.Code,
				"event_update_request": eventUpdateRequest,
			})

			if err := eiu.d.Repository.Event.UpdatePartial(ctx, event.Code, &eventUpdateRequest); err != nil {
				logger.SetErrorContext(ctx, &logger.ErrorContext{
					Type:      "UpdateEventError",
					Code:      "UPDATE_EVENT_ERROR",
					Message:   err.Error(),
					Retriable: false,
				})
				return nil, errorc.Error(err)
			}
		}
	}

	// Step 9: Bulk create all instances
	// GORM automatically uses transaction context if called within Atomic()
	if err := eiu.d.Repository.EventInstance.BulkCreate(ctx, instances); err != nil {
		return nil, errorc.Error(err)
	}

	logger.EnrichContextWith(ctx, map[string]any{
		"instances_created": len(instances),
		"event_code":        event.Code,
	})

	return &instances, nil
}

// validateInstanceSchedule validates all time-related constraints for an instance
func validateInstanceSchedule(ctx context.Context, schedule models.InstanceSchedule) error {
	now := time.Now()

	// Registration window validation
	if !schedule.RegisterStartAt.Before(schedule.RegisterEndAt) {
		return errorc.Error(errorc.ErrorInvalidInput,
			"registration start must be before registration end")
	}

	if !schedule.RegisterEndAt.Before(schedule.StartAt) {
		return errorc.Error(errorc.ErrorInvalidInput,
			"registration must close before event starts")
	}

	// Verification window validation
	if !schedule.VerifyStartAt.Before(schedule.VerifyEndAt) {
		return errorc.Error(errorc.ErrorInvalidInput,
			"verification start must be before verification end")
	}

	// Event time validation
	if !schedule.StartAt.Before(schedule.EndAt) {
		return errorc.Error(errorc.ErrorInvalidInput,
			"event start must be before event end")
	}

	// Warn if registration is already closed (but allow for backdating)
	if schedule.RegisterEndAt.Before(now) {
		logger.EnrichContext(ctx, "registration_already_closed", true)
	}

	return nil
}

// validateInstanceType validates instance type and related configurations
func validateInstanceType(ctx context.Context, request *models.InstanceRequest) error {
	switch request.InstanceType {
	case "registration":
		// Must have registration configuration
		if len(request.RegistrationConfig.Methods) == 0 {
			return errorc.Error(errorc.ErrorInvalidInput,
				"registration instances must have registration methods")
		}
		if request.RegistrationConfig.Capacity == 0 {
			return errorc.Error(errorc.ErrorInvalidInput,
				"registration instances must have capacity")
		}
		logger.EnrichContext(ctx, "instance_type", "registration")

	case "announcement":
		// Clear registration config for announcement instances
		request.RegistrationConfig.Methods = nil
		request.RegistrationConfig.Capacity = 0
		request.RegistrationConfig.QuotaPerUser = 0
		request.RegistrationConfig.Flow = "free"
		request.RegistrationConfig.EnforceCommunityId = false
		request.RegistrationConfig.EnforceUniqueness = false
		request.RegistrationConfig.EnforceSelfRegistration = false
		logger.EnrichContext(ctx, "instance_type", "announcement")

	case "volunteer-attendance":
		// Volunteer instances need registration but with specific settings
		if len(request.RegistrationConfig.Methods) == 0 {
			// Default to QR-based attendance
			request.RegistrationConfig.Methods = []string{"personal-qr"}
		}
		if request.RegistrationConfig.Flow == "" {
			request.RegistrationConfig.Flow = "direct"
		}
		logger.EnrichContext(ctx, "instance_type", "volunteer-attendance")

	default:
		return errorc.Error(errorc.ErrorInvalidInput,
			"invalid instance type: %s (must be registration, announcement, or volunteer-attendance)",
			request.InstanceType)
	}

	return nil
}

// validateAgeConfiguration validates and auto-configures age-based settings
func (eiu *eventInstanceUsecase) validateAgeConfiguration(
	ctx context.Context,
	config *models.InstanceRegistrationConfig,
) error {
	// Validate age range if provided
	if config.MinAge != nil && config.MaxAge != nil {
		if *config.MinAge > *config.MaxAge {
			return errorc.Error(errorc.ErrorInvalidInput,
				"minimum age cannot be greater than maximum age")
		}
	}

	// Auto-require parent info based on configuration
	if config.MaxAge != nil && !config.RequireParentInfo {
		autoRequireAge := eiu.d.Config.Event.AutoRequireParentInfoAge
		if *config.MaxAge < autoRequireAge {
			config.RequireParentInfo = true
			logger.EnrichContextWith(ctx, map[string]any{
				"auto_require_parent_info": true,
				"max_age":                  *config.MaxAge,
				"threshold_age":            autoRequireAge,
			})
		}
	}

	// Validate family registration settings
	if config.IsFamilyRegistration {
		if config.MaxFamilyMembers == nil {
			// Default to 10 family members if not specified
			defaultMax := 10
			config.MaxFamilyMembers = &defaultMax
		}

		logger.EnrichContextWith(ctx, map[string]any{
			"family_registration_enabled": true,
			"max_family_members":          *config.MaxFamilyMembers,
		})
	}

	return nil
}
