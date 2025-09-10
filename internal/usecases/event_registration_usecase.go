package usecases

import (
	"context"
	"go-community/internal/common"
	"go-community/internal/config"
	"go-community/internal/constants"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
)

type EventRegistrationUsecase interface {
	Create(ctx context.Context, request models.CreateEventRegistrationRequest) (*models.EventRegistration, error)
}

type eventRegistrationUsecase struct {
	cfg                    *config.Configuration
	r                      pgsql.PostgreRepositories
	formAnswerUsecase      FormAnswerUsecase
	formAssociationUsecase FormAssociationUsecase
}

func NewEventRegistrationUsecase(cfg config.Configuration, r pgsql.PostgreRepositories, formAnswerUsecase FormAnswerUsecase, formAssociationUsecase FormAssociationUsecase) *eventRegistrationUsecase {
	return &eventRegistrationUsecase{
		cfg:                    &cfg,
		r:                      r,
		formAnswerUsecase:      formAnswerUsecase,
		formAssociationUsecase: formAssociationUsecase,
	}
}

func (eru *eventRegistrationUsecase) Create(ctx context.Context, request models.CreateEventRegistrationRequest) (response *models.EventRegistration, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	// 1. Basic input validation.
	if err = eru.validateRequestInput(request); err != nil {
		return nil, err
	}

	// 2. Handle validation specific to the 'personal-qr' method.
	if request.Method == "personal-qr" {
		if err = eru.validatePersonalQRMethod(ctx, request); err != nil {
			return nil, err
		}
	}

	// 3. Fetch event and instance data from the database.
	eventAndInstance, err := eru.r.Event.GetEventAndInstanceByCodes(ctx, request.EventCode, request.InstanceCode)
	if err != nil {
		return nil, err
	}
	if eventAndInstance == nil {
		return nil, models.ErrorDataNotFound
	}
	event, instance := divideEventAndInstance(eventAndInstance)

	// 4. Validate the event-specific rules.
	if err := eru.validateEvent(request, event); err != nil {
		return nil, err
	}

	// 5. Validate the instance-specific rules (timing, methods, etc.).
	if err := eru.validateInstance(request, instance); err != nil {
		return nil, err
	}

	// 6. For non-public events, validate user permissions.
	if event.Visibility != "public" && request.Method != "personal-qr" {
		if err := eru.validateUserPermissions(ctx, request, event); err != nil {
			return nil, err
		}
	}

	// All validations passed, proceed with registration logic (currently nil).
	return nil, nil
}

// validateRequestInput checks for basic request integrity.
func (eru *eventRegistrationUsecase) validateRequestInput(request models.CreateEventRegistrationRequest) error {
	// Ensure the instance code is derived from the event code.
	if request.EventCode != request.InstanceCode[:7] {
		return models.ErrorMismatchFields
	}
	return nil
}

// validatePersonalQRMethod handles checks specific to QR code registrations.
func (eru *eventRegistrationUsecase) validatePersonalQRMethod(ctx context.Context, request models.CreateEventRegistrationRequest) error {
	// Registrant must have a community ID.
	if request.Registrant.CommunityId == "" {
		return models.ErrorInvalidInput
	}

	// Check if the user exists.
	userExist, err := eru.r.User.GetByCommunityId(ctx, request.Registrant.CommunityId)
	if err != nil {
		return err
	}
	if &userExist == nil {
		return models.ErrorDataNotFound
	}

	// QR method only allows for a single attendee.
	if len(request.Attendees) > 1 {
		return models.ErrorQRForMoreThanOneRegister
	}

	return nil
}

// validateEvent checks if the event is valid and active.
func (eru *eventRegistrationUsecase) validateEvent(request models.CreateEventRegistrationRequest, event models.Event) error {
	// Event must exist and not be in a draft or inactive state.
	if event.ID == 0 || !common.CheckOneDataInList([]string{string(constants.EVENT_STATUS_DRAFT), string(constants.EVENT_STATUS_ACTIVE)}, []string{event.Status}) {
		return models.ErrorDataNotFound
	}
	// The event code in the request must match the retrieved event.
	if request.EventCode != event.Code {
		return models.ErrorEventNotValid
	}
	return nil
}

// validateInstance checks registration times, methods, and status for the event instance.
func (eru *eventRegistrationUsecase) validateInstance(request models.CreateEventRegistrationRequest, instance models.EventInstance) error {
	switch {
	// Instance must exist and be active.
	case instance.ID == 0 || !common.CheckOneDataInList([]string{string(constants.EVENT_STATUS_DRAFT), string(constants.EVENT_STATUS_ACTIVE)}, []string{instance.Status}):
		return models.ErrorDataNotFound
	// Registration method must be allowed by the instance.
	case !common.CheckOneDataInList([]string{request.Method}, instance.Methods):
		return models.ErrorDataNotFound
	// Registration start time must be before the end time.
	case instance.RegisterStartAt.After(instance.RegisterEndAt.In(common.GetLocation())):
		return models.ErrorRegistrationTimeDisabled
	// Cannot register before the registration window opens.
	case request.RegisterAt.Before(instance.RegisterStartAt.In(common.GetLocation())):
		return models.ErrorCannotRegisterYet
	// Cannot register after the registration window closes.
	case request.RegisterAt.After(instance.RegisterEndAt.In(common.GetLocation())):
		return models.ErrorRegistrationTimeDisabled
	// For QR codes, cannot register before the verification window opens.
	case request.Method == "personal-qr" && request.RegisterAt.Before(instance.VerifyStartAt.In(common.GetLocation())):
		return models.ErrorCannotRegisterYet
	// For QR codes, cannot register after the verification window closes.
	case request.Method == "personal-qr" && request.RegisterAt.After(instance.VerifyEndAt.In(common.GetLocation())):
		return models.ErrorRegistrationTimeDisabled
	}
	return nil
}

// validateUserPermissions checks if the registrant is allowed to register for a non-public event.
func (eru *eventRegistrationUsecase) validateUserPermissions(ctx context.Context, request models.CreateEventRegistrationRequest, event models.Event) error {
	userExist, err := eru.r.User.GetByCommunityId(ctx, request.Registrant.CommunityId)
	if err != nil {
		return err
	}
	if &userExist == nil {
		return models.ErrorDataNotFound
	}

	// Check if the user meets any of the allowed criteria (roles, user types, campus, or community).
	isAllowedRoles := common.CheckOneDataInList(event.AllowedRoles, userExist.Roles)
	isAllowedUsers := common.CheckOneDataInList(event.AllowedUserTypes, userExist.UserTypes)
	isAllowedCampus := common.CheckOneDataInList(event.AllowedCampuses, []string{userExist.CampusCode})
	isAllowedCommunity := common.CheckOneDataInList(event.AllowedCommunityIds, []string{userExist.CommunityID})

	if !isAllowedRoles && !isAllowedUsers && !isAllowedCampus && !isAllowedCommunity {
		return models.ErrorForbiddenRole
	}

	return nil
}

func divideEventAndInstance(eventAndInstance *models.GetEventAndInstanceByCodesDBOutput) (models.Event, models.EventInstance) {
	event := models.Event{
		ID:                   eventAndInstance.EventID,
		Code:                 eventAndInstance.EventCode,
		Title:                eventAndInstance.EventTitle,
		Topics:               eventAndInstance.EventTopics,
		Description:          eventAndInstance.EventDescription,
		TermsAndConditions:   eventAndInstance.EventTermsAndConditions,
		ImageLinks:           eventAndInstance.EventImageLinks,
		CreatedBy:            eventAndInstance.EventCreatedBy,
		LocationType:         eventAndInstance.EventLocationType,
		LocationOnlineLink:   eventAndInstance.EventLocationOnlineLink,
		LocationOfflineVenue: eventAndInstance.EventLocationOfflineVenue,
		Visibility:           eventAndInstance.EventVisibility,
		AllowedRoles:         eventAndInstance.EventAllowedRoles,
		AllowedUserTypes:     eventAndInstance.EventAllowedUserTypes,
		AllowedCampuses:      eventAndInstance.EventAllowedCampuses,
		AllowedCommunityIds:  eventAndInstance.EventAllowedCommunityIds,
		Recurrence:           eventAndInstance.EventRecurrence,
		StartAt:              eventAndInstance.EventStartAt,
		EndAt:                eventAndInstance.EventEndAt,
		PostDetails:          eventAndInstance.EventPostDetails,
		Status:               eventAndInstance.EventStatus,
	}

	instance := models.EventInstance{
		ID:                       eventAndInstance.InstanceID,
		Code:                     eventAndInstance.InstanceCode,
		Title:                    eventAndInstance.InstanceTitle,
		Description:              eventAndInstance.InstanceDescription,
		ValidateParentIdentifier: eventAndInstance.InstanceValidateParentIdentifier,
		ParentIdentifierInput:    eventAndInstance.InstanceParentIdentifierInput,
		ValidateChildIdentifier:  eventAndInstance.InstanceValidateChildIdentifier,
		ChildIdentifierInput:     eventAndInstance.InstanceChildIdentifierInput,
		EnforceCommunityId:       eventAndInstance.InstanceEnforceCommunityId,
		EnforceUniqueness:        eventAndInstance.InstanceEnforceUniqueness,
		Methods:                  eventAndInstance.InstanceMethods,
		Flow:                     eventAndInstance.InstanceFlow,
		StartAt:                  eventAndInstance.InstanceStartAt,
		EndAt:                    eventAndInstance.InstanceEndAt,
		RegisterStartAt:          eventAndInstance.InstanceRegisterStartAt,
		RegisterEndAt:            eventAndInstance.InstanceRegisterEndAt,
		VerifyStartAt:            eventAndInstance.InstanceVerifyStartAt,
		VerifyEndAt:              eventAndInstance.InstanceVerifyEndAt,
		Timezone:                 eventAndInstance.InstanceTimezone,
		LocationType:             eventAndInstance.InstanceLocationType,
		LocationOfflineVenue:     eventAndInstance.InstanceLocationOfflineVenue,
		LocationOnlineLink:       eventAndInstance.InstanceLocationOnlineLink,
		QuotaPerUser:             eventAndInstance.InstanceQuotaPerUser,
		Capacity:                 eventAndInstance.InstanceCapacity,
		PostDetails:              eventAndInstance.InstancePostDetails,
		Status:                   eventAndInstance.InstanceStatus,
	}

	return event, instance
}
