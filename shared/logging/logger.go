package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	UserIDKey    contextKey = "user_id"
	ServiceKey   contextKey = "service"
)

var Logger *slog.Logger

func Init(serviceName string) {
	// Determine output destination
	var writer io.Writer = os.Stdout

	// Check if LOG_FILE is specified
	if logFile := os.Getenv("LOG_FILE"); logFile != "" {
		// Create logs directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
			// Fall back to stdout if can't create directory
			writer = os.Stdout
		} else {
			// Open or create log file
			file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				// Fall back to stdout if can't open file
				writer = os.Stdout
			} else {
				// Use both stdout and file (for Docker logs + file logs)
				writer = io.MultiWriter(os.Stdout, file)
			}
		}
	}

	// Configure handler based on environment
	var handler slog.Handler
	logLevel := getLogLevel()

	if os.Getenv("LOG_FORMAT") == "text" {
		// Text format for development
		handler = slog.NewTextHandler(writer, &slog.HandlerOptions{
			Level: logLevel,
		})
	} else {
		// JSON format for production (default)
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
			Level: logLevel,
		})
	}

	// Create logger with service name
	Logger = slog.New(handler).With("service", serviceName)

	Logger.Info("Logger initialized", "service", serviceName, "level", logLevel.String())
}

func getLogLevel() slog.Level {
	level := strings.ToLower(os.Getenv("LOG_LEVEL"))
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithContext creates a logger with context fields
func WithContext(ctx context.Context) *slog.Logger {
	logger := Logger

	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		logger = logger.With("request_id", requestID)
	}

	if userID, ok := ctx.Value(UserIDKey).(int); ok {
		logger = logger.With("user_id", userID)
	}

	return logger
}

// Context helpers
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

func WithUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}
