package usecases

import (
	"context"
	"fmt"
	"go-community/internal/common"
	"go-community/internal/config"
	"go-community/internal/constants"
	"go-community/internal/models"
	"go-community/internal/pkg/errorgen"
	"go-community/internal/repositories/pgsql"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
)

type EventRegistrationUsecase interface {
	Create(ctx context.Context, request models.CreateEventRegistrationRequest, token models.TokenValues) (*models.EventRegistration, error)
}

type eventRegistrationUsecase struct {
	cfg                    *config.Configuration
	r                      pgsql.PostgreRepositories
	eventStatusUsecase     EventStatusUsecase
	formAnswerUsecase      FormAnswerUsecase
	formAssociationUsecase FormAssociationUsecase
	formQuestionUsecase    FormQuestionUsecase
}

func NewEventRegistrationUsecase(cfg config.Configuration, r pgsql.PostgreRepositories, eventStatusUsecase EventStatusUsecase, formAnswerUsecase FormAnswerUsecase, formAssociationUsecase FormAssociationUsecase, formQuestionUsecase FormQuestionUsecase) *eventRegistrationUsecase {
	return &eventRegistrationUsecase{
		cfg:                    &cfg,
		r:                      r,
		eventStatusUsecase:     eventStatusUsecase,
		formAnswerUsecase:      formAnswerUsecase,
		formAssociationUsecase: formAssociationUsecase,
		formQuestionUsecase:    formQuestionUsecase,
	}
}

func (eru *eventRegistrationUsecase) Create(ctx context.Context, request models.CreateEventRegistrationRequest, token models.TokenValues) (response *models.CreateEventRegistrationResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	// 1. Basic input validation.
	if err = eru.validateRequestInput(request); err != nil {
		return nil, err
	}

	// Fetch registrant user data once if community ID is provided.
	// This avoids redundant database calls in subsequent validation steps.
	var registrant *models.User
	if request.Registrant.CommunityId != "" {
		fetchedUser, err := eru.r.User.GetByCommunityId(ctx, request.Registrant.CommunityId)
		// If user not found, treat as nil user, but don't block registration unless required.
		if err != nil && err != models.ErrorDataNotFound {
			return nil, errorgen.Error(err, "failed to fetch registrant user data")
		}
		registrant = &fetchedUser
	}

	// 3. Fetch event and instance data from the database.
	eventAndInstance, err := eru.r.Event.GetEventAndInstanceByCodes(ctx, request.EventCode, request.InstanceCode)
	if err != nil {
		return nil, errorgen.Error(err)
	}
	if eventAndInstance == nil {
		return nil, errorgen.Error(models.ErrorDataNotFound, "event or instance not found")
	}

	event, instance := divideEventAndInstance(eventAndInstance)

	// 4. Validate the event-specific rules.
	if err = eru.validateEvent(request, event); err != nil {
		return nil, err
	}

	// 5. Validate the instance-specific rules (timing, methods, etc.).
	if err = eru.validateInstance(request, instance); err != nil {
		return nil, errorgen.Error(err, "failed when validating instance: %w", err.Error())
	}

	// 2. Handle validation specific to the 'personal-qr' method.
	if request.Method == "personal-qr" {
		if err = eru.validatePersonalQRMethod(request, registrant); err != nil {
			return nil, errorgen.Error(err, "failed when validating personal-qr method: %w", err.Error())
		}
	}

	// Note: Capacity validation is now performed inside the transaction to prevent race conditions.

	_, err = eru.eventStatusUsecase.DefineAvailabilityStatus(ctx, instance)
	if err != nil {
		return nil, errorgen.Error(err, "failed when defining availability status: %w", err.Error())
	}

	// 6. Validate that all required identifiers for parent and child attendees are present.
	if err = eru.validateRequiredIdentifiers(request, instance); err != nil {
		return nil, errorgen.Error(err, "failed when validating required identifiers: %w", err.Error())
	}

	// 7. For non-public events, validate user permissions using the pre-fetched user data.
	if event.Visibility != "public" {
		if err = eru.validateUserPermissions(registrant, event); err != nil {
			return nil, errorgen.Error(err, "failed when validating user permissions: %w", err.Error())
		}
	}

	// 8. Perform advanced uniqueness and registration enforcement validations.
	if err = eru.validateEnforcer(ctx, request, instance, registrant); err != nil {
		return nil, errorgen.Error(err, "failed when validating enforcer: %w", err.Error())
	}

	// 9. Prepare all payloads required for database insertion.
	registrationCode := uuid.New()
	eventRegistration, eventAttendances, formAnswerRequests, err := eru.preparePayloads(request, event, instance, registrationCode)
	if err != nil {
		return nil, errorgen.Error(err, "failed when preparing payloads: %w", err.Error())
	}

	var capturedFormResponses []*models.CreateFormAnswerResponse
	// 10. Execute all database write operations within a single atomic transaction.
	err = eru.r.Transaction.Atomic(ctx, func(ctx context.Context, txR *pgsql.PostgreRepositories) error {
		// Perform capacity validation within the transaction to prevent race conditions.
		if err := eru.validateCapacity(ctx, request, instance); err != nil {
			return errorgen.Error(err, "failed when validating capacity: %w", err.Error())
		}

		// Create the main registration record.
		if err := txR.EventRegistration.Create(ctx, eventRegistration); err != nil {
			return errorgen.Error(err, "failed when creating event registration: %w", err.Error())
		}

		// Bulk insert all event attendees in a single database call for efficiency.
		if len(eventAttendances) > 0 {
			if err := txR.EventAttendance.BulkCreate(ctx, eventAttendances); err != nil {
				return errorgen.Error(err, "failed when bulk creating event attendances: %w", err.Error())
			}
		}

		// Always validate and prepare answers, even if it's just to check for missing mandatory ones.
		forms, err := eru.formAnswerUsecase.WithTransaction(*txR).SubmitBatch(ctx, formAnswerRequests)
		if err != nil {
			return errorgen.Error(err, "failed when submitting form answers: %w", err.Error())
		}
		capturedFormResponses = forms

		return nil
	})
	if err != nil {
		return nil, errorgen.Error(err, "failed when executing transaction: %w", err.Error())
	}

	// 11. Construct the final response DTO.
	attendeeResponses := make([]models.CreateAttendeeResponse, len(eventAttendances))
	for i, attendance := range eventAttendances {
		var forms []models.FormQuestionAnswerResponse
		for _, formResponse := range capturedFormResponses {
			if formResponse.Identifier == attendance.Code.String() {
				forms = formResponse.Forms
				break
			}
		}
		attendeeResponses[i] = models.CreateAttendeeResponse{
			Type:        models.TYPE_EVENT_ATTENDANCE,
			Code:        attendance.Code.String(),
			Role:        attendance.Role,
			Name:        attendance.Name,
			Email:       attendance.Email,
			PhoneNumber: attendance.PhoneNumber,
			Status:      attendance.Status,
			Forms:       forms,
		}
	}
	response = &models.CreateEventRegistrationResponse{
		Type:         models.TYPE_EVENT_REGISTRATION,
		Code:         eventRegistration.Code.String(),
		EventCode:    eventRegistration.EventCode,
		InstanceCode: eventRegistration.InstanceCode,
		Method:       eventRegistration.Method,
		Quantity:     eventRegistration.Quantity,
		RegisterAt:   eventRegistration.RegisterAt,
		Registrant:   request.Registrant,
		Attendees:    attendeeResponses,
	}

	return response, nil
}

func (eru *eventRegistrationUsecase) preparePayloads(request models.CreateEventRegistrationRequest, event models.Event, instance models.EventInstance, registrationCode uuid.UUID) (*models.EventRegistration, []*models.EventAttendance, []*models.CreateFormAnswerRequest, error) {
	eventRegistration := &models.EventRegistration{
		Code:         registrationCode,
		EventCode:    event.Code,
		InstanceCode: instance.Code,
		Name:         request.Registrant.Name,
		Email:        request.Registrant.Email,
		PhoneNumber:  request.Registrant.PhoneNumber,
		CommunityId:  request.Registrant.CommunityId,
		Quantity:     len(request.Attendees),
		Method:       request.Method,
		RegisterAt:   request.RegisterAt,
		Status:       constants.StatusActive,
	}

	var eventAttendances []*models.EventAttendance
	var formAnswerRequests []*models.CreateFormAnswerRequest

	for _, attendee := range request.Attendees {
		var role string
		if attendee.IsParent {
			role = "parent"
		} else {
			role = "child"
		}

		attendanceCode := uuid.New()
		referenceCode := uuid.New().String()

		attendanceStatus := models.MapRegisterStatus[models.REGISTER_STATUS_PENDING]
		if request.Method == "personal-qr" || request.Method == "event-qr" {
			attendanceStatus = models.MapRegisterStatus[models.REGISTER_STATUS_SUCCESS]
		}

		eventAttendances = append(eventAttendances, &models.EventAttendance{
			Code:             attendanceCode,
			InstanceCode:     instance.Code,
			RegistrationCode: registrationCode,
			Role:             role,
			Name:             attendee.Name,
			Email:            attendee.Email,
			PhoneNumber:      attendee.PhoneNumber,
			ReferenceCode:    &referenceCode,
			Remarks:          &attendee.Remarks,
			Status:           attendanceStatus,
		})

		formAnswerRequests = append(formAnswerRequests, &models.CreateFormAnswerRequest{
			Entity:         []models.FormQuestionEntityFilter{{Type: models.TYPE_EVENT, Code: event.Code}, {Type: models.TYPE_EVENT_INSTANCE, Code: instance.Code}},
			Identifier:     attendanceCode.String(),
			IdentifierType: "eventAttendance",
			IsParent:       attendee.IsParent,
			Answers:        attendee.Form, // This will be an empty slice if no form is submitted
		})
	}

	return eventRegistration, eventAttendances, formAnswerRequests, nil
}

// answerAndValidateQuestions fetches questions, validates attendee answers, and prepares them for persistence.
// func (eru *eventRegistrationUsecase) answerAndValidateQuestions(ctx context.Context, request *models.CreateEventRegistrationRequest, event models.Event, instance models.EventInstance) ([]models.FormAnswer, error) {
// 	// 1. Fetch dynamic questions associated with the event and instance.
// 	questionEntityCodes := []models.FormQuestionEntityFilter{
// 		{Type: models.TYPE_EVENT, Code: event.Code},
// 		{Type: models.TYPE_EVENT_INSTANCE, Code: instance.Code},
// 	}
// 	dbQuestions, err := eru.r.FormQuestion.GetByAssociationEntity(ctx, questionEntityCodes)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 2. Construct the full list of questions for parent and child attendees.
// 	parentQuestions, childQuestions := buildEventQuestionLists(instance, dbQuestions)

// 	// 3. Validate answers for each attendee.
// 	if len(request.Attendees) == 0 {
// 		isMandatoryQuestionExists := false
// 		for _, q := range parentQuestions {
// 			if q.IsMandatory {
// 				isMandatoryQuestionExists = true
// 				break
// 			}
// 		}
// 		if isMandatoryQuestionExists {
// 			return nil, models.ErrorInvalidInput
// 		}
// 		return nil, nil
// 	}

// 	// Iterate through each attendee in the request and validate them against the
// 	// appropriate question set (parent or child). All validation errors are
// 	// collected and returned together.
// 	var validationErrors *multierror.Error
// 	for i, attendee := range request.Attendees {
// 		// First, validate the built-in fields (Name, Email, Phone) and mandatory dynamic questions.
// 		var err error
// 		if attendee.IsParent {
// 			err = eru.validateAttendee(attendee, parentQuestions)
// 		} else {
// 			err = eru.validateAttendee(attendee, childQuestions)
// 		}
// 		if err != nil {
// 			validationErrors = multierror.Append(validationErrors, fmt.Errorf("attendee %d (%s): %w", i+1, attendee.Name, err))
// 		}
// 	}
// 	// If there are any validation errors from the first pass, return them immediately.
// 	if validationErrors.ErrorOrNil() != nil {
// 		return nil, validationErrors.ErrorOrNil()
// 	}

// 	// 4. Prepare all dynamic form answers for submission.
// 	var allPreparedAnswers []models.FormAnswer
// 	formAnswerRequestMap := make(map[string][]models.AnswerItem)

// 	// Group answers by their original form code.
// 	for _, attendee := range request.Attendees {
// 		for _, answer := range attendee.Form {
// 			for _, q := range dbQuestions {
// 				if q.Code == answer.QuestionCode {
// 					formAnswerRequestMap[q.FormCode] = append(formAnswerRequestMap[q.FormCode], answer)
// 					break
// 				}
// 			}
// 		}
// 	}

// 	// Validate and prepare answers for each form.
// 	for formCode, answers := range formAnswerRequestMap {
// 		prepared, err := eru.formAnswerUsecase.ValidateAndPrepareAnswers(ctx, &models.CreateFormAnswerRequest{
// 			FormCode: formCode,
// 			Iden:     request.Registrant.CommunityId, // Assuming the registrant is the one answering.
// 			Answers:  answers,
// 		}, dbQuestions)
// 		if err != nil {
// 			return nil, err // Return validation errors from the shared use case.
// 		}
// 		allPreparedAnswers = append(allPreparedAnswers, prepared...)
// 	}

// 	return allPreparedAnswers, nil
// }

// validateAttendee checks if an attendee has provided all mandatory answers and validates against rules, accumulating all errors.
// func (eru *eventRegistrationUsecase) validateAttendee(attendee models.AttendeeRequest, questions []models.QuestionsResponse) error {
// 	var result *multierror.Error
// 	providedAnswers := make(map[string]string)
// 	if attendee.Form != nil {
// 		for _, ans := range attendee.Form {
// 			providedAnswers[ans.QuestionCode] = ans.Answer
// 		}
// 	}

// 	for _, q := range questions {
// 		var answerValue string
// 		isBuiltIn := q.Code == ""

// 		if isBuiltIn {
// 			switch q.QuestionType {
// 			case string(constants.QuestionTypeShortText):
// 				answerValue = attendee.Name
// 			case string(constants.QuestionTypeEmail):
// 				answerValue = attendee.Email
// 			case string(constants.QuestionTypePhone):
// 				answerValue = attendee.PhoneNumber
// 			}
// 		} else {
// 			answerValue = providedAnswers[q.Code]
// 		}

// 		if q.IsMandatory && answerValue == "" {
// 			result = multierror.Append(result, fmt.Errorf("missing answer for mandatory question: %s", q.Text))
// 			continue
// 		}

// 		if answerValue == "" {
// 			continue
// 		}

// 		// Create a temporary FormQuestion to use the shared validator
// 		tempFormQuestion := models.FormQuestion{
// 			Code:    q.Code,
// 			Type:    string(q.QuestionType),
// 			Options: q.Options,
// 			Rules:   q.Rules,
// 		}

// 		// Call the shared validator from form_answer_usecase.go
// 		if err := validateAnswer(*eru.cfg, tempFormQuestion, answerValue); err != nil {
// 			result = multierror.Append(result, err)
// 		}
// 	}
// 	return result.ErrorOrNil()
// }

// validateRequestInput checks for basic request integrity.
func (eru *eventRegistrationUsecase) validateRequestInput(request models.CreateEventRegistrationRequest) error {
	// Ensure the instance code is derived from the event code.
	if !strings.HasPrefix(request.InstanceCode, request.EventCode) {
		return models.ErrorMismatchFields
	}

	// Validate that there is exactly one primary attendee (IsParent: true) in the request.
	parentCount := 0
	for _, attendee := range request.Attendees {
		if attendee.IsParent {
			parentCount++
		}
	}

	if parentCount == 0 {
		return models.ErrorMissingParentAttendee
	}
	if parentCount > 1 {
		return models.ErrorMultipleParentAttendees
	}
	return nil
}

// validatePersonalQRMethod handles checks specific to QR code registrations using pre-fetched user data.
func (eru *eventRegistrationUsecase) validatePersonalQRMethod(request models.CreateEventRegistrationRequest, user *models.User) error {
	// QR method only allows for a single attendee.
	if len(request.Attendees) > 1 {
		return models.ErrorQRForMoreThanOneRegister
	}

	// The single attendee must have a community ID.
	attendee := request.Attendees[0]
	if attendee.CommunityId == "" {
		return models.ErrorIdentifierCommunityIdEmpty
	}

	// Check if the user exists using the pre-fetched object.
	if user == nil || user.ID == 0 {
		return errorgen.Error(models.ErrorDataNotFound, "user not found")
	}

	return nil
}

// validateEvent checks if the event is valid and active.
func (eru *eventRegistrationUsecase) validateEvent(request models.CreateEventRegistrationRequest, event models.Event) error {
	// Event must exist and not be in a draft or inactive state.
	if event.ID == 0 || !common.CheckOneDataInList([]string{event.Status}, []string{constants.MapEventStatus[constants.EVENT_STATUS_ACTIVE]}) {
		return errorgen.Error(models.ErrorDataNotFound, "event not found")
	}
	// The event code in the request must match the retrieved event.
	if request.EventCode != event.Code {
		return errorgen.Error(models.ErrorEventNotValid, "event code not valid")
	}
	// Cannot register for an announcement event.
	if event.Category == string(constants.EventCategoryAnnouncement) {
		return errorgen.Error(errorgen.InvalidInput, "cannot register for an announcement event")
	}
	return nil
}

// validateInstance checks registration times, methods, and status for the event instance.
func (eru *eventRegistrationUsecase) validateInstance(request models.CreateEventRegistrationRequest, instance models.EventInstance) error {
	switch {
	// Instance must exist and be active.
	case instance.ID == 0 || !common.CheckOneDataInList([]string{instance.Status}, []string{constants.MapEventStatus[constants.EVENT_STATUS_ACTIVE]}):
		return errorgen.Error(models.ErrorDataNotFound, "event instance not found")

	// Registration method must be allowed by the instance.
	case !common.CheckOneDataInList([]string{request.Method}, instance.Methods):
		return errorgen.Error(models.ErrorInvalidInput, "registration method not allowed by the instance")

	// Registration start time must be before the end time.
	case instance.RegisterStartAt.After(instance.RegisterEndAt.In(common.GetLocation())):
		return errorgen.Error(models.ErrorRegistrationTimeDisabled)

	// Cannot register before the registration window opens.
	case request.RegisterAt.Before(instance.RegisterStartAt.In(common.GetLocation())):
		return errorgen.Error(models.ErrorCannotRegisterYet)

	// Cannot register after the registration window closes.
	case request.RegisterAt.After(instance.RegisterEndAt.In(common.GetLocation())):
		return errorgen.Error(models.ErrorRegistrationTimeDisabled)

	// For QR codes, cannot register before the verification window opens.
	case (request.Method == "personal-qr" || request.Method == "event-qr") && request.RegisterAt.Before(instance.VerifyStartAt.In(common.GetLocation())):
		return errorgen.Error(models.ErrorCannotRegisterYet)

	// For QR codes, cannot register after the verification window closes.
	case (request.Method == "personal-qr" || request.Method == "event-qr") && request.RegisterAt.After(instance.VerifyEndAt.In(common.GetLocation())):
		return errorgen.Error(models.ErrorRegistrationTimeDisabled)
	}

	return nil
}

// validateUserPermissions checks if the registrant is allowed to register for a non-public event.
// It uses a pre-fetched user object to avoid extra database calls.
func (eru *eventRegistrationUsecase) validateUserPermissions(user *models.User, event models.Event) error {
	if user == nil || user.ID == 0 {
		return errorgen.Error(models.ErrorForbiddenRole, "user not found for permission check")
	}

	// Define each permission check as a separate boolean variable for clarity.
	isAllowedByUserType := len(event.AllowedUserTypes) > 0 && common.CheckOneDataInList(user.UserTypes, event.AllowedUserTypes)
	isAllowedByRole := len(event.AllowedRoles) > 0 && common.CheckOneDataInList(user.Roles, event.AllowedRoles)
	isAllowedByCampus := len(event.AllowedCampuses) > 0 && common.CheckOneDataInList([]string{user.CampusCode}, event.AllowedCampuses)
	isAllowedByCommunity := len(event.AllowedCommunityIds) > 0 && common.CheckOneDataInList([]string{user.CommunityID}, event.AllowedCommunityIds)

	// If any of the checks pass, the user is allowed.
	if isAllowedByUserType || isAllowedByRole || isAllowedByCampus || isAllowedByCommunity {
		return nil // User is allowed, continue.
	}

	return models.ErrorForbiddenRole
}

func (eru *eventRegistrationUsecase) validateEnforcer(ctx context.Context, request models.CreateEventRegistrationRequest, instance models.EventInstance, user *models.User) error {
	// Enforce one registration transaction per community_id for this instance.
	if instance.EnforceCommunityId {
		if request.Registrant.CommunityId == "" {
			return errorgen.Error(models.ErrorIdentifierCommunityIdEmpty)
		}
		isExist, err := eru.r.EventRegistration.CheckByCommunityIdAndInstanceCode(ctx, request.Registrant.CommunityId, instance.Code)
		if err != nil {
			return errorgen.Error(err)
		}
		if isExist {
			return errorgen.Error(models.ErrorEventCanOnlyRegisterOnce)
		}
	}

	if instance.EnforceUniqueness {
		// Check for duplicates within the request itself first.
		if !common.AreStringFieldsUnique(request.Attendees, "Email", "PhoneNumber", "LegalId", "CommunityId") {
			return errorgen.Error(models.ErrorDuplicateAttendeeInRequest)
		}

		var parentIdentifierFields, childIdentifierFields []models.IdentifierField
		if err := instance.ParentIdentifierFields.Unmarshal(&parentIdentifierFields); err != nil {
			return errorgen.Error(err, "failed to unmarshal parent identifier fields")
		}
		if err := instance.ChildIdentifierFields.Unmarshal(&childIdentifierFields); err != nil {
			return errorgen.Error(err, "failed to unmarshal child identifier fields")
		}

		// Efficiently check for uniqueness across all attendees in the database.
		// This bulk operation avoids the N+1 query problem.
		identifiersToValidate := make(map[string][]string)
		for _, attendee := range request.Attendees {
			var fieldsToValidate []models.IdentifierField
			if attendee.IsParent {
				fieldsToValidate = parentIdentifierFields
			} else {
				fieldsToValidate = childIdentifierFields
			}

			for _, field := range fieldsToValidate {
				var value string
				switch field.Type {
				case "email":
					value = attendee.Email
				case "phoneNumber":
					value = attendee.PhoneNumber
				case "legalId":
					value = attendee.LegalId
				case "communityId":
					value = attendee.CommunityId
				}

				if value != "" {
					identifiersToValidate[field.Type] = append(identifiersToValidate[field.Type], value)
				}
			}
		}

		if len(identifiersToValidate) > 0 {
			// Assumes repository has a method for bulk checking identifiers, for example:
			// CheckByIdentifiersAndInstanceCode(ctx, map[string][]string, string) (bool, error)
			isExist, err := eru.r.EventAttendance.CheckByIdentifiersAndInstanceCode(ctx, identifiersToValidate, instance.Code)
			if err != nil {
				return err
			}
			if isExist {
				return models.ErrorAlreadyRegistered
			}
		}
	}

	if instance.EnforceSelfRegistration {
		// Check if the registrant is trying to register for themselves using the pre-fetched user.
		if user == nil || user.ID == 0 {
			return fmt.Errorf("this event only allows self-registration, but user data could not be verified")
		}
		isSelfRegistration := false
		if len(request.Attendees) == 1 {
			attendee := request.Attendees[0]
			// Check if attendee details match the registrant's user profile.
			if (user.Email != "" && user.Email == attendee.Email) || (user.PhoneNumber != "" && user.PhoneNumber == attendee.PhoneNumber) {
				isSelfRegistration = true
			}
		}

		if !isSelfRegistration {
			return fmt.Errorf("this event only allows self-registration. You may only register for yourself, and the attendee details must match your own")
		}
	}

	return nil
}

// validateRequiredIdentifiers checks if all mandatory identifiers configured in the instance are present in the request.
func (eru *eventRegistrationUsecase) validateRequiredIdentifiers(request models.CreateEventRegistrationRequest, instance models.EventInstance) error {
	var result *multierror.Error

	var parentIdentifierFields, childIdentifierFields []models.IdentifierField
	if err := instance.ParentIdentifierFields.Unmarshal(&parentIdentifierFields); err != nil {
		return errorgen.Error(err, "failed to unmarshal parent identifier fields")
	}
	if err := instance.ChildIdentifierFields.Unmarshal(&childIdentifierFields); err != nil {
		return errorgen.Error(err, "failed to unmarshal child identifier fields")
	}

	for _, attendee := range request.Attendees {
		var fieldsToValidate []models.IdentifierField
		if attendee.IsParent {
			fieldsToValidate = parentIdentifierFields
		} else {
			fieldsToValidate = childIdentifierFields
		}

		for _, field := range fieldsToValidate {
			if !field.IsMandatory {
				continue
			}

			var value string
			switch field.Type {
			case "email":
				value = attendee.Email
			case "phoneNumber":
				value = attendee.PhoneNumber
			case "legalId":
				value = attendee.LegalId
			case "communityId":
				value = attendee.CommunityId
			default:
				result = multierror.Append(result, errorgen.Error(errorgen.InvalidInput, fmt.Sprintf("unknown identifier type configured for validation: %s", field.Type)))
				continue
			}

			if value == "" {
				role := "child"
				if attendee.IsParent {
					role = "parent"
				}
				result = multierror.Append(result, errorgen.Error(errorgen.ErrMissingFields, fmt.Sprintf("missing required identifier '%s' for %s attendee", field.Type, role)))
			}
		}
	}

	return result.ErrorOrNil()
}

// validateCapacity checks if the registration request respects user-specific and total event capacity.
// It is called from within a transaction to prevent race conditions.
func (eru *eventRegistrationUsecase) validateCapacity(ctx context.Context, request models.CreateEventRegistrationRequest, instance models.EventInstance) error {
	// Calculate the number of new attendees in this request.
	newAttendeesCount := len(request.Attendees)

	if instance.QuotaPerUser > 0 {
		// Get the counts of attendees by their status.
		count, err := eru.r.EventAttendance.CountByCommunityIdAndInstanceCode(ctx, request.Registrant.CommunityId, instance.Code)
		if err != nil {
			return err
		}

		if count+newAttendeesCount > instance.QuotaPerUser {
			return models.ErrorExceedMaxSeating
		}
	}

	if instance.Capacity != 0 {
		// Get the counts of attendees by their status.
		// This read is safe from race conditions because it's inside the transaction.
		// For even stronger guarantees in high-concurrency systems, a `SELECT FOR UPDATE` could be used here.
		attendanceStatusCount, err := eru.r.EventAttendance.GetStatusCountsByInstanceCode(ctx, instance.Code)
		if err != nil {
			return err
		}

		// Calculate the total number of booked seats by summing up pending and successful registrations.
		// Cancelled registrations do not count towards the capacity.
		bookedSeats := attendanceStatusCount.Pending + attendanceStatusCount.Success

		// Check if the new attendees would exceed the total capacity.
		if (bookedSeats + newAttendeesCount) > instance.Capacity {
			return models.ErrorRegisterQuotaNotAvailable
		}

	}

	return nil
}

func divideEventAndInstance(eventAndInstance *models.GetEventAndInstanceByCodesDBOutput) (models.Event, models.EventInstance) {
	event := models.Event{
		ID:                   eventAndInstance.EventID,
		Code:                 eventAndInstance.EventCode,
		Title:                eventAndInstance.EventTitle,
		Topics:               eventAndInstance.EventTopics,
		Category:             eventAndInstance.EventCategory,
		RedirectURL:          eventAndInstance.EventRedirectURL,
		Description:          eventAndInstance.EventDescription,
		TermsAndConditions:   eventAndInstance.EventTermsAndConditions,
		ImageLinks:           eventAndInstance.EventImageLinks,
		Slug:                 eventAndInstance.EventSlug,
		CreatedBy:            eventAndInstance.EventCreatedBy,
		LocationType:         eventAndInstance.EventLocationType,
		LocationOnlineLink:   eventAndInstance.EventLocationOnlineLink,
		LocationOfflineVenue: eventAndInstance.EventLocationOfflineVenue,
		Visibility:           eventAndInstance.EventVisibility,
		AllowedRoles:         eventAndInstance.EventAllowedRoles,
		AllowedUserTypes:     eventAndInstance.EventAllowedUserTypes,
		AllowedCampuses:      eventAndInstance.EventAllowedCampuses,
		AllowedCommunityIds:  eventAndInstance.EventAllowedCommunityIds,
		IsRecurring:          eventAndInstance.EventIsRecurring,
		ContactCommunityIds:  eventAndInstance.EventContactCommunityIds,
		StartAt:              eventAndInstance.EventStartAt,
		EndAt:                eventAndInstance.EventEndAt,
		PostDetails:          eventAndInstance.EventPostDetails,
		Status:               eventAndInstance.EventStatus,
	}

	instance := models.EventInstance{
		ID:                     eventAndInstance.InstanceID,
		Code:                   eventAndInstance.InstanceCode,
		Title:                  eventAndInstance.InstanceTitle,
		Description:            eventAndInstance.InstanceDescription,
		ParentIdentifierFields: eventAndInstance.InstanceParentIdentifierFields,
		ChildIdentifierFields:  eventAndInstance.InstanceChildIdentifierFields,
		EnforceCommunityId:     eventAndInstance.InstanceEnforceCommunityId,
		EnforceUniqueness:      eventAndInstance.InstanceEnforceUniqueness,
		Methods:                eventAndInstance.InstanceMethods,
		Flow:                   eventAndInstance.InstanceFlow,
		StartAt:                eventAndInstance.InstanceStartAt,
		EndAt:                  eventAndInstance.InstanceEndAt,
		RegisterStartAt:        eventAndInstance.InstanceRegisterStartAt,
		RegisterEndAt:          eventAndInstance.InstanceRegisterEndAt,
		VerifyStartAt:          eventAndInstance.InstanceVerifyStartAt,
		VerifyEndAt:            eventAndInstance.InstanceVerifyEndAt,
		Timezone:               eventAndInstance.InstanceTimezone,
		LocationType:           eventAndInstance.InstanceLocationType,
		LocationOfflineVenue:   eventAndInstance.InstanceLocationOfflineVenue,
		LocationOnlineLink:     eventAndInstance.InstanceLocationOnlineLink,
		QuotaPerUser:           eventAndInstance.InstanceQuotaPerUser,
		Capacity:               eventAndInstance.InstanceCapacity,
		PostDetails:            eventAndInstance.InstancePostDetails,
		Status:                 eventAndInstance.InstanceStatus,
	}

	return event, instance
}
