package usecases

import (
	"context"
	"fmt"
	"go-community/internal/models"
	"go-community/internal/pkg/generator"
	"go-community/internal/pkg/validator"
	"go-community/internal/repositories/pgsql"
	"strings"
)

type UserUsecase interface {
	CreateCool(ctx context.Context, request *models.CreateUserCoolRequest) (user *models.User, err error)
	CreateUser(ctx context.Context, request *models.CreateUserRequest) (user *models.User, err error)
	Login(ctx context.Context) (user *models.User, err error)
	GetByAccountNumber(ctx context.Context, accountNumber string) (user *models.User, err error)
}

type userUsecase struct {
	ur  pgsql.UserRepository
	cr  pgsql.CampusRepository
	ccr pgsql.CoolCategoryRepository
}

func NewUserUsecase(ur pgsql.UserRepository, cr pgsql.CampusRepository, ccr pgsql.CoolCategoryRepository) *userUsecase {
	return &userUsecase{
		ur:  ur,
		cr:  cr,
		ccr: ccr,
	}
}

func (uu *userUsecase) CreateCool(ctx context.Context, request *models.CreateUserCoolRequest) (user *models.User, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	existEmail, err := uu.ur.GetByEmail(ctx, request.Email)
	if err != nil {
		return nil, err
	}

	if existEmail.ID != 0 {
		return nil, models.ErrorAlreadyExist
	}

	existPhone, err := uu.ur.GetByPhoneNumber(ctx, request.PhoneNumber)
	if err != nil {
		return nil, err
	}

	if existPhone.ID != 0 {
		return nil, models.ErrorAlreadyExist
	}

	campus, err := uu.cr.GetByCode(ctx, request.CampusCode)
	if err != nil {
		return nil, err
	}

	if campus.ID == 0 {
		return nil, models.ErrorDataNotFound
	}

	coolCategory, err := uu.ccr.GetByCode(ctx, request.CoolCategoryCode)
	if err != nil {
		return nil, err
	}

	if coolCategory.ID == 0 {
		return nil, models.ErrorDataNotFound
	}

	accountNumber, err := generator.AccountNumber()
	if err != nil {
		return nil, err
	}

	input := models.User{
		AccountNumber:    accountNumber,
		Name:             request.Name,
		PhoneNumber:      fmt.Sprintf("+62%s", request.PhoneNumber),
		Email:            strings.ToLower(request.Email),
		UserType:         "REQUEST_COOL",
		Status:           "NOT_REGISTERED",
		Gender:           request.Gender,
		CampusCode:       request.CampusCode,
		CoolCategoryCode: request.CoolCategoryCode,
		MaritalStatus:    request.MaritalStatus,
		Age:              request.Age,
	}

	if err := uu.ur.Create(ctx, &input); err != nil {
		return nil, err
	}

	return &input, nil
}

func (uu *userUsecase) CreateUser(ctx context.Context, request *models.CreateUserRequest) (user *models.User, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	exist, err := uu.ur.GetOneByEmailPhone(ctx, request.PhoneNumber, request.Email)
	if err != nil {
		return nil, err
	}

	if exist.ID != 0 {
		return nil, models.ErrorAlreadyExist
	}

	campus, err := uu.cr.GetByCode(ctx, strings.ToUpper(request.CampusCode))
	if err != nil {
		return nil, err
	}

	if campus.ID == 0 {
		return nil, models.ErrorDataNotFound
	}

	accountNumber := generator.LuhnAccountNumber()
	isAccountNumberValid := validator.LuhnAccountNumber(accountNumber)
	if !isAccountNumberValid {
		return nil, models.ErrorInvalidAccountNumber
	}

	return
}

func (uu *userUsecase) GetByAccountNumber(ctx context.Context, accountNumber string) (user *models.User, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	data, err := uu.ur.GetOneByAccountNumber(ctx, accountNumber)
	if err != nil {
		return nil, models.ErrorDataNotFound
	}

	if data.ID == 0 {
		return nil, models.ErrorDataNotFound
	}

	return &data, nil
}

func (uu *userUsecase) CheckByEmail(ctx context.Context, email string) (isExist bool, user *models.User, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	data, err := uu.ur.GetByEmail(ctx, email)
	if err != nil {
		return false, nil, err
	}

	if data.ID != 0 && data.UserType != "REQUEST_COOL" {
		return true, nil, models.ErrorAlreadyExist
	}

	if data.ID != 0 && data.UserType == "REQUEST_COOL" {
		return true, &data, nil
	}

	return false, &data, nil
}
