package pgsql

import (
	"context"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type EventSessionRepository interface {
	GetAllByEventCode(ctx context.Context, eventCode string) (eventSessions []models.EventSession, err error)
	// // UpdateStatusByTimeOpen(ctx context.Context, changes string) (err error)
	// BulkUpdate(ctx context.Context, eventGeneral models.EventGeneral) (err error)
	// Update(ctx context.Context, eventGeneral models.EventGeneral) (err error)
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
