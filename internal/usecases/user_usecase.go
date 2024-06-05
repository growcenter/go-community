package usecases

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/pkg/generator"
	"go-community/internal/repositories/pgsql"
	"strings"
)

type UserUsecase interface {
	CreateCool(ctx context.Context, request *models.CreateUserCoolRequest) (user *models.User, err error)
}

type userUsecase struct {
	ur  pgsql.UserRepository
	cr  pgsql.CampusRepository
	ccr pgsql.CoolCategoryRepository
}

func NewUserUsecase(ur pgsql.UserRepository, cr pgsql.CampusRepository, ccr pgsql.CoolCategoryRepository) *userUsecase {
	return &userUsecase{
		ur: ur,
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

	accountNumber, err := generator.AccountNumber(&campus, &coolCategory)
	if err != nil {
		return nil, err
	}

	input := models.User{
		AccountNumber:    accountNumber,
		Name:             request.Name,
		PhoneNumber:      request.PhoneNumber,
		Email:            strings.ToLower(request.Email),
		UserType:         "REQUEST_COOL",
		Status:           "active",
		Gender:           request.Gender,
		CampusCode:       request.CampusCode,
		CoolCategoryCode: request.CoolCategoryCode,
		MaritalStatus:    request.MaritalStatus,
	}

	if err := uu.ur.Create(ctx, &input); err != nil {
		return nil, err
	}

	return &input, nil
}
