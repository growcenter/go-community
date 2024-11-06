package usecases

import (
	"context"
	"fmt"
	"go-community/internal/models"
	"go-community/internal/pkg/generator"
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
		CommunityID:    accountNumber,
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

	exist, err := uu.ur.GetByEmail(ctx, strings.ToLower(request.Email))
	if err != nil {
		return nil, err
	}

	campus, err := uu.cr.GetByCode(ctx, strings.ToUpper(request.CampusCode))
	if err != nil {
		return nil, err
	}

	if campus.ID == 0 {
		return nil, models.ErrorDataNotFound
	}

	switch {
	case exist.ID != 0 && exist.UserType == "REQUEST_COOL":
		token := "Trial"
		password := request.Password

		var input models.User

		if request.CoolCategoryCode != "" {
			input = models.User{
				UserType: "NON_KKJ_MEMBER",
				Status:   "NOT_REGISTERED",
				Roles:    "STANDARD_MEMBER",
				Password: password,
				Token:    token,
			}
		}

		coolCategory, err := uu.ccr.GetByCode(ctx, strings.ToUpper(request.CoolCategoryCode))
		if err != nil {
			return nil, err
		}

		if coolCategory.ID == 0 {
			return nil, models.ErrorDataNotFound
		}

		input = models.User{
			UserType:         "NON_KKJ_MEMBER",
			Status:           "NOT_REGISTERED",
			Roles:            "COOL_MEMBER",
			CoolCategoryCode: request.CoolCategoryCode,
			Password:         password,
			Token:            token,
		}

		if err := uu.ur.Update(ctx, &input); err != nil {
			return nil, err
		}

	case exist.ID != 0 && exist.UserType != "NON_KKJ_MEMBER":
		return nil, models.ErrorAlreadyExist
	default:
		accountNumber, err := generator.AccountNumber()
		if err != nil {
			return nil, err
		}

		input := models.User{
			CommunityID:    accountNumber,
			Name:             request.Name,
			PhoneNumber:      fmt.Sprintf("+62%s", request.PhoneNumber),
			Email:            strings.ToLower(request.Email),
			UserType:         "REQUEST_COOL",
			Status:           "NOT_REGISTERED",
			Gender:           request.Gender,
			CampusCode:       request.CampusCode,
			CoolCategoryCode: request.CoolCategoryCode,
			MaritalStatus:    request.MaritalStatus,
		}

		if err := uu.ur.Create(ctx, &input); err != nil {
			return nil, err
		}
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
