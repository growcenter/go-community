package usecases

import (
	"context"
	"go-community/internal/common"
	"go-community/internal/config"
	"go-community/internal/models"
	"go-community/internal/pkg/authorization"
	"go-community/internal/pkg/generator"
	"go-community/internal/pkg/hash"
	"go-community/internal/repositories/pgsql"
	"strings"
)

type UserUsecase interface {
	CreateVolunteer(ctx context.Context, request *models.CreateVolunteerRequest) (user *models.User, err error)
	CreateUserGeneral(ctx context.Context, request *models.CreateUserRequest) (user *models.User, err error)
	Login(ctx context.Context, request models.LoginUserRequest) (user *models.User, token string, err error)
	GetByAccountNumber(ctx context.Context, accountNumber string) (user *models.User, err error)
	Check(ctx context.Context, identifier string) (isExist bool, err error)
}

type userUsecase struct {
	ur  pgsql.UserRepository
	cr  pgsql.CampusRepository
	ccr pgsql.CoolCategoryRepository
	clr pgsql.CoolRepository
	cfg *config.Configuration
	a   authorization.Auth
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

	if request.Email == "" || request.PhoneNumber == "" {
		return nil, models.ErrorDataNotFound
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

	dataExist, err := uu.ur.CheckByEmailPhoneNumber(ctx, strings.ToLower(request.PhoneNumber), strings.ToLower(request.Email))
	if err != nil {
		return nil, err
	}

	if dataExist {
		var communityId string
		switch {
		case request.JemaatId != "" && request.KKJNumber != "":
			communityId = request.JemaatId
		case request.JemaatId != "" && request.KKJNumber == "":
			return nil, models.ErrorDidNotFillKKJNumber
		default:
			communityId = generator.LuhnAccountNumber()
		}

		salted := append([]byte(request.Password), uu.s...)
		password, err := hash.Generate(salted)
		if err != nil {
			return nil, err
		}

		input := models.User{
			CommunityID:   communityId,
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
			IsBaptized:    request.IsBaptized,
			IsKom100:      request.IsKOM100,
		}

		if err := uu.ur.Update(ctx, &input); err != nil {
			return nil, err
		}

		return &input, nil
	}

	var communityId string
	switch {
	case request.JemaatId != "" && request.KKJNumber != "":
		communityId = request.JemaatId
	case request.JemaatId != "" && request.KKJNumber == "":
		return nil, models.ErrorDidNotFillKKJNumber
	default:
		communityId = generator.LuhnAccountNumber()
	}

	salted := append([]byte(request.Password), uu.s...)
	password, err := hash.Generate(salted)
	if err != nil {
		return nil, err
	}

	input := models.User{
		CommunityID:   communityId,
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
		IsBaptized:    request.IsBaptized,
		IsKom100:      request.IsKOM100,
	}

	if err := uu.ur.Create(ctx, &input); err != nil {
		return nil, err
	}

	return &input, nil
}

func (uu *userUsecase) CreateUserGeneral(ctx context.Context, request *models.CreateUserRequest) (user *models.User, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	if request.Email == "" || request.PhoneNumber == "" {
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

	dataExist, err := uu.ur.CheckByEmailPhoneNumber(ctx, strings.ToLower(request.PhoneNumber), strings.ToLower(request.Email))
	if err != nil {
		return nil, err
	}

	if dataExist {
		return nil, models.ErrorAlreadyExist
	}

	var communityId string
	switch {
	case request.JemaatId != "" && request.KKJNumber != "":
		communityId = request.JemaatId
	case request.JemaatId != "" && request.KKJNumber == "":
		return nil, models.ErrorDidNotFillKKJNumber
	default:
		communityId = generator.LuhnAccountNumber()
	}

	salted := append([]byte(request.Password), uu.s...)
	password, err := hash.Generate(salted)
	if err != nil {
		return nil, err
	}

	input := models.User{
		CommunityID:   communityId,
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
		IsBaptized:    request.IsBaptized,
		IsKom100:      request.IsKom100,
	}

	if err := uu.ur.Create(ctx, &input); err != nil {
		return nil, err
	}

	return &input, nil
}

func (uu *userUsecase) Login(ctx context.Context, request models.LoginUserRequest) (usr *models.User, token string, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	user, err := uu.ur.GetOneByEmailPhone(ctx, strings.ToLower(request.Identifier))
	if err != nil {
		return nil, "", err
	}

	if user.ID == 0 {
		return nil, "", models.ErrorUserNotFound
	}

	salted := append([]byte(request.Password), uu.s...)
	if err = hash.Validate(user.Password, string(salted)); err != nil {
		return nil, "", models.ErrorInvalidPassword
	}

	tokenStatus := "active"
	bearerToken, err := uu.a.Generate(user.CommunityID, strings.ToLower(user.Roles), tokenStatus)
	if err != nil {
		return nil, "", err
	}

	return &user, bearerToken, nil
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
