package middleware

import (
	"go-community/internal/config"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (m *Middleware) corsMiddleware(config *config.Configuration) echo.MiddlewareFunc {
	// origin := fmt.Sprintf("http://%s:%d", config.Application.Host, config.Application.Port)
	// origin := fmt.Sprintf("https://%s", config.Application.Host)

	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{config.Frontend.Host},
		AllowMethods:     []string{http.MethodDelete, http.MethodGet, http.MethodOptions, http.MethodPatch, http.MethodPost, http.MethodPut, http.MethodPatch},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-API-Key", "X-Api-Key", "X-Refresh-Token"},
		AllowCredentials: true,
	})
}
