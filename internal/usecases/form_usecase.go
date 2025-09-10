package usecases

import (
	"context"
	"go-community/internal/constants"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"

	"github.com/google/uuid"
)

type FormUsecase interface {
	Create(ctx context.Context, request *models.CreateFormRequest) (response *models.CreateFormResponse, err error)
}

type formUsecase struct {
	r pgsql.PostgreRepositories
	q FormQuestionUsecase
}

func NewFormUsecase(r pgsql.PostgreRepositories, q FormQuestionUsecase) *formUsecase {
	return &formUsecase{
		r: r,
		q: q,
	}
}

func (fu *formUsecase) Create(ctx context.Context, request *models.CreateFormRequest) (response *models.CreateFormResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	form := models.Form{
		Code:        uuid.New(),
		Name:        request.Name,
		Description: request.Description,
		Status:      constants.StatusActive,
	}

	formAssociation := models.FormAssociation{
		FormCode:   form.Code,
		EntityCode: request.Entity.Code,
		EntityType: request.Entity.Type,
	}

	if err := fu.r.Form.Create(ctx, &form); err != nil {
		return nil, err
	}

	if err := fu.r.FormAssociation.Create(ctx, &formAssociation); err != nil {
		return nil, err
	}

	quesRes, err := fu.q.BulkCreate(ctx, &models.BulkCreateFormQuestionRequest{
		FormID:    form.Code.String(),
		Questions: request.Questions,
	})
	if err != nil {
		return nil, err
	}

	return &models.CreateFormResponse{
		Type:        "form",
		Code:        form.Code.String(),
		Name:        form.Name,
		Description: form.Description,
		FormEntityResponse: models.FormEntityResponse{
			Type: formAssociation.EntityType,
			Code: formAssociation.EntityCode,
		},
		Status:    form.Status,
		Questions: quesRes,
	}, nil
}
