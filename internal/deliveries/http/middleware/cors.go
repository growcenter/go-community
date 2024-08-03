package middleware

import (
	"fmt"
	"go-community/internal/config"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (m *Middleware) corsMiddleware(config *config.Configuration) echo.MiddlewareFunc {
	// origin := fmt.Sprintf("http://%s:%d", config.Frontend.Host, config.Frontend.Port)
	origin := fmt.Sprintf("http://%s", config.Frontend.Host)

	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{origin},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete, http.MethodOptions, http.MethodPut},
		AllowHeaders: []string{echo.HeaderAccept, echo.HeaderAcceptEncoding, echo.HeaderContentType, echo.HeaderAccept, "X-API-Key", echo.HeaderAuthorization},
	})
}
