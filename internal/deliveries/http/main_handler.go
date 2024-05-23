package http

import (
	"go-community/internal/deliveries/http/health"
	"go-community/internal/deliveries/http/middleware"
	v1 "go-community/internal/deliveries/http/v1"
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
func New(e *echo.Echo, u *usecases.Usecases) {
	// Middleware for Recover and Logging
	middleware := middleware.New(e)
	middleware.Default()

	// Input swagger initalization here


	e.GET("/", func(ctx echo.Context) error {
		message := "Welcome to GROW Center API Service"
		return ctx.String(http.StatusOK, message)
	})

	// API Grouping
	api := e.Group("/api")

	// Initialize Health & V1 Handlers
	health.NewHealhHandler(api, *u)
	v1.NewV1Handler(api, u)
}




// func (h *Handler) Start(ctx context.Context, contract *contract.Contract) *echo.Echo {
// 	router := echo.New()

// 	middleware := middleware.New(router, contract)
// 	middleware.Default()

// 	// Put Swagger HERE

// 	// Initialize router
// 	router.GET("/", func(ctx echo.Context) error {
// 		message := fmt.Sprintf("Welcome to %s", contract.Config.Application.Name)
// 		return ctx.String(http.StatusOK, message)
// 	})

// 	// Grouping API
// 	api := router.Group("/api")

// 	// Initialize handlers from Health and V1
// 	health.New(api, contract.Usecase.Health)

// 	return router
// }