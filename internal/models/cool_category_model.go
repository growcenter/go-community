package models

import "time"

var TYPE_COOL_CATEGORY = "coolCategory"

type CoolCategory struct {
	ID        int
	Code      string `gorm:"primaryKey"`
	Name      string
	AgeStart  int
	AgeEnd    int
	Status    string
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func (cc *CoolCategory) ToResponse() *CoolCategoryResponse {
	return &CoolCategoryResponse{
		Type:     TYPE_COOL_CATEGORY,
		Code:     cc.Code,
		Name:     cc.Name,
		AgeStart: cc.AgeStart,
		AgeEnd:   cc.AgeEnd,
		Status:   cc.Status,
	}
}

type CreateCoolCategoryRequest struct {
	Code     string `json:"code" validate:"required,min=3,max=3" example:"001"`
	Name     string `json:"name" validate:"required,min=1,max=50,nospecial,noStartEndSpaces" example:"Professionals"`
	AgeStart int    `json:"ageStart" validate:"required,noStartEndSpaces" example:"21"`
	AgeEnd   int    `json:"ageEnd" validate:"required,noStartEndSpaces" example:"32"`
	Status   string `json:"status" validate:"required,oneof=active inactive" example:"active"`
}

type CoolCategoryResponse struct {
	Type      string     `json:"type" example:"coolCategory"`
	ID        int        `json:"-" example:"1"`
	Code      string     `json:"code" example:"001"`
	Name      string     `json:"name" example:"Profesionals"`
	AgeStart  int        `json:"ageStart" example:"21"`
	AgeEnd    int        `json:"ageEnd" example:"32"`
	Status    string     `json:"status" example:"active"`
	CreatedAt *time.Time `json:"-" example:"2006-01-02 15:04:05"`
	UpdatedAt *time.Time `json:"-" example:"2006-01-02 15:04:05"`
}
