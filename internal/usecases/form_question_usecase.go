package usecases

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-community/internal/constants"
	"go-community/internal/models"
	"go-community/internal/pkg/errorgen"
	"go-community/internal/repositories/pgsql"
	"sort"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FormQuestionUsecase interface {
	BulkCreate(ctx context.Context, forms *models.Form, request *models.BulkCreateFormQuestionRequest) (responses []models.FormQuestionResponse, err error)
	// GetByAssociationEntity(ctx context.Context, entities []models.FormQuestionEntityFilter) (formQuestions []models.FormQuestion, err error)
}

type formQuestionUsecase struct {
	r pgsql.PostgreRepositories
}

func NewFormQuestionUsecase(r pgsql.PostgreRepositories) *formQuestionUsecase {
	return &formQuestionUsecase{
		r: r,
	}
}

func (fqu *formQuestionUsecase) BulkCreate(ctx context.Context, forms *models.Form, request *models.BulkCreateFormQuestionRequest) (responses []models.FormQuestionResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	formCode, err := uuid.Parse(request.FormID)
	if err != nil {
		return nil, models.ErrorInvalidInput
	}

	if forms == nil {
		_, err = fqu.r.Form.GetByCode(ctx, formCode)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorgen.Error(errorgen.DataNotFound)
			}
			return nil, err
		}
	}

	// Sort questions by display order to validate for duplicates and gaps.
	sort.Slice(request.Questions, func(i, j int) bool {
		return request.Questions[i].DisplayOrder < request.Questions[j].DisplayOrder
	})

	// Validate the display order sequence.
	if len(request.Questions) > 0 {
		// Check if the sequence starts from 1.
		if request.Questions[0].DisplayOrder != 1 {
			return nil, errorgen.Error(errorgen.InvalidInput, "display order must start from 1")
		}
		// Check for duplicates and gaps.
		for i := 1; i < len(request.Questions); i++ {
			if request.Questions[i].DisplayOrder != request.Questions[i-1].DisplayOrder+1 {
				return nil, errorgen.Error(errorgen.InvalidInput, "display order must be sequential and without duplicates or gaps.")
			}
		}
	}

	var questions []models.FormQuestion
	for _, q := range request.Questions {

		// Validation based on QuestionType
		switch q.QuestionType {
		case constants.QuestionTypeSingle, constants.QuestionTypeMultiple:
			if q.Options == nil || len(q.Options.Choices) == 0 {
				errMessage := fmt.Sprintf("options are required for question type %s", q.QuestionType)
				return nil, errorgen.Error(err, errMessage)
			}
		case constants.QuestionTypeShortText, constants.QuestionTypeLongText, constants.QuestionTypeEmail, constants.QuestionTypePhone, constants.QuestionTypeNumber, constants.QuestionTypeDate, constants.QuestionTypeTime:
			if q.Options != nil {
				errMessage := fmt.Sprintf("options are not allowed for question type %s", q.QuestionType)
				return nil, errorgen.Error(err, errMessage)
			}
		}

		// Ensure that a question cannot be mandatory for a group it doesn't apply to.
		// Every value in 'mandatoryFor' must also be present in 'applyFor'.
		for _, mandatoryTarget := range q.MandatoryFor {
			isApplicable := false
			for _, applyTarget := range q.ApplyFor {
				if mandatoryTarget == applyTarget {
					isApplicable = true
					break
				}
			}
			if !isApplicable {
				return nil, errorgen.Error(err, "invalid configuration: question is mandatory for '%s' but does not apply to them", mandatoryTarget)
			}
		}

		// Validate the rules themselves
		if q.Rules != nil {
			if err := fqu.validateQuestionRules(constants.QuestionType(q.QuestionType), q.Rules); err != nil {
				return nil, err
			}
		}

		var correctAnswer sql.NullString
		if q.CorrectAnswer != nil {
			correctAnswer = sql.NullString{String: *q.CorrectAnswer, Valid: true}
		}

		question := models.FormQuestion{
			Code:          uuid.New().String(),
			FormCode:      request.FormID,
			Text:          q.Text,
			Type:          string(q.QuestionType),
			MandatoryFor:  q.MandatoryFor,
			ApplyFor:      q.ApplyFor,
			Options:       q.Options,
			Rules:         q.Rules,
			CorrectAnswer: correctAnswer,
			DisplayOrder:  q.DisplayOrder,
		}
		questions = append(questions, question)
	}

	if err := fqu.r.FormQuestion.BulkCreate(ctx, &questions); err != nil {
		return nil, err
	}

	for _, q := range questions {
		responses = append(responses, *q.ToResponse())
	}

	return responses, nil
}

func (fqu *formQuestionUsecase) validateQuestionRules(questionType constants.QuestionType, rules *models.QuestionValidationRules) error {
	if rules.MinLength != nil && rules.MaxLength != nil {
		if *rules.MinLength > *rules.MaxLength {
			return fmt.Errorf("minLength cannot be greater than maxLength")
		}
	}

	if rules.MinValue != nil && rules.MaxValue != nil {
		if *rules.MinValue > *rules.MaxValue {
			return fmt.Errorf("minValue cannot be greater than maxValue")
		}
	}

	// Validate that rules apply to the correct question type
	switch questionType {
	case constants.QuestionTypeShortText, constants.QuestionTypeLongText, constants.QuestionTypeEmail, constants.QuestionTypePhone:
		if rules.MinValue != nil || rules.MaxValue != nil {
			return fmt.Errorf("min/max value rules are not applicable to text questions")
		}
		if rules.NotBefore != nil || rules.NotAfter != nil {
			return fmt.Errorf("date rules are not applicable to text questions")
		}
	case constants.QuestionTypeNumber:
		if rules.MinLength != nil || rules.MaxLength != nil || rules.Pattern != nil {
			return fmt.Errorf("text-based rules are not applicable to number questions")
		}
		if rules.NotBefore != nil || rules.NotAfter != nil {
			return fmt.Errorf("date rules are not applicable to number questions")
		}
	case constants.QuestionTypeDate, constants.QuestionTypeTime:
		if rules.MinLength != nil || rules.MaxLength != nil || rules.Pattern != nil {
			return fmt.Errorf("text-based rules are not applicable to date/time questions")
		}
		if rules.MinValue != nil || rules.MaxValue != nil {
			return fmt.Errorf("number rules are not applicable to date/time questions")
		}
	case constants.QuestionTypeSingle, constants.QuestionTypeMultiple:
		if rules.MinLength != nil || rules.MaxLength != nil || rules.Pattern != nil || rules.MinValue != nil || rules.MaxValue != nil || rules.NotBefore != nil || rules.NotAfter != nil {
			return fmt.Errorf("rules are not applicable to single/multiple choice questions")
		}
	}

	return nil
}
