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
	}
}

type (
	CreateEventUserManualRequest struct {
		Name        string `json:"name" validate:"required,min=1,max=50,nospecial,noStartEndSpaces" example:"Professionals"`
		PhoneNumber string `json:"phoneNumber" validate:"omitempty,noStartEndSpaces,phoneFormat"`
		Email       string `json:"email" validate:"omitempty,noStartEndSpaces,emailFormat" example:"jeremy@gmail.com"`
		Password    string `json:"password" validate:"required,min=6,max=50,noStartEndSpaces" example:"Professionals"`
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
		Role           string   `json:"role" validate:"required,oneof=user admin" example:"female"`
	}
	UpdateAccountRoleResponse struct {
		Type           string   `json:"type" example:"coolCategory"`
		AccountNumbers []string `json:"accountNumbers"`
		Role           string   `json:"role" validate:"required,oneof=user admin"`
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
