package main

import (
	"context"
	"go-community/internal/config"
	"go-community/internal/contract"
	"go-community/internal/pkg/logger"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	var (
		ctx = context.Background()
	)
	
	config, err := config.New(ctx)
	if err != nil {
		panic(err)
	}

	contract := contract.New(config)

	// Graceful shutdown
	go func() {
		if err := contract.Start(config.Application.Port); err != nil {
			log.Fatalf("server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := contract.Stop(c); err != nil {
        logger.Logger.Error("server shutdown failed: ", zap.Error(err))
    } else {
        logger.Logger.Info("Server gracefully stopped")
    }
}