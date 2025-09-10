package usecases

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
)

type FormAssociationUsecase interface {
	Create(ctx context.Context, request *models.CreateFormAssociationRequest) (*models.CreateFormAssociationResponse, error)
	GetByFormCode(ctx context.Context, formCode string) ([]*models.FormAssociation, error)
	GetByEntityCode(ctx context.Context, entityCode string, entityType string) ([]*models.FormAssociation, error)
	Delete(ctx context.Context, request *models.FormAssociation) error
}

type formAssociationUsecase struct {
	r pgsql.PostgreRepositories
}

func NewFormAssociationUsecase(r pgsql.PostgreRepositories) formAssociationUsecase {
	return formAssociationUsecase{r: r}
}

func (fau *formAssociationUsecase) Create(ctx context.Context, request *models.CreateFormAssociationRequest) (*models.CreateFormAssociationResponse, error) {
	association := &models.FormAssociation{
		FormCode:   request.FormCode,
		EntityCode: request.EntityCode,
		EntityType: request.EntityType,
	}

	if err := fau.r.FormAssociation.Create(ctx, association); err != nil {
		return nil, err
	}

	response := &models.CreateFormAssociationResponse{
		FormCode:   association.FormCode,
		EntityCode: association.EntityCode,
		EntityType: association.EntityType,
	}

	return response, nil
}

func (fau *formAssociationUsecase) GetByFormCode(ctx context.Context, formCode string) ([]*models.FormAssociation, error) {
	return fau.r.FormAssociation.GetByFormCode(ctx, formCode)
}

func (fau *formAssociationUsecase) GetByEntityCode(ctx context.Context, entityCode string, entityType string) ([]*models.FormAssociation, error) {
	return fau.r.FormAssociation.GetByEntityCode(ctx, entityCode, entityType)
}

func (fau *formAssociationUsecase) Delete(ctx context.Context, request *models.FormAssociation) error {
	return fau.r.FormAssociation.Delete(ctx, request)
}
