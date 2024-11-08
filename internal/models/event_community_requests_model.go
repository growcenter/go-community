package models

import (
	"time"
)

var TYPE_EVENT_COMMUNITY_REQUEST = "eventCommunityRequest"


type EventCommunityRequest struct {
	ID                int       `gorm:"primaryKey"`
	FullName          string    `gorm:"not null"`
	RequestType       string    `gorm:"not null;check:requestType IN ('Prayer', 'Grateful')"` // Enum-like validation
	Email             string    `gorm:"unique;not null"`
	PhoneNumber       string    `gorm:""`
	RequestInformation string   `gorm:"not null"`
	IsNeedContact     bool      `gorm:"not null;default:false"`
	AccountNumber     string    `gorm:"not null"`
	CreatedAt         time.Time `gorm:"not null;default:now()"`
	UpdatedAt         time.Time `gorm:"not null;default:now()"`
	DeletedAt         *time.Time `gorm:""`
}

// ToResponse converts the EventCommunityRequest model to a response format.
func (ecr *EventCommunityRequest) ToResponse() *EventCommunityRequestResponse {
	return &EventCommunityRequestResponse{
		Type:              TYPE_EVENT_COMMUNITY_REQUEST,
		ID:                ecr.ID,
		FullName:          ecr.FullName,
		RequestType:       ecr.RequestType,
		Email:             ecr.Email,
		PhoneNumber:       ecr.PhoneNumber,
		RequestInformation: ecr.RequestInformation,
		IsNeedContact:     ecr.IsNeedContact,
		AccountNumber:     ecr.AccountNumber,

	}
}


type CreateEventCommunityRequest struct {
	FullName          string `json:"fullName" validate:"required,min=1,max=100" example:"John Doe"`
	RequestType       string `json:"requestType" validate:"required,oneof=prayer grateful" example:"Prayer"`
	Email             string `json:"email" validate:"omitempty,noStartEndSpaces,emailFormat" example:"john.doe@example.com"`
	PhoneNumber       string `json:"phoneNumber" validate:"omitempty,noStartEndSpaces,phoneFormat" example:"021234567890"`
	RequestInformation string `json:"requestInformation" validate:"required" example:"Please pray for my family."`
	IsNeedContact     bool   `json:"isNeedContact" validate:"required" example:"true"`
	AccountNumber     string `json:"accountNumber" validate:"required" example:"123456789"`
}

// EventCommunityRequestResponse is the response format for community request.
type EventCommunityRequestResponse struct {
	Type               string     `json:"type" example:"eventCommunityRequest"`
	ID                 int        `json:"id"`
	FullName           string     `json:"fullName" example:"John Doe"`
	RequestType        string     `json:"requestType" example:"Prayer"`
	Email              string     `json:"email" example:"john.doe@example.com"`
	PhoneNumber        string     `json:"phoneNumber" example:"01234567890"`
	RequestInformation string     `json:"requestInformation" example:"Please pray for my family."`
	IsNeedContact      bool       `json:"isNeedContact" example:"true"`
	AccountNumber      string     `json:"accountNumber" example:"123456789"`

}
