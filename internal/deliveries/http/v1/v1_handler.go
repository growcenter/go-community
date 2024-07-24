package v1

import (
	"go-community/internal/config"
	"go-community/internal/usecases"

	"github.com/labstack/echo/v4"
)

func NewV1Handler(g *echo.Group, u *usecases.Usecases, c *config.Configuration) {
	v1 := g.Group("/v1")

	// Initialize handlers
	NewCampusHandler(v1, u)
	NewCoolHandler(v1, u)
	NewLocationHandler(v1, u)
	NewUserHandler(v1, u)
	NewEventUserHandler(v1, u, c)
	NewEventGeneralHandler(v1, u, c)
	NewEventSessionHandler(v1, u, c)
}
