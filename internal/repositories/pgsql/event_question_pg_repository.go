package pgsql

import (
	"context"
	"go-community/internal/models"
	"gorm.io/gorm"
)

type EventQuestionRepository interface {
	Create(ctx context.Context, question *models.EventQuestion) (err error)
	BulkCreate(ctx context.Context, questions *[]models.EventQuestion) (err error)
}

type eventQuestionRepository struct {
	db *gorm.DB
}

func NewEventQuestionRepository(db *gorm.DB) EventQuestionRepository {
	return &eventQuestionRepository{db: db}
}

func (eqr *eventQuestionRepository) Create(ctx context.Context, question *models.EventQuestion) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return eqr.db.Create(&question).Error
}

func (eqr *eventQuestionRepository) BulkCreate(ctx context.Context, questions *[]models.EventQuestion) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return eqr.db.Create(&questions).Error
}
