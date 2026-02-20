package usecases

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/pkg/errorc"
)

type FormAssociationUsecase interface {
	Create(ctx context.Context, request *models.CreateFormAssociationRequest) (*models.CreateFormAssociationResponse, error)
	GetByFormCode(ctx context.Context, formCode string) ([]models.FormAssociation, error)
	GetByEntity(ctx context.Context, entityType string, entityCode string) ([]models.FormAssociation, error)
	Delete(ctx context.Context, request *models.FormAssociation) error
	GetDummy(ctx context.Context, eventCode string, eventSessionCode string) ([]map[string]any, error)
}

type formAssociationUsecase struct {
	d *Dependencies
}

func NewFormAssociationUsecase(d *Dependencies) FormAssociationUsecase {
	return &formAssociationUsecase{d: d}
}

func (fau *formAssociationUsecase) Create(ctx context.Context, request *models.CreateFormAssociationRequest) (*models.CreateFormAssociationResponse, error) {
	association := &models.FormAssociation{
		FormCode:   request.FormCode,
		EntityCode: request.EntityCode,
		EntityType: request.EntityType,
	}

	if err := fau.d.Repository.FormAssociation.Create(ctx, association); err != nil {
		return nil, errorc.Error(err)
	}

	response := &models.CreateFormAssociationResponse{
		FormCode:   association.FormCode,
		EntityCode: association.EntityCode,
		EntityType: association.EntityType,
	}

	return response, nil
}

func (fau *formAssociationUsecase) GetByFormCode(ctx context.Context, formCode string) ([]models.FormAssociation, error) {
	return fau.d.Repository.FormAssociation.GetByFormCode(ctx, formCode)
}

func (fau *formAssociationUsecase) GetByEntity(ctx context.Context, entityType string, entityCode string) ([]models.FormAssociation, error) {
	return fau.d.Repository.FormAssociation.GetByEntity(ctx, entityType, entityCode)
}

func (fau *formAssociationUsecase) Delete(ctx context.Context, request *models.FormAssociation) error {
	return fau.d.Repository.FormAssociation.Delete(ctx, request.FormCode.String(), request.EntityCode)
}

func (fau *formAssociationUsecase) GetDummy(ctx context.Context, eventCode string, eventSessionCode string) ([]map[string]any, error) {
	return fau.d.Repository.FormAssociation.GetDummy(ctx, eventCode, eventSessionCode)
}
