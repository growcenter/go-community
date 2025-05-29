package pgsql

import (
	"context"
	"go-community/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CoolAttendanceRepository interface {
	Create(ctx context.Context, attendance *models.CoolAttendance) (err error)
	BulkCreate(ctx context.Context, attendance *[]models.CoolAttendance) (err error)
	CheckByMeetingId(ctx context.Context, meetingId uuid.UUID) (dataExist bool, err error)
}

type coolAttendanceRepository struct {
	db *gorm.DB
}

func NewCoolAttendanceRepository(db *gorm.DB) CoolAttendanceRepository {
	return &coolAttendanceRepository{db: db}
}

func (car *coolAttendanceRepository) Create(ctx context.Context, attendance *models.CoolAttendance) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return car.db.Create(&attendance).Error
}

func (car *coolAttendanceRepository) BulkCreate(ctx context.Context, attendance *[]models.CoolAttendance) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return car.db.Create(&attendance).Error
}

func (car *coolAttendanceRepository) CheckByMeetingId(ctx context.Context, meetingId uuid.UUID) (dataExist bool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return dataExist, car.db.Raw(queryCheckAttendanceOnMeetingId, meetingId).Scan(&dataExist).Error
}
