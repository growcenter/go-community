package http

import (
	echoSwagger "github.com/swaggo/echo-swagger"
	_ "go-community/docs"
	"go-community/internal/config"
	"go-community/internal/deliveries/http/health"
	"go-community/internal/deliveries/http/middleware"
	v1 "go-community/internal/deliveries/http/v1"
	v2 "go-community/internal/deliveries/http/v2"
	"go-community/internal/pkg/authorization"
	"go-community/internal/usecases"
	"net/http"

	"github.com/labstack/echo/v4"
)

// @title GO-COMMUNITY API DOCUMENTATION
// @version 1.0
// @description This is a go-community api docs.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api
// @schemes https
func New(e *echo.Echo, u *usecases.Usecases, c *config.Configuration, a *authorization.Auth) {
	// Middleware for Recover and Logging
	middleware := middleware.New(e)
	middleware.Default(c)

	e.GET("/", func(ctx echo.Context) error {
		message := "Welcome to GROW Community API Service!"
		return ctx.String(http.StatusOK, message)
	})

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// API Grouping
	api := e.Group("/api")

	// Initialize Health & V1 Handlers
	health.NewHealhHandler(api, *u)
	v1.NewV1Handler(api, u, c, a)
	v2.NewV2Handler(api, u, c, a)
}
