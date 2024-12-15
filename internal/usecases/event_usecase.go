package usecases

import (
	"context"
	"errors"
	"fmt"
	"go-community/internal/common"
	"go-community/internal/config"
	"go-community/internal/models"
	"go-community/internal/pkg/authorization"
	"go-community/internal/pkg/generator"
	"go-community/internal/repositories/pgsql"
	"gorm.io/gorm"
	"strings"
	"time"
)

type EventUsecase interface {
	Create(ctx context.Context, request models.CreateEventRequest) (response *models.CreateEventResponse, err error)
	GetAll(ctx context.Context, roles []string) (responses *[]models.GetAllEventsResponse, err error)
	GetByCode(ctx context.Context, code string) (response *models.GetEventByCodeResponse, err error)
	GetRegistered(ctx context.Context, communityIdOrigin string) (eventRegistrations []models.GetAllRegisteredUserResponse, err error)
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

	timeNowNano, err := common.NowWithNanoTime()
	if err != nil {
		return nil, err
	}

	eventStart, _ := time.Parse(time.RFC3339, request.EventStartAt)
	eventEnd, _ := time.Parse(time.RFC3339, request.EventEndAt)
	registerStart, _ := time.Parse(time.RFC3339, request.EventStartAt)
	registerEnd, _ := time.Parse(time.RFC3339, request.EventEndAt)

	code := fmt.Sprintf("event-%d-%d-%d", timeNowNano.UnixNano(), eventStart.UnixNano(), eventEnd.UnixNano())
	eventCode := generator.GenerateHashCode(code, 7)
	eventExist, err := eu.r.Event.CheckByCode(ctx, eventCode)
	if err != nil {
		return nil, err
	}

	if eventExist {
		return nil, models.ErrorAlreadyExist
	}

	var allowedRoles []string
	var allowedUsers []string
	var allowedCampuses []string
	switch {
	case request.AllowedFor == "public":
		allowedRoles = nil
		allowedUsers = nil
		allowedCampuses = nil
	case request.AllowedFor == "private" && request.AllowedCampuses != nil && request.AllowedRoles != nil && request.AllowedUsers != nil:
		countRole, err := eu.r.Role.CheckMultiple(ctx, request.AllowedRoles)
		if err != nil {
			return nil, err
		}

		if int(countRole) != len(request.AllowedRoles) {
			return nil, models.ErrorDataNotFound
		}

		countUserType, err := eu.r.UserType.CheckMultiple(ctx, request.AllowedUsers)
		if err != nil {
			return nil, err
		}

		if int(countUserType) != len(request.AllowedUsers) {
			return nil, models.ErrorDataNotFound
		}

		for i, str := range request.AllowedCampuses {
			request.AllowedCampuses[i] = strings.ToLower(str)
		}

		campusExist := common.CheckAllDataMapStructure(eu.cfg.Campus, request.AllowedCampuses)
		if !campusExist {
			return nil, models.ErrorDataNotFound
		}

		for i, str := range request.AllowedCampuses {
			request.AllowedCampuses[i] = strings.ToUpper(str)
		}

		allowedRoles = request.AllowedRoles
		allowedUsers = request.AllowedUsers
		allowedCampuses = request.AllowedCampuses
	default:
		return nil, models.ErrorViolateAllowedForPrivate
	}

	if eventStart.After(eventEnd) || registerStart.After(registerEnd) {
		return nil, models.ErrorStartDateLater
	}

	event := models.Event{
		Code:               eventCode,
		Title:              request.Title,
		Topics:             request.Topics,
		Description:        request.Description,
		TermsAndConditions: request.TermsAndConditions,
		AllowedFor:         request.AllowedFor,
		AllowedUsers:       allowedUsers,
		AllowedRoles:       allowedRoles,
		AllowedCampuses:    allowedCampuses,
		IsRecurring:        request.IsRecurring,
		Recurrence:         request.Recurrence,
		EventStartAt:       eventStart,
		EventEndAt:         eventEnd,
		RegisterStartAt:    registerStart,
		RegisterEndAt:      registerEnd,
		LocationType:       request.LocationType,
		LocationName:       request.LocationName,
		Status:             models.MapStatus[models.STATUS_ACTIVE],
	}

	instances := make([]models.EventInstance, 0)
	countInstance, err := eu.r.EventInstance.CountByCode(ctx, eventCode)
	if err != nil {
		return nil, err
	}

	if countInstance == 0 {
		countInstance = 1
	}

	for i, instanceRequest := range request.Instances {
		instanceStart, _ := time.Parse(time.RFC3339, instanceRequest.InstanceStartAt)
		instanceEnd, _ := time.Parse(time.RFC3339, instanceRequest.InstanceEndAt)
		instanceRegisterStart, _ := time.Parse(time.RFC3339, instanceRequest.RegisterStartAt)
		instanceRegisterEnd, _ := time.Parse(time.RFC3339, instanceRequest.RegisterEndAt)
		numberForCode := int(countInstance) + i
		code := fmt.Sprintf("instance-%s-%d-%d", eventCode, numberForCode, timeNowNano.UnixNano())
		instanceCode := fmt.Sprintf("%s-%s", eventCode, generator.GenerateHashCode(code, 7))

		if instanceStart.After(instanceEnd) || instanceRegisterStart.After(instanceRegisterEnd) {
			if instanceStart.After(instanceEnd) {

			}

			if instanceRegisterStart.After(instanceRegisterEnd) {

			}
			return nil, models.ErrorStartDateLater
		}

		if instanceRequest.RegisterFlow != models.MapRegisterFlow[models.REGISTER_FLOW_NONE] {
			if instanceRequest.MaxPerTransaction == 0 {
				if request.IsRecurring && instanceRequest.IsOnePerTicket {
					instanceRequest.MaxPerTransaction = 1
				}
				return nil, models.ErrorMaxPerTrxIsZero
			}

			if instanceRequest.CheckType == "" {
				return nil, models.ErrorAttendanceTypeWhenRequired
			}
		} else {
			instanceRequest.IsOnePerAccount = false
			instanceRequest.IsOnePerTicket = false
			instanceRequest.RegisterFlow = models.MapRegisterFlow[models.REGISTER_FLOW_NONE]
			instanceRequest.MaxPerTransaction = 0
			instanceRequest.CheckType = "none"
			instanceRequest.TotalSeats = 0
		}

		instance := models.EventInstance{
			Code:              instanceCode,
			EventCode:         eventCode,
			Title:             instanceRequest.Title,
			Description:       instanceRequest.Description,
			InstanceStartAt:   instanceStart,
			InstanceEndAt:     instanceEnd,
			RegisterStartAt:   instanceRegisterStart,
			RegisterEndAt:     instanceRegisterEnd,
			LocationType:      instanceRequest.LocationType,
			LocationName:      instanceRequest.LocationName,
			MaxPerTransaction: instanceRequest.MaxPerTransaction,
			IsOnePerAccount:   instanceRequest.IsOnePerAccount,
			IsOnePerTicket:    instanceRequest.IsOnePerTicket,
			RegisterFlow:      instanceRequest.RegisterFlow,
			CheckType:         instanceRequest.CheckType,
			TotalSeats:        instanceRequest.TotalSeats,
			Status:            models.MapStatus[models.STATUS_ACTIVE],
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
			Type:              models.TYPE_EVENT_INSTANCE,
			InstanceCode:      p.Code,
			EventCode:         p.EventCode,
			Title:             p.Title,
			Description:       p.Description,
			InstanceStartAt:   p.InstanceStartAt,
			InstanceEndAt:     p.InstanceEndAt,
			RegisterStartAt:   p.RegisterStartAt,
			RegisterEndAt:     p.RegisterEndAt,
			LocationType:      p.LocationType,
			LocationName:      p.LocationName,
			MaxPerTransaction: p.MaxPerTransaction,
			IsOnePerTicket:    p.IsOnePerTicket,
			IsOnePerAccount:   p.IsOnePerAccount,
			RegisterFlow:      p.RegisterFlow,
			CheckType:         p.CheckType,
			TotalSeats:        p.TotalSeats,
			Status:            p.Status,
		}
	}

	mainResponse := models.CreateEventResponse{
		Type:               models.TYPE_EVENT,
		Code:               event.Code,
		Title:              event.Title,
		Topics:             event.Topics,
		Description:        event.Description,
		TermsAndConditions: event.TermsAndConditions,
		AllowedFor:         event.AllowedFor,
		AllowedUsers:       event.AllowedUsers,
		AllowedRoles:       event.AllowedRoles,
		AllowedCampuses:    event.AllowedCampuses,
		IsRecurring:        event.IsRecurring,
		Recurrence:         event.Recurrence,
		EventStartAt:       event.EventStartAt,
		EventEndAt:         event.EventEndAt,
		RegisterStartAt:    event.RegisterStartAt,
		RegisterEndAt:      event.RegisterEndAt,
		LocationType:       event.LocationType,
		LocationName:       event.LocationName,
		Status:             event.Status,
		Instances:          instanceResponse,
	}

	return &mainResponse, nil
}

func (eu *eventUsecase) GetAll(ctx context.Context, roles []string, userTypes []string) (responses *[]models.GetAllEventsResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	events, err := eu.r.Event.GetAllByRolesAndUserTypes(ctx, roles, userTypes, models.MapStatus[models.STATUS_ACTIVE])
	if err != nil {
		return nil, err
	}

	list := make([]models.GetAllEventsResponse, len(events))
	for i, e := range events {
		availableStatus, err := models.DefineAvailabilityStatus(e)
		if err != nil {
			return nil, err
		}

		list[i] = models.GetAllEventsResponse{
			Type:                models.TYPE_EVENT,
			Code:                e.EventCode,
			Title:               e.EventTitle,
			Topics:              e.EventTopics,
			LocationType:        e.EventLocationType,
			AllowedFor:          e.EventAllowedFor,
			AllowedUsers:        e.EventAllowedUsers,
			AllowedRoles:        e.EventAllowedRoles,
			AllowedCampuses:     e.EventAllowedCampuses,
			IsRecurring:         e.EventIsRecurring,
			Recurrence:          e.EventRecurrence,
			EventStartAt:        e.EventStartAt,
			EventEndAt:          e.EventEndAt,
			RegisterStartAt:     e.EventRegisterStartAt,
			RegisterEndAt:       e.EventRegisterEndAt,
			TotalRemainingSeats: e.TotalRemainingSeats,
			AvailabilityStatus:  availableStatus,
		}
	}

	return &list, nil
}

func (eu *eventUsecase) GetByCode(ctx context.Context, code string, roles []string, userTypes []string) (response *models.GetEventByCodeResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	event, err := eu.r.Event.GetOneByCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, models.ErrorDataNotFound
		}
		return nil, err
	}

	if event.EventCode == "" || event.EventStatus != "active" {
		return nil, models.ErrorDataNotFound
	}

	if event.EventAllowedFor != "public" {
		isAllowedRoles := common.CheckOneDataInList(event.EventAllowedRoles, roles)
		isAllowedUsers := common.CheckOneDataInList(event.EventAllowedUsers, userTypes)

		if !isAllowedRoles && !isAllowedUsers {
			return nil, models.ErrorForbiddenRole
		}
	}

	availableStatus, err := models.DefineAvailabilityStatus(event)
	if err != nil {
		return nil, err
	}

	switch {
	case code != event.EventCode:
		return nil, models.ErrorEventNotValid
	case availableStatus == models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_UNAVAILABLE]:
		return nil, models.ErrorEventNotAvailable
	case availableStatus == models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_FULL]:
		return nil, models.ErrorRegisterQuotaNotAvailable
	case availableStatus == models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_SOON]:
		return nil, models.ErrorCannotRegisterYet
	}

	instances, err := eu.r.EventInstance.GetManyByEventCode(ctx, event.EventCode, models.MapStatus[models.STATUS_ACTIVE])
	if err != nil {
		return nil, err
	}

	if instances == nil {
		return nil, models.ErrorDataNotFound
	}

	instancesRes := make([]models.GetInstancesByEventCodeResponse, len(*instances))
	for i, p := range *instances {
		instanceAvailableStatus, err := models.DefineAvailabilityStatus(p)
		if err != nil {
			return nil, err
		}

		instancesRes[i] = models.GetInstancesByEventCodeResponse{
			Type:                models.TYPE_EVENT_INSTANCE,
			Code:                p.InstanceCode,
			Title:               p.InstanceTitle,
			Description:         p.InstanceDescription,
			InstanceStartAt:     p.InstanceStartAt,
			InstanceEndAt:       p.InstanceEndAt,
			RegisterStartAt:     p.InstanceRegisterStartAt,
			RegisterEndAt:       p.InstanceRegisterEndAt,
			LocationType:        p.InstanceLocationType,
			LocationName:        p.InstanceLocationName,
			MaxPerTransaction:   p.InstanceMaxPerTransaction,
			IsOnePerTicket:      p.InstanceIsOnePerTicket,
			IsOnePerAccount:     p.InstanceIsOnePerAccount,
			RegisterFlow:        p.InstanceRegisterFlow,
			CheckType:           p.InstanceCheckType,
			TotalSeats:          p.InstanceTotalSeats,
			BookedSeats:         p.InstanceBookedSeats,
			TotalRemainingSeats: p.TotalRemainingSeats,
			AvailabilityStatus:  instanceAvailableStatus,
		}
	}

	return &models.GetEventByCodeResponse{
		Type:               models.TYPE_EVENT,
		Code:               event.EventCode,
		Title:              event.EventTitle,
		Topics:             event.EventTopics,
		Description:        event.EventDescription,
		TermsAndConditions: event.EventTermsAndConditions,
		AllowedFor:         event.EventAllowedFor,
		AllowedUsers:       event.EventAllowedUsers,
		AllowedRoles:       event.EventAllowedRoles,
		AllowedCampuses:    event.EventAllowedCampuses,
		IsRecurring:        event.EventIsRecurring,
		Recurrence:         event.EventRecurrence,
		EventStartAt:       event.EventStartAt,
		EventEndAt:         event.EventEndAt,
		RegisterStartAt:    event.EventRegisterStartAt,
		RegisterEndAt:      event.EventRegisterEndAt,
		LocationType:       event.EventLocationType,
		LocationName:       event.EventLocationName,
		AvailabilityStatus: availableStatus,
		Instances:          instancesRes,
	}, nil
}

func (eu *eventUsecase) GetRegistered(ctx context.Context, communityIdOrigin string) (eventRegistrations []models.GetAllRegisteredUserResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	output, err := eu.r.Event.GetRegistered(ctx, communityIdOrigin)
	if err != nil {
		return nil, err
	}

	events := []models.GetAllRegisteredUserResponse{}
	for _, r := range output {
		e := models.GetAllRegisteredUserResponse{
			Type:               models.TYPE_EVENT,
			Code:               r.EventCode,
			Title:              r.EventTitle,
			Description:        r.EventDescription,
			TermsAndConditions: r.EventTermsAndConditions,
			StartAt:            r.EventStartAt,
			EndAt:              r.EventEndAt,
			LocationType:       r.EventLocationType,
			LocationName:       r.EventLocationName,
			Status:             r.EventStatus,
		}

		ei := models.InstancesForRegisteredRecordsResponse{
			Type:            models.TYPE_EVENT_INSTANCE,
			Code:            r.InstanceCode,
			Title:           r.InstanceTitle,
			Description:     r.InstanceDescription,
			InstanceStartAt: r.InstanceStartAt,
			InstanceEndAt:   r.InstanceEndAt,
			LocationType:    r.InstanceLocationType,
			LocationName:    r.InstanceLocationName,
			Status:          r.InstanceStatus,
		}

		var isPersonalQr bool
		if r.RegistrationRecordUpdatedBy == "user" {
			isPersonalQr = true
		}

		var verifiedAt string
		if !r.RegistrationRecordVerifiedAt.Time.IsZero() {
			verifiedAt = common.FormatDatetimeToString(r.RegistrationRecordVerifiedAt.Time, time.RFC3339)
		}

		rr := models.UserRegisteredRecordsResponse{
			Type:               models.TYPE_EVENT_REGISTRATION_RECORD,
			ID:                 r.RegistrationRecordID,
			Name:               r.RegistrationRecordName,
			Identifier:         r.RegistrationRecordIdentifier,
			CommunityId:        r.RegistrationRecordCommunityID,
			UpdatedBy:          r.RegistrationRecordUpdatedBy,
			RegisteredAt:       r.RegistrationRecordRegisteredAt,
			IsPersonalQr:       isPersonalQr,
			VerifiedAt:         verifiedAt,
			RegistrationStatus: r.RegistrationRecordStatus,
		}

		eventExist := false
		for j := range events {
			if events[j].Code == e.Code {
				instanceExist := false
				for k := range events[j].Instances {
					if events[j].Instances[k].Code == ei.Code {
						// Append registration record to the existing instance
						events[j].Instances[k].Registrants = append(events[j].Instances[k].Registrants, rr)
						instanceExist = true
						break
					}
				}

				// If instance doesn't exist, add it and include the registration
				if !instanceExist {
					ei.Registrants = append(ei.Registrants, rr)
					events[j].Instances = append(events[j].Instances, ei)
				}

				eventExist = true
				break
			}
		}

		if !eventExist {
			ei.Registrants = append(ei.Registrants, rr)
			e.Instances = append(e.Instances, ei)
			events = append(events, e)
		}
	}

	return events, nil
}
