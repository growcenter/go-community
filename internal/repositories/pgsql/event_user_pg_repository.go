package pgsql

import (
	"context"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type EventUserRepository interface {
	Create(ctx context.Context, eventUser *models.EventUser) (err error)
	Update(ctx context.Context, eventUser *models.EventUser) (err error)
	GetByEmail(ctx context.Context, email string) (eventUser models.EventUser, err error)
	GetByAccountNumber(ctx context.Context, accountNumber string) (eventUser models.EventUser, err error)
	GetByPhoneNumber(ctx context.Context, phoneNumber string) (eventUser models.EventUser, err error)
	GetByEmailPhone(ctx context.Context, identifier string) (eventUser models.EventUser, err error)
	BulkUpateRoleByAccountNumbers(ctx context.Context, accountNumbers []string, role string) (err error)
}

type eventUserRepository struct {
	db  *gorm.DB
	trx TransactionRepository
}

func NewEventUserRepository(db *gorm.DB, trx TransactionRepository) EventUserRepository {
	return &eventUserRepository{db: db, trx: trx}
}

func (eur *eventUserRepository) Create(ctx context.Context, eventUser *models.EventUser) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return eur.trx.Transaction(func(dtx *gorm.DB) error {
		return eur.db.Create(&eventUser).Error
	})
}

func (eur *eventUserRepository) Update(ctx context.Context, eventUser *models.EventUser) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return eur.trx.Transaction(func(dtx *gorm.DB) error {
		return eur.db.Save(&eventUser).Error
	})
}

func (eur *eventUserRepository) GetByAccountNumber(ctx context.Context, accountNumber string) (eventUser models.EventUser, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var eu models.EventUser
	err = eur.db.Where("account_number = ?", accountNumber).Find(&eu).Error

	return eu, err
}

func (eur *eventUserRepository) GetByEmail(ctx context.Context, email string) (eventUser models.EventUser, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var eu models.EventUser
	err = eur.db.Where("email = ?", email).Find(&eu).Error

	return eu, err
}

func (eur *eventUserRepository) GetByPhoneNumber(ctx context.Context, phoneNumber string) (eventUser models.EventUser, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var eu models.EventUser
	err = eur.db.Where("phone_number = ?", phoneNumber).Find(&eu).Error

	return eu, err
}

func (eur *eventUserRepository) GetByEmailPhone(ctx context.Context, identifier string) (eventUser models.EventUser, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var eu models.EventUser
	err = eur.db.Where("phone_number = ? OR email = ?", identifier, identifier).Find(&eu).Error

	return eu, err
}

func (eur *eventUserRepository) BulkUpateRoleByAccountNumbers(ctx context.Context, accountNumbers []string, role string) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return eur.trx.Transaction(func(dtx *gorm.DB) error {
		eventRegistration := models.EventRegistration{}
		return eur.db.Model(eventRegistration).Where("account_number IN ?", accountNumbers).Update("role", role).Error
	})
}
