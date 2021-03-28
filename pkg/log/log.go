package log

import (
	"go.uber.org/zap"
)

var rootLogger *zap.Logger

// GetLogger returns a named logger
func GetLogger(name string) *zap.SugaredLogger {
	return rootLogger.Named(name).Sugar()
}

func init() {
	level := zap.InfoLevel

	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	rootLogger, _ = cfg.Build()
	rootLogger.Sugar()
}
