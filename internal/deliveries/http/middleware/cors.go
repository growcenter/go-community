package middleware

import (
	"go-community/internal/config"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (m *Middleware) corsMiddleware(config *config.Configuration) echo.MiddlewareFunc {
	// origin := fmt.Sprintf("http://%s:%d", config.Frontend.Host, config.Frontend.Port)
	// origin := fmt.Sprintf("http://%s", config.Frontend.Host)

	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"*"},
		AllowHeaders: []string{"*"},
	})
}
