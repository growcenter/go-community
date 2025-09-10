package pgsql

import (
	"context"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type FormAnswerRepository interface {
	BulkCreate(ctx context.Context, answers *[]models.FormAnswer) error
}

type formAnswerRepository struct {
	db *gorm.DB
}

func NewFormAnswerRepository(db *gorm.DB) *formAnswerRepository {
	return &formAnswerRepository{
		db: db,
	}
}

func (far *formAnswerRepository) BulkCreate(ctx context.Context, answers *[]models.FormAnswer) error {
	return far.db.WithContext(ctx).Create(answers).Error
}
