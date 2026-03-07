package usecases

import (
	"context"
	"errors"
	"fmt"
	"go-community/internal/constants"
	"go-community/internal/models"
	"go-community/internal/pkg/errorc"
	"sort"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FormQuestionUsecase interface {
	// BulkCreate validates and persists questions using the default repository set.
	// When called inside an Atomic callback the transactional DB is embedded in ctx
	// by Atomic and picked up transparently by each repository — no special variant needed.
	BulkCreate(ctx context.Context, formCode string, request []models.BulkCreateFormQuestionRequest) (responses []models.FormQuestionResponse, err error)
}

type formQuestionUsecase struct {
	d *Dependencies
}

func NewFormQuestionUsecase(d *Dependencies) FormQuestionUsecase {
	return &formQuestionUsecase{
		d: d,
	}
}

func (fqu *formQuestionUsecase) BulkCreate(ctx context.Context, formCode string, request []models.BulkCreateFormQuestionRequest) (responses []models.FormQuestionResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	formCodeUUID, err := uuid.Parse(formCode)
	if err != nil {
		return nil, errorc.Error(errorc.ErrorInvalidInput, "invalid form code: %s", formCode)
	}

	// Verify the form exists when not called from within a CreateForm flow.
	_, err = fqu.d.Repository.Form.GetByCode(ctx, formCodeUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorc.Error(errorc.ErrorDataNotFound, "form not found: %s", formCode)
		}
		return nil, errorc.Error(err)
	}

	// Sort questions by display order to validate for duplicates and gaps.
	sort.Slice(request, func(i, j int) bool {
		return request[i].DisplayOrder < request[j].DisplayOrder
	})

	// Validate the display order sequence.
	if len(request) > 0 {
		if request[0].DisplayOrder != 1 {
			return nil, errorc.Error(errorc.ErrorInvalidInput, "display order must start from 1")
		}
		for i := 1; i < len(request); i++ {
			if request[i].DisplayOrder != request[i-1].DisplayOrder+1 {
				return nil, errorc.Error(errorc.ErrorInvalidInput, "display order must be sequential without duplicates or gaps")
			}
		}
	}

	var questions []models.FormQuestion
	for _, q := range request {

		// Validate options vs type: choice types require choices; others forbid them.
		switch q.QuestionType {
		case constants.QuestionTypeSingle, constants.QuestionTypeMultiple:
			if q.Options == nil || len(q.Options.Choices) == 0 {
				return nil, errorc.Error(errorc.ErrorInvalidInput,
					"options with at least one choice are required for question type %s", q.QuestionType)
			}
		case constants.QuestionTypeShortText, constants.QuestionTypeLongText,
			constants.QuestionTypeEmail, constants.QuestionTypePhone,
			constants.QuestionTypeNumber, constants.QuestionTypeDate, constants.QuestionTypeTime:
			if q.Options != nil {
				return nil, errorc.Error(errorc.ErrorInvalidInput,
					"options are not allowed for question type %s", q.QuestionType)
			}
		}

		// Every value in requiredFor must be present in visibleFor:
		// a question cannot be required for a group it doesn't apply to.
		visibleSet := make(map[string]struct{}, len(q.VisibleFor))
		for _, t := range q.VisibleFor {
			visibleSet[t] = struct{}{}
		}
		for _, required := range q.RequiredFor {
			if _, ok := visibleSet[required]; !ok {
				return nil, errorc.Error(errorc.ErrorInvalidInput,
					"question is required for '%s' but not visible to them", required)
			}
		}

		// Validate the rules themselves.
		if q.Rules != nil {
			if err := fqu.validateQuestionRules(q.QuestionType, q.Rules, q.Options); err != nil {
				return nil, err
			}
		}

		// Validate CorrectAnswer is a valid choice for choice-type questions.
		if q.CorrectAnswer != nil && *q.CorrectAnswer != "" {
			switch q.QuestionType {
			case constants.QuestionTypeSingle, constants.QuestionTypeMultiple:
				validChoice := false
				if q.Options != nil {
					for _, c := range q.Options.Choices {
						if c == *q.CorrectAnswer {
							validChoice = true
							break
						}
					}
				}
				if !validChoice {
					return nil, errorc.Error(errorc.ErrorInvalidInput,
						"correctAnswer '%s' is not one of the defined choices", *q.CorrectAnswer)
				}
			}
		}

		var optionsJSON models.JSONB
		if q.Options != nil {
			if err := optionsJSON.Marshal(q.Options); err != nil {
				return nil, errorc.Error(err)
			}
		}

		var rulesJSON models.JSONB
		if q.Rules != nil {
			if err := rulesJSON.Marshal(q.Rules); err != nil {
				return nil, errorc.Error(err)
			}
		}

		var correctAnswer string
		if q.CorrectAnswer != nil {
			correctAnswer = *q.CorrectAnswer
		}

		question := models.FormQuestion{
			Code:          uuid.New(),
			FormCode:      formCodeUUID,
			Text:          q.Text,
			Category:      string(q.QuestionType),
			RequiredFor:   q.RequiredFor,
			VisibleFor:    q.VisibleFor,
			Options:       optionsJSON,
			Rules:         rulesJSON,
			CorrectAnswer: correctAnswer,
			DisplayOrder:  q.DisplayOrder,
		}
		questions = append(questions, question)
	}

	if err := fqu.d.Repository.FormQuestion.BulkCreate(ctx, &questions); err != nil {
		return nil, errorc.Error(err)
	}

	responses = make([]models.FormQuestionResponse, 0, len(questions))
	for _, q := range questions {
		responses = append(responses, *q.ToResponse())
	}

	return responses, nil
}

// validateQuestionRules checks that the provided rules are semantically valid
// for the given question type and, where applicable, consistent with the provided options.
func (fqu *formQuestionUsecase) validateQuestionRules(
	questionType constants.QuestionType,
	rules *models.QuestionValidationRules,
	options *models.QuestionOptions,
) error {
	// Cross-field range checks.
	if rules.MinLength != nil && rules.MaxLength != nil && *rules.MinLength > *rules.MaxLength {
		return errorc.Error(errorc.ErrorInvalidInput, "minLength (%d) cannot be greater than maxLength (%d)", *rules.MinLength, *rules.MaxLength)
	}
	if rules.MinValue != nil && rules.MaxValue != nil && *rules.MinValue > *rules.MaxValue {
		return errorc.Error(errorc.ErrorInvalidInput, "minValue (%d) cannot be greater than maxValue (%d)", *rules.MinValue, *rules.MaxValue)
	}
	if rules.MinSelection != nil && rules.MaxSelection != nil && *rules.MinSelection > *rules.MaxSelection {
		return errorc.Error(errorc.ErrorInvalidInput, "minSelection (%d) cannot be greater than maxSelection (%d)", *rules.MinSelection, *rules.MaxSelection)
	}

	// Type-specific rule applicability.
	switch questionType {
	case constants.QuestionTypeShortText, constants.QuestionTypeLongText,
		constants.QuestionTypeEmail, constants.QuestionTypePhone:
		if rules.MinValue != nil || rules.MaxValue != nil {
			return fmt.Errorf("min/max value rules are not applicable to text questions")
		}
		if rules.NotBefore != nil || rules.NotAfter != nil {
			return fmt.Errorf("date rules are not applicable to text questions")
		}
		if rules.MinSelection != nil || rules.MaxSelection != nil {
			return fmt.Errorf("selection rules are not applicable to text questions")
		}

	case constants.QuestionTypeNumber:
		if rules.MinLength != nil || rules.MaxLength != nil || rules.Pattern != nil {
			return fmt.Errorf("text-based rules are not applicable to number questions")
		}
		if rules.NotBefore != nil || rules.NotAfter != nil {
			return fmt.Errorf("date rules are not applicable to number questions")
		}
		if rules.MinSelection != nil || rules.MaxSelection != nil {
			return fmt.Errorf("selection rules are not applicable to number questions")
		}

	case constants.QuestionTypeDate, constants.QuestionTypeTime:
		if rules.MinLength != nil || rules.MaxLength != nil || rules.Pattern != nil {
			return fmt.Errorf("text-based rules are not applicable to date/time questions")
		}
		if rules.MinValue != nil || rules.MaxValue != nil {
			return fmt.Errorf("number rules are not applicable to date/time questions")
		}
		if rules.MinSelection != nil || rules.MaxSelection != nil {
			return fmt.Errorf("selection rules are not applicable to date/time questions")
		}

	case constants.QuestionTypeSingle:
		if rules.MinLength != nil || rules.MaxLength != nil || rules.Pattern != nil ||
			rules.MinValue != nil || rules.MaxValue != nil ||
			rules.NotBefore != nil || rules.NotAfter != nil {
			return fmt.Errorf("text/number/date rules are not applicable to choice questions")
		}
		// single_choice: at most one selection is definitionally required.
		if rules.MaxSelection != nil && *rules.MaxSelection != 1 {
			return errorc.Error(errorc.ErrorInvalidInput, "maxSelection must be 1 for single_choice questions")
		}
		if options != nil && rules.MaxSelection != nil && *rules.MaxSelection > len(options.Choices) {
			return errorc.Error(errorc.ErrorInvalidInput,
				"maxSelection (%d) cannot exceed the number of choices (%d)", *rules.MaxSelection, len(options.Choices))
		}

	case constants.QuestionTypeMultiple:
		if rules.MinLength != nil || rules.MaxLength != nil || rules.Pattern != nil ||
			rules.MinValue != nil || rules.MaxValue != nil ||
			rules.NotBefore != nil || rules.NotAfter != nil {
			return fmt.Errorf("text/number/date rules are not applicable to choice questions")
		}
		if options != nil && rules.MaxSelection != nil && *rules.MaxSelection > len(options.Choices) {
			return errorc.Error(errorc.ErrorInvalidInput,
				"maxSelection (%d) cannot exceed the number of choices (%d)", *rules.MaxSelection, len(options.Choices))
		}
	}

	return nil
}
