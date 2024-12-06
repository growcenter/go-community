package pgsql

import (
	"context"
	"github.com/lib/pq"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type UserTypeRepository interface {
	Create(ctx context.Context, userType *models.UserType) (err error)
	GetByType(ctx context.Context, uType string) (userType models.UserType, err error)
	GetAll(ctx context.Context) (userTypes []models.UserType, err error)
	Check(ctx context.Context, uType string) (dataExist bool, err error)
	CheckMultiple(ctx context.Context, uTypes []string) (count int64, err error)
	GetByArray(ctx context.Context, array []string) (uType []models.UserType, err error)
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

func (utr *userTypeRepository) CheckMultiple(ctx context.Context, uTypes []string) (count int64, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = utr.db.Raw(queryMultipleCheckUserType, pq.Array(uTypes)).Scan(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (utr *userTypeRepository) GetByArray(ctx context.Context, array []string) (uType []models.UserType, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = utr.db.Raw(queryGetUserTypesByArray, pq.Array(array)).Scan(&uType).Error
	if err != nil {
		return nil, err
	}

	return uType, nil
}
