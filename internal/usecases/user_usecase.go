package usecases

import (
	"context"
	"fmt"
	"go-community/internal/common"
	"go-community/internal/config"
	"go-community/internal/models"
	"go-community/internal/pkg/authorization"
	"go-community/internal/pkg/generator"
	"go-community/internal/pkg/hash"
	"go-community/internal/repositories/pgsql"
	"strings"
	"time"
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
	utr pgsql.UserTypeRepository
	cfg *config.Configuration
	a   authorization.Auth
	s   []byte
}

func NewUserUsecase(ur pgsql.UserRepository, cr pgsql.CampusRepository, ccr pgsql.CoolCategoryRepository, clr pgsql.CoolRepository, utr pgsql.UserTypeRepository, cfg config.Configuration, s []byte) *userUsecase {
	return &userUsecase{
		ur:  ur,
		cr:  cr,
		ccr: ccr,
		clr: clr,
		utr: utr,
		cfg: &cfg,
		s:   s,
	}
}

func (uu *userUsecase) CreateVolunteer(ctx context.Context, request *models.CreateVolunteerRequest) (user *models.User, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	if request.Email == "" && request.PhoneNumber == "" {
		return nil, models.ErrorEmailPhoneNumberEmpty
	}

	_, departmentExist := uu.cfg.Department[strings.ToLower(request.DepartmentCode)]
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

	_, campusExist := uu.cfg.Campus[strings.ToLower(request.CampusCode)]
	if !campusExist {
		return nil, models.ErrorDataNotFound
	}

	userExist, err := uu.ur.GetOneByEmailPhoneNumber(ctx, common.StringTrimSpaceAndLower(request.Email), common.StringTrimSpaceAndLower(request.PhoneNumber))
	if err != nil {
		return nil, err
	}

	if userExist.ID != 0 {
		salted := append([]byte(request.Password), uu.s...)
		password, err := hash.Generate(salted)
		if err != nil {
			return nil, err
		}

		location, _ := time.LoadLocation("Asia/Jakarta")
		dob, err := common.ParseStringToDatetime("2006-01-02", request.DateOfBirth, location)
		if err != nil {
			return nil, err
		}

		input := models.User{
			//CommunityID:   userExist.CommunityID,
			Name:          strings.TrimSpace(common.CapitalizeFirstWord(request.Name)),
			PhoneNumber:   strings.TrimSpace(request.PhoneNumber),
			Email:         common.StringTrimSpaceAndLower(request.Email),
			Password:      password,
			UserType:      "volunteer",
			Status:        "active",
			Gender:        strings.ToLower(request.Gender),
			Address:       request.Address,
			CampusCode:    request.CampusCode,
			CoolID:        request.CoolID,
			Department:    request.DepartmentCode,
			PlaceOfBirth:  request.PlaceOfBirth,
			DateOfBirth:   &dob,
			MaritalStatus: request.MaritalStatus,
			KKJNumber:     request.KKJNumber,
			JemaatID:      request.JemaatId,
			IsBaptized:    request.IsBaptized,
			IsKom100:      request.IsKOM100,
		}

		if err := uu.ur.UpdateByEmailPhoneNumber(ctx, common.StringTrimSpaceAndLower(request.Email), strings.TrimSpace(request.PhoneNumber), &input); err != nil {
			return nil, err
		}

		input.CommunityID = userExist.CommunityID
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

	location, _ := time.LoadLocation("Asia/Jakarta")
	dob, err := common.ParseStringToDatetime("2006-01-02", request.DateOfBirth, location)
	if err != nil {
		return nil, err
	}

	fmt.Println(dob)

	input := models.User{
		CommunityID:   communityId,
		Name:          strings.TrimSpace(common.CapitalizeFirstWord(request.Name)),
		PhoneNumber:   strings.TrimSpace(request.PhoneNumber),
		Email:         common.StringTrimSpaceAndLower(request.Email),
		Password:      password,
		UserType:      "volunteer",
		Status:        "active",
		Gender:        strings.ToLower(request.Gender),
		Address:       request.Address,
		CampusCode:    request.CampusCode,
		CoolID:        request.CoolID,
		Department:    request.DepartmentCode,
		PlaceOfBirth:  request.PlaceOfBirth,
		DateOfBirth:   &dob,
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

	_, campusExist := uu.cfg.Campus[strings.ToLower(request.CampusCode)]
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
		CommunityID: communityId,
		Name:        common.CapitalizeFirstWord(request.Name),
		PhoneNumber: request.PhoneNumber,
		Email:       strings.ToLower(request.Email),
		Password:    password,
		UserType:    "VOLUNTEER",
		Status:      "active",
		//Roles:         "TODOYAGERALD",
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

func (uu *userUsecase) Login(ctx context.Context, request *models.LoginUserRequest) (usr *models.User, tokens *models.UserToken, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	user, err := uu.ur.GetOneByIdentifier(ctx, common.StringTrimSpaceAndLower(request.Identifier))
	if err != nil {
		return nil, nil, err
	}

	if user.ID == 0 {
		return nil, nil, models.ErrorUserNotFound
	}

	salted := append([]byte(request.Password), uu.s...)
	if err = hash.Validate(user.Password, string(salted)); err != nil {
		return nil, nil, models.ErrorInvalidPassword
	}

	userType, err := uu.utr.GetByType(ctx, strings.ToLower(user.UserType))
	if err != nil {
		return nil, nil, err
	}

	userRoles := models.CombineRoles(userType.Roles, user.Roles)
	tokens, err = uu.a.GenerateTokens(user.CommunityID, user.UserType, userRoles, "active")
	if err != nil {
		return nil, nil, err
	}

	return &user, tokens, nil
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

func (uu *userUsecase) Check(ctx context.Context, identifier string) (isExist bool, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	isExist, err = uu.ur.CheckByEmailPhoneNumber(ctx, identifier, identifier)
	if err != nil {
		return false, err
	}

	return isExist, nil
}
