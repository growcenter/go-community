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

		logger.Log.Warn("[SERVICE-ERROR]", logStatusError, logError)
	} else {
		logger.Log.Info("[SERVICE]", zap.String("status", "success"))
	}
}
