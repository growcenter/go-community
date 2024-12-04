package pgsql

import (
	"context"
	"github.com/lib/pq"
	"go-community/internal/models"
	"gorm.io/gorm"
)

type EventRepository interface {
	Create(ctx context.Context, event *models.Event) (err error)
	GetByCode(ctx context.Context, code string) (campus models.Event, err error)
	GetAll(ctx context.Context) (campus []models.Event, err error)
	GetAllByRoles(ctx context.Context, roles []string, status string) (output []models.GetAllEventsDBOutput, err error)
	CheckByCode(ctx context.Context, code string) (dataExist bool, err error)
	GetOneByCode(ctx context.Context, code string) (output *models.GetEventByCodeDBOutput, err error)
}

type eventRepository struct {
	db  *gorm.DB
	trx TransactionRepository
}

func NewEventRepository(db *gorm.DB, trx TransactionRepository) EventRepository {
	return &eventRepository{db: db, trx: trx}
}

func (er *eventRepository) Create(ctx context.Context, event *models.Event) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return er.trx.Transaction(func(dtx *gorm.DB) error {
		return er.db.Create(&event).Error
	})
}

func (er *eventRepository) GetByCode(ctx context.Context, code string) (campus models.Event, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var e models.Event
	err = er.db.Where("code = ?", code).Find(&e).Error

	return e, err
}

func (er *eventRepository) GetAll(ctx context.Context) (campus []models.Event, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var e []models.Event
	err = er.db.Find(&e).Error

	return e, err
}

func (er *eventRepository) CheckByCode(ctx context.Context, code string) (dataExist bool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = er.db.Raw(queryCheckEventByCode, code).Scan(&dataExist).Error
	if err != nil {
		return false, err
	}

	return dataExist, nil
}

func (er *eventRepository) GetAllByRoles(ctx context.Context, roles []string, status string) (output []models.GetAllEventsDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = er.db.Raw(queryGetAllEventsByRolesAndStatus, pq.Array(roles), status).Scan(&output).Error
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (er *eventRepository) GetOneByCode(ctx context.Context, code string) (output *models.GetEventByCodeDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = er.db.Raw(queryGetEventInstancesByEventCode, code).Scan(&output).Error
	if err != nil {
		return nil, err
	}

	return output, nil
}
