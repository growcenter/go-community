package pgsql

import (
	"context"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (err error)
	Update(ctx context.Context, user *models.User) (err error)
	UpdateByEmailPhoneNumber(ctx context.Context, email string, phoneNumber string, user *models.User) (err error)
	GetByAccountNumber(ctx context.Context, accountNumber string) (user models.User, err error)
	GetOneByAccountNumber(ctx context.Context, accountNumber string) (user models.User, err error)
	GetByEmail(ctx context.Context, email string) (user models.User, err error)
	GetByPhoneNumber(ctx context.Context, phoneNumber string) (user models.User, err error)
	GetOneByIdentifier(ctx context.Context, identifier string) (user models.User, err error)
	GetOneByEmailPhoneNumber(ctx context.Context, email string, phoneNumber string) (user models.User, err error)
	CheckByEmailPhoneNumber(ctx context.Context, email string, phoneNumber string) (dataExist bool, err error)
}

type userRepository struct {
	db  *gorm.DB
	trx TransactionRepository
}

func NewUserRepository(db *gorm.DB, trx TransactionRepository) UserRepository {
	return &userRepository{db: db, trx: trx}
}

func (ur *userRepository) Create(ctx context.Context, user *models.User) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return ur.trx.Transaction(func(dtx *gorm.DB) error {
		return ur.db.Create(&user).Error
	})
}

func (ur *userRepository) Update(ctx context.Context, user *models.User) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return ur.trx.Transaction(func(dtx *gorm.DB) error {
		return ur.db.Save(&user).Error
	})
}

func (ur *userRepository) UpdateByEmailPhoneNumber(ctx context.Context, email string, phoneNumber string, user *models.User) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	condition, args := ConditionExistOrNot(email, phoneNumber)
	return ur.trx.Transaction(func(dtx *gorm.DB) error {
		return ur.db.Model(&models.User{}).Where(condition, args...).Updates(user).Error
	})
}

func (ur *userRepository) GetByAccountNumber(ctx context.Context, accountNumber string) (user models.User, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var u models.User
	err = ur.db.Where("account_number = ?", accountNumber).Preload("Campus").Preload("CoolCategory").Find(&u).Error

	return u, err
}

func (ur *userRepository) GetOneByAccountNumber(ctx context.Context, accountNumber string) (user models.User, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var u models.User
	err = ur.db.Where("account_number = ?", accountNumber).Preload("Campus").Preload("CoolCategory").First(&u).Error

	return u, err
}

func (ur *userRepository) GetByEmail(ctx context.Context, email string) (user models.User, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var u models.User
	err = ur.db.Where("email = ?", email).Find(&u).Error

	return u, err
}

func (ur *userRepository) GetByPhoneNumber(ctx context.Context, phoneNumber string) (user models.User, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var u models.User
	err = ur.db.Where("phone_number = ?", phoneNumber).Find(&u).Error

	return u, err
}

func (ur *userRepository) GetOneByIdentifier(ctx context.Context, identifier string) (user models.User, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var u models.User
	err = ur.db.Raw(queryGetOneUserByIdentifier, identifier, identifier).Scan(&u).Error

	return u, err
}

func (ur *userRepository) GetOneByEmailPhoneNumber(ctx context.Context, email string, phoneNumber string) (user models.User, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var u models.User
	err = ur.db.Raw(queryGetOneUserByEmailPhoneNumber, email, email, phoneNumber, phoneNumber).Scan(&u).Error

	return u, err
}

func (ur *userRepository) CheckByEmailPhoneNumber(ctx context.Context, email string, phoneNumber string) (dataExist bool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = ur.db.Raw(queryCheckUserByEmailPhoneNumber, email, email, phoneNumber, phoneNumber).Scan(&dataExist).Error
	if err != nil {
		return false, err
	}

	return dataExist, nil
}
