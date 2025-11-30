package pgsql

import (
	"context"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type EventAttendanceRepository interface {
	Create(ctx context.Context, eventAttendance *models.EventAttendance) (err error)
	BulkCreate(ctx context.Context, eventAttendances []*models.EventAttendance) (err error)
	CheckByIdentifiersAndInstanceCode(ctx context.Context, identifiers map[string][]string, instanceCode string) (isExist bool, err error)
	CountByInstanceCode(ctx context.Context, instanceCode string) (count int64, err error)
	GetStatusCountsByInstanceCode(ctx context.Context, instanceCode string) (counts models.EventAttendanceStatusCount, err error)
	CheckByCode(ctx context.Context, code string) (isExist bool, err error)
	CountByCommunityIdAndInstanceCode(ctx context.Context, communityId string, instanceCode string) (count int, err error)
}

type eventAttendanceRepository struct {
	db *gorm.DB
}

func NewEventAttendanceRepository(db *gorm.DB) EventAttendanceRepository {
	return &eventAttendanceRepository{db: db}
}

func (r *eventAttendanceRepository) Create(ctx context.Context, eventAttendance *models.EventAttendance) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return r.db.Create(&eventAttendance).Error
}

func (r *eventAttendanceRepository) BulkCreate(ctx context.Context, eventAttendances []*models.EventAttendance) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return r.db.Create(&eventAttendances).Error
}

func (r *eventAttendanceRepository) CheckByIdentifiersAndInstanceCode(ctx context.Context, identifiers map[string][]string, instanceCode string) (isExist bool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	query, args := QueryCheckByIdentifiersAndInstanceCode(identifiers)
	if query == "" {
		return false, nil // No valid identifiers to check
	}

	// Prepend instanceCode to the arguments slice for the query's first placeholder.
	allArgs := append([]interface{}{instanceCode}, args...)

	var count int64
	err = r.db.Raw(query, allArgs...).Scan(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *eventAttendanceRepository) CountByInstanceCode(ctx context.Context, instanceCode string) (count int64, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = r.db.Raw(queryCountAttendanceByInstanceCode, instanceCode).Scan(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *eventAttendanceRepository) GetStatusCountsByInstanceCode(ctx context.Context, instanceCode string) (counts models.EventAttendanceStatusCount, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = r.db.Raw(queryCountAttendanceByStatus, instanceCode).Scan(&counts).Error
	if err != nil {
		return models.EventAttendanceStatusCount{}, err
	}

	return counts, nil
}

func (r *eventAttendanceRepository) CheckByCode(ctx context.Context, code string) (isExist bool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = r.db.Raw(queryCheckAttendanceByCode, code).Scan(&isExist).Error
	if err != nil {
		return false, err
	}

	return isExist, nil
}

func (r *eventAttendanceRepository) CountByCommunityIdAndInstanceCode(ctx context.Context, communityId string, instanceCode string) (count int, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = r.db.Raw(queryCountByCommunityIdAndInstanceCode, communityId, instanceCode).Scan(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}
