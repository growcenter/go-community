package pgsql

import (
	"context"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type EventGeneralRepository interface {
	GetAll(ctx context.Context) (eventGenerals []models.EventGeneral, err error)
	GetByCode(ctx context.Context, code string) (eventGeneral models.EventGeneral, err error)
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

func (egr *eventGeneralRepository) GetByCode(ctx context.Context, code string) (eventGeneral models.EventGeneral, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var eg models.EventGeneral
	err = egr.db.Where("code = ?", code).Find(&eg).Error

	return eg, err
}

func (egr *eventGeneralRepository) BulkUpdate(ctx context.Context, eventGeneral models.EventGeneral) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	// return egr.trx.Transaction(func(dtx *gorm.DB) error {
	// 	event := models.EventGeneral{}
	// 	return egr.db.Model(&event).Where("id = ?", eventGeneral.ID).Updates(eventGeneral).Error
	// })

	event := models.EventGeneral{}
	return egr.db.Model(&event).Where("id = ?", eventGeneral.ID).Updates(eventGeneral).Error

}

func (egr *eventGeneralRepository) Update(ctx context.Context, eventGeneral models.EventGeneral) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return egr.trx.Transaction(func(dtx *gorm.DB) error {
		return egr.db.Save(eventGeneral).Error
	})
}
