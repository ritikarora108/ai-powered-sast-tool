// backend/internal/logger/logger.go
package logger

import (
	"context"
	"net/http"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Global logger instance
	log *zap.Logger
	// Once guard to ensure the logger is only initialized once
	once sync.Once
)

// contextKey is a private type for request-scoped logger in context
type contextKey int

// loggerKey is the key for retrieving the logger from the context
const loggerKey contextKey = iota

// Init initializes the logger with production configuration
func Init() {
	once.Do(func() {
		// Determine if we're in development or production
		env := os.Getenv("APP_ENV")
		isDevelopment := env == "" || env == "development"

		// Configure logging level
		logLevel := zap.InfoLevel
		if os.Getenv("LOG_LEVEL") == "debug" {
			logLevel = zap.DebugLevel
		}

		// Create encoder config based on environment
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

		// Configure the logger
		config := zap.Config{
			Level:             zap.NewAtomicLevelAt(logLevel),
			Development:       isDevelopment,
			Encoding:          "json", // Use JSON in production
			EncoderConfig:     encoderConfig,
			OutputPaths:       []string{"stdout"},
			ErrorOutputPaths:  []string{"stderr"},
			DisableCaller:     !isDevelopment,
			DisableStacktrace: !isDevelopment,
		}

		// If in development mode, use a more readable console encoder
		if isDevelopment {
			config.Encoding = "console"
		}

		// Create the logger
		logger, err := config.Build()
		if err != nil {
			panic("Failed to initialize logger: " + err.Error())
		}

		logger.Info("Logger initialized",
			zap.Bool("development", isDevelopment),
			zap.String("level", logLevel.String()))

		log = logger
	})
}

// Get returns the global logger instance
func Get() *zap.Logger {
	if log == nil {
		Init()
	}
	return log
}

// With returns a child logger with additional fields
func With(fields ...zapcore.Field) *zap.Logger {
	return Get().With(fields...)
}

// WithRequest returns a logger with HTTP request details
func WithRequest(r *http.Request) *zap.Logger {
	return With(
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
		zap.String("user_agent", r.UserAgent()),
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("request_id", r.Header.Get("X-Request-ID")),
	)
}

// Debug logs a debug message
func Debug(msg string, fields ...zapcore.Field) {
	Get().Debug(msg, fields...)
}

// Info logs an info message
func Info(msg string, fields ...zapcore.Field) {
	Get().Info(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zapcore.Field) {
	Get().Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zapcore.Field) {
	Get().Error(msg, fields...)
}

// Fatal logs a fatal message and then calls os.Exit(1)
func Fatal(msg string, fields ...zapcore.Field) {
	Get().Fatal(msg, fields...)
}

// Sync flushes any buffered log entries
func Sync() error {
	return Get().Sync()
}

// WithContext adds the logger to the context
func WithContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext retrieves the logger from the context
func FromContext(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return logger
	}
	return Get()
}
