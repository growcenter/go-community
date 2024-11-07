package models

import (
	"time"
)

var TYPE_EVENT_COMMUNITY_REQUEST = "eventCommunityRequest"


type EventCommunityRequest struct {
	ID                int       `gorm:"primaryKey"`
	FullName          string    `gorm:"not null"`
	RequestType       string    `gorm:"not null;check:request_type IN ('Prayer', 'Grateful')"` // Enum-like validation
	Email             string    `gorm:"unique;not null"`
	PhoneNumber       string    `gorm:""`
	RequestInformation string   `gorm:"not null"`
	WantContacted     bool      `gorm:"not null;default:false"`
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
		WantContacted:     ecr.WantContacted,
		AccountNumber:     ecr.AccountNumber,
		CreatedAt:         ecr.CreatedAt,
		UpdatedAt:         ecr.UpdatedAt,
	}
}


type CreateEventCommunityRequest struct {
	FullName          string `json:"full_name" validate:"required,min=1,max=100" example:"John Doe"`
	RequestType       string `json:"request_type" validate:"required,oneof=Prayer Grateful" example:"Prayer"`
	Email             string `json:"email" validate:"required,email" example:"john.doe@example.com"`
	PhoneNumber       string `json:"phone_number" validate:"omitempty,e164" example:"021234567890"`
	RequestInformation string `json:"request_information" validate:"required" example:"Please pray for my family."`
	WantContacted     bool   `json:"want_contacted" validate:"required" example:"true"`
	AccountNumber     string `json:"account_number" validate:"required" example:"123456789"`
}

// EventCommunityRequestResponse is the response format for community request.
type EventCommunityRequestResponse struct {
	Type               string     `json:"type" example:"eventCommunityRequest"`
	ID                 int        `json:"id"`
	FullName           string     `json:"full_name" example:"John Doe"`
	RequestType        string     `json:"request_type" example:"Prayer"`
	Email              string     `json:"email" example:"john.doe@example.com"`
	PhoneNumber        string     `json:"phone_number" example:"01234567890"`
	RequestInformation string     `json:"request_information" example:"Please pray for my family."`
	WantContacted      bool       `json:"want_contacted" example:"true"`
	AccountNumber      string     `json:"account_number" example:"123456789"`
	CreatedAt          time.Time  `json:"created_at" example:"2024-11-07T10:30:00Z"`
	UpdatedAt          time.Time  `json:"updated_at" example:"2024-11-07T10:30:00Z"`
}
