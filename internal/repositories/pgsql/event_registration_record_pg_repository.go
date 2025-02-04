package pgsql

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/pkg/cursor"
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
	Download(ctx context.Context, param models.GetDownloadAllRegisteredParam) (output []models.GetDownloadAllRegisteredDBOutput, err error)
	//GetAllWithCursor(ctx context.Context, param models.GetAllRegisteredFilterOptions) (output []models.GetAllRegisteredRecordDBOutput, err error)
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

	// Set default limit if none provided
	if param.Limit <= 0 {
		param.Limit = 10 // Default limit
	}

	// Build the query
	queryList, paramList, err := BuildGetRegisteredQuery(param)
	if err != nil {
		return nil, "", "", 0, err
	}

	// Execute query
	var records []models.GetAllRegisteredRecordDBOutput
	err = errr.db.Raw(queryList, paramList...).Scan(&records).Error
	if err != nil {
		return nil, "", "", 0, err
	}

	queryCount, paramCount, err := BuildCountGetRegisteredQuery(param)
	if err != nil {
		return nil, "", "", 0, err
	}

	// Execute query
	err = errr.db.Raw(queryCount, paramCount...).Scan(&total).Error
	if err != nil {
		return nil, "", "", 0, err
	}

	if len(records) > 0 {
		hasMore := len(records) > param.Limit
		isForward := param.Direction != "prev" && param.Cursor != ""
		isBackward := !isForward && param.Direction == "prev" && param.Cursor != ""

		if hasMore {
			records = records[:param.Limit] // Remove the extra record
		}

		if isBackward {
			for i, j := 0, len(records)-1; i < j; i, j = i+1, j-1 {
				records[i], records[j] = records[j], records[i]
			}
		}

		if isBackward || hasMore {
			lastRecord := records[len(records)-1]
			nextCursor := models.GetAllRegisteredRecordCursor{
				CreatedAt: lastRecord.CreatedAt,
				ID:        lastRecord.ID,
			}
			next = cursor.EncryptCursorFromStruct(nextCursor)
		}

		if isForward || (hasMore && isBackward) {
			firstRecord := records[0]
			prevCursor := models.GetAllRegisteredRecordCursor{
				CreatedAt: firstRecord.CreatedAt,
				ID:        firstRecord.ID,
			}
			prev = cursor.EncryptCursorFromStruct(prevCursor)
		}
	}

	return records, prev, next, total, nil
}

func (errr *eventRegistrationRecordRepository) Download(ctx context.Context, param models.GetDownloadAllRegisteredParam) (output []models.GetDownloadAllRegisteredDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	// Build the query
	queryList, paramList, err := BuildDownloadGetRegisteredQuery(param)
	if err != nil {
		return nil, err
	}

	// Execute query
	var records []models.GetDownloadAllRegisteredDBOutput
	err = errr.db.Raw(queryList, paramList...).Scan(&records).Error
	if err != nil {
		return nil, err
	}

	return records, nil
}
