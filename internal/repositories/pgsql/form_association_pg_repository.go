package pgsql

import (
	"context"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type FormAssociationRepository interface {
	Create(ctx context.Context, association *models.FormAssociation) error
	GetByFormCode(ctx context.Context, formCode string) ([]*models.FormAssociation, error)
	GetByEntityCode(ctx context.Context, entityCode string, entityType string) ([]*models.FormAssociation, error)
	Delete(ctx context.Context, association *models.FormAssociation) error
}

type formAssociationRepository struct {
	db *gorm.DB
}

func NewFormAssociationRepository(db *gorm.DB) FormAssociationRepository {
	return &formAssociationRepository{db: db}
}

func (far *formAssociationRepository) Create(ctx context.Context, association *models.FormAssociation) error {
	return far.db.WithContext(ctx).Create(association).Error
}

func (far *formAssociationRepository) GetByFormCode(ctx context.Context, formCode string) ([]*models.FormAssociation, error) {
	var associations []*models.FormAssociation
	err := far.db.WithContext(ctx).Where("form_code = ?", formCode).Find(&associations).Error
	return associations, err
}

func (far *formAssociationRepository) GetByEntityCode(ctx context.Context, entityCode string, entityType string) ([]*models.FormAssociation, error) {
	var associations []*models.FormAssociation
	err := far.db.WithContext(ctx).Where("entity_code = ? AND entity_type = ?", entityCode, entityType).Find(&associations).Error
	return associations, err
}

func (far *formAssociationRepository) Delete(ctx context.Context, association *models.FormAssociation) error {
	return far.db.WithContext(ctx).Where("form_code = ? AND entity_code = ? AND entity_type = ?", association.FormCode, association.EntityCode, association.EntityType).Delete(&models.FormAssociation{}).Error
}
