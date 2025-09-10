package pgsql

import (
	"context"
	"go-community/internal/models"
	"gorm.io/gorm"

	"github.com/google/uuid"
)

type EventRegistrationRepository interface {
	Create(ctx context.Context, eventRegistration *models.EventRegistration) (err error)
	GetByCode(ctx context.Context, code uuid.UUID) (eventRegistration models.EventRegistration, err error)
}

type eventRegistrationRepository struct {
	db *gorm.DB
}

func NewEventRegistrationRepository(db *gorm.DB) EventRegistrationRepository {
	return &eventRegistrationRepository{db: db}
}

func (r *eventRegistrationRepository) Create(ctx context.Context, eventRegistration *models.EventRegistration) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return r.db.Create(&eventRegistration).Error
}

func (r *eventRegistrationRepository) GetByCode(ctx context.Context, code uuid.UUID) (eventRegistration models.EventRegistration, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var e models.EventRegistration
	err = r.db.Where("code = ?", code).First(&e).Error

	return e, err
}
