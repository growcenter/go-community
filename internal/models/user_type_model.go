package models

import (
	"github.com/lib/pq"
	"time"
)

var TYPE_USER_TYPE = "userType"

type UserType struct {
	ID          int
	Type        string
	Name        string
	Description string
	Roles       pq.StringArray `gorm:"type:text[]"`
	Category    string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

func (ut *UserType) ToResponse() *UserTypeResponse {
	return &UserTypeResponse{
		Type:        TYPE_USER_TYPE,
		UserType:    ut.Type,
		Name:        ut.Name,
		Roles:       ut.Roles,
		Description: ut.Description,
		Category:    ut.Category,
	}
}

type (
	CreateUserTypeRequest struct {
		UserType    string   `json:"userType" validate:"required" example:"volunteer"`
		Name        string   `json:"name" validate:"required" example:"Volunteer"`
		Roles       []string `json:"roles" validate:"required" example:"event-view-volunteer, event-edit-volunteer"`
		Description string   `json:"description" example:"General Volunteer"`
		Category    string   `json:"category" validate:"required,oneof=general internal cool"`
	}
	UserTypeResponse struct {
		Type        string   `json:"type" example:"userType"`
		UserType    string   `json:"userType" example:"volunteer"`
		Name        string   `json:"name" example:"Volunteer"`
		Description string   `json:"description" example:"Volunteer"`
		Roles       []string `json:"roles" example:"event-view-event-viewer"`
		Category    string   `json:"category" example:"general"`
	}
)

type UserTypeSummaryResponse struct {
	Type     string `json:"type" example:"userType"`
	UserType string `json:"userType" example:"volunteer"`
	Name     string `json:"name,omitempty" example:"Volunteer"`
}
