package pgsql

import (
	"context"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type CoolCategoryRepository interface {
	Create(ctx context.Context, coolDivision *models.CoolCategory) (err error)
	GetByCode(ctx context.Context, code string) (coolDivision models.CoolCategory, err error)
}

type coolCategoryRepository struct {
	db	*gorm.DB
	trx TransactionRepository
}

func NewCoolCategoryRepository(db *gorm.DB, trx TransactionRepository) CoolCategoryRepository {
	return &coolCategoryRepository{db: db, trx: trx}
}

func (cdr *coolCategoryRepository) Create(ctx context.Context, coolDivision *models.CoolCategory) (err error) {
	defer func ()  {
		LogRepository(ctx, err)
	}()

	return cdr.trx.Transaction(func(dtx *gorm.DB) error {
		return cdr.db.Create(&coolDivision).Error
	})
}

func (cdr *coolCategoryRepository) GetByCode(ctx context.Context, code string) (coolDivision models.CoolCategory, err error) {
	defer func ()  {
		LogRepository(ctx, err)
	}()

	var cd models.CoolCategory
	err = cdr.db.Where("code = ?", code).Find(&cd).Error

	return cd, err
}
