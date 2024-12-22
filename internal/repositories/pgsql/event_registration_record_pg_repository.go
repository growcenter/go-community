package pgsql

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/pkg/cursor"
	"time"

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
	GetEventAttendance(ctx context.Context, communityId, startDate string, endDate string) (output []models.GetEventAttendanceDBOutput, err error)
	GetAllWithCursor(ctx context.Context, param models.GetAllRegisteredCursorParam) (output []models.GetAllRegisteredRecordDBOutput, prev string, next string, total int, err error)
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

func (errr *eventRegistrationRecordRepository) GetEventAttendance(ctx context.Context, communityId, startDate string, endDate string) (output []models.GetEventAttendanceDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = errr.db.Raw(queryGetEventAttendance, communityId, startDate, endDate).Scan(&output).Error
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (errr *eventRegistrationRecordRepository) GetAllWithCursor(ctx context.Context, param models.GetAllRegisteredCursorParam) (output []models.GetAllRegisteredRecordDBOutput, prev string, next string, total int, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var lastUpdatedAt time.Time
	var totalEntries int

	// Decrypt cursor if provided (based on updated_at)
	if param.Cursor != "" {
		lastUpdatedAt, err = cursor.DecryptCursor(param.Cursor)
		if err != nil {
			return nil, "", "", 0, err
		}
	}

	query, params, err := BuildEventRegistrationQuery(baseQueryGetRegisteredRecordList, param.EventCode, param.NameSearch, lastUpdatedAt, param.Direction)
	if err != nil {
		return nil, "", "", 0, err
	}

	// Adjust limit in parameters
	if param.Limit > 0 {
		params[len(params)-1] = param.Limit
	} else {
		params[len(params)-1] = 100 // Default limit
	}

	// Execute query
	var records []models.GetAllRegisteredRecordDBOutput
	err = errr.db.Raw(query, params...).Scan(&records).Error
	if err != nil {
		return nil, "", "", 0, err
	}

	// Get total count for pagination info
	//countQuery := `
	//	SELECT COUNT(*)
	//	FROM event_registration er
	//	WHERE 1=1
	//`
	countQuery, countParams, _ := BuildEventRegistrationQuery(queryCountEventAllRegistered, param.EventCode, param.NameSearch, time.Time{}, "")
	err = errr.db.Raw(countQuery, countParams...).Scan(&totalEntries).Error
	if err != nil {
		return nil, "", "", 0, err
	}

	// Generate cursors
	var nextCursor string
	var prevCursor string
	if len(records) > 0 {
		if param.Direction == "next" {
			nextCursor, err = cursor.EncryptCursor(records[len(records)-1].UpdatedAt.Format(time.RFC3339))
			if err != nil {
				return nil, "", "", 0, err
			}
		}
		if param.Direction == "prev" {
			prevCursor, err = cursor.EncryptCursor(records[0].UpdatedAt.Format(time.RFC3339))
			if err != nil {
				return nil, "", "", 0, err
			}
		}
	}

	return records, nextCursor, prevCursor, totalEntries, nil
}
