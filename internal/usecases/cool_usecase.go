package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	indonesiaAPI "go-community/internal/clients/indonesia-api"
	"go-community/internal/common"
	"go-community/internal/config"
	"go-community/internal/constants"
	"go-community/internal/models"
	"go-community/internal/pkg/errorgen"
	"go-community/internal/pkg/generator"
	"go-community/internal/repositories/pgsql"
	"strings"

	"gorm.io/gorm"
)

type CoolUsecase interface {
	Create(ctx context.Context, request models.CreateCoolRequest) (response *models.CreateCoolResponse, err error)
	GetAll(ctx context.Context) (response []models.GetAllCoolListResponse, err error)
	GetByCommunityId(ctx context.Context, communityId string) (response *models.GetCoolDetailResponse, err error)
	GetMemberByCode(ctx context.Context, param models.GetCoolMemberByCoolCodeParameter) (response []models.GroupedCoolMembers, err error)
	AddMemberByCode(ctx context.Context, coolCode string, requests []models.AddCoolMemberRequest) (response *models.AddCoolMemberResponse, err error)
}

type coolUsecase struct {
	r    pgsql.PostgreRepositories
	cfg  config.Configuration
	flag FeatureFlagUsecase
	i    indonesiaAPI.Client
}

func NewCoolUsecase(r pgsql.PostgreRepositories, cfg config.Configuration, flag FeatureFlagUsecase, i indonesiaAPI.Client) *coolUsecase {
	return &coolUsecase{
		r:    r,
		cfg:  cfg,
		flag: flag,
		i:    i,
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
		LocationType:            request.Location.Type,
		LocationAreaCode:        request.Location.AreaCode,
		LocationDistrictCode:    request.Location.DistrictCode,
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
		Location: models.CoolLocationResponse{
			Type:         cool.LocationType,
			AreaCode:     cool.LocationAreaCode,
			DistrictCode: cool.LocationDistrictCode,
		},
		Status: constants.MapStatus[constants.STATUS_ACTIVE],
	}

	return &res, nil
}

func (clu *coolUsecase) GetAll(ctx context.Context, header string, userType []string, communityId string) (response []models.GetAllCoolListResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	switch header {
	case "option":
		cools, err := clu.r.Cool.GetAllOptions(ctx)
		if err != nil {
			return nil, err
		}

		list := make([]models.GetAllCoolListResponse, len(cools))
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

			list[i] = models.GetAllCoolListResponse{
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
	case "list":
		isFacilitator := common.CheckOneDataInList(userType, []string{"cool-facilitator"})
		isAdmin := common.CheckOneDataInList(userType, []string{"cool-admin"})

		switch {
		case isFacilitator && isAdmin:
			// handle both
		case isFacilitator:
			// handle facilitator logic
			cools, err := clu.r.Cool.GetAllByFacilitatorCommunityId(ctx, communityId)
			if err != nil {
				return nil, errorgen.Error(err)
			}

			list := make([]models.GetAllCoolListResponse, len(cools))
			for i, e := range cools {
				var campusName string
				if e.CampusCode != "" {
					value, campus := clu.cfg.Campus[strings.ToLower(e.CampusCode)]
					if !campus {
						return nil, models.ErrorDataNotFound
					}
					campusName = value
				}

				leaderData, err := clu.r.User.GetManyNamesByCommunityId(ctx, e.LeaderCommunityIds)
				if err != nil {
					return nil, err
				}

				var leaders []models.CoolLeaderAndCoreResponse
				for _, v := range leaderData {
					leaders = append(leaders, models.CoolLeaderAndCoreResponse{
						Type:        models.TYPE_USER,
						CommunityId: v.CommunityId,
						Name:        v.Name,
					})
				}

				facilitatorData, err := clu.r.User.GetManyNamesByCommunityId(ctx, e.LeaderCommunityIds)
				if err != nil {
					return nil, err
				}

				var facilitators []models.CoolLeaderAndCoreResponse
				for _, v := range facilitatorData {
					facilitators = append(facilitators, models.CoolLeaderAndCoreResponse{
						Type:        models.TYPE_USER,
						CommunityId: v.CommunityId,
						Name:        v.Name,
					})
				}

				areaName, err := clu.i.GetCities(common.StringTrimSpaceAndLower(e.CampusCode), e.LocationAreaCode)
				if err != nil {
					return nil, errorgen.Error(err)
				}

				list[i] = models.GetAllCoolListResponse{
					Type:             models.TYPE_COOL,
					Code:             e.Code,
					Name:             e.Name,
					CampusCode:       e.CampusCode,
					CampusName:       campusName,
					Facilitators:     facilitators,
					Leaders:          leaders,
					Category:         e.Category,
					Gender:           e.Gender,
					LocationType:     e.LocationType,
					LocationAreaCode: e.LocationAreaCode,
					LocationAreaName: areaName[0].Name,
					Status:           e.Status,
				}
			}

			return list, nil
		case isAdmin:
			// handle admin logic
		default:
			return nil, errorgen.Error(errorgen.ForbiddenRole) // no valid role
		}
	default:
		return nil, errorgen.Error(errorgen.InvalidInput, "Invalid header value")
	}

	return
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
		Location: models.CoolLocationResponse{
			Type:         cool.LocationType,
			AreaCode:     cool.LocationAreaCode,
			DistrictCode: cool.LocationDistrictCode,
		},
		Status: cool.Status,
	}, nil
}

// func (clu *coolUsecase) GetMemberById(ctx context.Context, code string) (response []models.GroupedCoolMembers, err error) {
// 	defer func() {
// 		LogService(ctx, err)
// 	}()

// 	existCool, err := clu.r.Cool.CheckByCode(ctx, code)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if !existCool {
// 		return nil, models.ErrorDataNotFound
// 	}

// 	members, err := clu.r.Cool.GetCoolMemberByCode(ctx, code)
// 	if err != nil {
// 		return nil, err
// 	}

// 	facilitators, err := clu.r.Cool.GetCoolFacilitatorByCode(ctx, code)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(members) == 0 {
// 		return nil, models.ErrorDataNotFound
// 	}

// 	var nonFacilitatorResponse []models.GetCoolMemberResponse
// 	var facilitatorResponse []models.GetCoolMemberResponse
// 	if facilitators != nil {
// 		for _, facilitator := range facilitators {
// 			var userTypeOutputs []models.UserTypeDBOutput
// 			if err := json.Unmarshal(facilitator.UserTypes, &userTypeOutputs); err != nil {
// 				// Handle error
// 				return nil, err
// 			}

// 			facilitatorResponse = append(facilitatorResponse, models.GetCoolMemberResponse{
// 				Type:        models.TYPE_USER,
// 				CommunityId: facilitator.CommunityID,
// 				Name:        facilitator.Name,
// 				CoolCode:    facilitator.CoolCode,
// 				UserType:    userTypeOutputs,
// 			})
// 		}
// 	}

// 	for _, member := range members {
// 		var userTypeOutputs []models.UserTypeDBOutput
// 		if err := json.Unmarshal(member.UserTypes, &userTypeOutputs); err != nil {
// 			// Handle error
// 			return nil, err
// 		}

// 		nonFacilitatorResponse = append(nonFacilitatorResponse, models.GetCoolMemberResponse{
// 			Type:        models.TYPE_USER,
// 			CommunityId: member.CommunityID,
// 			Name:        member.Name,
// 			CoolCode:    member.CoolCode,
// 			UserType:    userTypeOutputs,
// 		})
// 	}

// 	var allCoolMembers []models.GroupedCoolMembers
// 	allCoolMembers = append(allCoolMembers, models.GroupMembersBySelectedTypes(facilitatorResponse, []string{constants.USER_TYPE_COOL_FACILITATOR})...)
// 	allCoolMembers = append(allCoolMembers, models.GroupMembersBySelectedTypes(nonFacilitatorResponse, []string{constants.USER_TYPE_COOL_LEADER, constants.USER_TYPE_COOL_CORE})...)
// 	return allCoolMembers, nil
// }

func (clu *coolUsecase) GetMemberByCode(ctx context.Context, param models.GetCoolMemberByCoolCodeParameter) (response []models.GroupedCoolMembers, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	existCool, err := clu.r.Cool.CheckByCode(ctx, param.Code)
	if err != nil {
		return nil, err
	}

	if !existCool {
		return nil, models.ErrorDataNotFound
	}

	members, err := clu.r.Cool.GetAllMembersByCode(ctx, param.Code)
	if err != nil {
		return nil, err
	}

	if len(members) == 0 {
		return nil, models.ErrorDataNotFound
	}

	var undividedRes []models.GetCoolMemberResponse
	for _, member := range members {
		var userTypeOutputs []models.UserTypeSimplifyResponse
		if err := json.Unmarshal(member.UserTypes, &userTypeOutputs); err != nil {
			// Handle error
			return nil, err
		}

		undividedRes = append(undividedRes, models.GetCoolMemberResponse{
			Type:        models.TYPE_USER,
			CommunityId: member.CommunityID,
			Name:        member.Name,
			CoolCode:    member.CoolCode,
			UserType:    userTypeOutputs,
		})
	}

	if len(param.Type) > 0 {
		coolUserTypes, found := constants.CoolUserType.LookupValuesArray(param.Type)
		if !found {
			return nil, models.ErrorInvalidInput
		}

		response = append(response, models.GroupMembersBySelectedTypes(undividedRes, coolUserTypes)...)
		return response, nil
	}

	allCoolTypes := constants.CoolUserType.GetAllKeys()
	response = append(response, models.GroupMembersBySelectedTypes(undividedRes, allCoolTypes)...)

	return response, nil
}

func (clu *coolUsecase) AddMemberByCode(ctx context.Context, requestingUserType []string, coolCode string, requests []models.AddCoolMemberRequest) (response *models.AddCoolMemberResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	// Get Data from COOL Table
	cool, err := clu.r.Cool.GetOneByCode(ctx, coolCode)
	if err != nil {
		return response, errorgen.Error(err)
	}

	// Check if COOL exist
	if cool.Code == "" {
		return response, errorgen.Error(errorgen.DataNotFound)
	}

	// Here we check user one by one
	var memberRes []models.AddedMemberResponse
	for _, request := range requests {
		// Get Community ID, User Type, Roles, and Cool Code here
		memberData, err := clu.r.User.GetRBAC(ctx, request.CommunityId)
		if err != nil {
			return nil, errorgen.Error(err)
		}

		// Check if user exist
		if memberData.CommunityId == "" {
			return nil, errorgen.Error(errorgen.DataNotFound)
		}

		// Logic: If user is already in another COOL, return error
		if memberData.CoolCode != "" || memberData.CoolCode == coolCode {
			return nil, errorgen.Error(errorgen.AlreadyExist, "User with communityId %s already in another COOL. Please contact the respective COOL Leader for the member adjustment", memberData.CommunityId)
		}

		// Check if user type is valid (only user type related to COOL)
		coolUserType, found := constants.CoolUserType.LookupValue(request.UserType)
		if !found {
			return nil, models.ErrorInvalidInput
		}

		existingUserTypes := []string(memberData.UserTypes) // convert pq.StringArray to []string
		// Check if user type already exist
		if common.CheckOneDataInList([]string{*coolUserType}, existingUserTypes) {
			fmt.Println("kena sini ye")
			continue
		}

		// combine the requested user type with existing user type in User Table
		userTypes := common.CombineMapStrings(existingUserTypes, []string{*coolUserType})

		// Logic to add community ids to cool table
		switch *coolUserType {
		case constants.USER_TYPE_COOL_FACILITATOR:
			if !common.CheckOneDataInList([]string{constants.USER_TYPE_COOL_ADMIN}, requestingUserType) {
				return nil, errorgen.Error(errorgen.ErrForbidden, "Only COOL Admin can add COOL Facilitator")
			}

			if common.CheckOneDataInList([]string{request.CommunityId}, cool.FacilitatorCommunityIds) && memberData.CoolCode == cool.Code {
				return nil, errorgen.Error(errorgen.AlreadyExist, "User with communityId %s already in another COOL. Please contact the respective COOL Leader for the member adjustment", memberData.CommunityId)
			}

			cool.FacilitatorCommunityIds = append(cool.FacilitatorCommunityIds, memberData.CommunityId)
		case constants.USER_TYPE_COOL_LEADER:
			if !common.CheckOneDataInList([]string{constants.USER_TYPE_COOL_ADMIN, constants.USER_TYPE_COOL_FACILITATOR}, requestingUserType) {
				return nil, errorgen.Error(errorgen.ErrForbidden, "Only COOL Admin and Facilitator can add COOL Leader")
			}

			if common.CheckOneDataInList([]string{request.CommunityId}, cool.LeaderCommunityIds) && memberData.CoolCode == cool.Code {
				return nil, errorgen.Error(errorgen.AlreadyExist, "User with communityId %s already in another COOL. Please contact the respective COOL Leader for the member adjustment", memberData.CommunityId)
			}

			cool.LeaderCommunityIds = append(cool.LeaderCommunityIds, memberData.CommunityId)
		case constants.USER_TYPE_COOL_CORE:
			if !common.CheckOneDataInList([]string{constants.USER_TYPE_COOL_LEADER}, requestingUserType) {
				return nil, errorgen.Error(errorgen.ForbiddenRole, "Only COOL Leader can add their core team")
			}

			if common.CheckOneDataInList([]string{request.CommunityId}, cool.CoreCommunityIds) && memberData.CoolCode == cool.Code {
				return nil, errorgen.Error(errorgen.AlreadyExist, "User with communityId %s already in another COOL. Please contact the respective COOL Leader for the member adjustment", memberData.CommunityId)
			}

			cool.CoreCommunityIds = append(cool.CoreCommunityIds, memberData.CommunityId)
		}

		// update user's user types and cool id into user table
		if err := clu.r.User.UpdateCoolTeamsByCommunityId(ctx, request.CommunityId, coolCode, userTypes); err != nil {
			return nil, errorgen.Error(err)
		}

		// update user's cool data here
		if err := clu.r.Cool.Update(ctx, &cool); err != nil {
			return nil, errorgen.Error(err)
		}

		memberRes = append(memberRes, models.AddedMemberResponse{
			Type:        models.TYPE_USER,
			CommunityId: request.CommunityId,
			UserType:    request.UserType,
		})
	}

	return &models.AddCoolMemberResponse{
		Type:         models.TYPE_COOL,
		CoolCode:     coolCode,
		AddedMembers: memberRes,
	}, nil
}

func (clu *coolUsecase) DeleteMemberByCode(ctx context.Context, requestingUserType []string, request models.DeleteCoolMemberRequest) (err error) {
	defer func() {
		LogService(ctx, err)
	}()

	// Get Data from COOL Table
	cool, err := clu.r.Cool.GetOneByCode(ctx, request.CoolCode)
	if err != nil {
		return errorgen.Error(err)
	}

	// Check if COOL exist
	if cool.Code == "" {
		return errorgen.Error(errorgen.DataNotFound)
	}

	// Get Community ID, User Type, Roles, and Cool Code here
	member, err := clu.r.User.GetRBAC(ctx, request.CommunityId)
	if err != nil {
		return errorgen.Error(err)
	}

	// Check if user is not exist
	if member.CommunityId == "" || member.CoolCode == "" {
		return errorgen.Error(errorgen.DataNotFound)
	}

	// Check if user's cool code is the same as the request
	if request.CoolCode != member.CoolCode {
		return errorgen.Error(errorgen.InvalidData, "Member with communityId %s is not in COOL %s", member.CommunityId, request.CoolCode)
	}

	// Remove community id from cool table (depend on user's cool user type)
	if common.CheckOneDataInList(cool.FacilitatorCommunityIds, []string{request.CommunityId}) {
		if !common.CheckOneDataInList([]string{constants.USER_TYPE_COOL_ADMIN}, requestingUserType) {
			return errorgen.Error(errorgen.ErrForbidden, "Only COOL Admin can delete COOL Facilitator")
		}

		cool.FacilitatorCommunityIds = common.RemoveSliceIfExact(cool.FacilitatorCommunityIds, []string{request.CommunityId})
	} else if common.CheckOneDataInList(cool.LeaderCommunityIds, []string{request.CommunityId}) {
		if !common.CheckOneDataInList([]string{constants.USER_TYPE_COOL_ADMIN, constants.USER_TYPE_COOL_FACILITATOR}, requestingUserType) {
			return errorgen.Error(errorgen.ErrForbidden, "Only COOL Admin and Facilitator can delete COOL Leader")
		}

		cool.LeaderCommunityIds = common.RemoveSliceIfExact(cool.LeaderCommunityIds, []string{request.CommunityId})
	} else if common.CheckOneDataInList(cool.CoreCommunityIds, []string{request.CommunityId}) {
		if !common.CheckOneDataInList([]string{constants.USER_TYPE_COOL_LEADER}, requestingUserType) {
			return errorgen.Error(errorgen.ErrForbidden, "Only COOL Leader can delete COOL Leader")
		}

		cool.CoreCommunityIds = common.RemoveSliceIfExact(cool.CoreCommunityIds, []string{request.CommunityId})
	}

	existingUserTypes := []string(member.UserTypes) // convert pq.StringArray to []string
	coolUserTypes := []string{constants.USER_TYPE_COOL_CORE, constants.USER_TYPE_COOL_FACILITATOR, constants.USER_TYPE_COOL_LEADER, constants.USER_TYPE_COOL_MEMBER}

	// Check if user's user types have one of cool user types
	if !common.CheckOneDataInList(coolUserTypes, existingUserTypes) {
		return errorgen.Error(errorgen.InvalidData, "Member with communityId %s is not in COOL %s", member.CommunityId, request.CoolCode)
	}

	// Remove user's cool user types in user table
	newUserTypes := common.RemoveSliceIfExact(existingUserTypes, coolUserTypes)
	if err := clu.r.User.UpdateCoolTeamsByCommunityId(ctx, request.CommunityId, "", newUserTypes); err != nil {
		return errorgen.Error(err)
	}

	// update the cool data in cool table (the removed community id)
	if err := clu.r.Cool.Update(ctx, &cool); err != nil {
		return errorgen.Error(err)
	}

	return nil
}

func (clu *coolUsecase) UpdateMember(ctx context.Context, parameter models.UpdateRoleMemberParameter, request models.UpdateRoleMemberRequest, requestingUserType []string) (response *models.UpdateRoleMemberResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	// Get Data from COOL Table
	cool, err := clu.r.Cool.GetOneByCode(ctx, parameter.CoolCode)
	if err != nil {
		return nil, errorgen.Error(err)
	}

	// Check if COOL exist
	if cool.Code == "" {
		return nil, errorgen.Error(errorgen.DataNotFound)
	}

	// Get Community ID, User Type, Roles, and Cool Code here
	member, err := clu.r.User.GetRBAC(ctx, parameter.CommunityId)
	if err != nil {
		return nil, errorgen.Error(err)
	}

	// Check if user is not exist
	if member.CommunityId == "" || member.CoolCode == "" {
		return nil, errorgen.Error(errorgen.DataNotFound)
	}

	// Check if user's cool code is the same as the request
	if parameter.CoolCode != member.CoolCode {
		return nil, errorgen.Error(errorgen.InvalidData, "Member with communityId %s is not in COOL %s", member.CommunityId, parameter.CoolCode)
	}

	previousData := models.PreviousAfterUpdateRoleMember{
		CoolCode: member.CoolCode,
		UserType: member.UserTypes,
	}

	existingUserTypes := []string(member.UserTypes) // convert pq.StringArray to []string
	coolUserTypes := []string{constants.USER_TYPE_COOL_CORE, constants.USER_TYPE_COOL_FACILITATOR, constants.USER_TYPE_COOL_LEADER, constants.USER_TYPE_COOL_MEMBER}

	// Check if user's user types have one of cool user types
	if !common.CheckOneDataInList(coolUserTypes, existingUserTypes) {
		return nil, errorgen.Error(errorgen.InvalidData, "Member with communityId %s is not in COOL %s", member.CommunityId, parameter.CoolCode)
	}

	// Remove user's cool user types in user table
	changedUserTypes := common.RemoveSliceIfExact(existingUserTypes, coolUserTypes)

	// Remove community id from cool table (depend on user's cool user type)
	if common.CheckOneDataInList(cool.FacilitatorCommunityIds, []string{parameter.CommunityId}) {
		if !common.CheckOneDataInList([]string{constants.USER_TYPE_COOL_ADMIN}, requestingUserType) {
			return nil, errorgen.Error(errorgen.ErrForbidden, "Only COOL Admin can delete COOL Facilitator")
		}

		cool.FacilitatorCommunityIds = common.RemoveSliceIfExact(cool.FacilitatorCommunityIds, []string{parameter.CommunityId})
	} else if common.CheckOneDataInList(cool.LeaderCommunityIds, []string{parameter.CommunityId}) {
		if !common.CheckOneDataInList([]string{constants.USER_TYPE_COOL_ADMIN, constants.USER_TYPE_COOL_FACILITATOR}, requestingUserType) {
			return nil, errorgen.Error(errorgen.ErrForbidden, "Only COOL Admin and Facilitator can delete COOL Leader")
		}

		cool.LeaderCommunityIds = common.RemoveSliceIfExact(cool.LeaderCommunityIds, []string{parameter.CommunityId})
	} else if common.CheckOneDataInList(cool.CoreCommunityIds, []string{parameter.CommunityId}) {
		if !common.CheckOneDataInList([]string{constants.USER_TYPE_COOL_LEADER}, requestingUserType) {
			return nil, errorgen.Error(errorgen.ErrForbidden, "Only COOL Leader can delete COOL Leader")
		}

		cool.CoreCommunityIds = common.RemoveSliceIfExact(cool.CoreCommunityIds, []string{parameter.CommunityId})
	}

	switch request.UserType {
	case constants.USER_TYPE_COOL_FACILITATOR:
		if !common.CheckOneDataInList([]string{constants.USER_TYPE_COOL_ADMIN}, requestingUserType) {
			return nil, errorgen.Error(errorgen.ErrForbidden, "Only COOL Admin can delete COOL Facilitator")
		}

		cool.Code = clu.cfg.Cool.FacilitatorCode
		changedUserTypes = append(changedUserTypes, constants.USER_TYPE_COOL_FACILITATOR)
		cool.FacilitatorCommunityIds = append(cool.FacilitatorCommunityIds, member.CommunityId)
	case constants.USER_TYPE_COOL_LEADER:
		if !common.CheckOneDataInList([]string{constants.USER_TYPE_COOL_ADMIN, constants.USER_TYPE_COOL_FACILITATOR}, requestingUserType) {
			return nil, errorgen.Error(errorgen.ErrForbidden, "Only COOL Admin and Facilitator can delete COOL Leader")
		}

		changedUserTypes = append(changedUserTypes, constants.USER_TYPE_COOL_LEADER)
		cool.LeaderCommunityIds = append(cool.LeaderCommunityIds, member.CommunityId)
	case constants.USER_TYPE_COOL_CORE:
		if !common.CheckOneDataInList([]string{constants.USER_TYPE_COOL_LEADER}, requestingUserType) {
			return nil, errorgen.Error(errorgen.ErrForbidden, "Only COOL Leader can delete COOL Leader")
		}

		changedUserTypes = append(changedUserTypes, constants.USER_TYPE_COOL_CORE)
		cool.CoreCommunityIds = append(cool.CoreCommunityIds, member.CommunityId)
	case constants.USER_TYPE_COOL_MEMBER:
		changedUserTypes = append(changedUserTypes, constants.USER_TYPE_COOL_MEMBER)
	}

	if err := clu.r.User.UpdateCoolTeamsByCommunityId(ctx, member.CommunityId, "", changedUserTypes); err != nil {
		return nil, errorgen.Error(err)
	}
	if err := clu.r.Cool.Update(ctx, &cool); err != nil {
		return nil, errorgen.Error(err)
	}

	afterData := models.PreviousAfterUpdateRoleMember{
		CoolCode: member.CoolCode,
		UserType: changedUserTypes,
	}

	return &models.UpdateRoleMemberResponse{
		Type:        models.TYPE_USER,
		CommunityId: member.CommunityId,
		Previous:    previousData,
		After:       afterData,
	}, nil
}
