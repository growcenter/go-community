package pgsql

import (
	"context"
	"go-community/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FormRepository interface {
	Create(ctx context.Context, form *models.Form) error
	GetByCode(ctx context.Context, code uuid.UUID) (form models.Form, err error)
	GetByCodes(ctx context.Context, codes []uuid.UUID) (forms []models.Form, err error)
}

type formRepository struct {
	db *gorm.DB
}

func NewFormRepository(db *gorm.DB) *formRepository {
	return &formRepository{
		db: db,
	}
}

func (fr *formRepository) Create(ctx context.Context, form *models.Form) error {
	return fr.db.WithContext(ctx).Create(form).Error
}

func (fr *formRepository) GetByCodes(ctx context.Context, codes []uuid.UUID) (forms []models.Form, err error) {
	err = fr.db.WithContext(ctx).Where("code IN ?", codes).Find(&forms).Error
	return forms, err
}

func (fr *formRepository) GetByCode(ctx context.Context, code uuid.UUID) (form models.Form, err error) {
	err = fr.db.WithContext(ctx).Where("code = ?", code).First(&form).Error
	return form, err
}
