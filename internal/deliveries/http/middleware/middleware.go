package middleware

import (
	"go-community/internal/config"

	"github.com/labstack/echo/v4"
)

type Middleware struct {
	e   *echo.Echo
	cfg *config.Configuration
}

func New(e *echo.Echo, cfg *config.Configuration) Middleware {
	return Middleware{
		e:   e,
		cfg: cfg,
	}
}
