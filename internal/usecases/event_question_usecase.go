package usecases

import (
	"context"
	"github.com/google/uuid"
	"go-community/internal/common"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
)

type EventQuestionUsecase interface {
	Create(ctx context.Context, request models.CreateQuestionRequest) (response []models.CreateQuestionResponse, err error)
}

type eventQuestionUsecase struct {
	r pgsql.PostgreRepositories
}

func NewEventQuestionUsecase(r pgsql.PostgreRepositories) *eventQuestionUsecase {
	return &eventQuestionUsecase{
		r: r,
	}
}

func (equ *eventQuestionUsecase) Create(ctx context.Context, request models.CreateQuestionRequest) (response []models.CreateQuestionResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	// Validate the request
	existEvent, err := equ.r.Event.CheckByCode(ctx, common.StringTrimSpaceAndLower(request.EventCode))
	if err != nil {
		return nil, err
	}

	if !existEvent {
		return nil, models.ErrorDataNotFound
	}

	if request.InstanceCode != nil {
		for _, instanceCode := range request.InstanceCode {
			if request.EventCode != (instanceCode)[:7] {
				return nil, models.ErrorMismatchFields
			}
		}

		existInstance, err := equ.r.EventInstance.CheckMultiple(ctx, request.InstanceCode)
		if err != nil {
			return nil, err
		}

		if int(existInstance) != len(request.InstanceCode) {
			return nil, models.ErrorDataNotFound
		}
	}

	questionDetails := make([]models.EventQuestion, 0)
	for _, questionDetail := range request.Questions {
		questionType, description, err := models.CreateQuestionSetup(questionDetail.Type, questionDetail.Description, questionDetail.Options, *questionDetail.Rules)
		if err != nil {
			return nil, err
		}

		question := models.EventQuestion{
			ID:                    uuid.New(),
			EventCode:             request.EventCode,
			InstanceCode:          request.InstanceCode,
			Question:              questionDetail.Question,
			Description:           description,
			Type:                  *questionType,
			Options:               questionDetail.Options,
			IsMainRequired:        questionDetail.IsMainRequired,
			IsRegistrantRequired:  questionDetail.IsRegistrantRequired,
			DisplayOrder:          questionDetail.DisplayOrder,
			IsVisibleToRegistrant: questionDetail.IsVisibleToRegistrant,
			Status:                questionDetail.Status,
		}

		questionDetails = append(questionDetails, question)
	}

	if err = equ.r.EventQuestion.BulkCreate(ctx, &questionDetails); err != nil {
		return nil, err
	}

	questionResponse := make([]models.CreateQuestionResponse, len(questionDetails))
	for i, p := range questionDetails {
		questionResponse[i] = models.CreateQuestionResponse{
			Type:                  models.TYPE_EVENT_QUESTION,
			ID:                    p.ID,
			EventCode:             p.EventCode,
			InstanceCode:          p.InstanceCode,
			Question:              p.Question,
			Description:           p.Description,
			QuestionType:          p.Type,
			IsMainRequired:        p.IsMainRequired,
			IsRegistrantRequired:  p.IsRegistrantRequired,
			Options:               p.Options,
			DisplayOrder:          p.DisplayOrder,
			IsVisibleToRegistrant: p.IsVisibleToRegistrant,
			Rules:                 p.Rules,
			Status:                p.Status,
		}
	}

	return questionResponse, nil
}
