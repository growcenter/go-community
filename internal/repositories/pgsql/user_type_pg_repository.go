package pgsql

import (
	"context"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type UserTypeRepository interface {
	Create(ctx context.Context, userType *models.UserType) (err error)
	GetByType(ctx context.Context, uType string) (userType models.UserType, err error)
	GetAll(ctx context.Context) (userTypes []models.UserType, err error)
	Check(ctx context.Context, uType string) (dataExist bool, err error)
}

type userTypeRepository struct {
	db  *gorm.DB
	trx TransactionRepository
}

func NewUserTypeRepository(db *gorm.DB, trx TransactionRepository) UserTypeRepository {
	return &userTypeRepository{db: db, trx: trx}
}

func (utr *userTypeRepository) Create(ctx context.Context, userType *models.UserType) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return utr.trx.Transaction(func(dtx *gorm.DB) error {
		return utr.db.Create(&userType).Error
	})
}

func (utr *userTypeRepository) GetByType(ctx context.Context, uType string) (userType models.UserType, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var ut models.UserType
	err = utr.db.Where("type = ?", uType).Find(&ut).Error

	return ut, err
}

func (utr *userTypeRepository) GetAll(ctx context.Context) (userTypes []models.UserType, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var ut []models.UserType
	err = utr.db.Find(&ut).Error

	return ut, err
}

func (utr *userTypeRepository) Check(ctx context.Context, uType string) (dataExist bool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = utr.db.Raw(querySingleCheckUserType, uType).Scan(&dataExist).Error
	if err != nil {
		return false, err
	}

	return dataExist, nil
}
