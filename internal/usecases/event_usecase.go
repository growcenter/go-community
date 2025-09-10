package usecases

import (
	"context"
	"errors"
	"fmt"
	"go-community/internal/common"
	"go-community/internal/config"
	"go-community/internal/constants"
	"go-community/internal/models"
	"go-community/internal/pkg/authorization"
	"go-community/internal/pkg/errorgen"
	"go-community/internal/pkg/generator"
	"go-community/internal/repositories/pgsql"
	"strings"
	"time"

	"gorm.io/gorm"
)

type EventUsecase interface {
	Create(ctx context.Context, request models.CreateEventRequest) (response *models.CreateEventResponse, err error)
	GetAll(ctx context.Context, roles []string) (responses *[]models.GetAllEventsResponse, err error)
	GetByCode(ctx context.Context, code string, roles []string, userTypes []string, communityId string) (detail *models.GetEventByCodeResponse, data []models.GetInstancesByEventCodeResponse, err error)
	GetRegistered(ctx context.Context, communityIdOrigin string) (eventRegistrations []models.GetAllRegisteredUserResponse, err error)
	GetTitles(ctx context.Context) (eventTitles []models.GetEventTitlesResponse, err error)
	GetSummary(ctx context.Context, code string) (detail *models.GetEventSummaryResponse, data []models.GetInstanceSummaryResponse, err error)
}

type eventUsecase struct {
	cfg  *config.Configuration
	a    authorization.Auth
	r    pgsql.PostgreRepositories
	flag FeatureFlagUsecase
	ei   EventInstanceUsecase
	f    FormUsecase
}

func NewEventUsecase(cfg config.Configuration, a authorization.Auth, r pgsql.PostgreRepositories, flag FeatureFlagUsecase, ei EventInstanceUsecase, f FormUsecase) *eventUsecase {
	return &eventUsecase{
		cfg:  &cfg,
		a:    a,
		r:    r,
		flag: flag,
		ei:   ei,
		f:    f,
	}
}

func (eu *eventUsecase) Create(ctx context.Context, request models.CreateEventRequest, createdBy string) (response *models.CreateEventResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	eventTimes, _ := common.ParseMultipleTime([]string{request.TimeConfig.StartAt, request.TimeConfig.EndAt}, "Asia/Jakarta", time.RFC3339)
	if eventTimes[0].After(eventTimes[1]) {
		return nil, errorgen.Error(errorgen.ErrInvalidDate, "start time cannot be later than end time")
	}

	timeNowNano, _ := common.NowWithNanoTime()
	eventCode := generator.GenerateHashCode(fmt.Sprintf("event-%d-%d-%d", timeNowNano.UnixNano(), eventTimes[0].UnixNano(), eventTimes[1].UnixNano()), 7)
	eventExist, err := eu.r.Event.CheckByCode(ctx, eventCode)
	if err != nil {
		return nil, err
	}

	if eventExist {
		return nil, errorgen.Error(errorgen.AlreadyExist, "event code already exist")
	}

	var allowedUsers, allowedRoles, allowedCampuses, allowedUserTypes []string
	switch request.AccessConfig.Visibility {
	case "public":
		break
	case "private":
		if request.AccessConfig.Campuses == nil && request.AccessConfig.CommunityIds == nil && request.AccessConfig.Roles == nil && request.AccessConfig.UserTypes == nil {
			return nil, errorgen.Error(errorgen.ErrMissingFields, "one of the fields is required for private events")
		}

		if err := eu.validatePrivateEventConstraint(ctx, request.AccessConfig.Roles, eu.r.Role.CheckMultiple, "roles"); err != nil {
			return nil, err
		}
		allowedRoles = request.AccessConfig.Roles

		if err := eu.validatePrivateEventConstraint(ctx, request.AccessConfig.CommunityIds, eu.r.User.CheckMultiple, "users"); err != nil {
			return nil, err
		}
		allowedUsers = request.AccessConfig.CommunityIds

		if err := eu.validatePrivateEventConstraint(ctx, request.AccessConfig.UserTypes, eu.r.UserType.CheckMultiple, "user types"); err != nil {
			return nil, err
		}
		allowedUserTypes = request.AccessConfig.UserTypes

		if request.AccessConfig.Campuses != nil {
			lowerCampuses := make([]string, len(request.AccessConfig.Campuses))
			for i, c := range request.AccessConfig.Campuses {
				lowerCampuses[i] = strings.ToLower(c)
			}

			if !common.CheckAllDataMapStructure(eu.cfg.Campus, lowerCampuses) {
				return nil, errorgen.Error(errorgen.DataNotFound, "one of the campuses don't exist")
			}
			allowedCampuses = request.AccessConfig.Campuses
		}
	default:
		return nil, errorgen.Error(errorgen.ErrMissingFields, "one of the fields is required for private events")
	}

	var eventStatus string
	if request.IsPublish {
		eventStatus = string(constants.EVENT_STATUS_ACTIVE)
	} else {
		eventStatus = string(constants.EVENT_STATUS_DRAFT)
	}

	event := models.Event{
		Code:                 eventCode,
		Title:                request.Title,
		Topics:               request.Topics,
		Description:          request.Description,
		TermsAndConditions:   request.TermsAndConditions,
		ImageLinks:           request.ImageLinks,
		RedirectLink:         request.RedirectLink,
		CreatedBy:            createdBy,
		Visibility:           request.AccessConfig.Visibility,
		AllowedCommunityIds:  allowedUsers,
		AllowedRoles:         allowedRoles,
		AllowedCampuses:      allowedCampuses,
		AllowedUserTypes:     allowedUserTypes,
		Recurrence:           request.TimeConfig.Recurrence,
		StartAt:              eventTimes[0].In(common.GetLocation()),
		EndAt:                eventTimes[1].In(common.GetLocation()),
		LocationType:         request.Location.Type,
		LocationOfflineVenue: request.Location.OfflineVenue,
		LocationOnlineLink:   request.Location.OnlineLink,
		Status:               eventStatus,
	}

	var instanceRes []models.CreateInstanceResponse
	var questionRes []models.FormQuestionResponse
	err = eu.r.Transaction.Atomic(ctx, func(ctx context.Context, r *pgsql.PostgreRepositories) error {
		if err := eu.r.Event.Create(ctx, &event); err != nil {
			return nil
		}

		instanceRes, err = eu.ei.Create(ctx, &event, request.Instances)
		if err != nil {
			return nil
		}

		if request.Questions != nil {
			form := models.CreateFormRequest{
				Name:        event.Title,
				Description: event.Description,
				Questions:   request.Questions,
			}

			formRes, err := eu.f.Create(ctx, &form)
			if err != nil {
				return nil
			}

			questionRes = formRes.Questions
		}
		return nil
	})

	return &models.CreateEventResponse{
		Type:               models.TYPE_EVENT,
		Code:               event.Code,
		Title:              event.Title,
		Topics:             event.Topics,
		Description:        event.Description,
		TermsAndConditions: event.TermsAndConditions,
		ImageLinks:         event.ImageLinks,
		RedirectLink:       event.RedirectLink,
		AccessConfig: models.EventAccessConfigResponse{
			Visibility:   event.Visibility,
			CommunityIds: event.AllowedCommunityIds,
			Roles:        event.AllowedRoles,
			UserTypes:    event.AllowedUserTypes,
			Campuses:     event.AllowedCampuses},
		TimeConfig: models.EventTimeConfigResponse{
			StartAt:    event.StartAt.Format(time.RFC3339),
			EndAt:      event.EndAt.Format(time.RFC3339),
			Recurrence: event.Recurrence,
		},
		Location: models.EventLocationResponse{
			Type:         event.LocationType,
			OfflineVenue: event.LocationOfflineVenue,
			OnlineLink:   event.LocationOnlineLink,
		},
		Status:    event.Status,
		Instances: instanceRes,
		Questions: questionRes,
	}, nil
}

// validatePrivateEventConstraint checks if the provided IDs exist in the database.
func (eu *eventUsecase) validatePrivateEventConstraint(ctx context.Context, ids []string, checkFunc func(context.Context, []string) (int64, error), entityName string) error {
	if ids == nil {
		return nil
	}

	count, err := checkFunc(ctx, ids)
	if err != nil {
		return errorgen.Error(err)
	}

	if int(count) != len(ids) {
		return errorgen.Error(errorgen.DataNotFound, fmt.Sprintf("one of the %s don't exist", entityName))
	}

	return nil
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

	events, err := eu.r.Event.GetAllByRolesAndUserTypes(ctx, roles, userTypes, isNotGeneral, constants.MapStatus[constants.STATUS_ACTIVE])
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

	instances, err := eu.r.EventInstance.GetManyByEventCode(ctx, event.EventCode, constants.MapStatus[constants.STATUS_ACTIVE])
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
