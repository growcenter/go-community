package usecases

import (
	"context"
	"fmt"
	"go-community/internal/common"
	"go-community/internal/config"
	"go-community/internal/models"
	"go-community/internal/pkg/authorization"
	"go-community/internal/repositories/pgsql"
	"strings"
	"time"
)

type EventUsecase interface {
	Create(ctx context.Context, request models.CreateEventRequest) (response *models.CreateEventResponse, err error)
	GetAll(ctx context.Context, roles []string) (responses *[]models.GetAllEventsResponse, err error)
}

type eventUsecase struct {
	cfg *config.Configuration
	a   authorization.Auth
	r   pgsql.PostgreRepositories
}

func NewEventUsecase(cfg config.Configuration, a authorization.Auth, r pgsql.PostgreRepositories) *eventUsecase {
	return &eventUsecase{
		cfg: &cfg,
		a:   a,
		r:   r,
	}
}

func (eu *eventUsecase) Create(ctx context.Context, request models.CreateEventRequest) (response *models.CreateEventResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	eventExist, err := eu.r.Event.CheckByCode(ctx, request.Code)
	if err != nil {
		return nil, err
	}

	if eventExist {
		return nil, models.ErrorAlreadyExist
	}

	countRole, err := eu.r.Role.CheckMultiple(ctx, request.AllowedRoles)
	if err != nil {
		return nil, err
	}

	if int(countRole) != len(request.AllowedRoles) {
		return nil, models.ErrorDataNotFound
	}

	for i, str := range request.CampusCode {
		request.CampusCode[i] = strings.ToLower(str)
	}

	campusExist := common.CheckDataMapStructure(eu.cfg.Campus, request.CampusCode)
	if !campusExist {
		return nil, models.ErrorDataNotFound
	}

	eventStart, _ := time.Parse(time.RFC3339, request.EventStartAt)
	eventEnd, _ := time.Parse(time.RFC3339, request.EventEndAt)
	registerStart, _ := time.Parse(time.RFC3339, request.EventStartAt)
	registerEnd, _ := time.Parse(time.RFC3339, request.EventEndAt)

	if eventStart.After(eventEnd) || registerStart.After(registerEnd) {
		return nil, models.ErrorStartDateLater
	}

	for i, str := range request.CampusCode {
		request.CampusCode[i] = strings.ToUpper(str)
	}

	event := models.Event{
		Code:            common.StringTrimSpaceAndUpper(request.Code),
		Title:           request.Title,
		Location:        request.Location,
		Description:     request.Description,
		CampusCode:      request.CampusCode,
		AllowedRoles:    request.AllowedRoles,
		IsRecurring:     request.IsRecurring,
		Recurrence:      request.Recurrence,
		EventStartAt:    eventStart,
		EventEndAt:      eventEnd,
		RegisterStartAt: registerStart,
		RegisterEndAt:   registerEnd,
		Status:          "active",
	}

	instances := make([]models.EventInstance, 0)
	countInstance, err := eu.r.EventInstance.CountByCode(ctx, request.Code)
	if err != nil {
		return nil, err
	}

	if countInstance == 0 {
		countInstance = 1
	}

	for i, instanceRequest := range request.Instances {
		numberForCode := int(countInstance) + i
		instanceStart, _ := time.Parse(time.RFC3339, instanceRequest.InstanceStartAt)
		instanceEnd, _ := time.Parse(time.RFC3339, instanceRequest.InstanceEndAt)
		instanceRegisterStart, _ := time.Parse(time.RFC3339, instanceRequest.RegisterStartAt)
		instanceRegisterEnd, _ := time.Parse(time.RFC3339, instanceRequest.RegisterEndAt)
		instanceCode := fmt.Sprintf("%s-%d", common.StringTrimSpaceAndUpper(request.Code), numberForCode)

		instance := models.EventInstance{
			Code:            instanceCode,
			Title:           instanceRequest.Title,
			Location:        instanceRequest.Location,
			EventCode:       common.StringTrimSpaceAndUpper(request.Code),
			InstanceStartAt: instanceStart,
			InstanceEndAt:   instanceEnd,
			RegisterStartAt: instanceRegisterStart,
			RegisterEndAt:   instanceRegisterEnd,
			Description:     instanceRequest.Description,
			MaxRegister:     instanceRequest.MaxRegister,
			TotalSeats:      instanceRequest.TotalSeats,
			IsRequired:      instanceRequest.IsRequired,
			Status:          "active",
		}

		instances = append(instances, instance)
	}

	if err := eu.r.Event.Create(ctx, &event); err != nil {
		return nil, err
	}

	if err = eu.r.EventInstance.BulkCreate(ctx, &instances); err != nil {
		return nil, err
	}

	instanceResponse := make([]models.CreateInstanceResponse, len(instances))
	for i, p := range instances {
		instanceResponse[i] = models.CreateInstanceResponse{
			Type:            models.TYPE_EVENT_INSTANCE,
			InstanceCode:    p.Code,
			Title:           p.Title,
			Location:        p.Location,
			EventCode:       common.StringTrimSpaceAndUpper(request.Code),
			InstanceStartAt: p.InstanceStartAt,
			InstanceEndAt:   p.InstanceEndAt,
			RegisterStartAt: p.RegisterStartAt,
			RegisterEndAt:   p.RegisterEndAt,
			Description:     p.Description,
			MaxRegister:     p.MaxRegister,
			TotalSeats:      p.TotalSeats,
			IsRequired:      p.IsRequired,
		}
	}

	mainResponse := models.CreateEventResponse{
		Type:            models.TYPE_EVENT,
		Code:            common.StringTrimSpaceAndUpper(request.Code),
		Title:           request.Title,
		Location:        request.Location,
		Description:     request.Description,
		CampusCode:      request.CampusCode,
		AllowedRoles:    request.AllowedRoles,
		IsRecurring:     request.IsRecurring,
		Recurrence:      request.Recurrence,
		EventStartAt:    eventStart,
		EventEndAt:      eventEnd,
		RegisterStartAt: registerStart,
		RegisterEndAt:   registerEnd,
		Instances:       instanceResponse,
	}

	return &mainResponse, nil
}

func (eu *eventUsecase) GetAll(ctx context.Context, roles []string) (responses *[]models.GetAllEventsResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()
	fmt.Println("roles are ", roles)
	events, err := eu.r.Event.GetAllByRoles(ctx, roles, "active")
	if err != nil {
		return nil, err
	}

	fmt.Println("events is ", events)

	list := make([]models.GetAllEventsResponse, len(events))
	for i, e := range events {
		var availableStatus string
		switch {
		case e.TotalRemainingSeats <= 0 && e.InstanceIsRequired == true && e.EventIsRecurring == false:
			availableStatus = "full"
		case common.Now().Before(e.EventRegisterStartAt.In(common.GetLocation())):
			availableStatus = "soon"
		case common.Now().After(e.EventRegisterEndAt.In(common.GetLocation())):
			availableStatus = "unavailable"
		default:
			availableStatus = "available"
		}

		list[i] = models.GetAllEventsResponse{
			Type:               models.TYPE_EVENT,
			Code:               e.EventCode,
			Title:              e.EventTitle,
			Location:           e.EventLocation,
			CampusCode:         e.EventCampusCode,
			IsRecurring:        e.EventIsRecurring,
			Recurrence:         e.EventRecurrence,
			EventStartAt:       e.EventStartAt,
			EventEndAt:         e.EventEndAt,
			RegisterStartAt:    e.EventRegisterStartAt,
			RegisterEndAt:      e.EventRegisterEndAt,
			AvailabilityStatus: availableStatus,
		}
	}

	return &list, nil
}
