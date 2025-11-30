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

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type EventUsecase interface {
	Create(ctx context.Context, request models.CreateEventRequest, createdBy string) (response *models.CreateEventResponse, err error)
	GetAll(ctx context.Context, claims *models.TokenValues, params models.GetAllEventsParams) (responses *[]models.GetAllEventsResponse, err error)
	GetByCode(ctx context.Context, code string, roles []string, userTypes []string, communityId string) (detail *models.GetEventByCodeResponse, data []models.GetInstancesByEventCodeResponse, err error)
	GetQuestions(ctx context.Context, parameter models.GetEventQuestionParameter) (response *models.GetEventQuestionResponse, err error)
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

	// 1. Parse and validate event start and end times.
	eventTimes, _ := common.ParseMultipleTime([]string{request.TimeConfig.StartAt, request.TimeConfig.EndAt}, request.TimeConfig.Timezone, time.RFC3339)
	if eventTimes[0].After(eventTimes[1]) {
		return nil, errorgen.Error(errorgen.ErrInvalidDate, "start time cannot be later than end time")
	}

	// 2. Generate a unique event code and a URL-friendly slug.
	timeNowNano, _ := common.NowWithNanoTime()
	eventCode := generator.GenerateHashCode(fmt.Sprintf("event-%d-%d-%d", timeNowNano.UnixNano(), eventTimes[0].UnixNano(), eventTimes[1].UnixNano()), 7)
	// If no slug is provided, generate one from the event title.
	var slugExist bool
	if request.Slug != "" {
		// 3. Ensure the slug meets the minimum length requirement.
		if len(request.Slug) < 7 {
			// Calculate how many characters are needed to meet the minimum length of 7.
			neededChars := 7 - len(request.Slug)
			// Append a hyphen and a random string of the required length.
			request.Slug = fmt.Sprintf("%s-%s", request.Slug, generator.GenerateHashCode(time.Now().String(), neededChars))
		}
		// 4. Check for existing events with the same code or slug to prevent duplicates.
		var err error
		slugExist, err = eu.r.Event.CheckByCodeOrSlug(ctx, eventCode, request.Slug)
		if err != nil {
			return nil, err
		}
	} else {
		request.Slug = strings.ToLower(strings.ReplaceAll(request.Title, " ", "-"))
	}

	// If a slug or code already exists, return a conflict error.
	if slugExist {
		return nil, errorgen.Error(errorgen.AlreadyExist, "slug or code already exist")
	}

	// 5. Validate access configuration. For private events, at least one access control list must be defined.
	var allowedUsers, allowedRoles, allowedCampuses, allowedUserTypes []string
	switch request.AccessConfig.Visibility {
	case "public":
		break
	case "private":
		if request.AccessConfig.Campuses == nil && request.AccessConfig.CommunityIds == nil && request.AccessConfig.Roles == nil && request.AccessConfig.UserTypes == nil {
			return nil, errorgen.Error(errorgen.ErrMissingFields, "one of the fields is required for private events")
		}
	default:
		return nil, errorgen.Error(errorgen.ErrMissingFields, "one of the fields is required for private events")
	}

	// 6. Validate and assign access control lists (ACLs) if provided.
	if request.AccessConfig.Roles != nil {
		if err = eu.validatePrivateEventConstraint(ctx, request.AccessConfig.Roles, eu.r.Role.CheckMultiple, "roles"); err != nil {
			return nil, err
		}
		allowedRoles = request.AccessConfig.Roles
	}

	if request.AccessConfig.CommunityIds != nil {
		if err = eu.validatePrivateEventConstraint(ctx, request.AccessConfig.CommunityIds, eu.r.User.CheckMultiple, "users"); err != nil {
			return nil, err
		}
		allowedUsers = request.AccessConfig.CommunityIds
	}

	if request.AccessConfig.UserTypes != nil {
		if err = eu.validatePrivateEventConstraint(ctx, request.AccessConfig.UserTypes, eu.r.UserType.CheckMultiple, "user types"); err != nil {
			return nil, err
		}
		allowedUserTypes = request.AccessConfig.UserTypes
	}

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

	// 6. Validate and assign contact community ids if provided.
	var contactCommunityIds []string
	if request.ContactCommunityIds != nil {
		if err = eu.validatePrivateEventConstraint(ctx, request.ContactCommunityIds, eu.r.User.CheckMultiple, "contact community ids"); err != nil {
			return nil, err
		}
		contactCommunityIds = request.ContactCommunityIds
	}

	if err = eu.validatePrivateEventConstraint(ctx, request.OrganizerCommunityIds, eu.r.User.CheckMultiple, "organizer community ids"); err != nil {
		return nil, err
	}
	organizerCommunityIds := request.OrganizerCommunityIds

	if request.Location.Visibility == "" {
		request.Location.Visibility = "all"
	}

	// 7. Set the event status based on whether it's published.
	var eventStatus string
	if request.IsPublish {
		eventStatus = constants.MapEventStatus[constants.EVENT_STATUS_ACTIVE]
	} else {
		eventStatus = constants.MapEventStatus[constants.EVENT_STATUS_DRAFT]
		// If event is draft, all instances should be draft as well
		for i := range request.Instances {
			request.Instances[i].IsPublish = false
		}
	}

	if request.Category == string(constants.EventCategoryAnnouncement) {
		if request.RedirectURL == "" {
			return nil, errorgen.Error(errorgen.ErrMissingFields, "redirectUrl is required for announcement category")
		}
	}

	// 8. Construct the event model with the processed data.
	event := models.Event{
		Code:                  eventCode,
		Title:                 request.Title,
		Topics:                pq.StringArray(request.Topics),
		Category:              request.Category,
		Description:           request.Description,
		TermsAndConditions:    request.TermsAndConditions,
		ImageLinks:            pq.StringArray(request.ImageLinks),
		Slug:                  request.Slug,
		RedirectURL:           request.RedirectURL,
		CreatedBy:             createdBy,
		Visibility:            request.AccessConfig.Visibility,
		AllowedCommunityIds:   pq.StringArray(allowedUsers),
		AllowedRoles:          pq.StringArray(allowedRoles),
		AllowedCampuses:       pq.StringArray(allowedCampuses),
		AllowedUserTypes:      pq.StringArray(allowedUserTypes),
		ContactCommunityIds:   pq.StringArray(contactCommunityIds),
		OrganizerCommunityIds: pq.StringArray(organizerCommunityIds),
		IsRecurring:           request.TimeConfig.IsRecurring,
		StartAt:               eventTimes[0].In(common.GetLocation()),
		EndAt:                 eventTimes[1].In(common.GetLocation()),
		LocationType:          request.Location.Type,
		LocationOfflineVenue:  request.Location.OfflineVenue,
		LocationOnlineLink:    request.Location.OnlineLink,
		LocationDetail:        request.Location.Detail,
		LocationVisibility:    request.Location.Visibility,
		Timezone:              request.TimeConfig.Timezone,
		Status:                eventStatus,
	}

	// 9. Use a database transaction to create the event, its instances, and any associated forms atomically.
	var instanceRes []models.CreateInstanceResponse
	var questionRes []models.FormQuestionResponse
	err = eu.r.Transaction.Atomic(ctx, func(ctx context.Context, r *pgsql.PostgreRepositories) error {
		if err = r.Event.Create(ctx, &event); err != nil {
			return errorgen.Error(err, "failed to create event: %s", err.Error())
		}

		if len(request.Instances) > 0 {
			instanceRes, err = eu.ei.Create(ctx, &event, request.Instances)
			if err != nil {
				return errorgen.Error(err, "failed to create event instances: %s", err.Error())
			}
		}

		// If there are questions in the request, create a form for the event.
		if request.Questions != nil {
			form := models.CreateFormRequest{
				Name:        event.Title,
				Description: event.Description,
				Questions:   request.Questions,
				Entity: models.FormEntityRequest{
					Type: "event",
					Code: event.Code,
				},
			}

			formRes, err := eu.f.Create(ctx, &form)
			if err != nil {
				return errorgen.Error(errorgen.DataNotFound, "failed to create event form: %s", err.Error())
			}

			questionRes = formRes.Questions
		}
		return nil
	})

	if err != nil {
		return nil, errorgen.Error(err, "failed to create event atomically: %s", err.Error())
	}

	// 10. Build and return the successful response object.
	return &models.CreateEventResponse{
		Type:               models.TYPE_EVENT,
		Code:               event.Code,
		Title:              event.Title,
		Topics:             event.Topics,
		Category:           event.Category,
		Description:        event.Description,
		TermsAndConditions: event.TermsAndConditions,
		ImageLinks:         event.ImageLinks,
		Slug:               event.Slug,
		RedirectURL:        event.RedirectURL,
		AccessConfig: models.EventAccessConfigResponse{
			Visibility:   event.Visibility,
			CommunityIds: event.AllowedCommunityIds,
			Roles:        event.AllowedRoles,
			UserTypes:    event.AllowedUserTypes,
			Campuses:     event.AllowedCampuses,
		},
		TimeConfig: models.EventTimeConfigResponse{
			StartAt:  event.StartAt.Format(time.RFC3339),
			EndAt:    event.EndAt.Format(time.RFC3339),
			Timezone: event.Timezone,
		},
		Location: models.EventLocationResponse{
			Type:         event.LocationType,
			OfflineVenue: event.LocationOfflineVenue,
			OnlineLink:   event.LocationOnlineLink,
			Detail:       event.LocationDetail,
			Visibility:   event.LocationVisibility,
		},
		Status:    event.Status,
		Instances: instanceRes,
		Questions: questionRes,
	}, nil
}

// validatePrivateEventConstraint is a generic function to check if all provided IDs (e.g., for roles, users, user types)
// exist in the database. This is crucial for ensuring the integrity of access control lists for private events.
//
// Parameters:
//   - ctx: The context for the database operation.
//   - ids: A slice of strings representing the identifiers to be validated.
//   - checkFunc: A function that takes a context and a slice of IDs, and returns the count of existing records and an error.
//     This makes the function reusable for different entity types.
//   - entityName: A user-friendly string (e.g., "roles", "users") used in the error message if validation fails.
func (eu *eventUsecase) validatePrivateEventConstraint(ctx context.Context, ids []string, checkFunc func(context.Context, []string) (int64, error), entityName string) error {
	// If the list of IDs is nil, there's nothing to validate.
	if ids == nil {
		return nil
	}

	// Call the provided check function to query the database for the count of existing IDs.
	count, err := checkFunc(ctx, ids)
	if err != nil {
		return errorgen.Error(err)
	}

	// If the count of existing records does not match the number of IDs provided, return an error.
	if int(count) != len(ids) {
		return errorgen.Error(errorgen.DataNotFound, fmt.Sprintf("one of the %s don't exist", entityName))
	}

	return nil
}

func (eu *eventUsecase) GetAll(ctx context.Context, claims *models.TokenValues, params models.GetAllEventsParams) (response *[]models.GetAllEventsResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	params.AllowedRoles = claims.Roles
	params.AllowedUserTypes = claims.UserTypes
	params.AllowedCommunityIDs = []string{claims.Id}
	params.AllowedCampuses = []string{claims.CampusCode}

	if params.Status == "" {
		params.Status = constants.MapStatus[constants.STATUS_ACTIVE]
	}

	userTypeData, err := eu.r.UserType.GetByArray(ctx, params.AllowedUserTypes)
	if err != nil {
		return nil, err
	}
	params.UserTypes = userTypeData

	events, err := eu.r.Event.GetAllEvents(ctx, params)
	if err != nil {
		return nil, err
	}

	list := make([]models.GetAllEventsResponse, len(events))
	for i, e := range events {

		list[i] = models.GetAllEventsResponse{
			Type:               models.TYPE_EVENT,
			Code:               e.Code,
			Title:              e.Title,
			Topics:             e.Topics,
			Category:           e.Category,
			Description:        e.Description,
			TermsAndConditions: e.TermsAndConditions,
		}
	}

	return &list, nil
}

func (eu *eventUsecase) GetByCode(ctx context.Context, parameter *models.GetEventByCodeParameter, claims *models.TokenValues) (detail *models.GetEventByCodeResponse, data []models.GetInstancesByEventCodeResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	code := parameter.Code
	fmt.Println(code)
	codeLength := len(code)
	fmt.Println(codeLength)
	var eventCode, eventSlug string
	switch {
	case codeLength <= 7:
		eventCode, eventSlug = code, ""
	case codeLength > 7:
		eventCode, eventSlug = "", code
	default:
		return nil, nil, models.ErrorEventNotValid
	}

	event, err := eu.r.Event.GetOneByCodeOrSlug(ctx, eventCode, eventSlug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errorgen.Error(errorgen.DataNotFound)
		}
		return nil, nil, err
	}

	switch {
	case event == nil:
		return nil, nil, errorgen.Error(errorgen.DataNotFound)
	// case event.EventCode == "" || event.EventStatus != constants.EventStatusActive:
	case event.EventCode == "":
		return nil, nil, errorgen.Error(errorgen.DataNotFound)
	case event.EventVisibility != constants.EventVisibilityPublic:
		isAllowedRoles := common.CheckOneDataInList(event.EventAllowedRoles, claims.Roles)
		isAllowedUsers := common.CheckOneDataInList(event.EventAllowedUserTypes, claims.UserTypes)
		isAllowedCampuses := common.CheckOneDataInList(event.EventAllowedCampuses, []string{claims.CampusCode})

		if !isAllowedRoles && !isAllowedUsers && !isAllowedCampuses {
			return nil, nil, errorgen.Error(errorgen.ForbiddenRole)
		}
	case eventCode != "" && event.EventCode != eventCode:
		return nil, nil, models.ErrorEventNotValid
	case eventSlug != "" && event.EventSlug != eventSlug:
		return nil, nil, models.ErrorEventNotValid
	}

	createdByUser, err := eu.r.User.GetUserNameByCommunityId(ctx, event.EventCreatedBy)
	if err != nil {
		fmt.Println(err)
		return nil, nil, errorgen.Error(err)
	}

	contacts, err := eu.r.User.GetIdentifierByCommunityIds(ctx, event.EventContactCommunityIds)
	if err != nil {
		return nil, nil, errorgen.Error(err)
	}

	organizers, err := eu.r.User.GetIdentifierByCommunityIds(ctx, event.EventOrganizerCommunityIds)
	if err != nil {
		return nil, nil, errorgen.Error(err)
	}

	contactsRes := make([]models.UserIdentifierResponse, len(contacts))
	for i, p := range contacts {
		contactsRes[i] = models.UserIdentifierResponse{
			CommunityId: p.CommunityId,
			Name:        p.Name,
			Email:       p.Email,
			PhoneNumber: p.PhoneNumber,
		}
	}

	organizersRes := make([]models.UserIdentifierResponse, len(organizers))
	for i, p := range organizers {
		organizersRes[i] = models.UserIdentifierResponse{
			CommunityId: p.CommunityId,
			Name:        p.Name,
			Email:       p.Email,
			PhoneNumber: p.PhoneNumber,
		}
	}

	instancesRes := make([]models.GetInstancesByEventCodeResponse, len(event.Instances))
	for i, p := range event.Instances {
		var parentIdentifierFields, childIdentifierFields []models.IdentifierField
		if err := p.ParentIdentifierFields.Unmarshal(&parentIdentifierFields); err != nil {
			return nil, nil, errorgen.Error(err, "failed to unmarshal parent identifier fields")
		}
		if err := p.ChildIdentifierFields.Unmarshal(&childIdentifierFields); err != nil {
			return nil, nil, errorgen.Error(err, "failed to unmarshal child identifier fields")
		}

		instancesRes[i] = models.GetInstancesByEventCodeResponse{
			Type:        models.TYPE_EVENT_INSTANCE,
			Code:        p.Code,
			EventCode:   p.EventCode,
			Title:       p.Title,
			Description: p.Description,
			IdentifierConfig: models.InstanceIdentifierConfigResponse{
				ParentIdentifierFields: parentIdentifierFields,
				ChildIdentifierFields:  childIdentifierFields,
			},
			TimeConfig: models.InstanceTimeConfigResponse{
				StartAt:         common.FormatDatetimeToStringWithTimezone(p.StartAt, time.RFC3339, p.Timezone),
				EndAt:           common.FormatDatetimeToStringWithTimezone(p.EndAt, time.RFC3339, p.Timezone),
				RegisterStartAt: common.FormatDatetimeToStringWithTimezone(p.RegisterStartAt, time.RFC3339, p.Timezone),
				RegisterEndAt:   common.FormatDatetimeToStringWithTimezone(p.RegisterEndAt, time.RFC3339, p.Timezone),
				VerifyStartAt:   common.FormatDatetimeToStringWithTimezone(p.VerifyStartAt, time.RFC3339, p.Timezone),
				VerifyEndAt:     common.FormatDatetimeToStringWithTimezone(p.VerifyEndAt, time.RFC3339, p.Timezone),
				Timezone:        p.Timezone,
			},
			Location: models.EventLocationResponse{
				Type:         p.LocationType,
				OfflineVenue: p.LocationOfflineVenue,
				OnlineLink:   p.LocationOnlineLink,
				Detail:       p.LocationDetail,
			},
			RegistrationConfig: models.InstanceRegistrationConfigResponse{
				QuotaPerUser:       p.QuotaPerUser,
				Capacity:           p.Capacity,
				Flow:               p.Flow,
				Methods:            p.Methods,
				EnforceCommunityId: p.EnforceCommunityID,
				EnforceUniqueness:  p.EnforceUniqueness,
			},
		}
	}

	return &models.GetEventByCodeResponse{
		Type:               models.TYPE_EVENT,
		Code:               event.EventCode,
		Title:              event.EventTitle,
		Topics:             event.EventTopics,
		Category:           event.EventCategory,
		Description:        event.EventDescription,
		TermsAndConditions: event.EventTermsAndConditions,
		ImageLinks:         event.EventImageLinks,
		Slug:               event.EventSlug,
		RedirectURL:        event.EventRedirectURL,
		Contacts:           contactsRes,
		Organizers:         organizersRes,
		CreatedBy:          event.EventCreatedBy,
		CreatedByName:      createdByUser.Name,
		Location: models.EventLocationResponse{
			Type:         event.EventLocationType,
			OfflineVenue: event.EventLocationOfflineVenue,
			OnlineLink:   event.EventLocationOnlineLink,
			Detail:       event.EventLocationDetail,
		},
		AccessConfig: models.EventAccessConfigResponse{
			Visibility:   event.EventVisibility,
			CommunityIds: event.EventAllowedCommunityIds,
			UserTypes:    event.EventAllowedUserTypes,
			Roles:        event.EventAllowedRoles,
			Campuses:     event.EventAllowedCampuses,
		},
		TimeConfig: models.EventTimeConfigResponse{
			IsRecurring: event.EventIsRecurring,
			StartAt:     common.FormatDatetimeToString(event.EventStartAt, time.RFC3339),
			EndAt:       common.FormatDatetimeToString(event.EventEndAt, time.RFC3339),
		},
		Status:    event.EventStatus,
		Instances: instancesRes,
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

func (eu *eventUsecase) GetQuestions(ctx context.Context, parameter models.GetEventQuestionParameter) (response *models.GetEventQuestionResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	// Split the instance code to get the event code part.
	// Example: "EVT1234-INST567" -> "EVT1234"
	eventCodeParts := strings.Split(parameter.InstanceCode, "-")
	if len(eventCodeParts) < 1 {
		return nil, errorgen.Error(errorgen.InvalidInput, "invalid instance code format")
	}
	eventCode := eventCodeParts[0]
	eventExist, err := eu.r.Event.CheckByCode(ctx, eventCode)
	if err != nil {
		return nil, errorgen.Error(err)
	}

	if !eventExist {
		return nil, errorgen.Error(errorgen.DataNotFound)
	}

	instance, err := eu.r.EventInstance.GetByCode(ctx, parameter.InstanceCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorgen.Error(errorgen.DataNotFound)
		}
		return nil, errorgen.Error(err)
	}

	if instance.ID == 0 || instance.Status != constants.StatusActive {
		return nil, errorgen.Error(errorgen.DataNotFound)
	}

	questionEntityCodes := []models.FormQuestionEntityFilter{
		{
			Type: models.TYPE_EVENT,
			Code: eventCode,
		},
		{
			Type: models.TYPE_EVENT_INSTANCE,
			Code: parameter.InstanceCode,
		},
	}

	questions, err := eu.r.FormQuestion.GetByAssociationEntity(ctx, questionEntityCodes)
	if err != nil {
		return nil, err
	}

	// Use the centralized builder function to construct question lists.
	parentQuestions, childQuestions, err := buildEventQuestionLists(instance, questions)
	if err != nil {
		return nil, err
	}

	var parentIdentifierFields, childIdentifierFields []models.IdentifierField
	if err := instance.ParentIdentifierFields.Unmarshal(&parentIdentifierFields); err != nil {
		return nil, errorgen.Error(err, "failed to unmarshal parent identifier fields")
	}
	if err := instance.ChildIdentifierFields.Unmarshal(&childIdentifierFields); err != nil {
		return nil, errorgen.Error(err, "failed to unmarshal child identifier fields")
	}

	response = &models.GetEventQuestionResponse{
		EventCode:    instance.EventCode,
		InstanceCode: instance.Code,
		Title:        instance.Title,
		IdentifierConfig: models.InstanceIdentifierConfigResponse{
			ParentIdentifierFields: parentIdentifierFields,
			ChildIdentifierFields:  childIdentifierFields,
		},
		TimeConfig: models.InstanceTimeConfigResponse{
			StartAt:         common.FormatDatetimeToStringWithTimezone(instance.StartAt, time.RFC3339, instance.Timezone),
			EndAt:           common.FormatDatetimeToStringWithTimezone(instance.EndAt, time.RFC3339, instance.Timezone),
			RegisterStartAt: common.FormatDatetimeToStringWithTimezone(instance.RegisterStartAt, time.RFC3339, instance.Timezone),
			RegisterEndAt:   common.FormatDatetimeToStringWithTimezone(instance.RegisterEndAt, time.RFC3339, instance.Timezone),
			VerifyStartAt:   common.FormatDatetimeToStringWithTimezone(instance.VerifyStartAt, time.RFC3339, instance.Timezone),
			VerifyEndAt:     common.FormatDatetimeToStringWithTimezone(instance.VerifyEndAt, time.RFC3339, instance.Timezone),
			Timezone:        instance.Timezone,
		},
		Location: models.EventLocationResponse{
			Type:         instance.LocationType,
			OfflineVenue: instance.LocationOfflineVenue,
			OnlineLink:   instance.LocationOnlineLink,
			Detail:       instance.LocationDetail,
		},
		RegistrationConfig: models.InstanceRegistrationConfigResponse{
			QuotaPerUser:       instance.QuotaPerUser,
			Capacity:           instance.Capacity,
			Flow:               instance.Flow,
			Methods:            instance.Methods,
			EnforceCommunityId: instance.EnforceCommunityId,
			EnforceUniqueness:  instance.EnforceUniqueness,
		},
		ParentQuestions: parentQuestions,
		ChildQuestions:  childQuestions,
	}

	return response, nil
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

// buildQuestionLists constructs the question sets for parent and child attendees based on instance configuration and dynamic questions.
func buildEventQuestionLists(instance models.EventInstance, dbQuestions []models.FormQuestion) (parentQuestions []models.QuestionsResponse, childQuestions []models.QuestionsResponse, err error) {
	// 1. Add the default "Name" question, which is always mandatory.
	nameQuestion := models.QuestionsResponse{
		Type:         models.TYPE_FORM_QUESTION,
		EntityType:   models.TYPE_EVENT,
		Text:         "Please input your name",
		QuestionType: string(constants.QuestionTypeShortText),
		IsMandatory:  true,
	}
	parentQuestions = append(parentQuestions, nameQuestion)
	childQuestions = append(childQuestions, nameQuestion)

	// 2. Define configurations for standard identifier questions (email, phone).
	identifierConfigs := map[string]struct {
		text         string
		questionType constants.QuestionType
	}{
		"email": {text: "Please input valid e-mail address", questionType: constants.QuestionTypeEmail},
		"phone": {text: "Please input valid phone number", questionType: constants.QuestionTypePhone},
	}

	var parentIdentifierFields, childIdentifierFields []models.IdentifierField
	if err = instance.ParentIdentifierFields.Unmarshal(&parentIdentifierFields); err != nil {
		return nil, nil, errorgen.Error(err, "failed to unmarshal parent identifier fields")
	}
	if err = instance.ChildIdentifierFields.Unmarshal(&childIdentifierFields); err != nil {
		return nil, nil, errorgen.Error(err, "failed to unmarshal child identifier fields")
	}

	// 3. Loop through identifier configs to build and append questions dynamically.
	for idType, config := range identifierConfigs {
		question := models.QuestionsResponse{
			Type:         models.TYPE_FORM_QUESTION,
			EntityType:   models.TYPE_EVENT,
			Text:         config.text,
			QuestionType: string(config.questionType),
		}

		for _, field := range parentIdentifierFields {
			if field.Type == idType {
				q := question // Create a copy for the parent list
				q.IsMandatory = field.IsMandatory
				parentQuestions = append(parentQuestions, q)
				break
			}
		}

		for _, field := range childIdentifierFields {
			if field.Type == idType {
				q := question // Create a copy for the child list
				q.IsMandatory = field.IsMandatory
				childQuestions = append(childQuestions, q)
				break
			}
		}
	}

	// 4. Append dynamic questions fetched from the database.
	for _, q := range dbQuestions {
		question := models.QuestionsResponse{
			Type:         models.TYPE_FORM_QUESTION,
			Code:         q.Code,
			FormCode:     q.FormCode,
			EntityType:   models.TYPE_FORM_QUESTION,
			Text:         q.Text,
			QuestionType: string(q.Type),
			Options:      q.Options,
			Rules:        q.Rules,
		}
		// Add to parent list if applicable, creating a copy to set mandatory status independently.
		if common.CheckOneDataInList(q.ApplyFor, []string{"parent"}) {
			parentQ := question
			parentQ.IsMandatory = common.CheckOneDataInList(q.MandatoryFor, []string{"parent"})
			parentQuestions = append(parentQuestions, parentQ)
		}
		// Add to child list if applicable, creating a copy to set mandatory status independently.
		if common.CheckOneDataInList(q.ApplyFor, []string{"child"}) {
			childQ := question
			childQ.IsMandatory = common.CheckOneDataInList(q.MandatoryFor, []string{"child"})
			childQuestions = append(childQuestions, childQ)
		}
	}
	return
}
