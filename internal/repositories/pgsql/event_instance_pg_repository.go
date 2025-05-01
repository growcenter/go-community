package pgsql

import (
	"context"
	"github.com/lib/pq"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type EventInstanceRepository interface {
	Create(ctx context.Context, event *models.EventInstance) (err error)
	BulkCreate(ctx context.Context, events *[]models.EventInstance) (err error)
	GetByCode(ctx context.Context, code string) (campus models.EventInstance, err error)
	GetAll(ctx context.Context) (campus []models.EventInstance, err error)
	CountByCode(ctx context.Context, code string) (count int64, err error)
	GetManyByEventCode(ctx context.Context, eventCode string, status string) (outputs *[]models.GetInstanceByEventCodeDBOutput, err error)
	GetOneByCode(ctx context.Context, code string, status string) (output *models.GetInstanceByCodeDBOutput, err error)
	GetSeatsNamesByCode(ctx context.Context, code string) (output *models.GetSeatsAndNamesByInstanceCodeDBOutput, err error)
	UpdateBookedSeatsByCode(ctx context.Context, code string, event *models.GetSeatsAndNamesByInstanceCodeDBOutput) (err error)
	UpdateScannedSeatsByCode(ctx context.Context, code string, event *models.GetSeatsAndNamesByInstanceCodeDBOutput) (err error)
	UpdateSeatsByCode(ctx context.Context, code string, event *models.GetSeatsAndNamesByInstanceCodeDBOutput) (err error)
	GetSummary(ctx context.Context, eventCode string) (output []models.GetInstanceSummaryDBOutput, err error)
	CheckByCode(ctx context.Context, code string) (dataExist bool, err error)
	CheckMultiple(ctx context.Context, codes []string) (count int64, err error)
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

func (eir *eventInstanceRepository) GetManyByEventCode(ctx context.Context, eventCode string, status string) (outputs *[]models.GetInstanceByEventCodeDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = eir.db.Raw(queryGetSessionsByEventCode, eventCode, status).Scan(&outputs).Error
	if err != nil {
		return nil, err
	}

	return outputs, nil
}

func (eir *eventInstanceRepository) GetOneByCode(ctx context.Context, code string, status string) (output *models.GetInstanceByCodeDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = eir.db.Raw(queryGetSessionByCode, code, status).Scan(&output).Error
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (eir *eventInstanceRepository) GetSeatsNamesByCode(ctx context.Context, code string) (output *models.GetSeatsAndNamesByInstanceCodeDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = eir.db.Raw(queryGetSeatsByInstanceCode, code).Scan(&output).Error
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (eir *eventInstanceRepository) UpdateBookedSeatsByCode(ctx context.Context, code string, event *models.GetSeatsAndNamesByInstanceCodeDBOutput) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return eir.db.Model(&models.EventInstance{}).Where("code = ?", code).Update("booked_seats", event.BookedSeats).Error
}

func (eir *eventInstanceRepository) UpdateScannedSeatsByCode(ctx context.Context, code string, event *models.GetSeatsAndNamesByInstanceCodeDBOutput) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return eir.db.Model(&models.EventInstance{}).Where("code = ?", code).Update("scanned_seats", event.ScannedSeats).Error
}

func (eir *eventInstanceRepository) UpdateSeatsByCode(ctx context.Context, code string, event *models.GetSeatsAndNamesByInstanceCodeDBOutput) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return eir.db.Model(&models.EventInstance{}).Where("code = ?", code).Updates(map[string]interface{}{
		"scanned_seats": event.ScannedSeats,
		"booked_seats":  event.BookedSeats,
	}).Error
}

func (eir *eventInstanceRepository) GetSummary(ctx context.Context, eventCode string) (output []models.GetInstanceSummaryDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = eir.db.Raw(queryGetInstanceSummary, eventCode).Scan(&output).Error
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (eir *eventInstanceRepository) CheckByCode(ctx context.Context, code string) (dataExist bool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = eir.db.Raw(queryCheckEventInstanceByCode, code).Scan(&dataExist).Error
	if err != nil {
		return false, err
	}

	return dataExist, nil
}

func (eir *eventInstanceRepository) CheckMultiple(ctx context.Context, codes []string) (count int64, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = eir.db.Raw(queryMultipleCheckEventInstance, pq.Array(codes)).Scan(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}
