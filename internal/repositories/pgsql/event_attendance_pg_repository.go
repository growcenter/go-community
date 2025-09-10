package pgsql

import (
	"context"
	"go-community/internal/models"
	"gorm.io/gorm"
)

type EventAttendanceRepository interface {
	Create(ctx context.Context, eventAttendance *models.Attendee) (err error)
}

type eventAttendanceRepository struct {
	db *gorm.DB
}

func NewEventAttendanceRepository(db *gorm.DB) EventAttendanceRepository {
	return &eventAttendanceRepository{db: db}
}

func (r *eventAttendanceRepository) Create(ctx context.Context, eventAttendance *models.Attendee) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return r.db.Create(&eventAttendance).Error
}
