package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"strings"
)

type contextKey string

const (
	RequestIDKey contextKey = "requestID"
	UserIDKey    contextKey = "userID"
)

var (
	defaultLogger *slog.Logger
	debugMode     bool
)

// Init initializes the global logger with the specified level.
// When level is "debug", log messages include source file and line number.
func Init(level string) {
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
		debugMode = true
	case "info":
		logLevel = slog.LevelInfo
		debugMode = false
	case "warn", "warning":
		logLevel = slog.LevelWarn
		debugMode = false
	case "error":
		logLevel = slog.LevelError
		debugMode = false
	default:
		logLevel = slog.LevelInfo
		debugMode = false
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	if debugMode {
		opts.AddSource = true
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

// WithContext creates a logger with context values (requestID, userID) attached.
func WithContext(ctx context.Context) *slog.Logger {
	logger := defaultLogger
	if logger == nil {
		logger = slog.Default()
	}

	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		logger = logger.With("requestID", requestID)
	}

	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		logger = logger.With("userID", userID)
	}

	return logger
}

// Debug logs at debug level with context.
func Debug(ctx context.Context, msg string, args ...any) {
	logger := WithContext(ctx)
	if debugMode {
		args = appendSource(args)
	}
	logger.Debug(msg, args...)
}

// Info logs at info level with context.
func Info(ctx context.Context, msg string, args ...any) {
	logger := WithContext(ctx)
	if debugMode {
		args = appendSource(args)
	}
	logger.Info(msg, args...)
}

// Warn logs at warn level with context.
func Warn(ctx context.Context, msg string, args ...any) {
	logger := WithContext(ctx)
	if debugMode {
		args = appendSource(args)
	}
	logger.Warn(msg, args...)
}

// Error logs at error level with context.
func Error(ctx context.Context, msg string, args ...any) {
	logger := WithContext(ctx)
	if debugMode {
		args = appendSource(args)
	}
	logger.Error(msg, args...)
}

// appendSource adds caller file:line to log arguments when debug mode is enabled.
func appendSource(args []any) []any {
	_, file, line, ok := runtime.Caller(2)
	if ok {
		// Shorten file path to just the last two segments
		parts := strings.Split(file, "/")
		if len(parts) > 2 {
			file = strings.Join(parts[len(parts)-2:], "/")
		}
		args = append(args, "caller", slog.GroupValue(
			slog.String("file", file),
			slog.Int("line", line),
		))
	}
	return args
}

// ContextWithRequestID adds a request ID to the context.
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// ContextWithUserID adds a user ID to the context.
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetRequestID retrieves the request ID from context.
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}
