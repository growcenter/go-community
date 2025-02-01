package models

import "time"

var TYPE_ROLE = "role"

type Role struct {
	ID          int
	Role        string
	Description string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

func (r *Role) ToResponse() *RoleResponse {
	return &RoleResponse{
		Type:        TYPE_ROLE,
		Role:        r.Role,
		Description: r.Description,
	}
}

type (
	CreateRoleRequest struct {
		Role        string `json:"role" validate:"required" example:"event-view-volunteer"`
		Description string `json:"description" example:"View specifically for event"`
	}
	RoleResponse struct {
		Type        string `json:"type" example:"role"`
		Role        string `json:"role" example:"event-view-volunteer"`
		Description string `json:"description,omitempty" example:"View specifically for event"`
	}
)
