package models

import "time"

var TYPE_COOL_CATEGORY = "coolCategory"

type CoolCategory struct {
	ID          int
	Code		string
	Name        string
	AgeStart   	int
	AgeEnd 		int
	Status		string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

func (cd *CoolCategory) ToResponse() *CoolCategoryResponse {
	return &CoolCategoryResponse{
		Type:        TYPE_COOL_CATEGORY,
		Code: 		 cd.Code,
		Name:        cd.Name,
		AgeStart: 	 cd.AgeStart,
		AgeEnd:		 cd.AgeEnd,
		Status:      cd.Status,
	}
}

type CreateCoolCategoryRequest struct {
	Code        string	`json:"code" validate:"required,numerical,min=3,max=3" example:"001"`
	Name        string 	`json:"name" validate:"required,min=1,max=50,nospecial,noStartEndSpaces" example:"Professionals"`
	AgeStart	int		`json:"ageStart" validate:"required,numerical,nospecial,noStartEndSpaces" example:"21"`
	AgeEnd		int		`json:"ageEnd" validate:"required,numerical,nospecial,noStartEndSpaces" example:"32"`
	Status      string	`json:"status" validate:"required,oneof=active inactive" example:"active"`
}

type CoolCategoryResponse struct {
	Type        string     	`json:"type" example:"coolCategory"`
	ID          int        	`json:"-" example:"1"`
	Code		string		`json:"code" example:"001"`
	Name        string     	`json:"name" example:"Profesionals"`
	AgeStart 	int     	`json:"ageStart" example:"21"`
	AgeEnd		int			`json:"ageEnd" example:"32"`
	Status      string     	`json:"status" example:"active"`
	CreatedAt   *time.Time 	`json:"-" example:"2006-01-02 15:04:05"`
	UpdatedAt   *time.Time 	`json:"-" example:"2006-01-02 15:04:05"`
}