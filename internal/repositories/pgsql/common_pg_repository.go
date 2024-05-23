package pgsql

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/pkg/logger"

	"go.uber.org/zap"
)

func LogRepository(ctx context.Context, err error) {
	if err != nil {
		if err == models.ErrorNoRows {
			logger.Logger.Warn("[REPOSITORY-ERROR]", zap.String("status", "error"), zap.Error(err))
		} else {
			logger.Logger.Error("[REPOSITORY-ERROR]", zap.String("status", "error"), zap.Error(err))
		}
	} else {
		logger.Logger.Info("[REPOSITORY]", zap.String("status", "success"))
	}
}