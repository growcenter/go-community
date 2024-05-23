package v1

import (
	"go-community/internal/usecases"

	"github.com/labstack/echo/v4"
)

func NewV1Handler(g *echo.Group, u *usecases.Usecases) {
    v1 := g.Group("/v1")

    // Initialize user handler
    NewCampusHandler(v1, u)
}