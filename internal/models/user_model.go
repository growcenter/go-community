package models

import (
	"database/sql"
	"github.com/lib/pq"
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
	UserTypes        pq.StringArray `gorm:"type:text[]"`
	Status           string
	Roles            pq.StringArray `gorm:"type:text[]"`
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

func (u *CreateUserResponse) ToCreateUser() *CreateUserResponse {
	return &CreateUserResponse{
		Type:           TYPE_USER,
		CommunityId:    u.CommunityId,
		Name:           u.Name,
		PhoneNumber:    u.PhoneNumber,
		Email:          u.Email,
		UserTypes:      u.UserTypes,
		CampusCode:     u.CampusCode,
		PlaceOfBirth:   u.PlaceOfBirth,
		DateOfBirth:    u.DateOfBirth,
		Address:        u.Address,
		Gender:         u.Gender,
		DepartmentCode: u.DepartmentCode,
		CoolID:         u.CoolID,
		KKJNumber:      u.KKJNumber,
		JemaatId:       u.JemaatId,
		IsKOM100:       u.IsKOM100,
		IsBaptized:     u.IsBaptized,
		MaritalStatus:  u.MaritalStatus,
		Status:         u.Status,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
	}
}

type (
	CreateUserRequest struct {
		Name           string   `json:"name" validate:"required,min=1,max=50,nospecial" example:"Professionals"`
		Email          string   `json:"email" validate:"omitempty,emailFormat,emailOrPhoneField" example:"jeremy@gmail.com"`
		PhoneNumber    string   `json:"phoneNumber" validate:"omitempty,phoneFormat"`
		Password       string   `json:"password" validate:"required,min=6,max=50,noStartEndSpaces" example:"Professionals"`
		UserTypes      []string `json:"userTypes" validate:"required" example:"volunteer"`
		CampusCode     string   `json:"campusCode" validate:"omitempty,min=3,max=3" example:"001"`
		PlaceOfBirth   string   `json:"placeOfBirth" validate:"required"`
		DateOfBirth    string   `json:"dateOfBirth" validate:"required,yyymmddFormat"`
		Address        string   `json:"address"`
		Gender         string   `json:"gender" validate:"omitempty,oneof=male female"`
		DepartmentCode string   `json:"department_code" example:"MUSIC"`
		CoolID         int      `json:"coolId" example:"1"`
		KKJNumber      string   `json:"kkjNumber,omitempty"`
		JemaatId       string   `json:"jemaatId,omitempty"`
		IsKOM100       bool     `json:"isKom100"`
		IsBaptized     bool     `json:"isBaptized"`
		MaritalStatus  string   `json:"maritalStatus" validate:"omitempty,oneof=single married others" example:"active"`
	}
	CreateUserResponse struct {
		Type           string     `json:"type" example:"user"`
		ID             int        `json:"-" example:"1"`
		CommunityId    string     `json:"communityId"`
		Name           string     `json:"name" example:"Profesionals"`
		PhoneNumber    string     `json:"phoneNumber"`
		Email          string     `json:"email"`
		UserTypes      []string   `json:"userTypes"`
		CampusCode     string     `json:"campusCode"`
		PlaceOfBirth   string     `json:"placeOfBirth"`
		DateOfBirth    *time.Time `json:"dateOfBirth"`
		Address        string     `json:"address"`
		Gender         string     `json:"gender"`
		DepartmentCode string     `json:"departmentCode"`
		CoolID         int        `json:"coolId" example:"1"`
		KKJNumber      string     `json:"kkjNumber,omitempty"`
		JemaatId       string     `json:"jemaatId,omitempty"`
		IsKOM100       bool       `json:"isKom100"`
		IsBaptized     bool       `json:"isBaptized"`
		MaritalStatus  string     `json:"maritalStatus"`
		Status         string     `json:"status" example:"active"`
		CreatedAt      *time.Time `json:"-" example:"2006-01-02 15:04:05"`
		UpdatedAt      *time.Time `json:"-" example:"2006-01-02 15:04:05"`
	}
)

func (u *CreateVolunteerResponse) ToCreateVolunteer() *CreateVolunteerResponse {
	return &CreateVolunteerResponse{
		Type:           TYPE_USER,
		CommunityId:    u.CommunityId,
		Name:           u.Name,
		PhoneNumber:    u.PhoneNumber,
		Email:          u.Email,
		UserTypes:      u.UserTypes,
		CampusCode:     u.CampusCode,
		PlaceOfBirth:   u.PlaceOfBirth,
		DateOfBirth:    u.DateOfBirth,
		Address:        u.Address,
		Gender:         u.Gender,
		DepartmentCode: u.DepartmentCode,
		CoolID:         u.CoolID,
		KKJNumber:      u.KKJNumber,
		JemaatId:       u.JemaatId,
		IsKOM100:       u.IsKOM100,
		IsBaptis:       u.IsBaptis,
		MaritalStatus:  u.MaritalStatus,
		Status:         u.Status,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
	}
}

type (
	CreateVolunteerRequest struct {
		Name           string   `json:"name" validate:"required,min=1,max=50,nospecial" example:"Professionals"`
		PhoneNumber    string   `json:"phoneNumber" validate:"omitempty,phoneFormat"`
		Email          string   `json:"email" validate:"omitempty,emailFormat" example:"jeremy@gmail.com"`
		Password       string   `json:"password" validate:"required,min=6,max=50,noStartEndSpaces" example:"Professionals"`
		UserTypes      []string `json:"userTypes" validate:"required" example:"volunteer"`
		CampusCode     string   `json:"campusCode" validate:"omitempty,min=3,max=3" example:"001"`
		PlaceOfBirth   string   `json:"placeOfBirth" validate:"required"`
		DateOfBirth    string   `json:"dateOfBirth" validate:"required,yyymmddFormat"`
		Address        string   `json:"address"`
		Gender         string   `json:"gender" validate:"omitempty,oneof=male female"`
		DepartmentCode string   `json:"department_code" validate:"required,noStartEndSpaces" example:"MUSIC"`
		CoolID         int      `json:"coolId" validate:"required" example:"1"`
		KKJNumber      string   `json:"kkjNumber,omitempty"`
		JemaatId       string   `json:"jemaatId,omitempty"`
		IsKOM100       bool     `json:"isKom100" validate:"required"`
		IsBaptized     bool     `json:"isBaptized,omitempty" validate:"required"`
		MaritalStatus  string   `json:"maritalStatus" validate:"omitempty,oneof=single married others" example:"active"`
	}
	CreateVolunteerResponse struct {
		Type           string     `json:"type" example:"coolCategory"`
		ID             int        `json:"-" example:"1"`
		CommunityId    string     `json:"communityId"`
		Name           string     `json:"name" example:"Profesionals"`
		PhoneNumber    string     `json:"phoneNumber"`
		Email          string     `json:"email"`
		UserTypes      []string   `json:"userTypes"`
		CampusCode     string     `json:"campusCode"`
		PlaceOfBirth   string     `json:"placeOfBirth"`
		DateOfBirth    *time.Time `json:"dateOfBirth"`
		Address        string     `json:"address"`
		Gender         string     `json:"gender"`
		DepartmentCode string     `json:"departmentCode"`
		CoolID         int        `json:"coolId" example:"1"`
		KKJNumber      string     `json:"kkjNumber,omitempty"`
		JemaatId       string     `json:"jemaatId,omitempty"`
		IsKOM100       bool       `json:"isKom100"`
		IsBaptis       bool       `json:"isBaptized"`
		MaritalStatus  string     `json:"maritalStatus"`
		Status         string     `json:"status" example:"active"`
		CreatedAt      *time.Time `json:"-" example:"2006-01-02 15:04:05"`
		UpdatedAt      *time.Time `json:"-" example:"2006-01-02 15:04:05"`
	}
)

func (u *LoginUserResponse) ToLogin() LoginUserResponse {
	return LoginUserResponse{
		Type:           TYPE_USER,
		Name:           u.Name,
		Email:          u.Email,
		PhoneNumber:    u.PhoneNumber,
		CommunityId:    u.CommunityId,
		UserTypes:      u.UserTypes,
		CampusCode:     u.CampusCode,
		DateOfBirth:    u.DateOfBirth,
		Gender:         u.Gender,
		DepartmentCode: u.DepartmentCode,
		CoolID:         u.CoolID,
		KKJNumber:      u.KKJNumber,
		JemaatId:       u.JemaatId,
		IsKOM100:       u.IsKOM100,
		IsBaptized:     u.IsBaptized,
		MaritalStatus:  u.MaritalStatus,
		Roles:          u.Roles,
		Token:          u.Token,
		Status:         u.Status,
	}
}

type (
	LoginUserRequest struct {
		Identifier string `json:"identifier" validate:"required,emailPhoneFormat"`
		Password   string `json:"password" validate:"required,noStartEndSpaces"`
	}
	LoginUserResponse struct {
		Type           string        `json:"type" example:"coolCategory"`
		Name           string        `json:"name" example:"Profesionals"`
		PhoneNumber    string        `json:"phoneNumber"`
		Email          string        `json:"email"`
		CommunityId    string        `json:"communityId"`
		UserTypes      []string      `json:"userTypes"`
		CampusCode     string        `json:"campusCode"`
		PlaceOfBirth   string        `json:"placeOfBirth"`
		DateOfBirth    *time.Time    `json:"dateOfBirth"`
		Address        string        `json:"address"`
		Gender         string        `json:"gender"`
		DepartmentCode string        `json:"departmentCode"`
		CoolID         int           `json:"coolId" example:"1"`
		KKJNumber      string        `json:"kkjNumber,omitempty"`
		JemaatId       string        `json:"jemaatId,omitempty"`
		IsKOM100       bool          `json:"isKom100"`
		IsBaptized     bool          `json:"isBaptized"`
		MaritalStatus  string        `json:"maritalStatus"`
		Roles          []string      `json:"roles"`
		Token          []interface{} `json:"tokens"`
		Status         string        `json:"status" example:"active"`
	}
)

func (u *CheckUserExistResponse) ToCheck() *CheckUserExistResponse {
	return &CheckUserExistResponse{
		Type:       TYPE_USER,
		Identifier: u.Identifier,
		User:       u.User,
	}
}

type CheckUserExistResponse struct {
	Type       string `json:"type" example:"user"`
	Identifier string `json:"identifier"`
	User       bool   `json:"user"`
}

func (u *GetOneByCommunityIdResponse) ToGetOneByCommunityId() *GetOneByCommunityIdResponse {
	return &GetOneByCommunityIdResponse{
		Type:           TYPE_USER,
		Name:           u.Name,
		Email:          u.Email,
		PhoneNumber:    u.PhoneNumber,
		CommunityId:    u.CommunityId,
		UserTypes:      u.UserTypes,
		CampusCode:     u.CampusCode,
		CampusName:     u.CampusName,
		DateOfBirth:    u.DateOfBirth,
		Gender:         u.Gender,
		DepartmentCode: u.DepartmentCode,
		DepartmentName: u.DepartmentName,
		CoolID:         u.CoolID,
		CoolName:       u.CoolName,
		KKJNumber:      u.KKJNumber,
		JemaatId:       u.JemaatId,
		IsKOM100:       u.IsKOM100,
		IsBaptized:     u.IsBaptized,
		MaritalStatus:  u.MaritalStatus,
		Roles:          u.Roles,
		Status:         u.Status,
	}
}

type (
	GetOneByCommunityIdParameter struct {
		CommunityId string `json:"communityId" validate:"required,communityId"`
	}
	GetOneByCommunityIdResponse struct {
		Type           string         `json:"type" example:"coolCategory"`
		Name           string         `json:"name" example:"Profesionals"`
		PhoneNumber    string         `json:"phoneNumber"`
		Email          string         `json:"email"`
		CommunityId    string         `json:"communityId"`
		UserTypes      []string       `json:"userTypes"`
		CampusCode     string         `json:"campusCode"`
		CampusName     string         `json:"campusName"`
		PlaceOfBirth   string         `json:"placeOfBirth"`
		DateOfBirth    *time.Time     `json:"dateOfBirth"`
		Address        string         `json:"address"`
		Gender         string         `json:"gender"`
		DepartmentCode string         `json:"departmentCode"`
		DepartmentName string         `json:"departmentName"`
		CoolID         int            `json:"coolId" example:"1"`
		CoolName       string         `json:"coolName"`
		KKJNumber      string         `json:"kkjNumber,omitempty"`
		JemaatId       string         `json:"jemaatId,omitempty"`
		IsKOM100       bool           `json:"isKom100"`
		IsBaptized     bool           `json:"isBaptized"`
		MaritalStatus  string         `json:"maritalStatus"`
		Roles          []RoleResponse `json:"roles"`
		Status         string         `json:"status" example:"active"`
	}
)

func (u *User) ToUpdatePassword() UpdateUserPasswordResponse {
	return UpdateUserPasswordResponse{
		Type:        TYPE_USER,
		Name:        u.Name,
		Email:       u.Email,
		PhoneNumber: u.PhoneNumber,
		CommunityId: u.CommunityID,
		UserTypes:   u.UserTypes,
		CampusCode:  u.CampusCode,
		Roles:       u.Roles,
		Status:      u.Status,
	}
}

type (
	UpdateUserPasswordParam struct {
		Identifier string `json:"identifier" validate:"required,emailPhoneFormat"`
	}
	UpdateUserPasswordRequest struct {
		Password        string `json:"password" validate:"required,min=6,max=50,noStartEndSpaces"`
		ConfirmPassword string `json:"confirmPassword" validate:"required,min=6,max=50,noStartEndSpaces,eqfield=Password"`
	}
	UpdateUserPasswordResponse struct {
		Type        string   `json:"type" example:"coolCategory"`
		Name        string   `json:"name" example:"Profesionals"`
		Email       string   `json:"email,omitempty"`
		PhoneNumber string   `json:"phoneNumber,omitempty"`
		CommunityId string   `json:"communityId"`
		UserTypes   []string `json:"userTypes"`
		CampusCode  string   `json:"campusCode"`
		Roles       []string `json:"roles"`
		Status      string   `json:"status" example:"active"`
	}
)

type (
	GetNameOnUserDBOutput struct {
		Name        string
		CommunityId string
	}
)

type (
	GetAllUserDBOutput struct {
		ID            int
		CommunityID   string
		Name          string
		PhoneNumber   string
		Email         string
		UserTypes     pq.StringArray `gorm:"type:text[]"`
		Roles         pq.StringArray `gorm:"type:text[]"`
		Status        string
		Gender        string
		Address       string
		CampusCode    string
		CoolID        int
		CoolName      string
		Department    string
		DateOfBirth   *time.Time
		PlaceOfBirth  string
		MaritalStatus string
		KKJNumber     string
		JemaatID      string
		IsBaptized    bool
		IsKom100      bool
		CreatedAt     *time.Time
		UpdatedAt     *time.Time
		DeletedAt     sql.NullTime
	}
	GetAllUserCursorParam struct {
		Direction  string `query:"direction"`
		Cursor     string `query:"cursor"`
		Limit      int    `query:"limit"`
		Search     string `query:"search"`
		SearchBy   string `query:"searchBy" validate:"omitempty,oneof=communityId name phoneNumber email"`
		CampusCode string `query:"campusCode"`
		CoolId     int    `query:"coolId"`
		Department string `query:"department"`
	}
	GetAllUserCursorResponse struct {
		Type           string     `json:"type"`
		Name           string     `json:"name"`
		CommunityID    string     `json:"communityId"`
		PhoneNumber    string     `json:"phoneNumber"`
		Email          string     `json:"email"`
		UserTypes      []string   `json:"userTypes"`
		Roles          []string   `json:"roles"`
		Status         string     `json:"status"`
		Gender         string     `json:"gender"`
		Address        string     `json:"address"`
		CampusCode     string     `json:"campusCode"`
		CampusName     string     `json:"campusName"`
		CoolID         int        `json:"coolId"`
		CoolName       string     `json:"coolName"`
		DepartmentCode string     `json:"departmentCode"`
		DepartmentName string     `json:"departmentName"`
		DateOfBirth    *time.Time `json:"dateOfBirth"`
		PlaceOfBirth   string     `json:"placeOfBirth"`
		MaritalStatus  string     `json:"maritalStatus"`
		KKJNumber      string     `json:"kkjNumber"`
		JemaatID       string     `json:"jemaatId"`
		IsBaptized     bool       `json:"isBaptized"`
		IsKom100       bool       `json:"isKom100"`
		CreatedAt      time.Time  `json:"createdAt"`
		UpdatedAt      time.Time  `json:"updatedAt"`
		DeletedAt      string     `json:"deletedAt"`
	}
)
