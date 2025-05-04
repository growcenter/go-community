package models

import (
	"database/sql"
	"github.com/lib/pq"
	"time"
)

var TYPE_COOL = "cool"

type Cool struct {
	ID                      int
	Name                    string
	Description             string
	CampusCode              string
	FacilitatorCommunityIds pq.StringArray `gorm:"type:text[]"`
	LeaderCommunityIds      pq.StringArray `gorm:"type:text[]"`
	CoreCommunityIds        pq.StringArray `gorm:"type:text[]"`
	Category                string
	Gender                  string
	Recurrence              string
	LocationType            string
	LocationName            string
	Status                  string
	CreatedAt               *time.Time
	UpdatedAt               *time.Time
	DeletedAt               sql.NullTime
}

func (c *CreateCoolResponse) ToResponse() CreateCoolResponse {
	return CreateCoolResponse{
		Type:         TYPE_COOL,
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
		LocationType: c.LocationType,
		LocationName: c.LocationName,
		Status:       c.Status,
	}
}

type (
	CreateCoolRequest struct {
		Name                    string   `json:"name" validate:"required,min=1,max=50,nospecial" example:"Professionals"`
		Description             *string  `json:"description"`
		CampusCode              string   `json:"campusCode" validate:"required,min=3,max=3"`
		FacilitatorCommunityIds []string `json:"facilitatorCommunityIds" validate:"required"`
		LeaderCommunityIds      []string `json:"leaderCommunityIds" validate:"required"`
		CoreCommunityIds        []string `json:"coreCommunityIds"`
		Category                string   `json:"category" validate:"required"`
		Gender                  *string  `json:"gender" validate:"omitempty,oneof=male female all"`
		Recurrence              *string  `json:"recurrence"`
		LocationType            string   `json:"locationType" validate:"required,oneof=offline onsite hybrid"`
		LocationName            *string  `json:"locationName"`
	}
	CreateCoolResponse struct {
		Type         string                      `json:"type"`
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
		LocationType string                      `json:"locationType"`
		LocationName string                      `json:"locationName"`
		Status       string                      `json:"status"`
	}
	CoolLeaderAndCoreResponse struct {
		Type        string `json:"type"`
		CommunityId string `json:"communityId"`
		Name        string `json:"name"`
	}
)

func (c *GetAllCoolOptionsResponse) ToResponse() GetAllCoolOptionsResponse {
	return GetAllCoolOptionsResponse{
		Type:       TYPE_COOL,
		ID:         c.ID,
		Name:       c.Name,
		CampusCode: c.CampusCode,
		Leaders:    c.Leaders,
		Status:     c.Status,
	}
}

type (
	GetAllCoolOptionsDBOutput struct {
		ID                 int
		Name               string
		CampusCode         string
		LeaderCommunityIds pq.StringArray `gorm:"type:text[]"`
		Status             string
	}
	GetAllCoolOptionsResponse struct {
		Type       string                      `json:"type"`
		ID         int                         `json:"id"`
		Name       string                      `json:"name"`
		CampusCode string                      `json:"campusCode"`
		CampusName string                      `json:"campusName"`
		Leaders    []CoolLeaderAndCoreResponse `json:"leaders"`
		Status     string                      `json:"status"`
	}
)

type GetCoolDetailResponse struct {
	Type         string                      `json:"type"`
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
	LocationType string                      `json:"locationType"`
	LocationName string                      `json:"locationName"`
	Status       string                      `json:"status"`
}
