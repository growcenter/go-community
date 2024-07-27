package pgsql

import (
	"context"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type EventSessionRepository interface {
	GetAllByEventCode(ctx context.Context, eventCode string) (eventSessions []models.EventSession, err error)
	GetByCode(ctx context.Context, code string) (session models.EventSession, err error)
	BulkUpdate(ctx context.Context, eventSession models.EventSession) (err error)
	Update(ctx context.Context, eventSession models.EventSession) (err error)
}

type eventSessionRepository struct {
	db  *gorm.DB
	trx TransactionRepository
}

func NewEventSessionRepository(db *gorm.DB, trx TransactionRepository) EventSessionRepository {
	return &eventSessionRepository{db, trx}
}

func (esr *eventSessionRepository) GetAllByEventCode(ctx context.Context, eventCode string) (eventSessions []models.EventSession, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var es []models.EventSession
	err = esr.db.Where("event_code = ?", eventCode).Find(&es).Error

	return es, err
}

func (esr *eventSessionRepository) GetByCode(ctx context.Context, code string) (session models.EventSession, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var es models.EventSession
	err = esr.db.Where("code = ?", code).Find(&es).Error

	return es, err
}

func (esr *eventSessionRepository) BulkUpdate(ctx context.Context, eventSession models.EventSession) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return esr.trx.Transaction(func(dtx *gorm.DB) error {
		session := models.EventSession{}
		return esr.db.Model(&session).Where("id = ?", eventSession.ID).Updates(eventSession).Error
	})
}

func (esr *eventSessionRepository) Update(ctx context.Context, eventSession models.EventSession) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return esr.trx.Transaction(func(dtx *gorm.DB) error {
		return esr.db.Save(eventSession).Error
	})
}
