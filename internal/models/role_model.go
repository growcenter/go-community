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
		Description string `json:"description" example:"View specifically for event"`
	}
)

func CombineRoles(userTypeRoles, additionalRoles []string) []string {
	uniqueRoles := make(map[string]bool)

	// Add roles from userTypeRoles
	for _, role := range userTypeRoles {
		uniqueRoles[role] = true
	}

	// Add roles from additionalRoles
	for _, role := range additionalRoles {
		uniqueRoles[role] = true
	}

	// Convert map keys back to a slice
	var allRoles []string
	for role := range uniqueRoles {
		allRoles = append(allRoles, role)
	}

	return allRoles
}
