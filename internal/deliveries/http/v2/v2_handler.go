package v2

import (
	"go-community/internal/config"
	"go-community/internal/deliveries/http/middleware"
	"go-community/internal/pkg/authorization"
	"go-community/internal/usecases"

	"github.com/labstack/echo/v4"
)

func NewV2Handler(g *echo.Group, u *usecases.Usecases, c *config.Configuration, a *authorization.Auth) {
	v2 := g.Group("/v2")
	v2.Use(middleware.InternalMiddleware(c))

	// Initialize handlers
	NewEventHandler(v2, u, c, a)
	NewUserHandler(v2, u, c)
	NewTokenHandler(v2, a, c)
	NewRoleHandler(v2, u)
	NewConfigHandler(v2, c)
	NewCoolHandler(v2, u)
}
