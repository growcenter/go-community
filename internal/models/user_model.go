package models

import (
	"database/sql"
	"time"
)

var TYPE_USER = "user"

type User struct {
	ID               int
	CommunityID      string
	Name             string
	PhoneNumber      string
	Email            string
	Password         string
	UserType         string
	Status           string
	Roles            string
	Token            string
	Gender           string
	Address          string
	CampusCode       string
	CoolCategoryCode string
	CoolID           int
	Department       string
	DateOfBirth      *time.Time
	PlaceOfBirth     string
	MaritalStatus    string
	DateOfMarriage   *time.Time
	EmploymentStatus string
	EducationLevel   string
	KKJNumber        string
	JemaatID         string
	IsBaptized       bool
	IsKom100         bool
	Age              int
	CreatedAt        *time.Time
	UpdatedAt        *time.Time
	DeletedAt        sql.NullTime

	Campus       Campus       `gorm:"foreignKey:CampusCode"`
	CoolCategory CoolCategory `gorm:"foreignKey:CoolCategoryCode"`
}

//func (u *User) ToCreateUserCool() *CreateUserCoolResponse {
//	return &CreateUserCoolResponse{
//		Type:             TYPE_USER,
//		CommunityId:      u.CommunityID,
//		Name:             u.Name,
//		Gender:           u.Gender,
//		Age:              u.Age,
//		PhoneNumber:      u.PhoneNumber,
//		Email:            u.Email,
//		CampusCode:       u.CampusCode,
//		CoolCategoryCode: u.CoolCategoryCode,
//		MaritalStatus:    u.MaritalStatus,
//		Status:           u.Status,
//	}
//}

//type (
//	CreateUserCoolRequest struct {
//		Name             string `json:"name" validate:"required,min=1,max=50,nospecial,noStartEndSpaces" example:"Professionals"`
//		Gender           string `json:"gender" validate:"required,oneof=male female" example:"female"`
//		Age              int    `json:"age" validate:"required,noStartEndSpaces" example:"21"`
//		PhoneNumber      string `json:"phoneNumber" validate:"required,noStartEndSpaces,phoneFormat"`
//		Email            string `json:"email" validate:"required,noStartEndSpaces,emailFormat" example:"32"`
//		CampusCode       string `json:"campusCode" validate:"required,min=3,max=3" example:"001"`
//		CoolCategoryCode string `json:"coolCategoryCode" validate:"required,min=3,max=3" example:"001"`
//		MaritalStatus    string `json:"maritalStatus" validate:"required,oneof=single married others" example:"active"`
//	}
//	CreateUserCoolResponse struct {
//		Type             string     `json:"type" example:"coolCategory"`
//		ID               int        `json:"-" example:"1"`
//		CommunityId      string     `json:"communityId"`
//		Name             string     `json:"name" example:"Profesionals"`
//		Gender           string     `json:"gender"`
//		Age              int        `json:"age"`
//		PhoneNumber      string     `json:"phoneNumber"`
//		Email            string     `json:"email"`
//		CampusCode       string     `json:"campusCode"`
//		CoolCategoryCode string     `json:"coolCategoryCode"`
//		MaritalStatus    string     `json:"maritalStatus"`
//		Status           string     `json:"status" example:"active"`
//		CreatedAt        *time.Time `json:"-" example:"2006-01-02 15:04:05"`
//		UpdatedAt        *time.Time `json:"-" example:"2006-01-02 15:04:05"`
//	}
//)

func (u *User) ToCreateUser() *CreateUserResponse {
	return &CreateUserResponse{
		Type:             TYPE_USER,
		Name:             u.Name,
		Email:            u.Email,
		Gender:           u.Gender,
		Age:              u.Age,
		PhoneNumber:      u.PhoneNumber,
		CampusCode:       u.CampusCode,
		CoolCategoryCode: u.CoolCategoryCode,
		MaritalStatus:    u.MaritalStatus,
	}
}

type (
	CreateUserRequest struct {
		Name           string    `json:"name" validate:"required,min=1,max=50,nospecial,noStartEndSpaces" example:"Professionals"`
		PhoneNumber    string    `json:"phoneNumber" validate:"omitempty,noStartEndSpaces,phoneFormat"`
		Email          string    `json:"email" validate:"omitempty,noStartEndSpaces,emailFormat" example:"jeremy@gmail.com"`
		Password       string    `json:"password" validate:"required,min=6,max=50,noStartEndSpaces" example:"Professionals"`
		PlaceOfBirth   string    `json:"placeOfBirth" validate:"required"`
		DateOfBirth    time.Time `json:"dateOfBirth" validate:"required,yyyymmddFormat"`
		Address        string    `json:"address"`
		Gender         string    `json:"gender" validate:"omitempty,oneof=male female"`
		DepartmentCode string    `json:"department_code" validate:"required,noStartEndSpaces" example:"MUSIC"`
		CoolID         int       `json:"coolId" validate:"required" example:"1"`
		KKJNumber      string    `json:"kkjNumber,omitempty"`
		JemaatId       string    `json:"jemaatId,omitempty"`
		IsKom100       bool      `json:"isKom100" validate:"required"`
		IsBaptized     bool      `json:"isBaptized,omitempty" validate:"required"`
		CampusCode     string    `json:"campusCode" validate:"omitempty,min=3,max=3" example:"001"`
		MaritalStatus  string    `json:"maritalStatus" validate:"omitempty,oneof=single married others" example:"active"`
	}
	CreateUserResponse struct {
		Type             string `json:"type" example:"coolCategory"`
		Name             string `json:"name"`
		Email            string `json:"email"`
		Gender           string `json:"gender"`
		Age              int    `json:"age"`
		PhoneNumber      string `json:"phoneNumber"`
		CampusCode       string `json:"campusCode"`
		CoolCategoryCode string `json:"coolCategoryCode"`
		MaritalStatus    string `json:"maritalStatus"`
	}
)

func (u *User) ToCreateVolunteer() *CreateVolunteerResponse {
	return &CreateVolunteerResponse{
		Type:        TYPE_USER,
		CommunityId: u.CommunityID,
		Name:        u.Name,
		Email:       u.Email,
		Gender:      u.Gender,
		PhoneNumber: u.PhoneNumber,
		CampusCode:  u.CampusCode,
	}
}

type (
	CreateVolunteerRequest struct {
		Name           string    `json:"name" validate:"required,min=1,max=50,nospecial,noStartEndSpaces" example:"Professionals"`
		PhoneNumber    string    `json:"phoneNumber" validate:"omitempty,noStartEndSpaces,phoneFormat"`
		Email          string    `json:"email" validate:"omitempty,noStartEndSpaces,emailFormat" example:"jeremy@gmail.com"`
		Password       string    `json:"password" validate:"required,min=6,max=50,noStartEndSpaces" example:"Professionals"`
		CampusCode     string    `json:"campusCode" validate:"omitempty,min=3,max=3" example:"001"`
		PlaceOfBirth   string    `json:"placeOfBirth" validate:"required"`
		DateOfBirth    time.Time `json:"dateOfBirth" validate:"required,yyyymmddFormat"`
		Address        string    `json:"address"`
		Gender         string    `json:"gender" validate:"omitempty,oneof=male female"`
		DepartmentCode string    `json:"department_code" validate:"required,noStartEndSpaces" example:"MUSIC"`
		CoolID         int       `json:"coolId" validate:"required" example:"1"`
		KKJNumber      string    `json:"kkjNumber,omitempty"`
		JemaatId       string    `json:"jemaatId,omitempty"`
		IsKOM100       bool      `json:"isKom100" validate:"required"`
		IsBaptized     bool      `json:"isBaptized,omitempty" validate:"required"`
		MaritalStatus  string    `json:"maritalStatus" validate:"omitempty,oneof=single married others" example:"active"`
	}
	CreateVolunteerResponse struct {
		Type           string     `json:"type" example:"coolCategory"`
		ID             int        `json:"-" example:"1"`
		CommunityId    string     `json:"communityId"`
		Name           string     `json:"name" example:"Profesionals"`
		PhoneNumber    string     `json:"phoneNumber"`
		Email          string     `json:"email"`
		CampusCode     string     `json:"campusCode"`
		PlaceOfBirth   string     `json:"placeOfBirth"`
		DateOfBirth    time.Time  `json:"dateOfBirth"`
		Address        string     `json:"address"`
		Gender         string     `json:"gender"`
		DepartmentCode string     `json:"departmentCode"`
		CoolID         int        `json:"coolId" example:"1"`
		KKJNumber      string     `json:"kkjNumber,omitempty"`
		JemaatId       string     `json:"jemaatId,omitempty"`
		IsKOM100       bool       `json:"isKom100"`
		IsBaptis       bool       `json:"isBaptized"`
		MaritalStatus  string     `json:"maritalStatus"`
		Role           string     `json:"role"`
		Status         string     `json:"status" example:"active"`
		CreatedAt      *time.Time `json:"-" example:"2006-01-02 15:04:05"`
		UpdatedAt      *time.Time `json:"-" example:"2006-01-02 15:04:05"`
	}
)

func (eu *LoginUserResponse) ToLogin() LoginUserResponse {
	return LoginUserResponse{
		Type:           TYPE_USER,
		Name:           eu.Name,
		Email:          eu.Email,
		PhoneNumber:    eu.PhoneNumber,
		CommunityId:    eu.CommunityId,
		CampusCode:     eu.CampusCode,
		DateOfBirth:    eu.DateOfBirth,
		Gender:         eu.Gender,
		DepartmentCode: eu.DepartmentCode,
		CoolID:         eu.CoolID,
		KKJNumber:      eu.KKJNumber,
		JemaatId:       eu.JemaatId,
		IsKOM100:       eu.IsKOM100,
		IsBaptized:     eu.IsBaptized,
		MaritalStatus:  eu.MaritalStatus,
		Role:           eu.Role,
		Token:          eu.Token,
		Status:         eu.Status,
	}
}

type (
	LoginUserRequest struct {
		Identifier string `json:"identifier" validate:"required,emailPhoneFormat,noStartEndSpaces"`
		Password   string `json:"password" validate:"required,noStartEndSpaces"`
	}
	LoginUserResponse struct {
		Type           string    `json:"type" example:"coolCategory"`
		Name           string    `json:"name" example:"Profesionals"`
		PhoneNumber    string    `json:"phoneNumber"`
		Email          string    `json:"email"`
		CommunityId    string    `json:"communityId"`
		CampusCode     string    `json:"campusCode"`
		DateOfBirth    time.Time `json:"dateOfBirth"`
		Gender         string    `json:"gender"`
		DepartmentCode string    `json:"departmentCode"`
		CoolID         int       `json:"coolId" example:"1"`
		KKJNumber      string    `json:"kkjNumber,omitempty"`
		JemaatId       string    `json:"jemaatId,omitempty"`
		IsKOM100       bool      `json:"isKom100"`
		IsBaptized     bool      `json:"isBaptized"`
		MaritalStatus  string    `json:"maritalStatus"`
		Role           string    `json:"role"`
		Token          string    `json:"token"`
		Status         string    `json:"status" example:"active"`
	}
)

func (u *CheckUserEmailResponse) ToCheck() *CheckUserEmailResponse {
	return &CheckUserEmailResponse{
		Type:     TYPE_USER,
		Email:    u.Email,
		IsExist:  false,
		UserType: u.UserType,
	}
}

type CheckUserEmailResponse struct {
	Type     string `json:"type" example:"coolCategory"`
	Email    string `json:"email"`
	IsExist  bool   `json:"isExist"`
	UserType string `json:"userType"`
}

func (u *User) ToGetUserByAccountNumber() *GetUserByAccountNumber {
	return &GetUserByAccountNumber{
		Type:             TYPE_USER,
		CommunityId:      u.CommunityID,
		Name:             u.Name,
		Gender:           u.Gender,
		Age:              u.Age,
		PhoneNumber:      u.PhoneNumber,
		Email:            u.Email,
		CampusCode:       u.CampusCode,
		CampusName:       u.Campus.Name,
		CoolCategoryCode: u.CoolCategoryCode,
		CoolCategoryName: u.CoolCategory.Name,
		MaritalStatus:    u.MaritalStatus,
		Status:           u.Status,
	}
}

type GetUserByAccountNumber struct {
	Type             string `json:"type" example:"coolCategory"`
	ID               int    `json:"-" example:"1"`
	CommunityId      string `json:"CommunityId"`
	Name             string `json:"name" example:"Profesionals"`
	Gender           string `json:"gender"`
	Age              int    `json:"age"`
	PhoneNumber      string `json:"phoneNumber"`
	Email            string `json:"email"`
	CampusCode       string `json:"campusCode"`
	CampusName       string `json:"campusCodeName"`
	CoolCategoryCode string `json:"coolCategoryCode"`
	CoolCategoryName string `json:"coolCategoryName"`
	Roles            string `json:"roles"`
	MaritalStatus    string `json:"maritalStatus"`
	Status           string `json:"status" example:"active"`
}
