package usecases

import (
	"context"
	"go-community/internal/pkg/logger"

	"go.uber.org/zap"
)

func LogService(ctx context.Context, err error) {
	if err != nil {
		logStatusError := zap.String("status", "error")
		logError := zap.Error(err)

		logger.Logger.Warn("[SERVICE-ERROR]", logStatusError, logError)
	} else {
		logger.Logger.Info("[SERVICE]", zap.String("status", "success"))
	}
}