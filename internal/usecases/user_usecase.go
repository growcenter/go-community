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

	fmt.Println("1ok")

	if existEmail.ID != 0 {
		return nil, models.ErrorAlreadyExist
	}

	existPhone, err := uu.ur.GetByPhoneNumber(ctx, request.PhoneNumber)
	if err != nil {
		return nil, err
	}
	fmt.Println("2ok")
	if existPhone.ID != 0 {
		return nil, models.ErrorAlreadyExist
	}
	fmt.Println("3ok")
	campus, err := uu.cr.GetByCode(ctx, request.CampusCode)
	if err != nil {
		return nil, err
	}
	fmt.Println("4ok")
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
