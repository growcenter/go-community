package pgsql

import (
	"context"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type LocationRepository interface {
	Create(ctx context.Context, location *models.Location) (err error)
	GetByCode(ctx context.Context, code string) (location models.Location, err error)
	GetByCampusCode(ctx context.Context, campusCode string) (locations []models.Location, err error)
	GetAll(ctx context.Context) (locations []models.Location, err error)
}

type locationRepository struct {
	db	*gorm.DB
	trx TransactionRepository
}

func NewLocationRepository(db *gorm.DB, trx TransactionRepository) LocationRepository {
	return &locationRepository{db: db, trx: trx}
}

func (lr *locationRepository) Create(ctx context.Context, location *models.Location) (err error) {
	defer func ()  {
		LogRepository(ctx, err)
	}()

	return lr.trx.Transaction(func(dtx *gorm.DB) error {
		return lr.db.Create(&location).Error
	})
}

func (lr *locationRepository) GetByCode(ctx context.Context, code string) (location models.Location, err error) {
	defer func ()  {
		LogRepository(ctx, err)
	}()

	var l models.Location
	err = lr.db.Where("code = ?", code).Find(&l).Error

	return l, err
}

func (lr *locationRepository) GetByCampusCode(ctx context.Context, campusCode string) (locations []models.Location, err error) {
	defer func ()  {
		LogRepository(ctx, err)
	}()

	var l []models.Location
	err = lr.db.Where("campus_code = ?", campusCode).Find(&l).Error

	return l, err
}

func (lr *locationRepository) GetAll(ctx context.Context) (locations []models.Location, err error) {
	defer func ()  {
		LogRepository(ctx, err)
	}()

	var l []models.Location
	err = lr.db.Find(&l).Error

	return l, err
}