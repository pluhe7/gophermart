package logger

import (
	"go.uber.org/zap"
)

var Log = zap.NewNop()

func InitLogger(level string) {
	atomicLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		Log.Error("parse logger level", zap.Error(err))
		return
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = atomicLevel

	logger, err := cfg.Build()
	if err != nil {
		Log.Error("build logger config", zap.Error(err))
		return
	}

	Log = logger
}
