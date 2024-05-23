package middleware

import (
	"go-community/internal/pkg/logger"

	"github.com/labstack/echo/v4/middleware"
)

func (m *Middleware) Default() {
	m.e.Use(middleware.Recover())
	m.e.Use(m.LoggingMiddleware(logger.Logger))
}