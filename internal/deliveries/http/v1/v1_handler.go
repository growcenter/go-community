package v1

import (
	"go-community/internal/config"
	"go-community/internal/deliveries/http/middleware"
	"go-community/internal/usecases"

	"github.com/labstack/echo/v4"
)

func NewV1Handler(g *echo.Group, u *usecases.Usecases, c *config.Configuration) {
	v1 := g.Group("/v1")
	v1.Use(middleware.InternalMiddleware(c))

	// Initialize handlers
	NewCampusHandler(v1, u)
	NewCoolHandler(v1, u)
	NewLocationHandler(v1, u)
	NewUserHandler(v1, u)
	NewEventUserHandler(v1, u, c)
	NewEventGeneralHandler(v1, u, c)
	NewEventSessionHandler(v1, u, c)
	NewEventRegistrationHandler(v1, u, c)
	NewEventInternalHandler(v1, u, c)

	v1noGuard := g.Group("/v1")
	NewEventGoogleHandler(v1noGuard, u)
}
