package pgsql

import (
	"context"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type CampusRepository interface {
	Create(ctx context.Context, campus *models.Campus) (err error)
	GetByCode(ctx context.Context, code string) (campus models.Campus, err error)
	GetAll(ctx context.Context) (campus []models.Campus, err error)
}

type campusRepository struct {
	db	*gorm.DB
	trx TransactionRepository
}

func NewCampusRepository(db *gorm.DB, trx TransactionRepository) CampusRepository {
	return &campusRepository{db: db, trx: trx}
}

func (cr *campusRepository) Create(ctx context.Context, campus *models.Campus) (err error) {
	defer func ()  {
		LogRepository(ctx, err)
	}()

	return cr.trx.Transaction(func(dtx *gorm.DB) error {
		return cr.db.Create(&campus).Error
	})
}

func (cr *campusRepository) GetByCode(ctx context.Context, code string) (campus models.Campus, err error) {
	defer func ()  {
		LogRepository(ctx, err)
	}()

	var c models.Campus
	err = cr.db.Where("code = ?", code).Find(&c).Error

	return c, err
}

func (cr *campusRepository) GetAll(ctx context.Context) (campus []models.Campus, err error) {
	defer func ()  {
		LogRepository(ctx, err)
	}()

	var c []models.Campus
	err = cr.db.Find(&c).Error

	return c, err
}