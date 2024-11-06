package usecases

import (
	"context"
	"fmt"
	"go-community/internal/common"
	"go-community/internal/config"
	"go-community/internal/models"
	"go-community/internal/pkg/generator"
	"go-community/internal/pkg/hash"
	"go-community/internal/repositories/pgsql"
	"strings"
)

type UserUsecase interface {
	CreateUser(ctx context.Context, request *models.CreateUserRequest) (user *models.User, err error)
	Login(ctx context.Context) (user *models.User, err error)
	GetByAccountNumber(ctx context.Context, accountNumber string) (user *models.User, err error)
}

type userUsecase struct {
	ur  pgsql.UserRepository
	cr  pgsql.CampusRepository
	ccr pgsql.CoolCategoryRepository
	clr pgsql.CoolRepository
	cfg *config.Configuration
	s   []byte
}

func NewUserUsecase(ur pgsql.UserRepository, cr pgsql.CampusRepository, ccr pgsql.CoolCategoryRepository, clr pgsql.CoolRepository, cfg config.Configuration, s []byte) *userUsecase {
	return &userUsecase{
		ur:  ur,
		cr:  cr,
		ccr: ccr,
		clr: clr,
		cfg: &cfg,
		s:   s,
	}
}

func (uu *userUsecase) CreateVolunteer(ctx context.Context, request *models.CreateVolunteerRequest) (user *models.User, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	dataExist, err := uu.ur.CheckByEmailPhoneNumber(ctx, strings.ToLower(request.PhoneNumber), strings.ToLower(request.Email))
	if err != nil {
		return nil, err
	}

	if dataExist {
		return nil, models.ErrorAlreadyExist
	}

	_, departmentExist := uu.cfg.Department[strings.ToUpper(request.DepartmentCode)]
	if !departmentExist {
		return nil, models.ErrorDataNotFound
	}

	coolExist, err := uu.clr.CheckById(ctx, request.CoolID)
	if err != nil {
		return nil, err
	}

	if !coolExist {
		return nil, models.ErrorDataNotFound
	}

	_, campusExist := uu.cfg.Campus[strings.ToUpper(request.CampusCode)]
	if !campusExist {
		return nil, models.ErrorDataNotFound
	}

	var accountNumber string
	//if request.JemaatId != "" && request.KKJNumber != "" {
	//	accountNumber = request.JemaatId
	//} else {
	//	accountNumber = generator.LuhnAccountNumber()
	//}

	switch {
	case request.JemaatId != "" && request.KKJNumber != "":
		accountNumber = request.JemaatId
	case request.JemaatId != "" && request.KKJNumber == "":
		return nil, models.ErrorDidNotFillKKJNumber
	default:
		accountNumber = generator.LuhnAccountNumber()
	}

	salted := append([]byte(request.Password), uu.s...)
	password, err := hash.Generate(salted)
	if err != nil {
		return nil, err
	}

	input := models.User{
		CommunityID:   accountNumber,
		Name:          common.CapitalizeFirstWord(request.Name),
		PhoneNumber:   request.PhoneNumber,
		Email:         strings.ToLower(request.Email),
		Password:      password,
		UserType:      "VOLUNTEER",
		Status:        "active",
		Roles:         "TODOYAGERALD",
		Gender:        strings.ToLower(request.Gender),
		Address:       request.Address,
		CampusCode:    request.CampusCode,
		CoolID:        request.CoolID,
		Department:    request.DepartmentCode,
		PlaceOfBirth:  request.PlaceOfBirth,
		DateOfBirth:   &request.DateOfBirth,
		MaritalStatus: request.MaritalStatus,
		KKJNumber:     request.KKJNumber,
		JemaatID:      request.JemaatId,
		IsBaptized:    request.Baptis,
		IsKom100:      request.KOM100,
	}

	if err := uu.ur.Create(ctx, &input); err != nil {
		return nil, err
	}

	// TODO: TOKEN
	//tokenStatus := "active"
	//bearerToken, err := euu.a.Generate(accountNumber, input.Role, tokenStatus)
	//if err != nil {
	//	return nil, "", err
	//}
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
			CommunityID:      accountNumber,
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
