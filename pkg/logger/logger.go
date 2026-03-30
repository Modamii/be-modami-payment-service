package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	once   sync.Once
	global *zap.Logger
)

// Init initializes the global logger.
func Init(env string) {
	once.Do(func() {
		var cfg zap.Config
		if env == "production" {
			cfg = zap.NewProductionConfig()
		} else {
			cfg = zap.NewDevelopmentConfig()
			cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		l, err := cfg.Build()
		if err != nil {
			panic(err)
		}
		global = l
	})
}

// L returns the global logger, initializing with development config if not yet initialized.
func L() *zap.Logger {
	if global == nil {
		Init("development")
	}
	return global
}

// Sync flushes buffered log entries.
func Sync() {
	if global != nil {
		_ = global.Sync()
	}
}
