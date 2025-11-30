package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"go-community/internal/common"
	"go-community/internal/config"
	"go-community/internal/constants"
	"go-community/internal/models"
	"go-community/internal/pkg/authorization"
	"go-community/internal/pkg/errorgen"
	"go-community/internal/pkg/generator"
	"go-community/internal/repositories/pgsql"
	"time"
)

type EventInstanceUsecase interface {
	Create(ctx context.Context, event *models.Event, requests []models.CreateInstanceRequest) (response []models.CreateInstanceResponse, err error)
}

type eventInstanceUsecase struct {
	cfg *config.Configuration
	a   authorization.Auth
	r   pgsql.PostgreRepositories
	f   FormUsecase
}

func NewEventInstanceUsecase(cfg config.Configuration, a authorization.Auth, r pgsql.PostgreRepositories, f FormUsecase) *eventInstanceUsecase {
	return &eventInstanceUsecase{
		cfg: &cfg,
		a:   a,
		r:   r,
		f:   f,
	}
}

// SECTION: Create Instance - Create iterates through multiple instance requests to prepare them for creation.
func (eiu *eventInstanceUsecase) Create(ctx context.Context, event *models.Event, requests []models.CreateInstanceRequest) (response []models.CreateInstanceResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	var instances []models.EventInstance
	for i, request := range requests {
		// ANCHOR: - Step 1: Fetch parent event if not provided. This is useful when creating instances for an already existing event.
		if event == nil {
			eventExist, err := eiu.r.Event.GetByCode(ctx, request.EventCode)
			if err != nil {
				return nil, errorgen.Error(err)
			}

			if eventExist.ID == 0 {
				return nil, errorgen.Error(errorgen.DataNotFound)
			}

			event = &eventExist
		}

		// ANCHOR: - Step 2: Determine the instance's status based on the 'IsPublish' flag from the request.
		var instanceStatus string
		if request.IsPublish {
			instanceStatus = constants.EventStatusActive
		} else {
			instanceStatus = constants.EventStatusDraft
		}

		// ANCHOR: - Step 3: Count existing instances for this event to help generate a unique instance code.
		countInstance, err := eiu.r.EventInstance.CountByCode(ctx, event.Code)
		if err != nil {
			return nil, errorgen.Error(err)
		}

		// ANCHOR: - Step 4: Parse all time-related strings from the request into proper time.Time objects, applying the specified timezone.
		instanceTimes, _ := common.ParseMultipleTime([]string{request.TimeConfig.StartAt, request.TimeConfig.EndAt, request.TimeConfig.RegisterStartAt, request.TimeConfig.RegisterEndAt, request.TimeConfig.VerifyStartAt, request.TimeConfig.VerifyEndAt}, request.TimeConfig.Timezone, time.RFC3339)

		// ANCHOR: - Step 5: Generate a unique and non-sequential code for the new instance.
		timeNowNano, _ := common.NowWithNanoTime()
		numberForCode := int(countInstance) + i
		instanceCode := fmt.Sprintf("%s-%s", event.Code, generator.GenerateHashCode(fmt.Sprintf("instance-%s-%d-%d", event.Code, numberForCode, timeNowNano.UnixNano()), 7))

		// ANCHOR: - Step 6: Validate the parsed time ranges to ensure start times are not after their corresponding end times.
		if instanceTimes[0].After(instanceTimes[1]) || instanceTimes[2].After(instanceTimes[3]) || instanceTimes[4].After(instanceTimes[5]) {
			return nil, errorgen.Error(errorgen.ErrInvalidDate, "start time cannot be later than end time")
		}

		// ANCHOR: - Step 7: Validate the registration configuration.
		if len(request.RegistrationConfig.Methods) != 0 {
			// If registration methods are specified, capacity cannot be zero.
			if request.RegistrationConfig.Capacity == 0 {
				// A special case for recurring events where quota might be set to 1 if uniqueness is enforced.
				if event.IsRecurring && request.RegistrationConfig.EnforceUniqueness {
					request.RegistrationConfig.QuotaPerUser = 1
				}
				return nil, errorgen.Error(errorgen.ErrInvalidInput, "capacity cannot be zero")
			}

			// If methods are present, a registration flow (e.g., 'direct', 'staged') must also be defined.
			if request.RegistrationConfig.Flow == "" {
				return nil, errorgen.Error(errorgen.ErrMissingFields, "registration flow cannot be empty")
			}
		} else {
			request.RegistrationConfig.EnforceCommunityId = false
			request.RegistrationConfig.EnforceUniqueness = false
			request.RegistrationConfig.EnforceSelfRegistration = false
			request.RegistrationConfig.Capacity = 0
			request.RegistrationConfig.Flow = "free"
			request.RegistrationConfig.QuotaPerUser = 0
		}

		// ANCHOR: - Step 8: If 'IsFollowEvent' is true, the instance inherits its main details (title, description, time, location) from the parent event.
		if request.IsFollowEvent {
			request.Title = event.Title
			request.Description = event.Description
			instanceTimes[0] = event.StartAt
			instanceTimes[1] = event.EndAt
			request.Location.Type = event.LocationType
			request.Location.OfflineVenue = event.LocationOfflineVenue
			request.Location.OnlineLink = event.LocationOnlineLink
		}

		parentIdentifierFields, err := json.Marshal(request.IdentifierConfig.ParentIdentifierFields)
		if err != nil {
			return nil, errorgen.Error(err, "failed to marshal parent identifier fields")
		}
		childIdentifierFields, err := json.Marshal(request.IdentifierConfig.ChildIdentifierFields)
		if err != nil {
			return nil, errorgen.Error(err, "failed to marshal child identifier fields")
		}

		// ANCHOR: - Step 9: Construct the final EventInstance model to be saved to the database.
		instance := models.EventInstance{
			Code:                    instanceCode,
			EventCode:               event.Code,
			Title:                   request.Title,
			Description:             request.Description,
			ParentIdentifierFields:  models.JSONB(parentIdentifierFields),
			ChildIdentifierFields:   models.JSONB(childIdentifierFields),
			StartAt:                 instanceTimes[0].In(common.GetLocation()),
			EndAt:                   instanceTimes[1].In(common.GetLocation()),
			RegisterStartAt:         instanceTimes[2].In(common.GetLocation()),
			RegisterEndAt:           instanceTimes[3].In(common.GetLocation()),
			VerifyStartAt:           instanceTimes[4].In(common.GetLocation()),
			VerifyEndAt:             instanceTimes[5].In(common.GetLocation()),
			LocationType:            request.Location.Type,
			LocationOfflineVenue:    request.Location.OfflineVenue,
			LocationOnlineLink:      request.Location.OnlineLink,
			LocationDetail:          request.Location.Detail,
			Timezone:                request.TimeConfig.Timezone,
			Capacity:                request.RegistrationConfig.Capacity,
			QuotaPerUser:            request.RegistrationConfig.QuotaPerUser,
			EnforceCommunityId:      request.RegistrationConfig.EnforceCommunityId,
			EnforceUniqueness:       request.RegistrationConfig.EnforceUniqueness,
			EnforceSelfRegistration: request.RegistrationConfig.EnforceSelfRegistration,
			Methods:                 request.RegistrationConfig.Methods,
			Flow:                    request.RegistrationConfig.Flow,
			Status:                  instanceStatus,
		}
		instances = append(instances, instance)

		// ANCHOR: - Step 10: If requested, update the parent event's start and end times to match this instance's times.
		if request.IsUpdateEventTime {
			if event.StartAt != instanceTimes[0].In(common.GetLocation()) && event.EndAt != instanceTimes[1].In(common.GetLocation()) {
				event.StartAt = instanceTimes[0].In(common.GetLocation())
				event.EndAt = instanceTimes[1].In(common.GetLocation())

				if err := eiu.r.Event.Update(ctx, event); err != nil {
					return nil, errorgen.Error(err)
				}
			}
		}

		// ANCHOR: - Step 11: Build the response object for this newly created instance.
		response = append(response, models.CreateInstanceResponse{
			Type:         models.TYPE_EVENT_INSTANCE,
			InstanceCode: instanceCode,
			EventCode:    instance.EventCode,
			Title:        instance.Title,
			Description:  instance.Description,
			IdentifierConfig: models.InstanceIdentifierConfigResponse{
				ParentIdentifierFields: request.IdentifierConfig.ParentIdentifierFields,
				ChildIdentifierFields:  request.IdentifierConfig.ChildIdentifierFields,
			},
			TimeConfig: models.InstanceTimeConfigResponse{
				StartAt:         instance.StartAt.Format(time.RFC3339),
				EndAt:           instance.EndAt.Format(time.RFC3339),
				RegisterStartAt: instance.RegisterStartAt.Format(time.RFC3339),
				RegisterEndAt:   instance.RegisterEndAt.Format(time.RFC3339),
				VerifyStartAt:   instance.VerifyStartAt.Format(time.RFC3339),
				VerifyEndAt:     instance.VerifyEndAt.Format(time.RFC3339),
				Timezone:        instance.Timezone,
			},
			Location: models.EventLocationResponse{
				Type:         instance.LocationType,
				OfflineVenue: instance.LocationOfflineVenue,
				OnlineLink:   instance.LocationOnlineLink,
				Detail:       instance.LocationDetail,
			},
			RegistrationConfig: models.InstanceRegistrationConfigResponse{
				Capacity:                instance.Capacity,
				QuotaPerUser:            instance.QuotaPerUser,
				EnforceCommunityId:      instance.EnforceCommunityId,
				EnforceUniqueness:       instance.EnforceUniqueness,
				EnforceSelfRegistration: instance.EnforceSelfRegistration,
				Methods:                 instance.Methods,
				Flow:                    instance.Flow,
			},
			Status: constants.MapStatus[constants.STATUS_ACTIVE],
		})
	}

	// ANCHOR: - Step 12: Check if this operation is part of a larger database transaction. If it is, the creation will be handled by the parent transaction. If not, it will create the instances now.
	isTransactionActive := eiu.r.Transaction.IsTransactionActive()
	if isTransactionActive {
		// LINK internal/usecases/event_instance_usecase.go:229
		// ANCHOR: - Step 13: If in a transaction, call createAll to handle the bulk insert and form creation.
		err = eiu.createAll(ctx, instances, requests, response)
		if err != nil {
			return nil, errorgen.Error(err, "failed to create event instance: %s", err.Error())
		}

		return response, nil
	}

	// If not in a transaction, create the instances and associated forms directly.
	err = eiu.createAll(ctx, instances, requests, response)
	if err != nil {
		return nil, errorgen.Error(err, "failed to create event instance: %s", err.Error())
	}

	return response, nil
}

// SECTION createAll handles the database insertion for instances and their associated forms.
func (eiu *eventInstanceUsecase) createAll(ctx context.Context, instances []models.EventInstance, requests []models.CreateInstanceRequest, response []models.CreateInstanceResponse) error {
	// ANCHOR: - Step 1: Bulk insert all the prepared event instances into the database.
	if err := eiu.r.EventInstance.BulkCreate(ctx, &instances); err != nil {
		return errorgen.Error(err)
	}

	for i, request := range requests {
		// ANCHOR: - Step 2: Validate that if questions are provided, the registration flow is appropriate (not direct QR methods).
		if (common.CheckOneDataInList([]string{models.RegisterFlowPersonal, models.RegisterFlowEvent}, request.RegistrationConfig.Methods) || request.RegistrationConfig.Flow == "direct") && request.Questions != nil {
			return errorgen.Error(errorgen.ErrHaveToUseRegistrationQrForQuestion)
		}

		// ANCHOR: - Step 3: If questions are included in the request, create a corresponding form and associate it with the new instance.
		if request.Questions != nil {
			form := models.CreateFormRequest{
				Name:        instances[i].Title,
				Description: instances[i].Description,
				Questions:   request.Questions,
				Entity: models.FormEntityRequest{
					Type: models.TYPE_EVENT_INSTANCE,
					Code: instances[i].Code,
				},
			}

			formRes, err := eiu.f.Create(ctx, &form)
			if err != nil {
				return errorgen.Error(err)
			}

			response[i].Questions = formRes.Questions
		}
	}
	return nil
}
