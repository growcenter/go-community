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
	GetTitles(ctx context.Context) (eventTitles []models.GetEventTitlesResponse, err error)
	GetSummary(ctx context.Context, code string) (detail *models.GetEventSummaryResponse, data []models.GetInstanceSummaryResponse, err error)
}

type eventUsecase struct {
	cfg  *config.Configuration
	a    authorization.Auth
	r    pgsql.PostgreRepositories
	flag FeatureFlagUsecase
}

func NewEventUsecase(cfg config.Configuration, a authorization.Auth, r pgsql.PostgreRepositories, flag FeatureFlagUsecase) *eventUsecase {
	return &eventUsecase{
		cfg:  &cfg,
		a:    a,
		r:    r,
		flag: flag,
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

	if eventStart.After(eventEnd) {
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
		instanceAllowVerifyAt, _ := time.Parse(time.RFC3339, instanceRequest.AllowVerifyAt)
		instanceDisallowVerifyAt, _ := time.Parse(time.RFC3339, instanceRequest.DisallowVerifyAt)

		numberForCode := int(countInstance) + i
		code := fmt.Sprintf("instance-%s-%d-%d", eventCode, numberForCode, timeNowNano.UnixNano())
		instanceCode := fmt.Sprintf("%s-%s", eventCode, generator.GenerateHashCode(code, 7))

		if instanceStart.After(instanceEnd) || instanceRegisterStart.After(instanceRegisterEnd) || instanceAllowVerifyAt.After(instanceDisallowVerifyAt) {
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
			AllowVerifyAt:     instanceAllowVerifyAt,
			DisallowVerifyAt:  instanceDisallowVerifyAt,
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
			AllowVerifyAt:     p.AllowVerifyAt,
			DisallowVerifyAt:  p.DisallowVerifyAt,
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

	userTypeInfos, err := eu.r.UserType.GetByArray(ctx, userTypes)
	if err != nil {
		return nil, err
	}

	isNotGeneral := common.ContainsValueInModel(userTypeInfos, func(userType models.UserType) bool {
		return userType.Category == "internal" || userType.Category == "cool"
	})

	events, err := eu.r.Event.GetAllByRolesAndUserTypes(ctx, roles, userTypes, isNotGeneral, models.MapStatus[models.STATUS_ACTIVE])
	if err != nil {
		return nil, err
	}

	enableUniversalFlag, err := eu.flag.IsFeatureEnabled(ctx, "event_be_universaleventaccess", "")
	if err != nil {
		return nil, err
	}

	list := make([]models.GetAllEventsResponse, len(events))
	if !enableUniversalFlag {
		for i, e := range events {
			availableStatus, err := models.DefineAvailabilityStatus(e)
			if err != nil {
				return nil, err
			}

			if e.InstanceTotalSeats == 0 {
				e.TotalRemainingSeats = 0
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
				RegisterStartAt:     &e.EventRegisterStartAt,
				RegisterEndAt:       &e.EventRegisterEndAt,
				TotalRemainingSeats: e.TotalRemainingSeats,
				ImagesLinks:         e.EventImageLinks,
				AvailabilityStatus:  availableStatus,
			}
		}

		return &list, nil
	}

	for i, e := range events {
		list[i] = models.GetAllEventsResponse{
			Type:            models.TYPE_EVENT,
			Code:            e.EventCode,
			Title:           e.EventTitle,
			Topics:          e.EventTopics,
			LocationType:    e.EventLocationType,
			AllowedFor:      e.EventAllowedFor,
			AllowedUsers:    e.EventAllowedUsers,
			AllowedRoles:    e.EventAllowedRoles,
			AllowedCampuses: e.EventAllowedCampuses,
			EventStartAt:    e.EventStartAt,
			EventEndAt:      e.EventEndAt,
			ImagesLinks:     e.EventImageLinks,
		}
	}

	return &list, nil

}

func (eu *eventUsecase) GetByCode(ctx context.Context, code string, roles []string, userTypes []string) (detail *models.GetEventByCodeResponse, data []models.GetInstancesByEventCodeResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	event, err := eu.r.Event.GetOneByCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, models.ErrorDataNotFound
		}
		return nil, nil, err
	}

	switch {
	case event == nil:
		return nil, nil, models.ErrorDataNotFound
	case event.EventCode == "" || event.EventStatus != "active":
		return nil, nil, models.ErrorDataNotFound
	case event.EventAllowedFor != "public":
		isAllowedRoles := common.CheckOneDataInList(event.EventAllowedRoles, roles)
		isAllowedUsers := common.CheckOneDataInList(event.EventAllowedUsers, userTypes)

		if !isAllowedRoles && !isAllowedUsers {
			return nil, nil, models.ErrorForbiddenRole
		}
	case code != event.EventCode:
		return nil, nil, models.ErrorEventNotValid
	}

	enableUniversalFlag, err := eu.flag.IsFeatureEnabled(ctx, "event_be_universaleventaccess", "")
	if err != nil {
		return nil, nil, err
	}

	if !enableUniversalFlag {
		availableStatus, err := models.DefineAvailabilityStatus(event)
		if err != nil {
			return nil, nil, err
		}

		switch {
		case code != event.EventCode:
			return nil, nil, models.ErrorEventNotValid
		case availableStatus == models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_UNAVAILABLE]:
			return nil, nil, models.ErrorEventNotAvailable
		case availableStatus == models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_FULL]:
			return nil, nil, models.ErrorRegisterQuotaNotAvailable
		case availableStatus == models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_SOON]:
			return nil, nil, models.ErrorCannotRegisterYet
		}
	}

	instances, err := eu.r.EventInstance.GetManyByEventCode(ctx, event.EventCode, models.MapStatus[models.STATUS_ACTIVE])
	if err != nil {
		return nil, nil, err
	}

	if instances == nil {
		return nil, nil, models.ErrorDataNotFound
	}

	enableShowDescriptionFlag, err := eu.flag.IsFeatureEnabled(ctx, "event_be_showdecriptionatinstance", "")
	if err != nil {
		return nil, nil, err
	}

	instancesRes := make([]models.GetInstancesByEventCodeResponse, len(*instances))
	if !enableShowDescriptionFlag {
		for i, p := range *instances {
			instanceAvailableStatus, err := models.DefineAvailabilityStatus(p)
			if err != nil {
				return nil, nil, err
			}

			instancesRes[i] = models.GetInstancesByEventCodeResponse{
				Type:                models.TYPE_EVENT_INSTANCE,
				Code:                p.InstanceCode,
				Title:               p.InstanceTitle,
				Description:         "",
				InstanceStartAt:     p.InstanceStartAt,
				InstanceEndAt:       p.InstanceEndAt,
				RegisterStartAt:     p.InstanceRegisterStartAt,
				RegisterEndAt:       p.InstanceRegisterEndAt,
				AllowVerifyAt:       p.InstanceAllowVerifyAt,
				DisallowVerifyAt:    p.InstanceDisallowVerifyAt,
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
			ImageLinks:         event.EventImageLinks,
		}, instancesRes, nil
	}

	for i, p := range *instances {
		instanceAvailableStatus, err := models.DefineAvailabilityStatus(p)
		if err != nil {
			return nil, nil, err
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
			AllowVerifyAt:       p.InstanceAllowVerifyAt,
			DisallowVerifyAt:    p.InstanceDisallowVerifyAt,
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
		ImageLinks:         event.EventImageLinks,
	}, instancesRes, nil
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
			ImageLinks:         r.EventImageLinks,
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

func (eu *eventUsecase) GetTitles(ctx context.Context) (eventTitles []models.GetEventTitlesResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	output, err := eu.r.Event.GetTitles(ctx)
	if err != nil {
		return nil, err
	}

	var res []models.GetEventTitlesResponse
	for _, i := range output {
		res = append(res, i.ToResponse())
	}

	return res, nil
}

func (eu *eventUsecase) GetSummary(ctx context.Context, code string) (detail *models.GetEventSummaryResponse, data []models.GetInstanceSummaryResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	event, err := eu.r.Event.GetSummary(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, models.ErrorDataNotFound
		}
		return nil, nil, err
	}

	if event == nil {
		return nil, nil, models.ErrorDataNotFound
	}

	switch event.EventAllowedFor {
	case "public":
		publicCount, err := eu.r.User.CountUserByUserTypeCategory(ctx, []string{"general", "cool", "internal"})
		if err != nil {
			return nil, nil, err
		}

		event.TotalUsers = int(publicCount)
	case "private":
		privateCount, err := eu.r.User.CountUserByUserTypeCategory(ctx, []string{"cool", "internal"})
		if err != nil {
			return nil, nil, err
		}

		event.TotalUsers = int(privateCount)
	default:
		return nil, nil, models.ErrorEventNotValid
	}

	instances, err := eu.r.EventInstance.GetSummary(ctx, event.EventCode)
	if err != nil {
		return nil, nil, err
	}

	var instanceRes []models.GetInstanceSummaryResponse
	for _, i := range instances {
		var totalRemainingSeats int
		switch {
		case event.EventAllowedFor == "private" && i.InstanceTotalSeats == 0:
			totalRemainingSeats = event.TotalUsers - i.InstanceBookedSeats
		case event.EventAllowedFor == "public" && i.InstanceTotalSeats == 0:
			totalRemainingSeats = event.TotalUsers - i.InstanceBookedSeats
		default:
			totalRemainingSeats = i.TotalRemainingSeats
		}

		i.AttendancePercentage = float64(i.InstanceScannedSeats) / float64(event.TotalUsers) * 100

		instanceRes = append(instanceRes, models.GetInstanceSummaryResponse{
			Type:                models.TYPE_EVENT_INSTANCE,
			EventCode:           event.EventCode,
			Code:                i.InstanceCode,
			Title:               i.InstanceTitle,
			RegisterFlow:        i.InstanceRegisterFlow,
			CheckType:           i.InstanceCheckType,
			TotalSeats:          i.InstanceTotalSeats,
			BookedSeats:         i.InstanceBookedSeats,
			ScannedSeats:        i.InstanceScannedSeats,
			TotalRemainingSeats: totalRemainingSeats,
			AttendPercentage:    i.AttendancePercentage,
			MaxPerTransaction:   i.InstanceMaxPerTransaction,
			Status:              i.InstanceStatus,
		})
	}

	return event.ToResponse(), instanceRes, nil
}
