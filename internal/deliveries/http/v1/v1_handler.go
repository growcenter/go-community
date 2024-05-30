package v1

import (
	"go-community/internal/usecases"

	"github.com/labstack/echo/v4"
)

func NewV1Handler(g *echo.Group, u *usecases.Usecases) {
    v1 := g.Group("/v1")

    // Initialize handlers
    NewCampusHandler(v1, u)
	NewCoolHandler(v1, u)
}