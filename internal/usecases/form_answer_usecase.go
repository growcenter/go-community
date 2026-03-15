package usecases

import (
	"context"
	"errors"
	"go-community/internal/constants"
	"go-community/internal/models"
	"go-community/internal/pkg/errorc"
	"go-community/internal/pkg/stringc"
	"go-community/internal/pkg/timec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FormAnswerUsecase interface {
	// Submit validates and persists a full form submission.
	// It enforces mandatory-question coverage, per-type answer rules (length, range,
	// date bounds, pattern, choice membership), and evaluates IsCorrect where applicable.
	Submit(ctx context.Context, request *models.CreateFormAnswerRequest) (*models.CreateFormAnswerResponse, error)
}

type formAnswerUsecase struct {
	d   *Dependencies
	loc *time.Location // timezone for date boundary resolution (defaults to Asia/Jakarta)
}

func NewFormAnswerUsecase(d *Dependencies) FormAnswerUsecase {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.UTC // safe fallback; should never occur on a standard OS
	}
	return &formAnswerUsecase{d: d, loc: loc}
}

func (fau *formAnswerUsecase) Submit(ctx context.Context, request *models.CreateFormAnswerRequest) (response *models.CreateFormAnswerResponse, err error) {
	// ── 1. Resolve form code ──────────────────────────────────────────────────
	// Either a direct formCode or an entity lookup must be provided.
	var formCode uuid.UUID
	if request.FormCode != "" {
		formCode, err = uuid.Parse(request.FormCode)
		if err != nil {
			return nil, errorc.Error(errorc.ErrorInvalidInput, "invalid form code")
		}
	} else if len(request.Entity) > 0 {
		// Resolve the form associated with the entity.  We take the first match.
		associations, err := fau.d.Repository.FormAssociation.GetByEntity(
			ctx,
			request.Entity[0].Type,
			request.Entity[0].Code,
		)
		if err != nil {
			return nil, errorc.Error(err)
		}
		if len(associations) == 0 {
			return nil, errorc.Error(errorc.ErrorDataNotFound, "no form found for the provided entity")
		}
		formCode = associations[0].FormCode
	} else {
		return nil, errorc.Error(errorc.ErrorInvalidInput, "either formCode or entity must be provided")
	}

	// ── 2. Load questions for this form ───────────────────────────────────────
	questions, err := fau.d.Repository.FormQuestion.GetByFormCode(ctx, formCode.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorc.Error(errorc.ErrorDataNotFound, "form questions not found")
		}
		return nil, errorc.Error(err)
	}
	if len(questions) == 0 {
		return nil, errorc.Error(errorc.ErrorDataNotFound, "this form has no questions")
	}

	// ── 3. Index questions by code ────────────────────────────────────────────
	questionMap := make(map[string]models.FormQuestion, len(questions))
	for _, q := range questions {
		questionMap[q.Code.String()] = q
	}

	// ── 4. Determine applicability tag for this submitter ─────────────────────
	// "parent" maps to the primary registrant; "child" to dependents.
	applicabilityTag := "child"
	if request.IsParent {
		applicabilityTag = "parent"
	}

	// ── 5. Check all mandatory questions are answered ─────────────────────────
	// Build a set of answered question codes.
	answeredCodes := make(map[string]struct{}, len(request.Answers))
	for _, a := range request.Answers {
		answeredCodes[a.QuestionCode] = struct{}{}
	}

	for _, q := range questions {
		// Only care about questions that apply to this registrant type.
		if !stringc.ContainsString(q.VisibleFor, applicabilityTag) {
			continue
		}
		if stringc.ContainsString(q.RequiredFor, applicabilityTag) {
			if _, answered := answeredCodes[q.Code.String()]; !answered {
				return nil, errorc.Error(errorc.ErrorInvalidInput,
					"mandatory question '%s' was not answered", q.Text)
			}
		}
	}

	// ── 6. Build and validate each answer ────────────────────────────────────
	submittedAt := time.Now().UTC()
	answers := make([]models.FormAnswer, 0, len(request.Answers))
	answersForResponse := make([]models.FormQuestionAnswerResponse, 0, len(request.Answers))

	for _, item := range request.Answers {
		q, found := questionMap[item.QuestionCode]
		if !found {
			return nil, errorc.Error(errorc.ErrorInvalidInput,
				"question code '%s' does not belong to this form", item.QuestionCode)
		}

		// Skip empty answers for optional questions without complaint.
		if item.Answer == "" && !stringc.ContainsString(q.RequiredFor, applicabilityTag) {
			continue
		}
		if item.Answer == "" && stringc.ContainsString(q.RequiredFor, applicabilityTag) {
			return nil, errorc.Error(errorc.ErrorInvalidInput,
				"mandatory question '%s' cannot have an empty answer", q.Text)
		}

		// Per-type answer validation.
		if err := fau.validateAnswer(q, item.Answer); err != nil {
			return nil, err
		}

		// Evaluate IsCorrect for graded questions.
		var isCorrect *bool
		if q.CorrectAnswer != "" {
			result := strings.EqualFold(q.CorrectAnswer, item.Answer)
			isCorrect = &result
		}

		qCode, _ := uuid.Parse(item.QuestionCode)
		answers = append(answers, models.FormAnswer{
			Code:           uuid.New(),
			FormCode:       formCode,
			QuestionCode:   qCode,
			IdentifierType: request.IdentifierType,
			IdentifierCode: request.Identifier,
			Answer:         item.Answer,
			IsCorrect:      isCorrect,
			Status:         constants.StatusActive,
			SubmittedAt:    submittedAt,
		})

		answerCode := answers[len(answers)-1].Code
		answersForResponse = append(answersForResponse, models.FormQuestionAnswerResponse{
			Question: *q.ToResponse(),
			Answer: models.AnswerResponse{
				Type:           "form_answer",
				Code:           answerCode.String(),
				IdentifierType: request.IdentifierType,
				Identifier:     request.Identifier,
				FormCode:       &formCode,
				QuestionCode:   qCode,
				Answer:         item.Answer,
				IsCorrect:      isCorrect,
				SubmittedAt:    submittedAt,
			},
		})
	}

	// ── 7. Persist atomically ─────────────────────────────────────────────────
	if err = fau.d.Repository.FormAnswer.BulkCreate(ctx, &answers); err != nil {
		return nil, errorc.Error(err)
	}

	return &models.CreateFormAnswerResponse{
		FormCode:       formCode.String(),
		Identifier:     request.Identifier,
		IdentifierType: request.IdentifierType,
		SubmittedAt:    submittedAt,
		Forms:          answersForResponse,
	}, nil
}

// validateAnswer checks that the raw answer string conforms to the question type
// and any QuestionValidationRules defined on the question.
func (fau *formAnswerUsecase) validateAnswer(q models.FormQuestion, answer string) error {
	qType := constants.QuestionType(q.Category)

	var rules *models.QuestionValidationRules
	if !q.Rules.IsNull() {
		_ = q.Rules.Unmarshal(&rules)
	}

	var options *models.QuestionOptions
	if !q.Options.IsNull() {
		_ = q.Options.Unmarshal(&options)
	}

	switch qType {
	// ── Text-family ──────────────────────────────────────────────────────────
	case constants.QuestionTypeShortText, constants.QuestionTypeLongText,
		constants.QuestionTypeEmail, constants.QuestionTypePhone:
		if rules == nil {
			return nil
		}
		length := len([]rune(answer)) // rune-length for CJK/emoji safety
		if rules.MinLength != nil && length < *rules.MinLength {
			return errorc.Error(errorc.ErrorInvalidData,
				"answer is too short (minimum %d characters)", *rules.MinLength)
		}
		if rules.MaxLength != nil && length > *rules.MaxLength {
			return errorc.Error(errorc.ErrorInvalidData,
				"answer is too long (maximum %d characters)", *rules.MaxLength)
		}
		if rules.Pattern != nil && *rules.Pattern != "" {
			matched, err := regexp.MatchString(*rules.Pattern, answer)
			if err != nil {
				return errorc.Error(errorc.ErrorInvalidData, "invalid pattern rule for question")
			}
			if !matched {
				return errorc.Error(errorc.ErrorInvalidData,
					"answer does not match the required pattern")
			}
		}

	// ── Number ────────────────────────────────────────────────────────────────
	case constants.QuestionTypeNumber:
		n, err := strconv.ParseFloat(answer, 64)
		if err != nil {
			return errorc.Error(errorc.ErrorInvalidData, "answer must be a number")
		}
		if rules == nil {
			return nil
		}
		if rules.MinValue != nil && n < float64(*rules.MinValue) {
			return errorc.Error(errorc.ErrorInvalidData,
				"answer must be at least %d", *rules.MinValue)
		}
		if rules.MaxValue != nil && n > float64(*rules.MaxValue) {
			return errorc.Error(errorc.ErrorInvalidData,
				"answer must be at most %d", *rules.MaxValue)
		}

	// ── Date ──────────────────────────────────────────────────────────────────
	case constants.QuestionTypeDate:
		const dateFmt = "2006-01-02"
		t, err := time.Parse(dateFmt, answer)
		if err != nil {
			return errorc.Error(errorc.ErrorInvalidData, "answer must be a date in YYYY-MM-DD format")
		}
		if rules == nil {
			return nil
		}
		if rules.NotBefore != nil {
			boundary, err := timec.ParseDateBoundary(*rules.NotBefore, dateFmt, fau.loc)
			if err != nil {
				return errorc.Error(errorc.ErrorInternalServer)
			}
			if t.Before(boundary) {
				return errorc.Error(errorc.ErrorInvalidData,
					"date must not be before %s", boundary.Format(dateFmt))
			}
		}
		if rules.NotAfter != nil {
			boundary, err := timec.ParseDateBoundary(*rules.NotAfter, dateFmt, fau.loc)
			if err != nil {
				return errorc.Error(errorc.ErrorInternalServer)
			}
			if t.After(boundary) {
				return errorc.Error(errorc.ErrorInvalidData,
					"date must not be after %s", boundary.Format(dateFmt))
			}
		}

	// ── Time ──────────────────────────────────────────────────────────────────
	case constants.QuestionTypeTime:
		if _, err := time.Parse("15:04", answer); err != nil {
			if _, err = time.Parse("15:04:05", answer); err != nil {
				return errorc.Error(errorc.ErrorInvalidData, "answer must be a time in HH:MM or HH:MM:SS format")
			}
		}

	// ── Single choice ─────────────────────────────────────────────────────────
	case constants.QuestionTypeSingle:
		if options == nil {
			return errorc.Error(errorc.ErrorInternalServer)
		}
		if !stringc.ContainsString(options.Choices, answer) {
			return errorc.Error(errorc.ErrorInvalidData,
				"'%s' is not a valid choice for this question", answer)
		}

	// ── Multiple choice ───────────────────────────────────────────────────────
	case constants.QuestionTypeMultiple:
		if options == nil {
			return errorc.Error(errorc.ErrorInternalServer)
		}
		// Multiple answers are submitted as a comma-separated list.
		selections := stringc.SplitTrimmed(answer, ",")
		if len(selections) == 0 {
			return errorc.Error(errorc.ErrorInvalidData, "at least one choice must be selected")
		}
		for _, sel := range selections {
			if !stringc.ContainsString(options.Choices, sel) {
				return errorc.Error(errorc.ErrorInvalidData,
					"'%s' is not a valid choice for this question", sel)
			}
		}
		if rules != nil {
			if rules.MinSelection != nil && len(selections) < *rules.MinSelection {
				return errorc.Error(errorc.ErrorInvalidData,
					"select at least %d option(s)", *rules.MinSelection)
			}
			if rules.MaxSelection != nil && len(selections) > *rules.MaxSelection {
				return errorc.Error(errorc.ErrorInvalidData,
					"select at most %d option(s)", *rules.MaxSelection)
			}
		}
	}
	return nil
}
