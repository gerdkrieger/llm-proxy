// Package logger provides a structured logging system based on zerolog.
// It supports JSON and console output, log levels, and request ID tracking.
package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ContextKey type for context keys to avoid collisions
type ContextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey ContextKey = "request_id"
)

// Config holds logging configuration
type Config struct {
	Level  string // debug, info, warn, error, fatal
	Format string // json or console
	Output io.Writer
}

// Logger wraps zerolog.Logger with additional functionality
type Logger struct {
	logger *zerolog.Logger
}

// New creates a new Logger instance with the given configuration
func New(cfg Config) *Logger {
	// Parse log level
	level := parseLevel(cfg.Level)
	zerolog.SetGlobalLevel(level)

	// Configure output writer
	output := cfg.Output
	if output == nil {
		output = os.Stdout
	}

	// Create logger based on format
	var zerologger zerolog.Logger
	if cfg.Format == "console" {
		// Console output with colors for development
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
		}
		zerologger = zerolog.New(output).With().
			Timestamp().
			Caller().
			Logger()
	} else {
		// JSON output for production
		zerologger = zerolog.New(output).With().
			Timestamp().
			Caller().
			Str("service", "llm-proxy").
			Logger()
	}

	return &Logger{
		logger: &zerologger,
	}
}

// parseLevel converts string level to zerolog.Level
func parseLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

// WithContext returns a logger with context values (e.g., request ID)
func (l *Logger) WithContext(ctx context.Context) *zerolog.Logger {
	logger := *l.logger

	// Add request ID if present
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		logger = logger.With().Str("request_id", requestID).Logger()
	}

	return &logger
}

// WithRequestID returns a logger with request ID
func (l *Logger) WithRequestID(requestID string) *zerolog.Logger {
	logger := l.logger.With().Str("request_id", requestID).Logger()
	return &logger
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

// Error logs an error message
func (l *Logger) Error(err error, msg string) {
	l.logger.Error().Err(err).Msg(msg)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(err error, format string, args ...interface{}) {
	l.logger.Error().Err(err).Msgf(format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}

// With returns a child logger with additional fields
func (l *Logger) With() *zerolog.Logger {
	return l.logger
}

// GetZerolog returns the underlying zerolog.Logger (for advanced usage)
func (l *Logger) GetZerolog() *zerolog.Logger {
	return l.logger
}

// Global logger instance (initialized by Init)
var defaultLogger *Logger

// Init initializes the global logger
func Init(cfg Config) {
	defaultLogger = New(cfg)
	log.Logger = *defaultLogger.logger
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	if defaultLogger == nil {
		// Fallback to default config if not initialized
		Init(Config{
			Level:  "info",
			Format: "json",
			Output: os.Stdout,
		})
	}
	return defaultLogger
}

// Convenience functions using the global logger

// Debug logs a debug message using the global logger
func Debug(msg string) {
	GetLogger().Debug(msg)
}

// Debugf logs a formatted debug message using the global logger
func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

// Info logs an info message using the global logger
func Info(msg string) {
	GetLogger().Info(msg)
}

// Infof logs a formatted info message using the global logger
func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

// Warn logs a warning message using the global logger
func Warn(msg string) {
	GetLogger().Warn(msg)
}

// Warnf logs a formatted warning message using the global logger
func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

// Error logs an error message using the global logger
func Error(err error, msg string) {
	GetLogger().Error(err, msg)
}

// Errorf logs a formatted error message using the global logger
func Errorf(err error, format string, args ...interface{}) {
	GetLogger().Errorf(err, format, args...)
}

// Fatal logs a fatal message using the global logger and exits
func Fatal(msg string) {
	GetLogger().Fatal(msg)
}

// Fatalf logs a formatted fatal message using the global logger and exits
func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}

// WithContext returns a logger with context values using the global logger
func WithContext(ctx context.Context) *zerolog.Logger {
	return GetLogger().WithContext(ctx)
}

// WithRequestID returns a logger with request ID using the global logger
func WithRequestID(requestID string) *zerolog.Logger {
	return GetLogger().WithRequestID(requestID)
}
