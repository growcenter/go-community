package contract

import (
	"context"
	"fmt"
	"go-community/internal/config"
	handler "go-community/internal/deliveries/http"
	"go-community/internal/pkg/authorization"
	"go-community/internal/pkg/cache"
	"go-community/internal/pkg/database/postgre"
	"go-community/internal/pkg/google"
	"go-community/internal/pkg/logger"
	"go-community/internal/repositories/pgsql"
	"go-community/internal/usecases"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

type Contract struct {
	echo *echo.Echo
}

func New(config *config.Configuration) *Contract {
	// Initialize Echo Framework
	var e = echo.New()

	// Initialize logger
	logger.Initialize()

	// Connect to PostgreSQL Database
	psql, err := postgre.ConnectWithGORM(config)
	if err != nil {
		logger.Instance.Fatal(context.Background(), fmt.Sprintf("[DATABASE_ERROR] Failed to setup database - %v", err), logger.Error(err))
	}

	// Check Database is working
	sql, err := psql.DB()
	if err != nil {
		logger.Instance.Fatal(context.Background(), fmt.Sprintf("[DATABASE_ERROR] Failed to setup database - %v", err), logger.Error(err))
	}

	if err = sql.Ping(); err != nil {
		logger.Instance.Fatal(context.Background(), fmt.Sprintf("[DATABASE_ERROR] Failed to connect the database - %v", err), logger.Error(err))
	}

	// Google
	oauthGoogle, err := google.NewGoogle(config)
	if err != nil {
		logger.Instance.Fatal(context.Background(), fmt.Sprintf("[GOOGLE_ERROR] Failed to setup google oauth - %v", err), logger.Error(err))
	}

	// Auth
	auth, err := authorization.NewAuthorization(config)
	if err != nil {
		logger.Instance.Fatal(context.Background(), fmt.Sprintf("[AUTH_ERROR] Failed to setup auth - %v", err), logger.Error(err))
	}

	// Register Repository
	postgreRepository := pgsql.New(psql)

	// Initialize Memory Cache with configurable TTL
	cacheTTL := time.Duration(config.Cache.TTLMinutes) * time.Minute
	memoryCache := cache.New(cacheTTL)

	// Register Service
	usecase := usecases.New(usecases.Dependencies{
		Repository:    postgreRepository,
		Google:        oauthGoogle,
		Authorization: auth,
		Config:        config,
		Cache:         memoryCache,
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
