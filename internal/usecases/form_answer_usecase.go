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
	"go-community/internal/pkg/validator"
	"go-community/internal/repositories/pgsql"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FormAnswerUsecase interface {
	Submit(ctx context.Context, request *models.CreateFormAnswerRequest) error
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

func (fau *formAnswerUsecase) Submit(ctx context.Context, request *models.CreateFormAnswerRequest) error {
	defer func() {
		LogService(ctx, nil)
	}()

	// 1. Validate Form exists
	formCode, err := uuid.Parse(request.FormCode)
	if err != nil {
		return models.ErrorInvalidInput
	}

	_, err = fau.r.Form.GetByCode(ctx, formCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.ErrorDataNotFound
		}
		return err
	}

	// 2. Validate User exists
	userExist, err := fau.r.User.CheckByCommunityId(ctx, request.CommunityID)
	if err != nil {
		return err
	}
	if !userExist {
		return models.ErrorDataNotFound
	}

	// 3. Get all questions for the form to validate answers
	questions, err := fau.r.FormQuestion.GetByFormCode(ctx, request.FormCode)
	if err != nil {
		return err
	}

	questionMap := make(map[string]models.FormQuestion)
	for _, q := range questions {
		questionMap[q.Code] = q
	}

	// 4. Validate and create FormAnswer records
	var answers []models.FormAnswer
	submittedAt := time.Now()
	for _, ans := range request.Answers {
		question, ok := questionMap[ans.QuestionCode]
		if !ok {
			// an answer was submitted for a question that doesn't belong to this form
			return models.ErrorInvalidInput
		}

		if err := validateAnswer(fau.cfg, question, ans.Answer); err != nil {
			return err
		}

		isCorrect := sql.NullBool{Bool: false, Valid: false}
		if question.CorrectAnswer.Valid {
			if common.StringTrimSpaceAndLower(question.CorrectAnswer.String) == common.StringTrimSpaceAndLower(ans.Answer) {
				isCorrect = sql.NullBool{Bool: true, Valid: true}
			}
		}

		answer := models.FormAnswer{
			ID:           uuid.New(),
			FormCode:     request.FormCode,
			CommunityID:  request.CommunityID,
			QuestionCode: ans.QuestionCode,
			Answer:       ans.Answer,
			IsCorrect:    isCorrect,
			SubmittedAt:  submittedAt,
		}

		answers = append(answers, answer)
	}

	if err := fau.r.FormAnswer.BulkCreate(ctx, &answers); err != nil {
		return err
	}

	return nil
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
	case constants.QuestionTypeCool:
		fmt.Println("cool")
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
