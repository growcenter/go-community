package pgsql

import (
	"context"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type EventRepository interface {
	Create(ctx context.Context, event *models.Event) (err error)
	GetByCode(ctx context.Context, code string) (campus models.Event, err error)
	GetEventAndInstanceByCodes(ctx context.Context, eventCode string, instanceCode string) (output *models.GetEventAndInstanceByCodesDBOutput, err error)
	GetAll(ctx context.Context) (campus []models.Event, err error)
	GetAllEvents(ctx context.Context, params models.GetAllEventsParams) (events []models.GetAllEventsDBOutput, err error)
	CheckByCode(ctx context.Context, code string) (dataExist bool, err error)
	CheckBySlug(ctx context.Context, slug string) (dataExist bool, err error)
	CheckByCodeOrSlug(ctx context.Context, code string, slug string) (dataExist bool, err error)
	// GetOneByCode(ctx context.Context, code string) (output *models.GetEventByCodeDBOutput, err error)
	GetOneByCodeOrSlug(ctx context.Context, code string, slug string) (output *models.GetEventWithInstancesDBOutput, err error)
	GetRegistered(ctx context.Context, communityIdOrigin string) (output []models.GetAllRegisteredUserDBOutput, err error)
	GetTitles(ctx context.Context) (output []models.GetEventTitlesDBOutput, err error)
	GetSummary(ctx context.Context, code string) (output *models.GetEventSummaryDBOutput, err error)
	Update(ctx context.Context, event *models.Event) (err error)
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

	return er.db.Create(&event).Error
}

func (er *eventRepository) GetByCode(ctx context.Context, code string) (campus models.Event, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var e models.Event
	err = er.db.Where("code = ?", code).Find(&e).Error

	return e, err
}

func (er *eventRepository) GetEventAndInstanceByCodes(ctx context.Context, eventCode string, instanceCode string) (output *models.GetEventAndInstanceByCodesDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = er.db.Raw(queryGetEventAndInstanceByCodes, eventCode, instanceCode).Scan(&output).Error
	if err != nil {
		return nil, err
	}

	return output, nil
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

func (er *eventRepository) CheckBySlug(ctx context.Context, slug string) (dataExist bool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = er.db.Raw(queryCheckEventBySlug, slug).Scan(&dataExist).Error
	if err != nil {
		return false, err
	}

	return dataExist, nil
}

func (er *eventRepository) CheckByCodeOrSlug(ctx context.Context, code string, slug string) (dataExist bool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = er.db.Raw(queryCheckEventByCodeOrSlug, code, slug).Scan(&dataExist).Error
	if err != nil {
		return false, err
	}

	return dataExist, nil
}

func (er *eventRepository) GetAllEvents(ctx context.Context, params models.GetAllEventsParams) (events []models.GetAllEventsDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	query, args := buildGetAllEventsQuery(params)
	err = er.db.Raw(query, args...).Scan(&events).Error
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (er *eventRepository) GetOneByCode(ctx context.Context, code string) (output *models.GetEventWithInstancesDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = er.db.Raw(queryGetEventWithInstancesByCode, code).Scan(&output).Error
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (er *eventRepository) GetOneByCodeOrSlug(ctx context.Context, code string, slug string) (output *models.GetEventWithInstancesDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	query, param := BuildQueryGetEventWithInstancesByCodeOrSlug(code, slug)
	err = er.db.Raw(query, param).Scan(&output).Error
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (er *eventRepository) GetRegistered(ctx context.Context, communityIdOrigin string) (output []models.GetAllRegisteredUserDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = er.db.Raw(queryGetRegisteredUserByCommunityIdOrigin, communityIdOrigin).Scan(&output).Error
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (er *eventRepository) GetTitles(ctx context.Context) (output []models.GetEventTitlesDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = er.db.Raw(queryGetEventTitles).Scan(&output).Error
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (er *eventRepository) GetSummary(ctx context.Context, code string) (output *models.GetEventSummaryDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = er.db.Raw(queryGetEventSummary, code).Scan(&output).Error
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (er *eventRepository) Update(ctx context.Context, event *models.Event) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return er.db.Save(&event).Error
}
