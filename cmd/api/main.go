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
		logger.Instance.Error(ctx, "server shutdown failed: ", logger.Error(err))
	} else {
		logger.Instance.Info(ctx, "Server gracefully stopped")
	}
}
