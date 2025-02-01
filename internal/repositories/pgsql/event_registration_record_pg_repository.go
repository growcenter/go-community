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
	CountByCommunityIdOriginAndInstanceCode(ctx context.Context, communityIdOrigin string, instanceCode string) (count int64, err error)
	CheckByIdentifier(ctx context.Context, identifier string) (isExist bool, err error)
	CheckByIdentifierAndInstanceCode(ctx context.Context, identifier string, instanceCode string) (isExist bool, err error)
	CheckByName(ctx context.Context, name string) (isExist bool, err error)
	CheckByNameAndInstanceCode(ctx context.Context, name string, instanceCode string) (isExist bool, err error)
	CheckByCommunityId(ctx context.Context, communityId string) (isExist bool, err error)
	CheckByCommunityIdAndInstanceCode(ctx context.Context, communityId string, instanceCode string) (isExist bool, err error)
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

func (errr *eventRegistrationRecordRepository) CountByCommunityIdOriginAndInstanceCode(ctx context.Context, communityIdOrigin string, instanceCode string) (count int64, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = errr.db.Raw(queryCountRecordByCommunityIdOriginAndInstanceCode, communityIdOrigin, instanceCode).Scan(&count).Error
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

func (errr *eventRegistrationRecordRepository) CheckByIdentifierAndInstanceCode(ctx context.Context, identifier string, instanceCode string) (isExist bool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = errr.db.Raw(queryCheckRecordByIdentifierAndInstanceCode, identifier, instanceCode).Scan(&isExist).Error
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

func (errr *eventRegistrationRecordRepository) CheckByNameAndInstanceCode(ctx context.Context, name string, instanceCode string) (isExist bool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = errr.db.Raw(queryCheckRecordByNameAndInstanceCode, name, instanceCode).Scan(&isExist).Error
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

func (errr *eventRegistrationRecordRepository) CheckByCommunityIdAndInstanceCode(ctx context.Context, communityId string, instanceCode string) (isExist bool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = errr.db.Raw(queryCheckRecordByCommunityIdAndInstanceCode, communityId, instanceCode).Scan(&isExist).Error
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

	// Decrypt cursor if provided (based on updated_at)
	if param.Cursor != "" {
		lastUpdatedAt, err = cursor.DecryptCursor(param.Cursor)
		if err != nil {
			return nil, "", "", 0, err
		}
	}

	// Set default limit if none provided
	limit := param.Limit
	if limit <= 0 {
		limit = 10 // Default limit
	}

	// Build the query
	query, params, err := BuildEventRegistrationQuery(
		baseQueryGetRegisteredRecordList,
		param.EventCode,
		param.InstanceCode,
		param.NameSearch,
		lastUpdatedAt,
		param.Direction,
		limit+1,
		param.CampusCode,
		param.DepartmentCode,
		param.CoolId,
	)
	if err != nil {
		return nil, "", "", 0, err
	}

	// Execute query
	var records []models.GetAllRegisteredRecordDBOutput
	err = errr.db.Raw(query, params...).Scan(&records).Error
	if err != nil {
		return nil, "", "", 0, err
	}

	// Get total count
	countQuery, countParams, _ := BuildEventRegistrationQuery(
		queryCountEventAllRegistered,
		param.EventCode,
		param.InstanceCode,
		param.NameSearch,
		time.Time{},
		"",
		0, // No limit needed for count query
		param.CampusCode,
		param.DepartmentCode,
		param.CoolId,
	)
	err = errr.db.Raw(countQuery, countParams...).Scan(&total).Error
	if err != nil {
		return nil, "", "", 0, err
	}

	// Check if there are more records
	hasMore := len(records) > limit
	if hasMore {
		records = records[:limit] // Remove the extra record
	}

	// Generate cursors
	if len(records) > 0 {
		// Generate prev cursor if we're not on the first page
		if param.Cursor != "" {
			prev, err = cursor.EncryptCursor(records[0].UpdatedAt.Format(time.RFC3339))
			if err != nil {
				return nil, "", "", 0, err
			}
		}

		// Generate next cursor only if we have more records
		if hasMore {
			next, err = cursor.EncryptCursor(records[len(records)-1].UpdatedAt.Format(time.RFC3339))
			if err != nil {
				return nil, "", "", 0, err
			}
		}
	}

	return records, prev, next, total, nil
}
