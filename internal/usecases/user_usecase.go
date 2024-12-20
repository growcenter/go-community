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
	"time"
)

type UserUsecase interface {
	Create(ctx context.Context, request *models.CreateUserRequest) (response *models.CreateUserResponse, err error)
	CreateVolunteer(ctx context.Context, request *models.CreateVolunteerRequest) (user *models.User, err error)
	Login(ctx context.Context, request models.LoginUserRequest) (user *models.User, token string, err error)
	GetByCommunityId(ctx context.Context, request models.GetOneByCommunityIdParameter) (response *models.GetOneByCommunityIdResponse, err error)
	Check(ctx context.Context, identifier string) (isExist bool, err error)
	UpdatePassword(ctx context.Context, param *models.UpdateUserPasswordParam, request *models.UpdateUserPasswordRequest) (user *models.User, err error)
}

type userUsecase struct {
	ur  pgsql.UserRepository
	cr  pgsql.CampusRepository
	ccr pgsql.CoolCategoryRepository
	clr pgsql.CoolRepository
	utr pgsql.UserTypeRepository
	rr  pgsql.RoleRepository
	cfg *config.Configuration
	a   authorization.Auth
	s   []byte
}

func NewUserUsecase(ur pgsql.UserRepository, cr pgsql.CampusRepository, ccr pgsql.CoolCategoryRepository, clr pgsql.CoolRepository, utr pgsql.UserTypeRepository, rr pgsql.RoleRepository, cfg config.Configuration, a authorization.Auth, s []byte) *userUsecase {
	return &userUsecase{
		ur:  ur,
		cr:  cr,
		ccr: ccr,
		clr: clr,
		utr: utr,
		rr:  rr,
		cfg: &cfg,
		a:   a,
		s:   s,
	}
}

func (uu *userUsecase) Create(ctx context.Context, request *models.CreateUserRequest) (response *models.CreateUserResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	if request.Email == "" && request.PhoneNumber == "" {
		return nil, models.ErrorEmailPhoneNumberEmpty
	}

	if request.DepartmentCode != "" {
		_, departmentExist := uu.cfg.Department[strings.ToLower(request.DepartmentCode)]
		if !departmentExist {
			return nil, models.ErrorDataNotFound
		}
	}

	if request.CoolID != 0 {
		coolExist, err := uu.clr.CheckById(ctx, request.CoolID)
		if err != nil {
			return nil, err
		}

		if !coolExist {
			return nil, models.ErrorDataNotFound
		}
	}

	if request.CampusCode != "" {
		_, campusExist := uu.cfg.Campus[strings.ToLower(request.CampusCode)]
		if !campusExist {
			return nil, models.ErrorDataNotFound
		}
	}

	userTypes, err := uu.utr.GetByArray(ctx, request.UserTypes)
	if err != nil {
		return nil, err
	}

	if len(userTypes) <= 0 {
		return nil, models.ErrorDataNotFound
	}

	userExist, err := uu.ur.GetOneByEmailPhoneNumber(ctx, common.StringTrimSpaceAndLower(request.Email), common.StringTrimSpaceAndLower(request.PhoneNumber))
	if err != nil {
		return nil, err
	}

	switch {
	case userExist.ID != 0:
		isInternal := common.ContainsValueInModel(userTypes, func(userType models.UserType) bool {
			return userType.Category == "internal" || userType.Category == "cool"
		})

		if !isInternal {
			return nil, models.ErrorAlreadyExist
		}

		if request.DepartmentCode == "" || request.CoolID == 0 {
			return nil, models.ErrorMissingDepartmentCool
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

		userExist.Name = strings.TrimSpace(common.CapitalizeFirstWord(request.Name))
		userExist.PhoneNumber = strings.TrimSpace(request.PhoneNumber)
		userExist.Email = common.StringTrimSpaceAndLower(request.Email)
		userExist.Password = password
		userExist.UserTypes = request.UserTypes
		userExist.Status = models.UserStatusActive
		userExist.Gender = strings.ToLower(request.Gender)
		userExist.Address = request.Address
		userExist.CampusCode = request.CampusCode
		userExist.CoolID = request.CoolID
		userExist.Department = strings.ToUpper(request.DepartmentCode)
		userExist.PlaceOfBirth = request.PlaceOfBirth
		userExist.DateOfBirth = &dob
		userExist.MaritalStatus = request.MaritalStatus
		userExist.KKJNumber = request.KKJNumber
		userExist.JemaatID = request.JemaatId
		userExist.IsBaptized = request.IsBaptized
		userExist.IsKom100 = request.IsKOM100

		if err := uu.ur.Update(ctx, &userExist); err != nil {
			return nil, err
		}

		response = &models.CreateUserResponse{
			Type:           models.TYPE_USER,
			CommunityId:    userExist.CommunityID,
			Name:           userExist.Name,
			PhoneNumber:    userExist.PhoneNumber,
			Email:          userExist.Email,
			UserTypes:      userExist.UserTypes,
			CampusCode:     userExist.CampusCode,
			PlaceOfBirth:   userExist.PlaceOfBirth,
			DateOfBirth:    userExist.DateOfBirth,
			Address:        userExist.Address,
			Gender:         userExist.Gender,
			DepartmentCode: userExist.Department,
			CoolID:         userExist.CoolID,
			KKJNumber:      userExist.KKJNumber,
			JemaatId:       userExist.JemaatID,
			IsBaptized:     userExist.IsBaptized,
			IsKOM100:       userExist.IsKom100,
			MaritalStatus:  userExist.MaritalStatus,
			Status:         userExist.Status,
		}

		return response, nil
	case userExist.ID == 0:
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

		input := models.User{
			CommunityID:   communityId,
			Name:          strings.TrimSpace(common.CapitalizeFirstWord(request.Name)),
			PhoneNumber:   strings.TrimSpace(request.PhoneNumber),
			Email:         common.StringTrimSpaceAndLower(request.Email),
			Password:      password,
			UserTypes:     request.UserTypes,
			Status:        models.UserStatusActive,
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

		response = &models.CreateUserResponse{
			Type:           models.TYPE_USER,
			CommunityId:    input.CommunityID,
			Name:           input.Name,
			PhoneNumber:    input.PhoneNumber,
			Email:          input.Email,
			UserTypes:      input.UserTypes,
			CampusCode:     input.CampusCode,
			PlaceOfBirth:   input.PlaceOfBirth,
			DateOfBirth:    input.DateOfBirth,
			Address:        input.Address,
			Gender:         input.Gender,
			DepartmentCode: input.Department,
			CoolID:         input.CoolID,
			KKJNumber:      input.KKJNumber,
			JemaatId:       input.JemaatID,
			IsBaptized:     input.IsBaptized,
			IsKOM100:       input.IsKom100,
			MaritalStatus:  input.MaritalStatus,
			Status:         input.Status,
		}

		return response, nil

	default:
		return nil, err
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

		//userExist.Name = strings.TrimSpace(common.CapitalizeFirstWord(request.Name))
		//userExist.PhoneNumber = strings.TrimSpace(request.PhoneNumber)
		//userExist.Email = common.StringTrimSpaceAndLower(request.Email)
		//userExist.Password = password
		//userExist.UserType = "volunteer"
		//userExist.Status = models.UserStatusActive
		//userExist.Gender = strings.ToLower(request.Gender)
		//userExist.Address = request.Address
		//userExist.CampusCode = strings.ToUpper(request.CampusCode)
		//userExist.CoolID = request.CoolID
		//userExist.Department = strings.ToUpper(request.DepartmentCode)
		//userExist.PlaceOfBirth = request.PlaceOfBirth
		//userExist.DateOfBirth = &dob
		//userExist.MaritalStatus = request.MaritalStatus
		//userExist.KKJNumber = request.KKJNumber
		//userExist.JemaatID = request.JemaatId
		//userExist.IsBaptized = request.IsBaptized
		//userExist.IsKom100 = request.IsKOM100
		//
		//if err := uu.ur.Update(ctx, &userExist); err != nil {
		//	return nil, err
		//}
		//
		//return &userExist, nil

		input := models.User{
			//CommunityID:   userExist.CommunityID,
			Name:          strings.TrimSpace(common.CapitalizeFirstWord(request.Name)),
			PhoneNumber:   strings.TrimSpace(request.PhoneNumber),
			Email:         common.StringTrimSpaceAndLower(request.Email),
			Password:      password,
			UserTypes:     request.UserTypes,
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

	input := models.User{
		CommunityID:   communityId,
		Name:          strings.TrimSpace(common.CapitalizeFirstWord(request.Name)),
		PhoneNumber:   strings.TrimSpace(request.PhoneNumber),
		Email:         common.StringTrimSpaceAndLower(request.Email),
		Password:      password,
		UserTypes:     request.UserTypes,
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

	userType, err := uu.utr.GetByArray(ctx, user.UserTypes)
	if err != nil {
		return nil, nil, err
	}

	rolesInUserType, err := common.GetUniqueFieldValuesFromModel(userType, "Roles")
	if err != nil {
		return nil, nil, err
	}

	userRoles := common.CombineMapStrings(rolesInUserType, user.Roles)
	user.Roles = userRoles
	tokens, err = uu.a.GenerateTokens(user.CommunityID, user.UserTypes, userRoles, "active")
	if err != nil {
		return nil, nil, err
	}

	return &user, tokens, nil
}

func (uu *userUsecase) GetByCommunityId(ctx context.Context, request models.GetOneByCommunityIdParameter) (response *models.GetOneByCommunityIdResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	user, err := uu.ur.GetOneByCommunityId(ctx, request.CommunityId)
	if err != nil {
		return nil, err
	}

	if user.ID == 0 {
		return nil, models.ErrorDataNotFound
	}

	campusName, campus := uu.cfg.Campus[strings.ToLower(user.CampusCode)]
	if !campus {
		return nil, models.ErrorDataNotFound
	}

	location, _ := time.LoadLocation("Asia/Jakarta")
	dob := user.DateOfBirth.In(location)

	var departmentName string
	if user.Department != "" {
		value, department := uu.cfg.Department[strings.ToLower(user.Department)]
		if !department {
			return nil, models.ErrorDataNotFound
		}
		departmentName = value
	}

	var coolName string
	if user.CoolID != 0 {
		cool, err := uu.clr.GetNameById(ctx, user.CoolID)
		if err != nil {
			return nil, err
		}

		coolName = cool.Name
	}

	userType, err := uu.utr.GetByArray(ctx, user.UserTypes)
	if err != nil {
		return nil, err
	}

	rolesInUserType, err := common.GetUniqueFieldValuesFromModel(userType, "Roles")
	if err != nil {
		return nil, err
	}

	userRoles := common.CombineMapStrings(rolesInUserType, user.Roles)
	roles, err := uu.rr.GetByArray(ctx, userRoles)
	if err != nil {
		return nil, err
	}

	var rolesRes []models.RoleResponse
	for _, v := range roles {
		rolesRes = append(rolesRes, *v.ToResponse())
	}

	data := models.GetOneByCommunityIdResponse{
		Type:           models.TYPE_USER,
		Name:           user.Name,
		PhoneNumber:    user.PhoneNumber,
		Email:          user.Email,
		CommunityId:    user.CommunityID,
		UserTypes:      user.UserTypes,
		CampusCode:     user.CampusCode,
		CampusName:     campusName,
		PlaceOfBirth:   user.PlaceOfBirth,
		DateOfBirth:    &dob,
		Address:        user.Address,
		Gender:         user.Gender,
		DepartmentCode: user.Department,
		DepartmentName: departmentName,
		CoolID:         user.CoolID,
		CoolName:       coolName,
		KKJNumber:      user.KKJNumber,
		JemaatId:       user.JemaatID,
		IsKOM100:       user.IsKom100,
		IsBaptized:     user.IsBaptized,
		MaritalStatus:  user.MaritalStatus,
		Roles:          rolesRes,
		Status:         user.Status,
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

func (uu *userUsecase) UpdatePassword(ctx context.Context, param *models.UpdateUserPasswordParam, request *models.UpdateUserPasswordRequest) (user *models.User, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	if request.Password != request.ConfirmPassword {
		return nil, models.ErrorMismatchFields
	}

	data, err := uu.ur.GetOneByIdentifier(ctx, strings.ToLower(param.Identifier))
	if err != nil {
		return nil, err
	}

	if data.ID == 0 {
		return nil, models.ErrorDataNotFound
	}

	salted := append([]byte(request.Password), uu.s...)
	password, err := hash.Generate(salted)
	if err != nil {
		return nil, err
	}

	data.Password = password
	if err := uu.ur.Update(ctx, &data); err != nil {
		return nil, err
	}

	return &data, nil
}
