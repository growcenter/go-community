package models

import (
	"time"
)

var TYPE_CAMPUS = "campus"

type Campus struct {
	ID        int
	Code      string `gorm:"primaryKey"`
	Region    string
	Name      string
	Location  string
	Address   string
	Status    string
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func (e *Campus) ToResponse() *CampusResponse {
	return &CampusResponse{
		Type:     TYPE_CAMPUS,
		Code:     e.Code,
		Region:   e.Region,
		Name:     e.Name,
		Location: e.Location,
		Address:  e.Address,
		Status:   e.Status,
	}
}

type CreateCampusRequest struct {
	Code     string `json:"code" validate:"required,min=3,max=3" example:"001"`
	Name     string `json:"name" validate:"required,min=1,max=50,nospecial,noStartEndSpaces" example:"Grow Community Jakarta"`
	Region   string `json:"region" validate:"required,min=3,max=10,nospecial,noStartEndSpaces" example:"Bekasi"`
	Location string `json:"location" validate:"required" example:"001"`
	Address  string `json:"address" validate:"required" example:"PT. Amartha Mikro Fintek"`
	Status   string `json:"status" validate:"required,oneof=active inactive" example:"active"`
}

type CampusResponse struct {
	Type      string     `json:"type" example:"campus"`
	ID        int        `json:"-" example:"1"`
	Code      string     `json:"code" example:"BKS"`
	Region    string     `json:"region" example:"Bekasi"`
	Name      string     `json:"name" example:"GROW Center Bekasi"`
	Location  string     `json:"description" example:"The Home - BTC Extension Lt. 3"`
	Address   string     `json:"address" example:"Jalan H. Djoyomartono"`
	Status    string     `json:"status" example:"active"`
	CreatedAt *time.Time `json:"-" example:"2006-01-02 15:04:05"`
	UpdatedAt *time.Time `json:"-" example:"2006-01-02 15:04:05"`
}
