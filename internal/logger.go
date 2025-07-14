package internal

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// CreateLogger sets up a production-grade zap logger with ISO8601 timestamps
// and stripped-down output (no stack traces, no caller info).
func CreateLogger() *zap.SugaredLogger {
	// Configure the encoder (controls log format)
	encoder := zap.NewProductionEncoderConfig()
	encoder.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder.TimeKey = "timestamp"
	encoder.StacktraceKey = ""

	// Configure logger behavior
	conf := zap.NewProductionConfig()
	conf.EncoderConfig = encoder
	conf.DisableStacktrace = true
	conf.DisableCaller = true

	// Build the logger and handle any error
	logger, err := conf.Build()
	if err != nil {
		// Fall back to default logger if initialization fails
		log.Fatalf("Failed to initialize zap logger: %v", err)
	}

	return logger.Sugar()
}
