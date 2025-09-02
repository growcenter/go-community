package contract

import (
	"context"
	"fmt"
	indonesiaAPI "go-community/internal/clients/indonesia-api"
	"go-community/internal/config"
	handler "go-community/internal/deliveries/http"
	"go-community/internal/pkg/authorization"
	"go-community/internal/pkg/database/postgre"
	"go-community/internal/pkg/google"
	"go-community/internal/pkg/logger"
	"go-community/internal/repositories/pgsql"
	"go-community/internal/usecases"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Contract struct {
	echo *echo.Echo
}

func New(config *config.Configuration) *Contract {
	// Initialize Echo Framework
	var e = echo.New()

	// Initialize logger
	logger.Init(config)

	// Connect to PostgreSQL Database
	psql, err := postgre.ConnectWithGORM(config)
	if err != nil {
		logger.Logger.Fatal(fmt.Sprintf("[DATABASE_ERROR] Failed to setup database - %v", err), zap.Error(err))
	}

	// Check Database is working
	sql, err := psql.DB()
	if err != nil {
		logger.Logger.Fatal(fmt.Sprintf("[DATABASE_ERROR] Failed to setup database - %v", err), zap.Error(err))
	}

	if err = sql.Ping(); err != nil {
		logger.Logger.Fatal(fmt.Sprintf("[DATABASE_ERROR] Failed to connect the database - %v", err), zap.Error(err))
	}

	// Google
	oauthGoogle, err := google.NewGoogle(config)
	if err != nil {
		logger.Logger.Fatal(fmt.Sprintf("[GOOGLE_ERROR] Failed to setup google oauth - %v", err), zap.Error(err))
	}

	// Auth
	auth, err := authorization.NewAuthorization(config)
	if err != nil {
		logger.Logger.Fatal(fmt.Sprintf("[AUTH_ERROR] Failed to setup auth - %v", err), zap.Error(err))
	}

	// Indonesia API
	indonesiaAPI := indonesiaAPI.NewClient(*config)
	// Register Repository
	postgreRepository := pgsql.New(psql)

	// Register Service
	usecase := usecases.New(usecases.Dependencies{
		Repository:    postgreRepository,
		Google:        oauthGoogle,
		Authorization: auth,
		Config:        config,
		Indonesia:     &indonesiaAPI,
	})

	// Register Handler
	handler.New(e, usecase, config, auth)

	return &Contract{
		echo: e,
	}
}

func (c *Contract) Start(port int) error {
	return c.echo.Start(":" + strconv.Itoa(port))
}

func (c *Contract) Stop(ctx context.Context) error {
	return c.echo.Shutdown(ctx)
}
