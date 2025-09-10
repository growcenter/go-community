package usecases

import (
	"context"
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
}

func NewEventInstanceUsecase(cfg config.Configuration, a authorization.Auth, r pgsql.PostgreRepositories) *eventInstanceUsecase {
	return &eventInstanceUsecase{
		cfg: &cfg,
		a:   a,
		r:   r,
	}
}

func (eiu *eventInstanceUsecase) Create(ctx context.Context, event *models.Event, requests []models.CreateInstanceRequest) (response []models.CreateInstanceResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	var instances []models.EventInstance
	for i, request := range requests {
		if event == nil {
			eventExist, err := eiu.r.Event.GetByCode(ctx, request.EventCode)
			if err != nil {
				return nil, err
			}

			if eventExist.ID == 0 {
				return nil, models.ErrorDataNotFound
			}

			event = &eventExist
		}

		var instanceStatus string
		if request.IsPublish {
			instanceStatus = string(constants.EVENT_STATUS_ACTIVE)
		} else {
			instanceStatus = string(constants.EVENT_STATUS_DRAFT)
		}

		countInstance, err := eiu.r.EventInstance.CountByCode(ctx, request.EventCode)
		if err != nil {
			return nil, err
		}

		instanceTimes, _ := common.ParseMultipleTime([]string{request.TimeConfig.StartAt, request.TimeConfig.EndAt, request.TimeConfig.RegisterStartAt, request.TimeConfig.RegisterEndAt, request.TimeConfig.VerifyStartAt, request.TimeConfig.VerifyEndAt}, "Asia/Jakarta", time.RFC3339)
		timeNowNano, _ := common.NowWithNanoTime()
		numberForCode := int(countInstance) + i
		instanceCode := fmt.Sprintf("%s-%s", request.EventCode, generator.GenerateHashCode(fmt.Sprintf("instance-%s-%d-%d", request.EventCode, numberForCode, timeNowNano.UnixNano()), 7))

		if instanceTimes[0].After(instanceTimes[1]) || instanceTimes[2].After(instanceTimes[3]) || instanceTimes[4].After(instanceTimes[5]) {
			return nil, models.ErrorStartDateLater
		}

		if len(request.RegistrationConfig.Methods) != 0 {
			if request.RegistrationConfig.Capacity == 0 {
				if event.Recurrence != "" && request.RegistrationConfig.EnforceUniqueness {
					request.RegistrationConfig.QuotaPerUser = 1
				}
				return nil, errorgen.Error(errorgen.ErrInvalidInput, "capacity cannot be zero")
			}

			if request.RegistrationConfig.Flow == "" {
				return nil, models.ErrorAttendanceTypeWhenRequired
			}
		} else {
			request.RegistrationConfig.EnforceCommunityId = false
			request.RegistrationConfig.EnforceUniqueness = false
			request.RegistrationConfig.Capacity = 0
			request.RegistrationConfig.Flow = "free"
			request.RegistrationConfig.QuotaPerUser = 0
		}

		instance := models.EventInstance{
			Code:                     instanceCode,
			EventCode:                request.EventCode,
			Title:                    request.Title,
			Description:              request.Description,
			ValidateParentIdentifier: request.IdentifierConfig.ValidateParentIdentifier,
			ParentIdentifierInput:    request.IdentifierConfig.ParentIdentifierInput,
			ValidateChildIdentifier:  request.IdentifierConfig.ValidateChildIdentifier,
			ChildIdentifierInput:     request.IdentifierConfig.ChildIdentifierInput,
			StartAt:                  instanceTimes[0].In(common.GetLocation()),
			EndAt:                    instanceTimes[1].In(common.GetLocation()),
			RegisterStartAt:          instanceTimes[2].In(common.GetLocation()),
			RegisterEndAt:            instanceTimes[3].In(common.GetLocation()),
			VerifyStartAt:            instanceTimes[4].In(common.GetLocation()),
			VerifyEndAt:              instanceTimes[5].In(common.GetLocation()),
			LocationType:             request.Location.Type,
			LocationOfflineVenue:     request.Location.OfflineVenue,
			LocationOnlineLink:       request.Location.OnlineLink,
			Timezone:                 request.TimeConfig.Timezone,
			Capacity:                 request.RegistrationConfig.Capacity,
			QuotaPerUser:             request.RegistrationConfig.QuotaPerUser,
			EnforceCommunityId:       request.RegistrationConfig.EnforceCommunityId,
			EnforceUniqueness:        request.RegistrationConfig.EnforceUniqueness,
			Methods:                  request.RegistrationConfig.Methods,
			Flow:                     request.RegistrationConfig.Flow,
			Status:                   instanceStatus,
		}
		instances = append(instances, instance)

		if request.IsUpdateEventTime {
			if event.StartAt != instanceTimes[0].In(common.GetLocation()) && event.EndAt != instanceTimes[1].In(common.GetLocation()) {
				event.StartAt = instanceTimes[0].In(common.GetLocation())
				event.EndAt = instanceTimes[1].In(common.GetLocation())

				if err := eiu.r.Event.Update(ctx, event); err != nil {
					return nil, err
				}
			}
		}

		response = append(response, models.CreateInstanceResponse{
			Type:         models.TYPE_EVENT_INSTANCE,
			InstanceCode: instanceCode,
			EventCode:    instance.EventCode,
			Title:        instance.Title,
			Description:  instance.Description,
			IdentifierConfig: models.InstanceIdentifierConfigResponse{
				ValidateParentIdentifier: instance.ValidateParentIdentifier,
				ParentIdentifierInput:    instance.ParentIdentifierInput,
				ValidateChildIdentifier:  instance.ValidateChildIdentifier,
				ChildIdentifierInput:     instance.ChildIdentifierInput,
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
			},
			RegistrationConfig: models.InstanceRegistrationConfigResponse{
				Capacity:           instance.Capacity,
				QuotaPerUser:       instance.QuotaPerUser,
				EnforceCommunityId: instance.EnforceCommunityId,
				EnforceUniqueness:  instance.EnforceUniqueness,
				Methods:            instance.Methods,
				Flow:               instance.Flow,
			},
			Status: constants.MapStatus[constants.STATUS_ACTIVE],
		})
	}

	if err := eiu.r.EventInstance.BulkCreate(ctx, &instances); err != nil {
		return nil, err
	}

	return response, nil
}
