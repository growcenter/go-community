package usecases

import (
	"context"
	"go-community/internal/common"
	"go-community/internal/config"
	"go-community/internal/constants"
	"go-community/internal/models"
	"go-community/internal/pkg/validator"
	"go-community/internal/repositories/pgsql"
	"strings"
	"time"
)

type CoolNewJoinerUsecase interface {
	Create(ctx context.Context, request *models.CreateCoolNewJoinerRequest) (response *models.CreateCoolNewJoinerResponse, err error)
	GetAll(ctx context.Context) (response []models.GetCoolNewJoinerResponse, info *models.CursorInfo, err error)
	UpdateStatus(ctx context.Context, request *models.UpdateCoolNewJoinerRequest) (response *models.UpdateCoolNewJoinerResponse, err error)
}

type coolNewJoinerUsecase struct {
	r     pgsql.PostgreRepositories
	cfg   *config.Configuration
	cfgDb configDBUsecase
}

func NewCoolNewJoinerUsecase(r pgsql.PostgreRepositories, cfg *config.Configuration, cfgDb configDBUsecase) *coolNewJoinerUsecase {
	return &coolNewJoinerUsecase{
		r:     r,
		cfg:   cfg,
		cfgDb: cfgDb,
	}
}

func (cnju *coolNewJoinerUsecase) Create(ctx context.Context, request *models.CreateCoolNewJoinerRequest) (response *models.CreateCoolNewJoinerResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	phoneNumber, err := validator.PhoneNumber("ID", request.PhoneNumber)
	if err != nil {
		return nil, err
	}

	maritalStatus, found := constants.MaritalStatus.LookupValue(common.StringTrimSpaceAndLower(request.MaritalStatus))
	if !found {
		return nil, models.ErrorDataNotFound
	}

	communityOfInterest, found := constants.CommunityOfInterest.LookupValue(common.StringTrimSpaceAndLower(request.CommunityOfInterest))
	if !found {
		return nil, models.ErrorDataNotFound
	}

	_, campusExist := cnju.cfg.Campus[common.StringTrimSpaceAndLower(request.CampusCode)]
	if !campusExist {
		return nil, models.ErrorDataNotFound
	}

	locationExist, err := cnju.cfgDb.IsLocationExist(ctx, request.Location)
	if err != nil {
		return nil, err
	}

	if !locationExist {
		return nil, models.ErrorDataNotFound
	}

	input := models.CoolNewJoiner{
		Name:                common.CapitalizeFirstWord(strings.TrimSpace(request.Name)),
		MaritalStatus:       *maritalStatus,
		Gender:              request.Gender,
		YearOfBirth:         request.YearOfBirth,
		PhoneNumber:         *phoneNumber,
		Address:             request.Address,
		CommunityOfInterest: *communityOfInterest,
		CampusCode:          common.StringTrimSpaceAndUpper(request.CampusCode),
		Location:            common.CapitalizeFirstWord(request.Location),
		UpdatedBy:           nil,
		Status:              constants.CoolJoinerStatusPending,
	}

	if err := cnju.r.CoolNewJoiner.Create(ctx, &input); err != nil {
		return nil, err
	}

	return &models.CreateCoolNewJoinerResponse{
		Type:                models.TYPE_COOL_NEW_JOINER,
		Name:                input.Name,
		MaritalStatus:       input.MaritalStatus,
		Gender:              request.Gender,
		YearOfBirth:         input.YearOfBirth,
		PhoneNumber:         input.PhoneNumber,
		Address:             input.Address,
		CommunityOfInterest: input.CommunityOfInterest,
		CampusCode:          input.CampusCode,
		Location:            input.Location,
		Status:              input.Status,
	}, nil
}

func (cnju *coolNewJoinerUsecase) GetAll(ctx context.Context, param models.GetAllCoolNewJoinerCursorParam) (res []models.GetCoolNewJoinerResponse, info *models.CursorInfo, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	if param.MaritalStatus != "" {
		maritalStatus, found := constants.MaritalStatus.LookupValue(common.StringTrimSpaceAndLower(param.MaritalStatus))
		if !found {
			return nil, nil, models.ErrorDataNotFound
		}
		param.MaritalStatus = *maritalStatus
	}

	if param.CommunityOfInterest != "" {
		communityOfInterest, found := constants.CommunityOfInterest.LookupValue(common.StringTrimSpaceAndLower(param.CommunityOfInterest))
		if !found {
			return nil, nil, models.ErrorDataNotFound
		}
		param.CommunityOfInterest = *communityOfInterest
	}

	if param.Location != "" {
		locationExist, err := cnju.cfgDb.IsLocationExist(ctx, param.Location)
		if err != nil {
			return nil, nil, err
		}

		if !locationExist {
			return nil, nil, models.ErrorDataNotFound
		}
	}

	if param.CampusCode != "" {
		_, campusExist := cnju.cfg.Campus[common.StringTrimSpaceAndLower(param.CampusCode)]
		if !campusExist {
			return nil, nil, models.ErrorDataNotFound
		}
	}

	if param.PhoneNumber != "" {
		phoneNumber, err := validator.PhoneNumber("ID", param.PhoneNumber)
		if err != nil {
			return nil, nil, err
		}
		param.PhoneNumber = *phoneNumber
	}

	output, pagination, err := cnju.r.CoolNewJoiner.GetAll(ctx, param)
	if err != nil {
		return nil, nil, err
	}

	var response []models.GetCoolNewJoinerResponse
	for _, item := range output {
		var deletedAt string
		if !item.DeletedAt.Time.IsZero() {
			deletedAt = common.FormatDatetimeToString(item.DeletedAt.Time, time.RFC3339)
		}

		var campusName string
		if item.CampusCode != "" {
			value, department := cnju.cfg.Campus[strings.ToLower(item.CampusCode)]
			if !department {
				return nil, nil, models.ErrorDataNotFound
			}
			campusName = value
		}

		response = append(response, models.GetCoolNewJoinerResponse{
			Type:                models.TYPE_COOL_NEW_JOINER,
			ID:                  item.ID,
			Name:                item.Name,
			MaritalStatus:       item.MaritalStatus,
			Gender:              item.Gender,
			YearOfBirth:         item.YearOfBirth,
			PhoneNumber:         item.PhoneNumber,
			Address:             item.Address,
			CommunityOfInterest: item.CommunityOfInterest,
			CampusCode:          item.CampusCode,
			CampusName:          campusName,
			Location:            item.Location,
			UpdatedBy:           item.UpdatedBy,
			Status:              item.Status,
			CreatedAt:           item.CreatedAt,
			UpdatedAt:           item.UpdatedAt,
			DeletedAtString:     deletedAt,
		})
	}

	info = &models.CursorInfo{
		PreviousCursor: pagination.Prev,
		NextCursor:     pagination.Next,
		TotalData:      pagination.Total,
	}

	return response, info, nil
}

func (cnju *coolNewJoinerUsecase) UpdateStatus(ctx context.Context, request *models.UpdateCoolNewJoinerRequest) (response *models.UpdateCoolNewJoinerResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	newJoiner, err := cnju.r.CoolNewJoiner.GetById(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	if newJoiner.ID == 0 {
		return nil, models.ErrorDataNotFound
	}

	if newJoiner.Status == request.Status {
		return nil, models.ErrorInvalidInput
	}

	updater, err := cnju.r.User.GetUserNameByCommunityId(ctx, request.UpdatedBy)
	if err != nil {
		return nil, err
	}

	if err := cnju.r.CoolNewJoiner.Update(ctx, &models.CoolNewJoiner{
		ID:                  newJoiner.ID,
		Name:                newJoiner.Name,
		MaritalStatus:       newJoiner.MaritalStatus,
		Gender:              newJoiner.Gender,
		YearOfBirth:         newJoiner.YearOfBirth,
		PhoneNumber:         newJoiner.PhoneNumber,
		Address:             newJoiner.Address,
		CommunityOfInterest: newJoiner.CommunityOfInterest,
		CampusCode:          newJoiner.CampusCode,
		Location:            newJoiner.Location,
		UpdatedBy:           &updater.Name,
		Status:              request.Status,
	}); err != nil {
		return nil, err
	}

	var deletedAt string
	if !newJoiner.DeletedAt.Time.IsZero() {
		deletedAt = common.FormatDatetimeToString(newJoiner.DeletedAt.Time, time.RFC3339)
	}

	var campusName string
	value, department := cnju.cfg.Campus[strings.ToLower(newJoiner.CampusCode)]
	if !department {
		return nil, models.ErrorDataNotFound
	}
	campusName = value

	return &models.UpdateCoolNewJoinerResponse{
		Type:                models.TYPE_COOL_NEW_JOINER,
		ID:                  newJoiner.ID,
		Name:                newJoiner.Name,
		MaritalStatus:       newJoiner.MaritalStatus,
		Gender:              newJoiner.Gender,
		YearOfBirth:         newJoiner.YearOfBirth,
		PhoneNumber:         newJoiner.PhoneNumber,
		Address:             newJoiner.Address,
		CommunityOfInterest: newJoiner.CommunityOfInterest,
		CampusCode:          newJoiner.CampusCode,
		CampusName:          campusName,
		Location:            newJoiner.Location,
		UpdatedBy:           updater.Name,
		Status:              request.Status,
		CreatedAt:           *newJoiner.CreatedAt,
		UpdatedAt:           *newJoiner.UpdatedAt,
		DeletedAt:           deletedAt,
	}, nil
}
