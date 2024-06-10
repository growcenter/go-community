package models

import "time"

var TYPE_LOCATION = "location"

type Location struct {
	ID         int
	Code       string
	CampusCode string
	Name       string
	Region     string
	Status     string
	CreatedAt  *time.Time
	UpdatedAt  *time.Time
}

func (l *Location) ToResponse() *LocationResponse {
	return &LocationResponse{
		Type:       TYPE_LOCATION,
		Name:       l.Name,
		Code:       l.Code,
		CampusCode: l.CampusCode,
		Region:     l.Region,
		Status:     l.Status,
	}
}

type CreateLocationRequest struct {
	Name       string `json:"name" validate:"required,min=1,max=50,nospecial,noStartEndSpaces" example:"Professionals"`
	Code       string `json:"code" validate:"required,min=5,max=5" example:"PSGRH"`
	CampusCode string `json:"campusCode" validate:"required,min=3,max=3" example:"001"`
	Status     string `json:"status" validate:"required,oneof=active inactive" example:"active"`
}

type LocationResponse struct {
	Type       string     `json:"type" example:"location"`
	ID         int        `json:"-" example:"1"`
	Name       string     `json:"name" example:"Profesionals"`
	Code       string     `json:"code" example:"PSGRH"`
	CampusCode string     `json:"campusCode" example:"BKS"`
	Region     string     `json:"region" example:"Bekasi"`
	Status     string     `json:"status" example:"active"`
	CreatedAt  *time.Time `json:"-" example:"2006-01-02 15:04:05"`
	UpdatedAt  *time.Time `json:"-" example:"2006-01-02 15:04:05"`
}
