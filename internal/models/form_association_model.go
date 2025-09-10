package models

import "github.com/google/uuid"

type FormAssociation struct {
	ID         int
	FormCode   uuid.UUID `json:"form_code"`
	EntityCode string    `json:"entity_code"`
	EntityType string    `json:"entity_type"`
}

type (
	CreateFormAssociationRequest struct {
		FormCode   uuid.UUID `json:"formCode" validate:"required"`
		EntityCode string    `json:"entityCode" validate:"required"`
		EntityType string    `json:"entityType" validate:"required"`
	}

	CreateFormAssociationResponse struct {
		FormCode   uuid.UUID `json:"formCode"`
		EntityCode string    `json:"entityCode"`
		EntityType string    `json:"entityType"`
	}
)
