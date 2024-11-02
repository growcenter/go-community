package models

import (
	"database/sql"
	"time"
)

var TYPE_EVENT_USER = "eventUser"

type EventUser struct {
	ID            int
	AccountNumber string
	Name          string
	PhoneNumber   string
	Email         string
	Password      string
	Address       string
	Status        string
	State         string
	Role          string
	Gender        string    // Optional Gender field
	MarriageStatus string   // Optional Marriage Status field
	Department    string    // Optional Department field
	KKJ           string      // Optional KKJ field (true or false)
	COOL		  string
	KOM100        bool      // Optional KOM100 field (true or false)
	Baptis        bool      // Optional Baptis field (true or false)
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
	DeletedAt     sql.NullTime
}

func (eu *EventUser) ToResponse() *GetEventUserResponse {
	return &GetEventUserResponse{
		Type:          TYPE_EVENT_USER,
		Name:          eu.Name,
		AccountNumber: eu.AccountNumber,
		Email:         eu.Email,
		PhoneNumber:   eu.PhoneNumber,
		Address:       eu.Address,
		Status:        eu.Status,
		Role:          eu.Role,
		Gender:         eu.Gender,           // Include optional Gender in response
		MarriageStatus: eu.MarriageStatus,   // Include optional MarriageStatus in response
		Department:     eu.Department,
		COOL: 			eu.COOL,       // Include optional Cool in response
		KKJ:            eu.KKJ,              // Include optional KKJ in response
		KOM100:         eu.KOM100,           // Include optional KOM100 in response
		Baptis:         eu.Baptis,           // Include optional Baptis in response
		
	}
}

type GetEventUserResponse struct {
	Type          string `json:"type" example:"coolCategory"`
	Name          string `json:"name"`
	Email         string `json:"email,omitempty"`
	PhoneNumber   string `json:"phoneNumber,omitempty"`
	AccountNumber string `json:"accountNumber"`
	Address       string `json:"address"`
	Role          string `json:"role"`
	Status        string `json:"status" example:"active"`
	Gender         string `json:"gender,omitempty"`           // Optional Gender field
	MarriageStatus string `json:"marriageStatus,omitempty"`   // Optional MarriageStatus field
	Department     string `json:"department,omitempty"`       // Optional Department field
	COOL     	  string `json:"cool,omitempty"`       			  // Optional COOL field
	KKJ            string   `json:"kkj,omitempty"`            // Optional KKJ field
	KOM100         bool   `json:"kom100,omitempty"`           // Optional KOM100 field
	Baptis         bool   `json:"baptis,omitempty"`           // Optional Baptis field
}

func (eu *CreateEventUserResponse) ToCreateEventUser() *CreateEventUserResponse {
	return &CreateEventUserResponse{
		Type:          TYPE_EVENT_USER,
		Name:          eu.Name,
		Email:         eu.Email,
		AccountNumber: eu.AccountNumber,
		Role:          eu.Role,
		Token:         "token",
		Status:        eu.Status,
		Gender:         eu.Gender,           // Optional Gender
		MarriageStatus: eu.MarriageStatus,   // Optional MarriageStatus
		Department:     eu.Department,       // Optional Department
		COOL: 			eu.COOL,       // Include optional Cool in response
		KKJ:            eu.KKJ,              // Optional KKJ
		KOM100:         eu.KOM100,           // Optional KOM100
		Baptis:         eu.Baptis,           // Optional Baptis
	}
}

type (
	CreateEventUserRequest struct {
		Name  string `json:"name" validate:"required,min=1,max=50,nospecial,noStartEndSpaces" example:"Professionals"`
		Email string `json:"email" validate:"required,noStartEndSpaces,emailFormat" example:"jeremy@gmail.com"`
	}

	CreateEventUserResponse struct {
		Type          string `json:"type" example:"coolCategory"`
		Name          string `json:"name" example:"Profesionals"`
		Email         string `json:"email"`
		AccountNumber string `json:"accountNumber"`
		Role          string `json:"role"`
		Token         string `json:"token"`
		Status        string `json:"status" example:"active"`
		Gender         string `json:"gender,omitempty"`            // Optional Gender field
		MarriageStatus string `json:"marriageStatus,omitempty"`    // Optional MarriageStatus field
		Department     string `json:"department,omitempty"`        // Optional Department field
		COOL     	  string `json:"cool,omitempty"`       			  // Optional COOL field
		KKJ            string   `json:"kkj,omitempty"`               // Optional KKJ field
		KOM100         bool   `json:"kom100,omitempty"`            // Optional KOM100 field
		Baptis         bool   `json:"baptis,omitempty"`            // Optional Baptis field
	}
)

func (eu *CreateEventUserManualResponse) ToCreateEventUserManual() CreateEventUserManualResponse {
	return CreateEventUserManualResponse{
		Type:          TYPE_EVENT_USER,
		Name:          eu.Name,
		Email:         eu.Email,
		PhoneNumber:   eu.PhoneNumber,
		AccountNumber: eu.AccountNumber,
		Role:          eu.Role,
		Token:         eu.Token,
		Status:        eu.Status,
		Gender:         eu.Gender,           // Optional Gender
		MarriageStatus: eu.MarriageStatus,   // Optional MarriageStatus
		COOL: 			eu.COOL,       // Include optional Cool in response
		Department:     eu.Department,       // Optional Department
		KKJ:            eu.KKJ,              // Optional KKJ
		KOM100:         eu.KOM100,           // Optional KOM100
		Baptis:         eu.Baptis,           // Optional Baptis
	}
}

type (
	CreateEventUserManualRequest struct {
		Name        string `json:"name" validate:"required,min=1,max=50,nospecial,noStartEndSpaces" example:"Professionals"`
		PhoneNumber string `json:"phoneNumber" validate:"omitempty,noStartEndSpaces,phoneFormat"`
		Email       string `json:"email" validate:"omitempty,noStartEndSpaces,emailFormat" example:"jeremy@gmail.com"`
		Password    string `json:"password" validate:"required,min=6,max=50,noStartEndSpaces" example:"Professionals"`
		Gender         string `json:"gender,omitempty"`            // Optional Gender field
		MarriageStatus string `json:"marriageStatus,omitempty"`    // Optional MarriageStatus field
		Department     string `json:"department,omitempty"`        // Optional Department field
		COOL     	  string `json:"cool,omitempty"`       			  // Optional COOL field
		KKJ            string   `json:"kkj,omitempty"`               // Optional KKJ field
		KOM100         bool   `json:"kom100,omitempty"`            // Optional KOM100 field
		Baptis         bool   `json:"baptis,omitempty"`            // Optional Baptis field
		
	}
	CreateEventUserManualResponse struct {
		Type          string `json:"type" example:"coolCategory"`
		Name          string `json:"name" example:"Profesionals"`
		Email         string `json:"email,omitempty"`
		PhoneNumber   string `json:"phoneNumber,omitempty"`
		AccountNumber string `json:"accountNumber"`
		Role          string `json:"role"`
		Token         string `json:"token"`
		Status        string `json:"status" example:"active"`
		Gender         string `json:"gender,omitempty"`            // Optional Gender field
		MarriageStatus string `json:"marriageStatus,omitempty"`    // Optional MarriageStatus field
		Department     string `json:"department,omitempty"`        // Optional Department field
		COOL     	  string `json:"cool,omitempty"`       			  // Optional COOL field
		KKJ            string   `json:"kkj,omitempty"`               // Optional KKJ field
		KOM100         bool   `json:"kom100,omitempty"`            // Optional KOM100 field
		Baptis         bool   `json:"baptis,omitempty"`            // Optional Baptis field
	}
)

func (eu *LoginEventUserManualResponse) ToLoginEventUserManual() LoginEventUserManualResponse {
	return LoginEventUserManualResponse{
		Type:          TYPE_EVENT_USER,
		Name:          eu.Name,
		Email:         eu.Email,
		PhoneNumber:   eu.PhoneNumber,
		AccountNumber: eu.AccountNumber,
		Address:       eu.Address,
		Role:          eu.Role,
		Token:         eu.Token,
		Status:        eu.Status,
		KKJ: 		   eu.KKJ,
	}
}

type (
	LoginEventUserManualRequest struct {
		Identifier string `json:"identifier" validate:"required,noStartEndSpaces"`
		Password   string `json:"password" validate:"required,noStartEndSpaces"`
	}
	LoginEventUserManualResponse struct {
		Type          string `json:"type" example:"coolCategory"`
		Name          string `json:"name"`
		Email         string `json:"email,omitempty"`
		PhoneNumber   string `json:"phoneNumber,omitempty"`
		AccountNumber string `json:"accountNumber"`
		Address       string `json:"address"`
		Token         string `json:"token"`
		Role          string `json:"role"`
		Status        string `json:"status" example:"active"`
		KKJ           string `json:"kkj,omitempty"`              
	}
)

func (eu *UpdateAccountRoleResponse) ToUpdateAccountRole() UpdateAccountRoleResponse {
	return UpdateAccountRoleResponse{
		Type:           TYPE_EVENT_USER,
		AccountNumbers: eu.AccountNumbers,
		Role:           eu.Role,
	}
}

type (
	UpdateAccountRoleRequest struct {
		AccountNumbers []string `json:"accountNumbers" validate:"dive,required,numeric,noStartEndSpaces"`
		Role           string   `json:"role" validate:"required,oneof=user admin usher" example:"female"`
	}
	UpdateAccountRoleResponse struct {
		Type           string   `json:"type" example:"coolCategory"`
		AccountNumbers []string `json:"accountNumbers"`
		Role           string   `json:"role"`
	}
)

func (u *LogoutEventUserResponse) ToLogout() *LogoutEventUserResponse {
	return &LogoutEventUserResponse{
		Type:          TYPE_EVENT_USER,
		AccountNumber: u.AccountNumber,
		IsLoggedOut:   false,
		Token:         u.Token,
	}
}

type LogoutEventUserResponse struct {
	Type          string `json:"type" example:"coolCategory"`
	AccountNumber string `json:"accountNumber"`
	IsLoggedOut   bool   `json:"isLoggedOut"`
	Token         string `json:"token"`
}

func (eu *EventUser) ToUpdatePassword() *UpdatePasswordResponse {
	return &UpdatePasswordResponse{
		Type:          TYPE_EVENT_USER,
		Name:          eu.Name,
		AccountNumber: eu.AccountNumber,
		Email:         eu.Email,
		PhoneNumber:   eu.PhoneNumber,
		Status:        eu.Status,
		Role:          eu.Role,
	}
}

type (
	UpdatePasswordRequest struct {
		Identifier string `json:"identifier" validate:"required,noStartEndSpaces,emailPhoneFormat"`
		Password   string `json:"password" validate:"required,min=6,max=50,noStartEndSpaces" example:"Professionals"`
	}
	UpdatePasswordResponse struct {
		Type          string `json:"type" example:"coolCategory"`
		Name          string `json:"name" example:"Profesionals"`
		Email         string `json:"email,omitempty"`
		PhoneNumber   string `json:"phoneNumber,omitempty"`
		AccountNumber string `json:"accountNumber"`
		Role          string `json:"role"`
		Status        string `json:"status" example:"active"`
	}
)
