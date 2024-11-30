package pgsql

import (
	"context"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type EventInstanceRepository interface {
	Create(ctx context.Context, event *models.EventInstance) (err error)
	BulkCreate(ctx context.Context, events *[]models.EventInstance) (err error)
	GetByCode(ctx context.Context, code string) (campus models.EventInstance, err error)
	GetAll(ctx context.Context) (campus []models.EventInstance, err error)
	CountByCode(ctx context.Context, code string) (count int64, err error)
}

type eventInstanceRepository struct {
	db  *gorm.DB
	trx TransactionRepository
}

func NewEventInstanceRepository(db *gorm.DB, trx TransactionRepository) EventInstanceRepository {
	return &eventInstanceRepository{db: db, trx: trx}
}

func (eir *eventInstanceRepository) Create(ctx context.Context, event *models.EventInstance) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return eir.trx.Transaction(func(dtx *gorm.DB) error {
		return eir.db.Create(&event).Error
	})
}

func (eir *eventInstanceRepository) BulkCreate(ctx context.Context, events *[]models.EventInstance) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return eir.trx.Transaction(func(dtx *gorm.DB) error {
		return eir.db.Create(&events).Error
	})
}

func (eir *eventInstanceRepository) GetByCode(ctx context.Context, code string) (campus models.EventInstance, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var ei models.EventInstance
	err = eir.db.Where("code = ?", code).Find(&ei).Error

	return ei, err
}

func (eir *eventInstanceRepository) GetAll(ctx context.Context) (campus []models.EventInstance, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var e []models.EventInstance
	err = eir.db.Find(&e).Error

	return e, err
}

func (eir *eventInstanceRepository) CountByCode(ctx context.Context, code string) (count int64, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = eir.db.Raw(queryCountEventInstanceByCode, code).Scan(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}
