package usecases

import (
	"context"
	"fmt"
	"go-community/internal/common"
	"go-community/internal/config"
	"go-community/internal/models"
	"go-community/internal/pkg/authorization"
	"go-community/internal/pkg/generator"
	"go-community/internal/repositories/pgsql"
	"time"
)

type EventInstanceUsecase interface {
	Create(ctx context.Context, request models.CreateInstanceExistingEventRequest) (response *models.CreateInstanceResponse, err error)
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

func (eiu *eventInstanceUsecase) Create(ctx context.Context, request models.CreateInstanceExistingEventRequest) (response *models.CreateInstanceResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	eventExist, err := eiu.r.Event.GetByCode(ctx, request.EventCode)
	if err != nil {
		return nil, err
	}

	if eventExist.ID == 0 {
		return nil, models.ErrorDataNotFound
	}

	timeNowNano, err := common.NowWithNanoTime()
	if err != nil {
		return nil, err
	}

	countInstance, err := eiu.r.EventInstance.CountByCode(ctx, request.EventCode)
	if err != nil {
		return nil, err
	}

	instanceStart, _ := time.Parse(time.RFC3339, request.InstanceStartAt)
	instanceEnd, _ := time.Parse(time.RFC3339, request.InstanceEndAt)
	instanceRegisterStart, _ := time.Parse(time.RFC3339, request.RegisterStartAt)
	instanceRegisterEnd, _ := time.Parse(time.RFC3339, request.RegisterEndAt)
	instanceAllowVerifyAt, _ := time.Parse(time.RFC3339, request.AllowVerifyAt)
	instanceDisallowVerifyAt, _ := time.Parse(time.RFC3339, request.DisallowVerifyAt)
	numberForCode := int(countInstance) + 1
	code := fmt.Sprintf("instance-%s-%d-%d", request.EventCode, numberForCode, timeNowNano.UnixNano())
	instanceCode := fmt.Sprintf("%s-%s", request.EventCode, generator.GenerateHashCode(code, 7))

	if instanceStart.After(instanceEnd) || instanceRegisterStart.After(instanceRegisterEnd) || instanceAllowVerifyAt.After(instanceDisallowVerifyAt) {
		return nil, models.ErrorStartDateLater
	}

	if request.RegisterFlow != models.MapRegisterFlow[models.REGISTER_FLOW_NONE] {
		if request.MaxPerTransaction == 0 {
			if eventExist.IsRecurring && request.IsOnePerTicket {
				request.MaxPerTransaction = 1
			}
			return nil, models.ErrorMaxPerTrxIsZero
		}

		if request.CheckType == "" {
			return nil, models.ErrorAttendanceTypeWhenRequired
		}
	} else {
		request.IsOnePerAccount = false
		request.IsOnePerTicket = false
		request.RegisterFlow = models.MapRegisterFlow[models.REGISTER_FLOW_NONE]
		request.MaxPerTransaction = 0
		request.CheckType = "none"
		request.TotalSeats = 0
	}

	instance := models.EventInstance{
		Code:              instanceCode,
		EventCode:         request.EventCode,
		Title:             request.Title,
		Description:       request.Description,
		InstanceStartAt:   instanceStart,
		InstanceEndAt:     instanceEnd,
		RegisterStartAt:   instanceRegisterStart,
		RegisterEndAt:     instanceRegisterEnd,
		AllowVerifyAt:     instanceAllowVerifyAt,
		DisallowVerifyAt:  instanceDisallowVerifyAt,
		LocationType:      request.LocationType,
		LocationName:      request.LocationName,
		MaxPerTransaction: request.MaxPerTransaction,
		IsOnePerAccount:   request.IsOnePerAccount,
		IsOnePerTicket:    request.IsOnePerTicket,
		RegisterFlow:      request.RegisterFlow,
		CheckType:         request.CheckType,
		TotalSeats:        request.TotalSeats,
		Status:            models.MapStatus[models.STATUS_ACTIVE],
	}

	if err := eiu.r.EventInstance.Create(ctx, &instance); err != nil {
		return nil, err
	}

	if request.IsUpdateEventTime {
		eventExist.EventStartAt = instanceStart
		eventExist.EventEndAt = instanceEnd
		eventExist.RegisterStartAt = instanceRegisterStart
		eventExist.RegisterEndAt = instanceRegisterEnd

		if err := eiu.r.Event.Update(ctx, &eventExist); err != nil {
			return nil, err
		}
	}

	res := models.CreateInstanceResponse{
		Type:              models.TYPE_EVENT_INSTANCE,
		InstanceCode:      instanceCode,
		EventCode:         request.EventCode,
		Title:             request.Title,
		Description:       request.Description,
		InstanceStartAt:   instanceStart,
		InstanceEndAt:     instanceEnd,
		RegisterStartAt:   instanceRegisterStart,
		RegisterEndAt:     instanceRegisterEnd,
		AllowVerifyAt:     instanceAllowVerifyAt,
		DisallowVerifyAt:  instanceDisallowVerifyAt,
		LocationType:      request.LocationType,
		LocationName:      request.LocationName,
		MaxPerTransaction: request.MaxPerTransaction,
		IsOnePerAccount:   request.IsOnePerAccount,
		IsOnePerTicket:    request.IsOnePerTicket,
		RegisterFlow:      request.RegisterFlow,
		CheckType:         request.CheckType,
		TotalSeats:        request.TotalSeats,
		Status:            models.MapStatus[models.STATUS_ACTIVE],
	}

	return &res, nil
}
