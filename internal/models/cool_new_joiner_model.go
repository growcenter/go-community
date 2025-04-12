package models

import (
	"database/sql"
	"time"
)

var TYPE_COOL_NEW_JOINER string = "coolNewJoiner"

type CoolNewJoiner struct {
	ID                  int
	Name                string
	MaritalStatus       string
	Gender              string
	YearOfBirth         int
	PhoneNumber         string
	Address             string
	CommunityOfInterest string
	CampusCode          string
	Location            string
	UpdatedBy           *string
	Status              string
	CreatedAt           *time.Time
	UpdatedAt           *time.Time
	DeletedAt           sql.NullTime
}

func (e *CreateCoolNewJoinerResponse) ToResponse() *CreateCoolNewJoinerResponse {
	return &CreateCoolNewJoinerResponse{
		Type:                TYPE_COOL_NEW_JOINER,
		Name:                e.Name,
		MaritalStatus:       e.MaritalStatus,
		Gender:              e.Gender,
		YearOfBirth:         e.YearOfBirth,
		PhoneNumber:         e.PhoneNumber,
		Address:             e.Address,
		CommunityOfInterest: e.CommunityOfInterest,
		CampusCode:          e.CampusCode,
		Location:            e.Location,
		Status:              e.Status,
	}
}

type (
	CreateCoolNewJoinerRequest struct {
		Name                string `json:"name" validate:"required,min=1,max=255" example:"John Doe"`
		MaritalStatus       string `json:"maritalStatus" validate:"required,oneof=single married divorced widowed" example:"single"`
		Gender              string `json:"gender" validate:"required,oneof=male female"`
		YearOfBirth         int    `json:"yearOfBirth" validate:"required,min=1900,max=2023" example:"1990"`
		PhoneNumber         string `json:"phoneNumber" validate:"required,phoneFormat0862" example:"+628123456789"`
		Address             string `json:"address" validate:"required,min=1,max=255" example:"123 Main St, Jakarta"`
		CommunityOfInterest string `json:"communityOfInterest" validate:"required,min=1,max=255" example:"Technology, Sports"`
		CampusCode          string `json:"campusCode" validate:"required,min=3,max=3" example:"001"`
		Location            string `json:"location" validate:"required" example:"Bekasi Timur"`
	}
	CreateCoolNewJoinerResponse struct {
		Type                string `json:"type" example:"coolNewJoiner"`
		Name                string `json:"name" example:"John Doe"`
		MaritalStatus       string `json:"maritalStatus" example:"single"`
		Gender              string `json:"gender" example:"male"`
		YearOfBirth         int    `json:"yearOfBirth" example:"1990"`
		PhoneNumber         string `json:"phoneNumber" example:"+628123456789"`
		Address             string `json:"address" example:"123 Main St, Jakarta"`
		CommunityOfInterest string `json:"communityOfInterest" example:"Technology, Sports"`
		CampusCode          string `json:"campusCode" example:"001"`
		Location            string `json:"location" example:"Bekasi Timur"`
		Status              string `json:"status" example:"active"`
	}
)

type (
	UpdateCoolNewJoinerRequest struct {
		Status    string `json:"status" validate:"required,oneof=pending followed completed" example:"active"`
		Id        int    `json:"id" validate:"required,min=1" example:"1"`
		UpdatedBy string `json:"updatedBy" validate:"required,min=1,max=255" example:"admin"`
	}
	UpdateCoolNewJoinerResponse struct {
		Type                string    `json:"type" example:"coolNewJoiner"`
		ID                  int       `json:"id" example:"1"`
		Name                string    `json:"name" example:"John Doe"`
		MaritalStatus       string    `json:"maritalStatus" example:"single"`
		Gender              string    `json:"gender"`
		YearOfBirth         int       `json:"yearOfBirth" example:"1990"`
		PhoneNumber         string    `json:"phoneNumber" example:"+628123456789"`
		Address             string    `json:"address" example:"123 Main St, Jakarta"`
		CommunityOfInterest string    `json:"communityOfInterest" example:"Technology, Sports"`
		CampusCode          string    `json:"campusCode" example:"001"`
		CampusName          string    `json:"campusName" example:"Campus Name"`
		Location            string    `json:"location" example:"Bekasi Timur"`
		UpdatedBy           string    `json:"updatedBy" example:"admin"`
		Status              string    `json:"status" example:"followed"`
		CreatedAt           time.Time `json:"createdAt" example:"2023-01-01T00:00:00Z"`
		UpdatedAt           time.Time `json:"updatedAt" example:"2023-01-01T00:00:00Z"`
		DeletedAt           string    `json:"deletedAt" example:"2023-01-01T00:00:00Z"`
	}
)

type (
	GetAllCoolNewJoinerCursor struct {
		ID        int
		CreatedAt time.Time
	}
	GetAllCoolNewJoinerCursorParam struct {
		Direction           string `query:"direction"`
		Cursor              string `query:"cursor"`
		Limit               int    `query:"limit"`
		Name                string `query:"name"`
		PhoneNumber         string `query:"phoneNumber" validate:"omitempty,phoneFormat0862"`
		CampusCode          string `query:"campusCode" validate:"omitempty,min=3,max=3"`
		MaritalStatus       string `query:"maritalStatus"`
		CommunityOfInterest string `query:"communityOfInterest"`
		Status              string `query:"status"validate:"omitempty,oneof=pending followed completed"`
		Gender              string `query:"gender" validate:"omitempty,oneof=male female"`
		Location            string `query:"location"`
	}
	GetCoolNewJoinerResponse struct {
		Type                string       `json:"type" example:"coolNewJoiner"`
		ID                  int          `json:"id" example:"1"`
		Name                string       `json:"name" example:"John Doe"`
		MaritalStatus       string       `json:"maritalStatus" example:"single"`
		Gender              string       `json:"gender"`
		YearOfBirth         int          `json:"yearOfBirth" example:"1990"`
		PhoneNumber         string       `json:"phoneNumber" example:"+628123456789"`
		Address             string       `json:"address" example:"123 Main St, Jakarta"`
		CommunityOfInterest string       `json:"communityOfInterest" example:"Technology, Sports"`
		CampusCode          string       `json:"campusCode" example:"001"`
		CampusName          string       `json:"campusName" example:"Campus Name"`
		Location            string       `json:"location" example:"Bekasi Timur"`
		UpdatedBy           string       `json:"updatedBy" example:"admin"`
		Status              string       `json:"status" example:"followed"`
		CreatedAt           time.Time    `json:"createdAt" example:"2023-01-01T00:00:00Z"`
		UpdatedAt           time.Time    `json:"updatedAt" example:"2023-01-01T00:00:00Z"`
		DeletedAt           sql.NullTime `json:"-" example:"2023-01-01T00:00:00Z"`
		DeletedAtString     string       `json:"deletedAt" example:"2023-01-01T00:00:00Z"`
	}
)
