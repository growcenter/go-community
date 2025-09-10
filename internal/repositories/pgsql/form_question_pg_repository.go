package pgsql

import (
	"context"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type FormQuestionRepository interface {
	Create(ctx context.Context, formQuestion *models.FormQuestion) error
	BulkCreate(ctx context.Context, formQuestions *[]models.FormQuestion) error
	GetByFormCode(ctx context.Context, formCode string) (formQuestions []models.FormQuestion, err error)
	GetByFormCodes(ctx context.Context, formCodes []string) (formQuestions []models.FormQuestion, err error)
}

type formQuestionRepository struct {
	db *gorm.DB
}

func NewFormQuestionRepository(db *gorm.DB) *formQuestionRepository {
	return &formQuestionRepository{
		db: db,
	}
}

func (fqr *formQuestionRepository) Create(ctx context.Context, formQuestion *models.FormQuestion) error {
	return fqr.db.WithContext(ctx).Create(formQuestion).Error
}

func (fqr *formQuestionRepository) BulkCreate(ctx context.Context, formQuestions *[]models.FormQuestion) error {
	return fqr.db.WithContext(ctx).Create(formQuestions).Error
}

func (fqr *formQuestionRepository) GetByFormCode(ctx context.Context, formCode string) (formQuestions []models.FormQuestion, err error) {
	err = fqr.db.WithContext(ctx).Where("form_code = ?", formCode).Find(&formQuestions).Error
	return formQuestions, err
}

func (fqr *formQuestionRepository) GetByFormCodes(ctx context.Context, formCodes []string) (formQuestions []models.FormQuestion, err error) {
	err = fqr.db.WithContext(ctx).Where("form_code IN ?", formCodes).Find(&formQuestions).Error
	return formQuestions, err
}
