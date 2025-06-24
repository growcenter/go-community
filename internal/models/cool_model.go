package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

var (
	TYPE_COOL        = "cool"
	TYPE_COOL_MEMBER = "coolMember"
)

type Cool struct {
	ID                      int
	Code                    string
	Name                    string
	Description             *string
	CampusCode              string
	FacilitatorCommunityIds pq.StringArray `gorm:"type:text[]"`
	LeaderCommunityIds      pq.StringArray `gorm:"type:text[]"`
	CoreCommunityIds        pq.StringArray `gorm:"type:text[]"`
	Category                string
	Gender                  *string
	Recurrence              *string
	LocationType            string
	LocationAreaCode        string
	LocationDistrictCode    string
	Status                  string
	CreatedAt               *time.Time
	UpdatedAt               *time.Time
	DeletedAt               sql.NullTime
}

func (c *CreateCoolResponse) ToResponse() CreateCoolResponse {
	return CreateCoolResponse{
		Type:         TYPE_COOL,
		Code:         c.Code,
		Name:         c.Name,
		Description:  c.Description,
		CampusCode:   c.CampusCode,
		CampusName:   c.CampusName,
		Facilitators: c.Facilitators,
		Leaders:      c.Leaders,
		CoreTeam:     c.CoreTeam,
		Category:     c.Category,
		Gender:       c.Gender,
		Recurrence:   c.Recurrence,
		Location:     c.Location,
		Status:       c.Status,
	}
}

type (
	CoolLocationRequest struct {
		Type         string `json:"type" validate:"required,oneof=offline onsite hybrid"`
		AreaCode     string `json:"areaCode" validate:"required"`
		DistrictCode string `json:"districtCode" validate:"required"`
	}
	CoolLocationResponse struct {
		Type         string `json:"type"`
		AreaCode     string `json:"areaCode"`
		AreaName     string `json:"areaName,omitempty"`
		DistrictCode string `json:"districtCode"`
		DistrictName string `json:"districtName,omitempty"`
	}
)

type (
	CreateCoolRequest struct {
		Name                    string              `json:"name" validate:"required,min=1,max=50,nospecial" example:"Professionals"`
		Description             string              `json:"description"`
		CampusCode              string              `json:"campusCode" validate:"required,min=3,max=3"`
		FacilitatorCommunityIds []string            `json:"facilitatorCommunityIds" validate:"required"`
		LeaderCommunityIds      []string            `json:"leaderCommunityIds" validate:"required"`
		CoreCommunityIds        []string            `json:"coreCommunityIds" validate:"omitempty"`
		Category                string              `json:"category" validate:"required"`
		Gender                  string              `json:"gender" validate:"omitempty,oneof=male female all"`
		Recurrence              string              `json:"recurrence"`
		Location                CoolLocationRequest `json:"location" validate:"required"`
	}
	CreateCoolResponse struct {
		Type         string                      `json:"type"`
		Code         string                      `json:"code"`
		Name         string                      `json:"name"`
		Description  string                      `json:"description"`
		CampusCode   string                      `json:"campusCode"`
		CampusName   string                      `json:"campusName"`
		Facilitators []CoolLeaderAndCoreResponse `json:"facilitators"`
		Leaders      []CoolLeaderAndCoreResponse `json:"leaders"`
		CoreTeam     []CoolLeaderAndCoreResponse `json:"coreTeam"`
		Category     string                      `json:"category"`
		Gender       string                      `json:"gender"`
		Recurrence   string                      `json:"recurrence"`
		Location     CoolLocationResponse        `json:"location"`
		Status       string                      `json:"status"`
	}
	CoolLeaderAndCoreResponse struct {
		Type        string `json:"type"`
		CommunityId string `json:"communityId"`
		Name        string `json:"name"`
	}
)

type (
	GetAllCoolOptionsDBOutput struct {
		ID                 int
		Code               string
		Name               string
		CampusCode         string
		LeaderCommunityIds pq.StringArray `gorm:"type:text[]"`
		Status             string
	}
)

type GetCoolDetailResponse struct {
	Type         string                      `json:"type"`
	Code         string                      `json:"code"`
	Name         string                      `json:"name"`
	Description  string                      `json:"description"`
	CampusCode   string                      `json:"campusCode"`
	CampusName   string                      `json:"campusName"`
	Facilitators []CoolLeaderAndCoreResponse `json:"facilitators"`
	Leaders      []CoolLeaderAndCoreResponse `json:"leaders"`
	CoreTeam     []CoolLeaderAndCoreResponse `json:"coreTeam"`
	Category     string                      `json:"category"`
	Gender       string                      `json:"gender"`
	Recurrence   string                      `json:"recurrence"`
	Location     CoolLocationResponse        `json:"location"`
	Status       string                      `json:"status"`
}

type (
	GetCoolMembersByIdDBOutput struct {
		CommunityID   string          `gorm:"column:community_id"`
		Name          string          `gorm:"column:name"`
		CoolCode      string          `gorm:"column:cool_code"`
		UserTypes     json.RawMessage `gorm:"column:user_types"` // For the JSON data
		IsFacilitator bool
	}
	// Optional: Define a UserType struct to unmarshal the JSON into
	UserTypeDBOutput struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}
	GetCoolMemberByCoolCodeParameter struct {
		Code string   `json:"code" validate:"required"`
		Type []string `json:"type" validate:"omitempty,dive,oneof=facilitator leader core member"`
	}
	GetCoolMemberResponse struct {
		Type        string                     `json:"type"`
		CommunityId string                     `json:"communityId"`
		Name        string                     `json:"name"`
		CoolCode    string                     `json:"coolCode"`
		UserType    []UserTypeSimplifyResponse `json:"-"`
	}
	GroupedCoolMembers struct {
		Type     string                  `json:"type"`
		UserType string                  `json:"userType"`
		Members  []GetCoolMemberResponse `json:"members"`
	}
)

type (
	AddCoolMemberRequest struct {
		CommunityId string `json:"communityId" validate:"required,communityId"`
		UserType    string `json:"userType" validate:"required,oneof=facilitator leader core member"`
	}
	AddCoolMemberResponse struct {
		Type         string                `json:"type"`
		CoolCode     string                `json:"coolCode"`
		AddedMembers []AddedMemberResponse `json:"addedMembers"`
	}
	AddedMemberResponse struct {
		Type        string `json:"type"`
		CommunityId string `json:"communityId"`
		UserType    string `json:"userType"`
	}
)

type (
	DeleteCoolMemberRequest struct {
		CoolCode    string `json:"coolCode" validate:"required"`
		CommunityId string `json:"communityId" validate:"required,communityId"`
	}
)

type (
	UpdateRoleMemberRequest struct {
		CommunityId string `json:"communityId" validate:"required,communityId"`
		UserType    string `json:"userType" validate:"required,oneof=facilitator leader core member"`
	}
	PreviousAfterUpdateRoleMember struct {
		CoolCode string   `json:"coolCode"`
		Role     []string `json:"role"`
		UserType []string `json:"userType"`
	}
	UpdateRoleMemberResponse struct {
		Type        string                        `json:"type"`
		CommunityId string                        `json:"communityId"`
		Previous    PreviousAfterUpdateRoleMember `json:"previous"`
		After       PreviousAfterUpdateRoleMember `json:"after"`
	}
)

type (
	GetAllCoolListDBOutput struct {
		Code                    string
		Name                    string
		CampusCode              string
		FacilitatorCommunityIds pq.StringArray `gorm:"type:text[]"`
		LeaderCommunityIds      pq.StringArray `gorm:"type:text[]"`
		Category                string
		Gender                  string
		LocationType            string
		LocationAreaCode        string
		Status                  string
	}
	GetAllCoolListResponse struct {
		Type             string                      `json:"type"`
		Code             string                      `json:"code"`
		Name             string                      `json:"name"`
		CampusCode       string                      `json:"campusCode"`
		CampusName       string                      `json:"campusName"`
		Facilitators     []CoolLeaderAndCoreResponse `json:"facilitators,omitempty"`
		Leaders          []CoolLeaderAndCoreResponse `json:"leaders"`
		Category         string                      `json:"category,omitempty"`
		Gender           string                      `json:"gender,omitempty"`
		LocationType     string                      `json:"locationType,omitempty"`
		LocationAreaCode string                      `json:"locationAreaCode,omitempty"`
		LocationAreaName string                      `json:"locationAreaName,omitempty"`
		Status           string                      `json:"status"`
	}
)

// Group logic
func GroupMembersBySelectedTypes(
	members []GetCoolMemberResponse,
	selectedTypes []string,
) []GroupedCoolMembers {
	groupMap := make(map[string][]GetCoolMemberResponse)
	allowed := make(map[string]bool)

	// Convert slice to map for fast lookup
	for _, t := range selectedTypes {
		allowed[t] = true
	}

	for _, member := range members {
		simple := GetCoolMemberResponse{
			Type:        member.Type,
			CommunityId: member.CommunityId,
			Name:        member.Name,
			CoolCode:    member.CoolCode,
			UserType:    member.UserType,
		}

		for _, userType := range member.UserType {
			if allowed[userType.UserType] {
				groupMap[userType.UserType] = append(groupMap[userType.UserType], simple)
			}
		}
	}

	var result []GroupedCoolMembers
	for _, t := range selectedTypes {
		if members, exists := groupMap[t]; exists {
			result = append(result, GroupedCoolMembers{
				Type:     TYPE_USER_TYPE,
				UserType: t,
				Members:  members,
			})
		}
	}

	return result
}
