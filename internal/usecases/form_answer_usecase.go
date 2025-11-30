package usecases

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-community/internal/common"
	"go-community/internal/config"
	"go-community/internal/constants"
	"go-community/internal/models"
	"go-community/internal/pkg/errorgen"
	"go-community/internal/pkg/validator"
	"go-community/internal/repositories/pgsql"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"gorm.io/gorm"
)

type FormAnswerUsecase interface {
	Submit(ctx context.Context, request *models.CreateFormAnswerRequest) (*models.CreateFormAnswerResponse, error)
	SubmitBatch(ctx context.Context, requests []*models.CreateFormAnswerRequest) ([]*models.CreateFormAnswerResponse, error)
	ValidateAndPrepareAnswers(ctx context.Context, request *models.CreateFormAnswerRequest, questions []models.FormQuestion) ([]models.FormAnswer, error)
	WithTransaction(txr pgsql.PostgreRepositories) FormAnswerUsecase
}

type formAnswerUsecase struct {
	r   pgsql.PostgreRepositories
	cfg config.Configuration
}

func NewFormAnswerUsecase(r pgsql.PostgreRepositories, cfg config.Configuration) *formAnswerUsecase {
	return &formAnswerUsecase{
		r:   r,
		cfg: cfg,
	}
}

func (fau *formAnswerUsecase) WithTransaction(txr pgsql.PostgreRepositories) FormAnswerUsecase {
	return NewFormAnswerUsecase(txr, fau.cfg)
}

// Submit handles a single form submission, ensuring the entire process is atomic by creating its own transaction.
func (fau *formAnswerUsecase) Submit(ctx context.Context, request *models.CreateFormAnswerRequest) (*models.CreateFormAnswerResponse, error) {
	var response *models.CreateFormAnswerResponse
	err := fau.r.Transaction.Atomic(ctx, func(ctx context.Context, txR *pgsql.PostgreRepositories) error {
		responses, batchErr := fau.WithTransaction(*txR).SubmitBatch(ctx, []*models.CreateFormAnswerRequest{request})
		if batchErr != nil {
			return batchErr
		}
		if len(responses) > 0 {
			response = responses[0]
		}
		return nil
	})
	return response, err
}

// SubmitBatch handles multiple form submissions. It assumes the caller has already started a transaction
// and that this usecase instance was created using WithTransaction.
func (fau *formAnswerUsecase) SubmitBatch(ctx context.Context, requests []*models.CreateFormAnswerRequest) ([]*models.CreateFormAnswerResponse, error) {
	var allAnswersToCreate []models.FormAnswer
	allQuestionsMap := make(map[string]models.FormQuestion)
	answersByIdentifier := make(map[string][]models.FormAnswer)
	questionsByRequestIdentifier := make(map[string][]models.FormQuestion)

	// Step 1: Collect all unique form codes and entities to fetch questions in bulk.
	formCodes := make(map[string]bool)
	questionsByEntityKey := make(map[string][]models.FormQuestion)

	for _, req := range requests {
		if req.FormCode != "" {
			formCodes[req.FormCode] = true
		}
	}

	// Step 2: Fetch all questions in bulk.
	if len(formCodes) > 0 {
		var fc []string
		for code := range formCodes {
			fc = append(fc, code)
		}
		// Assumes a repository method `GetByFormCodes` exists for bulk fetching.
		questions, err := fau.r.FormQuestion.GetByFormCodes(ctx, fc)
		if err != nil {
			return nil, err
		}
		for _, q := range questions {
			allQuestionsMap[q.Code] = q
		}
	}

	// Step 3: Validate and prepare answers for each request.
	for _, request := range requests {
		var currentRequestQuestions []models.FormQuestion
		if request.FormCode != "" {
			for _, q := range allQuestionsMap {
				if q.FormCode == request.FormCode {
					currentRequestQuestions = append(currentRequestQuestions, q)
				}
			}
		} else {
			entityKey := ""
			for _, e := range request.Entity {
				entityKey += e.Type + e.Code
			}

			cachedQuestions, ok := questionsByEntityKey[entityKey]
			if !ok {
				fetchedQuestions, err := fau.r.FormQuestion.GetByAssociationEntity(ctx, request.Entity) // This can return gorm.ErrRecordNotFound
				if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, fmt.Errorf("failed to get questions for identifier %s: %w", request.Identifier, err)
				}
				questionsByEntityKey[entityKey] = fetchedQuestions
				currentRequestQuestions = fetchedQuestions
			} else {
				currentRequestQuestions = cachedQuestions
			}
		}

		// Ensure all fetched questions are in the main map for the response builder.
		for _, q := range currentRequestQuestions {
			allQuestionsMap[q.Code] = q
		}

		preparedAnswers, err := fau.ValidateAndPrepareAnswers(ctx, request, currentRequestQuestions)
		if err != nil {
			return nil, fmt.Errorf("validation failed for identifier %s: %w", request.Identifier, err)
		}

		allAnswersToCreate = append(allAnswersToCreate, preparedAnswers...)
		answersByIdentifier[request.Identifier] = preparedAnswers
		questionsByRequestIdentifier[request.Identifier] = currentRequestQuestions
	}

	// Step 4: Persist all answers in a single bulk create operation.
	if len(allAnswersToCreate) > 0 {
		if err := fau.r.FormAnswer.BulkCreate(ctx, &allAnswersToCreate); err != nil {
			return nil, err
		}
	}

	// Step 5: Build responses for each original request.
	var responses []*models.CreateFormAnswerResponse
	for _, request := range requests {
		response, err := fau.buildCreateFormAnswerResponse(request, answersByIdentifier[request.Identifier], questionsByRequestIdentifier[request.Identifier])
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// buildCreateFormAnswerResponse is a helper to construct the response object.
func (fau *formAnswerUsecase) buildCreateFormAnswerResponse(request *models.CreateFormAnswerRequest, answers []models.FormAnswer, questions []models.FormQuestion) (*models.CreateFormAnswerResponse, error) {
	var submittedAt time.Time
	if len(answers) > 0 {
		submittedAt = answers[0].SubmittedAt
	}

	response := &models.CreateFormAnswerResponse{
		FormCode:       request.FormCode,
		Identifier:     request.Identifier,
		IdentifierType: request.IdentifierType,
		SubmittedAt:    submittedAt,
		Forms:          make([]models.FormQuestionAnswerResponse, 0, len(answers)),
	}

	questionMap := make(map[string]models.FormQuestion)
	for _, q := range questions {
		questionMap[q.Code] = q
	}

	attendeeRole := "child"
	if request.IsParent {
		attendeeRole = "parent"
	}

	for _, ans := range answers {
		question, ok := questionMap[ans.QuestionCode.String()]
		if !ok {
			continue // Should not happen if validation passed
		}

		var isCorrect *bool
		if ans.IsCorrect.Valid {
			isCorrect = &ans.IsCorrect.Bool
		}
		isMandatoryForResponse := common.CheckOneDataInList(question.MandatoryFor, []string{attendeeRole})

		response.Forms = append(response.Forms, models.FormQuestionAnswerResponse{
			Question: models.QuestionsResponse{
				Type:         models.TYPE_FORM_QUESTION,
				Code:         question.Code,
				FormCode:     question.FormCode,
				Text:         question.Text,
				QuestionType: question.Type,
				IsMandatory:  isMandatoryForResponse,
				Options:      question.Options,
				Rules:        question.Rules,
			},
			Answer: models.AnswerResponse{
				Type:           models.TYPE_FORM_ANSWER,
				Code:           ans.ID.String(),
				IdentifierType: ans.IdentifierType,
				Identifier:     ans.Identifier,
				FormCode:       ans.FormCode,
				QuestionCode:   ans.QuestionCode,
				Answer:         ans.Answer,
				IsCorrect:      isCorrect,
				SubmittedAt:    ans.SubmittedAt,
			},
		})
	}
	return response, nil
}

func (fau *formAnswerUsecase) ValidateAndPrepareAnswers(ctx context.Context, request *models.CreateFormAnswerRequest, questions []models.FormQuestion) ([]models.FormAnswer, error) {
	fmt.Printf("\n--- [DEBUG] Starting ValidateAndPrepareAnswers for identifier: %s ---\n", request.Identifier)
	fmt.Printf("\n--- [DEBUG] Questions collected: %v ---\n", questions)

	if request.FormCode != "" {
		formCode, err := uuid.Parse(request.FormCode)
		if err != nil {
			return nil, models.ErrorInvalidInput
		}

		_, err = fau.r.Form.GetByCode(ctx, formCode)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, models.ErrorDataNotFound
			}
			return nil, err
		}
	}

	// 2. Validate Identifier exists
	switch request.IdentifierType {
	case "eventAttendance":
		attendanceExists, err := fau.r.EventAttendance.CheckByCode(ctx, request.Identifier)
		if err != nil {
			return nil, err
		}
		if !attendanceExists {
			return nil, models.ErrorDataNotFound
		}
	case "communityId":
		// 2.2 Validate Community exists
		userExist, err := fau.r.User.CheckByCommunityId(ctx, request.Identifier)
		if err != nil {
			return nil, err
		}
		if !userExist {
			return nil, models.ErrorDataNotFound
		}
	default:
		return nil, models.ErrorInvalidInput
	}

	// Determine the role of the attendee ("parent" or "child") from the request.
	var attendeeRole string
	if request.IsParent {
		attendeeRole = "parent"
	} else {
		attendeeRole = "child"
	}
	fmt.Printf("[DEBUG] Attendee role determined as: %s\n", attendeeRole)
	var validationErrors *multierror.Error

	// Create a map of all questions for quick lookup by code.
	allQuestionsMap := make(map[string]models.FormQuestion)
	for _, q := range questions {
		allQuestionsMap[q.Code] = q
	}

	// Create a map of provided answers for quick lookup.
	providedAnswers := make(map[string]string)
	for _, ans := range request.Answers {
		providedAnswers[ans.QuestionCode] = ans.Answer
	}

	// Validate all questions relevant to the attendee's role.
	for _, q := range questions {
		fmt.Printf("[DEBUG] Processing question code: %s (Text: %s)\n", q.Code, q.Text)

		// Check if the question applies to the current attendee's role.
		isApplicable := common.CheckOneDataInList(q.ApplyFor, []string{attendeeRole})
		if !isApplicable {
			fmt.Printf("[DEBUG] -> Question %s is NOT applicable for role '%s'. Skipping.\n", q.Code, attendeeRole)
			continue // Skip questions that don't apply.
		}
		fmt.Printf("[DEBUG] -> Question %s IS applicable for role '%s'.\n", q.Code, attendeeRole)

		providedAnswer, answerExists := providedAnswers[q.Code]

		// Check if the question is mandatory for this role and if an answer is missing.
		isMandatoryForThisRole := common.CheckOneDataInList(q.MandatoryFor, []string{attendeeRole})
		fmt.Printf("[DEBUG] -> Is mandatory for this role: %v. Answer exists: %v. Provided answer: '%s'\n", isMandatoryForThisRole, answerExists, providedAnswer)
		if isMandatoryForThisRole && (!answerExists || providedAnswer == "") {
			err := fmt.Errorf("missing answer for mandatory question: %s", q.Text)
			fmt.Printf("[DEBUG] -> VALIDATION FAILED: %s\n", err.Error())
			validationErrors = multierror.Append(validationErrors, err)
			continue // Continue to check other questions
		}

		// If an answer is provided, validate it against the question's rules.
		if answerExists && providedAnswer != "" {
			fmt.Printf("[DEBUG] -> Validating provided answer for question %s...\n", q.Code)
			if err := validateAnswer(fau.cfg, q, providedAnswer); err != nil {
				validationErrors = multierror.Append(validationErrors, err)
			}
		}
	}

	if validationErrors.ErrorOrNil() != nil {
		return nil, validationErrors.ErrorOrNil()
	}

	// Now, prepare the FormAnswer records for persistence.
	var answers []models.FormAnswer
	submittedAt := time.Now()
	for _, ans := range request.Answers {
		// If the question code is empty, just ignore this answer and continue.
		if ans.QuestionCode == "" {
			continue
		}

		question, ok := allQuestionsMap[ans.QuestionCode]
		if !ok {
			// This should be caught by the validation loop above, but it's a good safeguard.
			return nil, errorgen.Error(errorgen.InvalidInput, fmt.Sprintf("question code %s not found", ans.QuestionCode))
		}

		isCorrect := sql.NullBool{Bool: false, Valid: false}
		if question.CorrectAnswer.Valid {
			if common.StringTrimSpaceAndLower(question.CorrectAnswer.String) == common.StringTrimSpaceAndLower(ans.Answer) {
				isCorrect = sql.NullBool{Bool: true, Valid: true}
			}
		}

		questionCodeUUID, err := uuid.Parse(ans.QuestionCode)
		if err != nil {
			return nil, errorgen.Error(errorgen.InvalidInput, fmt.Sprintf("invalid question code format: %s", ans.QuestionCode))
		}

		var formCodeUUID *uuid.UUID
		if request.FormCode != "" {
			parsedUUID, err := uuid.Parse(request.FormCode)
			if err != nil {
				return nil, errorgen.Error(errorgen.InvalidInput, fmt.Sprintf("invalid form code format: %s", request.FormCode))
			}
			formCodeUUID = &parsedUUID
		}

		answer := models.FormAnswer{
			ID:             uuid.New(),
			FormCode:       formCodeUUID,
			IdentifierType: request.IdentifierType,
			Identifier:     request.Identifier,
			QuestionCode:   questionCodeUUID,
			Answer:         ans.Answer,
			IsCorrect:      isCorrect,
			SubmittedAt:    submittedAt,
		}

		answers = append(answers, answer)
	}

	return answers, nil
}

func validateAnswer(cfg config.Configuration, question models.FormQuestion, answer string) error {
	switch constants.QuestionType(question.Type) {
	case constants.QuestionTypeEmail:
		if matched, _ := regexp.MatchString(`^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}`, answer); !matched {
			return fmt.Errorf("invalid email format for question %s", question.Code)
		}
	case constants.QuestionTypePhone:
		if _, err := validator.PhoneNumber("ID", answer); err != nil {
			return fmt.Errorf("invalid phone format for question %s: %w", question.Code, err)
		}
	case constants.QuestionTypeNumber:
		if _, err := strconv.Atoi(answer); err != nil {
			return fmt.Errorf("answer for question %s must be a number", question.Code)
		}
	case constants.QuestionTypeDate:
		if _, err := time.Parse("2006-01-02", answer); err != nil {
			return fmt.Errorf("invalid date format for question %s, expected YYYY-MM-DD", question.Code)
		}
	case constants.QuestionTypeTime:
		if _, err := time.Parse("15:04", answer); err != nil {
			return fmt.Errorf("invalid time format for question %s, expected HH:MM", question.Code)
		}
	case constants.QuestionTypeSingle:
		if !common.CheckOneDataInList(question.Options.Choices, []string{answer}) {
			return fmt.Errorf("answer for question %s is not a valid choice", question.Code)
		}
	case constants.QuestionTypeMultiple:
		answers := strings.Split(answer, ",")
		for _, ans := range answers {
			trimmedAns := strings.TrimSpace(ans)
			if !common.CheckOneDataInList(question.Options.Choices, []string{trimmedAns}) {
				return fmt.Errorf("answer '%s' for question %s is not a valid choice", trimmedAns, question.Code)
			}
		}
	case constants.QuestionTypeCampus:
		_, campusExist := cfg.Campus[common.StringTrimSpaceAndLower(answer)]
		if !campusExist {
			return models.ErrorDataNotFound
		}
	case constants.QuestionTypeDepartment:
		_, departmentExist := cfg.Department[common.StringTrimSpaceAndLower(answer)]
		if !departmentExist {
			return models.ErrorDataNotFound
		}
	}

	// Validate based on Rules
	if question.Rules != nil {
		rules := question.Rules
		if rules.MinLength != nil {
			if len(answer) < *rules.MinLength {
				return fmt.Errorf("answer for question %s must be at least %d characters long", question.Code, *rules.MinLength)
			}
		}
		if rules.MaxLength != nil {
			if len(answer) > *rules.MaxLength {
				return fmt.Errorf("answer for question %s must be at most %d characters long", question.Code, *rules.MaxLength)
			}
		}
		if rules.MinValue != nil {
			num, err := strconv.Atoi(answer)
			if err != nil {
				return fmt.Errorf("answer for question %s must be a number to be validated by min value", question.Code)
			}
			if num < *rules.MinValue {
				return fmt.Errorf("answer for question %s must be at least %d", question.Code, *rules.MinValue)
			}
		}
		if rules.MaxValue != nil {
			num, err := strconv.Atoi(answer)
			if err != nil {
				return fmt.Errorf("answer for question %s must be a number to be validated by max value", question.Code)
			}
			if num > *rules.MaxValue {
				return fmt.Errorf("answer for question %s must be at most %d", question.Code, *rules.MaxValue)
			}
		}
		if rules.Pattern != nil {
			matched, err := regexp.MatchString(*rules.Pattern, answer)
			if err != nil {
				return fmt.Errorf("invalid regex pattern for question %s", question.Code)
			}
			if !matched {
				return fmt.Errorf("answer for question %s does not match the required pattern", question.Code)
			}
		}
		if rules.NotBefore != nil {
			date, err := time.Parse("2006-01-02", answer)
			if err != nil {
				return fmt.Errorf("invalid date format for answer to question %s", question.Code)
			}
			var notBeforeDate time.Time
			if *rules.NotBefore == "today" {
				notBeforeDate = time.Now()
			} else {
				notBeforeDate, err = time.Parse("2006-01-02", *rules.NotBefore)
				if err != nil {
					return fmt.Errorf("invalid NotBefore date format in rule for question %s", question.Code)
				}
			}
			if date.Before(notBeforeDate) {
				return fmt.Errorf("date for question %s cannot be before %s", question.Code, *rules.NotBefore)
			}
		}
		if rules.NotAfter != nil {
			date, err := time.Parse("2006-01-02", answer)
			if err != nil {
				return fmt.Errorf("invalid date format for answer to question %s", question.Code)
			}
			var notAfterDate time.Time
			if *rules.NotAfter == "today" {
				notAfterDate = time.Now()
			} else {
				notAfterDate, err = time.Parse("2006-01-02", *rules.NotAfter)
				if err != nil {
					return fmt.Errorf("invalid NotAfter date format in rule for question %s", question.Code)
				}
			}
			if date.After(notAfterDate) {
				return fmt.Errorf("date for question %s cannot be after %s", question.Code, *rules.NotAfter)
			}
		}
	}
	return nil
}
