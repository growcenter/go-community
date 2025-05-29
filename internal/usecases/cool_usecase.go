package usecases

import (
	"context"
	"encoding/json"
	"go-community/internal/common"
	"go-community/internal/config"
	"go-community/internal/constants"
	"go-community/internal/models"
	"go-community/internal/pkg/generator"
	"go-community/internal/repositories/pgsql"
	"strings"

	"gorm.io/gorm"
)

type CoolUsecase interface {
	Create(ctx context.Context, request models.CreateCoolRequest) (response *models.CreateCoolResponse, err error)
	GetAll(ctx context.Context) (response []models.GetAllCoolOptionsResponse, err error)
	GetByCommunityId(ctx context.Context, communityId string) (response *models.GetCoolDetailResponse, err error)
	GetMemberById(ctx context.Context, id int) (response *models.GetCoolMemberByIdResponse, err error)
}

type coolUsecase struct {
	r    pgsql.PostgreRepositories
	cfg  config.Configuration
	flag FeatureFlagUsecase
}

func NewCoolUsecase(r pgsql.PostgreRepositories, cfg config.Configuration, flag FeatureFlagUsecase) *coolUsecase {
	return &coolUsecase{
		r:    r,
		cfg:  cfg,
		flag: flag,
	}
}

func (clu *coolUsecase) Create(ctx context.Context, request models.CreateCoolRequest) (response *models.CreateCoolResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	campusName, campusExist := clu.cfg.Campus[common.StringTrimSpaceAndLower(request.CampusCode)]
	if !campusExist {
		return nil, models.ErrorDataNotFound
	}

	countFacilitator, err := clu.r.User.CheckMultiple(ctx, request.FacilitatorCommunityIds)
	if err != nil {
		return nil, err
	}

	if int(countFacilitator) != len(request.FacilitatorCommunityIds) {
		return nil, models.ErrorDataNotFound
	}

	countLeader, err := clu.r.User.CheckMultiple(ctx, request.LeaderCommunityIds)
	if err != nil {
		return nil, err
	}

	if int(countLeader) != len(request.LeaderCommunityIds) {
		return nil, models.ErrorDataNotFound
	}

	if len(request.CoreCommunityIds) > 0 {
		countCore, err := clu.r.User.CheckMultiple(ctx, request.CoreCommunityIds)
		if err != nil {
			return nil, err
		}

		if int(countCore) != len(request.CoreCommunityIds) {
			return nil, models.ErrorDataNotFound
		}
	}

	category, found := constants.CommunityOfInterest.LookupValue(common.StringTrimSpaceAndLower(request.Category))
	if !found {
		return nil, models.ErrorDataNotFound
	}

	cool := models.Cool{
		Code:                    generator.TypeId("c"), // C stands for COOL
		Name:                    strings.TrimSpace(request.Name),
		Description:             &request.Description,
		CampusCode:              common.StringTrimSpaceAndUpper(request.CampusCode),
		FacilitatorCommunityIds: request.FacilitatorCommunityIds,
		LeaderCommunityIds:      request.LeaderCommunityIds,
		CoreCommunityIds:        request.CoreCommunityIds,
		Category:                *category,
		Gender:                  &request.Gender,
		Recurrence:              &request.Recurrence,
		LocationType:            request.LocationType,
		LocationName:            &request.LocationName,
		Status:                  constants.MapStatus[constants.STATUS_ACTIVE],
	}

	if err := clu.r.Cool.Create(ctx, &cool); err != nil {
		return nil, err
	}

	for _, facilitatorCommunityId := range request.FacilitatorCommunityIds {
		// Get User Type here
		userRbacs, err := clu.r.User.GetRBAC(ctx, facilitatorCommunityId)
		if err != nil {
			return nil, err
		}

		existingUserTypes := []string(userRbacs.UserTypes) // convert pq.StringArray to []string
		if common.CheckOneDataInList([]string{constants.USER_TYPE_COOL_FACILITATOR}, existingUserTypes) && userRbacs.CoolCode == clu.cfg.Cool.FacilitatorCode {
			continue
		}

		userTypes := common.CombineMapStrings(existingUserTypes, []string{constants.USER_TYPE_COOL_FACILITATOR})
		if err := clu.r.User.UpdateCoolTeamsByCommunityId(ctx, facilitatorCommunityId, clu.cfg.Cool.FacilitatorCode, userTypes); err != nil {
			return nil, err
		}
	}

	for _, leaderCommunityId := range request.LeaderCommunityIds {
		// Get User Type here
		userRbacs, err := clu.r.User.GetRBAC(ctx, leaderCommunityId)
		if err != nil {
			return nil, err
		}

		existingUserTypes := []string(userRbacs.UserTypes) // convert pq.StringArray to []string
		if common.CheckOneDataInList([]string{constants.USER_TYPE_COOL_LEADER}, existingUserTypes) && userRbacs.CoolCode == cool.Code {
			continue
		}

		userTypes := common.CombineMapStrings(existingUserTypes, []string{constants.USER_TYPE_COOL_LEADER})
		if err := clu.r.User.UpdateCoolTeamsByCommunityId(ctx, leaderCommunityId, cool.Code, userTypes); err != nil {
			return nil, err
		}
	}

	for _, coreCommunityId := range request.CoreCommunityIds {
		// Get User Type here
		userRbacs, err := clu.r.User.GetRBAC(ctx, coreCommunityId)
		if err != nil {
			return nil, err
		}

		existingUserTypes := []string(userRbacs.UserTypes) // convert pq.StringArray to []string
		if common.CheckOneDataInList([]string{constants.USER_TYPE_COOL_CORE}, existingUserTypes) && userRbacs.CoolCode == cool.Code {
			continue
		}

		userTypes := common.CombineMapStrings(existingUserTypes, []string{constants.USER_TYPE_COOL_CORE}) // add core team user type to the user
		// update user's user types and cool id here
		if err := clu.r.User.UpdateCoolTeamsByCommunityId(ctx, coreCommunityId, cool.Code, userTypes); err != nil {
			return nil, err
		}
	}

	facilitators, err := clu.r.User.GetUserNamesByMultipleCommunityId(ctx, request.FacilitatorCommunityIds)
	if err != nil {
		return nil, err
	}

	leaders, err := clu.r.User.GetUserNamesByMultipleCommunityId(ctx, request.LeaderCommunityIds)
	if err != nil {
		return nil, err
	}

	core, err := clu.r.User.GetUserNamesByMultipleCommunityId(ctx, request.CoreCommunityIds)
	if err != nil {
		return nil, err
	}

	var facRes []models.CoolLeaderAndCoreResponse
	for _, v := range facilitators {
		facRes = append(facRes, models.CoolLeaderAndCoreResponse{
			Type:        models.TYPE_USER,
			CommunityId: v.CommunityId,
			Name:        v.Name,
		})
	}

	var leadRes []models.CoolLeaderAndCoreResponse
	for _, v := range leaders {
		leadRes = append(leadRes, models.CoolLeaderAndCoreResponse{
			Type:        models.TYPE_USER,
			CommunityId: v.CommunityId,
			Name:        v.Name,
		})
	}

	var coreRes []models.CoolLeaderAndCoreResponse
	for _, v := range core {
		coreRes = append(coreRes, models.CoolLeaderAndCoreResponse{
			Type:        models.TYPE_USER,
			CommunityId: v.CommunityId,
			Name:        v.Name,
		})
	}

	res := models.CreateCoolResponse{
		Type:         models.TYPE_COOL,
		Code:         cool.Code,
		Name:         cool.Name,
		Description:  *cool.Description,
		CampusCode:   cool.CampusCode,
		CampusName:   campusName,
		Facilitators: facRes,
		Leaders:      leadRes,
		CoreTeam:     coreRes,
		Category:     cool.Category,
		Gender:       *cool.Gender,
		Recurrence:   *cool.Recurrence,
		LocationType: cool.LocationType,
		LocationName: *cool.LocationName,
		Status:       constants.MapStatus[constants.STATUS_ACTIVE],
	}

	return &res, nil
}

func (clu *coolUsecase) GetAll(ctx context.Context) (response []models.GetAllCoolOptionsResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	cools, err := clu.r.Cool.GetAllOptions(ctx)
	if err != nil {
		return nil, err
	}

	list := make([]models.GetAllCoolOptionsResponse, len(cools))
	for i, e := range cools {
		var campusName string
		if e.CampusCode != "" {
			value, campus := clu.cfg.Campus[strings.ToLower(e.CampusCode)]
			if !campus {
				return nil, models.ErrorDataNotFound
			}
			campusName = value
		}

		users, err := clu.r.User.GetManyNamesByCommunityId(ctx, e.LeaderCommunityIds)
		if err != nil {
			return nil, err
		}

		var leaders []models.CoolLeaderAndCoreResponse
		for _, v := range users {
			leaders = append(leaders, models.CoolLeaderAndCoreResponse{
				Type:        models.TYPE_USER,
				CommunityId: v.CommunityId,
				Name:        v.Name,
			})
		}

		list[i] = models.GetAllCoolOptionsResponse{
			Type:       models.TYPE_COOL,
			Code:       e.Code,
			Name:       e.Name,
			CampusCode: e.CampusCode,
			CampusName: campusName,
			Leaders:    leaders,
			Status:     e.Status,
		}
	}

	return list, nil
}

func (clu *coolUsecase) GetByCommunityId(ctx context.Context, communityId string) (response *models.GetCoolDetailResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	cool, err := clu.r.Cool.GetOneByCommunityId(ctx, communityId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrorDataNotFound
		}

		return nil, err
	}

	var campusName string
	if cool.CampusCode != "" {
		value, campus := clu.cfg.Campus[strings.ToLower(cool.CampusCode)]
		if !campus {
			return nil, models.ErrorDataNotFound
		}
		campusName = value
	}

	facilitators, err := clu.r.User.GetUserNamesByMultipleCommunityId(ctx, cool.FacilitatorCommunityIds)
	if err != nil {
		return nil, err
	}

	leaders, err := clu.r.User.GetUserNamesByMultipleCommunityId(ctx, cool.LeaderCommunityIds)
	if err != nil {
		return nil, err
	}

	core, err := clu.r.User.GetUserNamesByMultipleCommunityId(ctx, cool.CoreCommunityIds)
	if err != nil {
		return nil, err
	}

	var facRes []models.CoolLeaderAndCoreResponse
	for _, v := range facilitators {
		facRes = append(facRes, models.CoolLeaderAndCoreResponse{
			Type:        models.TYPE_USER,
			CommunityId: v.CommunityId,
			Name:        v.Name,
		})
	}

	var leadRes []models.CoolLeaderAndCoreResponse
	for _, v := range leaders {
		leadRes = append(leadRes, models.CoolLeaderAndCoreResponse{
			Type:        models.TYPE_USER,
			CommunityId: v.CommunityId,
			Name:        v.Name,
		})
	}

	var coreRes []models.CoolLeaderAndCoreResponse
	for _, v := range core {
		coreRes = append(coreRes, models.CoolLeaderAndCoreResponse{
			Type:        models.TYPE_USER,
			CommunityId: v.CommunityId,
			Name:        v.Name,
		})
	}

	return &models.GetCoolDetailResponse{
		Type:         models.TYPE_COOL,
		Code:         cool.Code,
		Name:         cool.Name,
		Description:  *cool.Description,
		CampusCode:   cool.CampusCode,
		CampusName:   campusName,
		Facilitators: facRes,
		Leaders:      leadRes,
		CoreTeam:     coreRes,
		Category:     cool.Category,
		Gender:       *cool.Gender,
		Recurrence:   *cool.Recurrence,
		LocationType: cool.LocationType,
		LocationName: *cool.LocationName,
		Status:       cool.Status,
	}, nil
}

func (clu *coolUsecase) GetMemberById(ctx context.Context, code string) (response []models.GetCoolMemberByIdResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	existCool, err := clu.r.Cool.CheckByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if !existCool {
		return nil, models.ErrorDataNotFound
	}

	members, err := clu.r.Cool.GetCoolMemberByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	if len(members) == 0 {
		return nil, models.ErrorDataNotFound
	}
	for _, member := range members {
		var userTypeOutputs []models.UserTypeDBOutput
		if err := json.Unmarshal(member.UserTypes, &userTypeOutputs); err != nil {
			// Handle error
			return nil, err
		}

		response = append(response, models.GetCoolMemberByIdResponse{
			Type:        models.TYPE_USER,
			CommunityId: member.CommunityID,
			Name:        member.Name,
			CoolCode:    member.CoolCode,
			UserType:    userTypeOutputs,
		})
	}

	return response, nil
}
