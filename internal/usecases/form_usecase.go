package usecases

import (
	"context"
	"go-community/internal/constants"
	"go-community/internal/models"
	"go-community/internal/pkg/errorc"
	"go-community/internal/repositories/pgsql"

	"github.com/google/uuid"
)

type FormUsecase interface {
	Create(ctx context.Context, request *models.CreateFormRequest) (response *models.CreateFormResponse, err error)
}

type formUsecase struct {
	d *Dependencies
}

func NewFormUsecase(d *Dependencies) FormUsecase {
	return &formUsecase{
		d: d,
	}
}

func (fu *formUsecase) Create(ctx context.Context, request *models.CreateFormRequest) (response *models.CreateFormResponse, err error) {
	code, err := uuid.NewV7()
	if err != nil {
		return nil, errorc.Error(err)
	}

	form := models.Form{
		Code:        code,
		Name:        request.Name,
		Description: request.Description,
		FormType:    string(constants.FormTypeRegistration),
		Status:      constants.StatusActive,
	}

	formAssociation := &models.CreateFormAssociationRequest{
		FormCode:   form.Code,
		EntityCode: request.Entity.Code,
		EntityType: request.Entity.Type,
	}

	var quesRes []models.FormQuestionResponse
	err = fu.d.Repository.Transaction.Atomic(ctx, func(ctx context.Context, r *pgsql.PostgreRepositories) error {
		if err = r.Form.Create(ctx, &form); err != nil {
			return errorc.Error(err)
		}

		if _, err = fu.d.FormAssociation.Create(ctx, formAssociation); err != nil {
			return errorc.Error(err)
		}

		quesRes, err = fu.d.FormQuestion.BulkCreate(ctx, form.Code.String(), request.Questions)
		if err != nil {
			return errorc.Error(err)
		}

		return nil
	})

	if err != nil {
		return nil, errorc.Error(err)
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
