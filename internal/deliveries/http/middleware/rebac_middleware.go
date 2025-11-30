package middleware

import (
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/models"
	"go-community/internal/usecases"

	"github.com/labstack/echo/v4"
)

// Can is a middleware that checks if a user has permission to perform an action on a resource.
// It uses a hybrid RBAC + ReBAC approach for efficiency.
//
// Parameters:
//   - usecase: The AccessRelationshipUsecase for checking permissions.
//   - permission: The permission to check (e.g., "edit").
//   - resourceType: The type of the resource (e.g., "event").
//   - resourceIDParam: The name of the URL parameter that contains the resource ID (e.g., "code").
func Can(usecase usecases.AccessRelationshipUsecase, permission string, resourceType string, resourceIDParam string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// --- RBAC Check (Fast Path) ---
			// First, check for a global "superadmin" role from the JWT claims.
			// This is a fast shortcut that avoids a database call for your most powerful users.
			if roles, ok := c.Get("roles").([]string); ok {
				for _, role := range roles {
					if role == "superadmin" {
						return next(c) // Super admin has all permissions, grant access immediately.
					}
				}
			}

			// --- ReBAC Check (Secure Fallback) ---
			// If the user is not a superadmin, proceed with the fine-grained ReBAC check.
			userID, ok := c.Get("id").(string)
			if !ok || userID == "" {
				return response.Error(c, models.ErrorUnauthorized)
			}

			resourceID := c.Param(resourceIDParam)
			if resourceID == "" {
				return response.Error(c, models.ErrorInvalidRequest)
			}

			// Perform the real-time permission check using the use case.
			allowed, err := usecase.CheckPermission(c.Request().Context(), "user", userID, permission, resourceType, resourceID)
			if err != nil {
				return response.Error(c, models.ErrorForbidden)
			}

			if !allowed {
				return response.Error(c, models.ErrorForbidden)
			}

			return next(c)
		}
	}
}
