package pgsql

import (
	"context"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type EventRegistrationRepository interface {
	Create(ctx context.Context, eventRegistration *models.EventRegistration) (err error)
	BulkCreate(ctx context.Context, eventRegistrations *[]models.EventRegistration) (err error)
	GetAll(ctx context.Context) (eventRegistrations []models.EventRegistration, err error)
	GetByIdentifier(ctx context.Context, identifier string) (eventRegistrations []models.EventRegistration, err error)
	GetByCode(ctx context.Context, code string) (eventRegistration models.EventRegistration, err error)
	GetByRegisteredBy(ctx context.Context, registeredBy string) (eventRegistration []models.EventRegistration, err error)
	BulkUpdate(ctx context.Context, eventRegistration models.EventRegistration) (err error)
	Update(ctx context.Context, eventRegistration models.EventRegistration) (err error)
}

type eventRegistrationRepository struct {
	db  *gorm.DB
	trx TransactionRepository
}

func NewEventRegistrationRepository(db *gorm.DB, trx TransactionRepository) EventRegistrationRepository {
	return &eventRegistrationRepository{db: db, trx: trx}
}

func (rer *eventRegistrationRepository) Create(ctx context.Context, eventRegistration *models.EventRegistration) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return rer.trx.Transaction(func(dtx *gorm.DB) error {
		return rer.db.Create(&eventRegistration).Error
	})
}

func (rer *eventRegistrationRepository) BulkCreate(ctx context.Context, eventRegistrations *[]models.EventRegistration) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return rer.trx.Transaction(func(dtx *gorm.DB) error {
		return rer.db.Create(&eventRegistrations).Error
	})
}

func (rer *eventRegistrationRepository) GetAll(ctx context.Context) (eventRegistrations []models.EventRegistration, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var er []models.EventRegistration
	err = rer.db.Find(&er).Error

	return er, err
}

func (rer *eventRegistrationRepository) GetByIdentifier(ctx context.Context, identifier string) (eventRegistrations []models.EventRegistration, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var er []models.EventRegistration
	err = rer.db.Where("identifier = ?", identifier).Find(&er).Error

	return er, err
}

func (rer *eventRegistrationRepository) GetByCode(ctx context.Context, code string) (eventRegistration models.EventRegistration, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var er models.EventRegistration
	err = rer.db.Where("code = ?", code).Find(&er).Error

	return er, err
}

func (rer *eventRegistrationRepository) GetByRegisteredBy(ctx context.Context, registeredBy string) (eventRegistration []models.EventRegistration, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var er []models.EventRegistration
	err = rer.db.Where("registered_by = ?", registeredBy).Find(&er).Error

	return er, err
}

func (rer *eventRegistrationRepository) BulkUpdate(ctx context.Context, eventRegistration models.EventRegistration) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return rer.trx.Transaction(func(dtx *gorm.DB) error {
		registration := models.EventRegistration{}
		return rer.db.Model(&registration).Where("id = ?", eventRegistration.ID).Updates(eventRegistration).Error
	})
}

func (rer *eventRegistrationRepository) Update(ctx context.Context, eventRegistration models.EventRegistration) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return rer.trx.Transaction(func(dtx *gorm.DB) error {
		return rer.db.Save(eventRegistration).Error
	})
}
