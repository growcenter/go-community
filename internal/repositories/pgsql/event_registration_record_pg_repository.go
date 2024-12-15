package pgsql

import (
	"context"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type EventRegistrationRecordRepository interface {
	Create(ctx context.Context, eventRegistrationRecord *models.EventRegistrationRecord) (err error)
	BulkCreate(ctx context.Context, eventRegistrationRecord *[]models.EventRegistrationRecord) (err error)
	GetById(ctx context.Context, id string) (eventRegistrationRecord models.EventRegistrationRecord, err error)
	GetAll(ctx context.Context) (eventRegistrationRecord []models.EventRegistrationRecord, err error)
	CountByIdentifierOriginAndStatus(ctx context.Context, identifierOrigin string, status string) (count int64, err error)
	CountByCommunityIdOrigin(ctx context.Context, communityIdOrigin string) (count int64, err error)
	CheckByIdentifier(ctx context.Context, identifier string) (isExist bool, err error)
	CheckByName(ctx context.Context, name string) (isExist bool, err error)
	CheckByCommunityId(ctx context.Context, communityId string) (isExist bool, err error)
	Update(ctx context.Context, eventRegistrationRecord models.EventRegistrationRecord) (err error)
}

type eventRegistrationRecordRepository struct {
	db  *gorm.DB
	trx TransactionRepository
}

func NewEventRegistrationRecordRepository(db *gorm.DB, trx TransactionRepository) EventRegistrationRecordRepository {
	return &eventRegistrationRecordRepository{db: db, trx: trx}
}

func (errr *eventRegistrationRecordRepository) Create(ctx context.Context, eventRegistrationRecord *models.EventRegistrationRecord) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return errr.trx.Transaction(func(dtx *gorm.DB) error {
		return errr.db.Create(&eventRegistrationRecord).Error
	})
}

func (errr *eventRegistrationRecordRepository) BulkCreate(ctx context.Context, eventRegistrationRecord *[]models.EventRegistrationRecord) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return errr.db.Create(&eventRegistrationRecord).Error
}

func (errr *eventRegistrationRecordRepository) GetById(ctx context.Context, id string) (eventRegistrationRecord models.EventRegistrationRecord, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var e models.EventRegistrationRecord
	err = errr.db.Where("id = ?", id).Find(&e).Error

	return e, err
}

func (errr *eventRegistrationRecordRepository) GetAll(ctx context.Context) (eventRegistrationRecord []models.EventRegistrationRecord, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var e []models.EventRegistrationRecord
	err = errr.db.Find(&e).Error

	return e, err
}

func (errr *eventRegistrationRecordRepository) CountByIdentifierOriginAndStatus(ctx context.Context, identifierOrigin string, status string) (count int64, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = errr.db.Raw(queryCountRecordByIdentifierOriginAndStatus, identifierOrigin, status).Scan(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (errr *eventRegistrationRecordRepository) CountByCommunityIdOrigin(ctx context.Context, communityIdOrigin string) (count int64, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = errr.db.Raw(queryCountRecordByCommunityIdOrigin, communityIdOrigin).Scan(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (errr *eventRegistrationRecordRepository) CheckByIdentifier(ctx context.Context, identifier string) (isExist bool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = errr.db.Raw(queryCheckRecordByIdentifier, identifier).Scan(&isExist).Error
	if err != nil {
		return false, err
	}

	return isExist, nil
}

func (errr *eventRegistrationRecordRepository) CheckByName(ctx context.Context, name string) (isExist bool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = errr.db.Raw(queryCheckRecordByName, name).Scan(&isExist).Error
	if err != nil {
		return false, err
	}

	return isExist, nil
}

func (errr *eventRegistrationRecordRepository) CheckByCommunityId(ctx context.Context, communityId string) (isExist bool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = errr.db.Raw(queryCheckRecordByCommunityId, communityId).Scan(&isExist).Error
	if err != nil {
		return false, err
	}

	return isExist, nil
}

func (errr *eventRegistrationRecordRepository) Update(ctx context.Context, eventRegistrationRecord models.EventRegistrationRecord) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return errr.db.Save(&eventRegistrationRecord).Error
}
