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
			logger.Log.Warn("[REPOSITORY-ERROR]", zap.Error(err))
		} else {
			logger.Log.Error("[REPOSITORY-ERROR]", zap.Error(err))
		}
	} else {
		logger.Log.Info("[REPOSITORY]", zap.String("status", "success"))
	}
}
