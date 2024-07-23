package pgsql

import (
	"context"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type EventGeneralRepository interface {
	GetAll(ctx context.Context) (eventGenerals []models.EventGeneral, err error)
	// UpdateStatusByTimeOpen(ctx context.Context, changes string) (err error)
	BulkUpdate(ctx context.Context, eventGeneral models.EventGeneral) (err error)
	Update(ctx context.Context, eventGeneral models.EventGeneral) (err error)
}

type eventGeneralRepository struct {
	db  *gorm.DB
	trx TransactionRepository
}

func NewEventGeneralRepository(db *gorm.DB, trx TransactionRepository) EventGeneralRepository {
	return &eventGeneralRepository{db: db, trx: trx}
}

func (egr *eventGeneralRepository) GetAll(ctx context.Context) (eventGenerals []models.EventGeneral, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var eg []models.EventGeneral
	err = egr.db.Find(&eg).Error

	return eg, err
}

func (egr *eventGeneralRepository) BulkUpdate(ctx context.Context, eventGeneral models.EventGeneral) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return egr.trx.Transaction(func(dtx *gorm.DB) error {
		// return egr.db.Save(eventGenerals).Error
		var eg models.EventGeneral
		return egr.db.Model(eg).Where("id = ?", eg.ID).Updates(eg).Error
	})
}

func (egr *eventGeneralRepository) Update(ctx context.Context, eventGeneral models.EventGeneral) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return egr.trx.Transaction(func(dtx *gorm.DB) error {
		return egr.db.Save(eventGeneral).Error
	})
}
