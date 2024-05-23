package logger

import (
	"fmt"
	"go-community/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func Init(config *config.Configuration) {
	var err error
	
	if config.Application.Environment == "dev" {
		Logger, err = zap.NewDevelopment()
	} else {
		prod := zap.NewProductionConfig()
		prod.EncoderConfig.TimeKey = "timestamp"
		prod.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
        Logger, err = prod.Build()
	}

	if err != nil {
		panic(fmt.Sprintf("[LOG_ERROR] Failed to setup logger - %v", err))
	}

	zap.ReplaceGlobals(Logger)

	// config := zap.NewDevelopmentConfig()
	// config.EncoderConfig= zapcore.EncoderConfig{
	// 	TimeKey:       "time",
    //     LevelKey:      "level",
    //     NameKey:       "logger",
    //     CallerKey:     "caller",
    //     MessageKey:    "message",
    //     StacktraceKey: "stacktrace",
    //     LineEnding:    zapcore.DefaultLineEnding,
	// 	EncodeLevel: zapcore.CapitalColorLevelEncoder,
	// 	EncodeTime: zapcore.ISO8601TimeEncoder,
	// 	EncodeCaller: zapcore.FullCallerEncoder,
	// 	EncodeName: zapcore.FullNameEncoder,
	// 	EncodeDuration: zapcore.SecondsDurationEncoder,

	// }

	// logger, err := config.Build()
	// if err != nil {
	// 	panic("failed to initialize logger")
	// }

	// Logger = logger
}