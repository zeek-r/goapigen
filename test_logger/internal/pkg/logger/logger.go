// Package logger provides structured logging configuration using zapctxd
package logger

import (
	"os"

	"github.com/bool64/ctxd"
	"github.com/bool64/zapctxd"
	"go.uber.org/zap/zapcore"
)

// New creates and returns a logger with production settings
func New() ctxd.Logger {
	return NewWithConfig(false, zapcore.InfoLevel)
}

// NewDevelopment creates and returns a logger with development settings
func NewDevelopment() ctxd.Logger {
	return NewWithConfig(true, zapcore.DebugLevel)
}

// NewWithConfig creates and returns a logger with custom settings
func NewWithConfig(development bool, level zapcore.Level) ctxd.Logger {
	config := zapctxd.Config{
		Level:   level,
		DevMode: development,
	}

	return zapctxd.New(config)
}

// NewFromEnv creates and returns a logger based on environment variables
// LOG_LEVEL: debug, info, warn, error (default: info)
// LOG_DEVELOPMENT: true, false (default: false)
func NewFromEnv() ctxd.Logger {
	level := zapcore.InfoLevel
	if levelStr := os.Getenv("LOG_LEVEL"); levelStr != "" {
		if parsedLevel, err := zapcore.ParseLevel(levelStr); err == nil {
			level = parsedLevel
		}
	}

	development := false
	if devStr := os.Getenv("LOG_DEVELOPMENT"); devStr == "true" {
		development = true
	}

	return NewWithConfig(development, level)
} 