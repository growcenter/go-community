package middleware

import (
	"go-community/internal/config"
	"go-community/internal/pkg/logger"

	"github.com/labstack/echo/v4/middleware"
)

func (m *Middleware) Default(config *config.Configuration) {
	m.e.Use(middleware.Recover())
	m.e.Use(m.LoggingMiddleware(logger.Logger))
	m.e.Use(m.corsMiddleware(config))
}
