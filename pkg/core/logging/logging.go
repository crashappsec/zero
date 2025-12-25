// Package logging provides structured logging for Zero using slog.
// It provides a consistent logging interface across all packages.
package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"runtime"
	"time"
)

// Level represents a logging level
type Level = slog.Level

// Logging levels
const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

// Logger wraps slog.Logger with Zero-specific functionality
type Logger struct {
	*slog.Logger
}

// defaultLogger is the global default logger
var defaultLogger = New(os.Stderr, LevelInfo)

// Default returns the default logger
func Default() *Logger {
	return defaultLogger
}

// SetDefault sets the default logger
func SetDefault(l *Logger) {
	defaultLogger = l
	slog.SetDefault(l.Logger)
}

// New creates a new logger that writes to w at the given level
func New(w io.Writer, level Level) *Logger {
	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Shorten time format
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(a.Value.Time().Format("15:04:05"))
			}
			// Shorten source path
			if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				if source != nil {
					source.File = shortPath(source.File)
				}
			}
			return a
		},
	}

	handler := slog.NewTextHandler(w, opts)
	return &Logger{slog.New(handler)}
}

// NewJSON creates a new logger that outputs JSON
func NewJSON(w io.Writer, level Level) *Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewJSONHandler(w, opts)
	return &Logger{slog.New(handler)}
}

// NewNop creates a logger that discards all output
func NewNop() *Logger {
	return &Logger{slog.New(slog.NewTextHandler(io.Discard, nil))}
}

// WithContext returns a logger with context values
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// Extract any context values we care about
	// This is a placeholder for future context propagation (trace IDs, etc.)
	return l
}

// With returns a logger with additional attributes
func (l *Logger) With(args ...any) *Logger {
	return &Logger{l.Logger.With(args...)}
}

// WithScanner returns a logger with scanner name attribute
func (l *Logger) WithScanner(name string) *Logger {
	return l.With("scanner", name)
}

// WithRepo returns a logger with repository attribute
func (l *Logger) WithRepo(repo string) *Logger {
	return l.With("repo", repo)
}

// WithOperation returns a logger with operation attribute
func (l *Logger) WithOperation(op string) *Logger {
	return l.With("op", op)
}

// WithError returns a logger with error attribute
func (l *Logger) WithError(err error) *Logger {
	if err == nil {
		return l
	}
	return l.With("error", err.Error())
}

// WithDuration returns a logger with duration attribute
func (l *Logger) WithDuration(d time.Duration) *Logger {
	return l.With("duration_ms", d.Milliseconds())
}

// Debug logs at debug level
func (l *Logger) Debug(msg string, args ...any) {
	l.Logger.Debug(msg, args...)
}

// Info logs at info level
func (l *Logger) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

// Warn logs at warn level
func (l *Logger) Warn(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

// Error logs at error level
func (l *Logger) Error(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}

// Debugf logs at debug level with formatted message
func (l *Logger) Debugf(format string, args ...any) {
	l.Logger.Debug(format, args...)
}

// Infof logs at info level with formatted message
func (l *Logger) Infof(format string, args ...any) {
	l.Logger.Info(format, args...)
}

// Warnf logs at warn level with formatted message
func (l *Logger) Warnf(format string, args ...any) {
	l.Logger.Warn(format, args...)
}

// Errorf logs at error level with formatted message
func (l *Logger) Errorf(format string, args ...any) {
	l.Logger.Error(format, args...)
}

// Package-level convenience functions using default logger

// Debug logs at debug level using default logger
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// Info logs at info level using default logger
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Warn logs at warn level using default logger
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// Error logs at error level using default logger
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// WithScanner returns a logger with scanner name
func WithScanner(name string) *Logger {
	return defaultLogger.WithScanner(name)
}

// WithRepo returns a logger with repository
func WithRepo(repo string) *Logger {
	return defaultLogger.WithRepo(repo)
}

// WithOperation returns a logger with operation
func WithOperation(op string) *Logger {
	return defaultLogger.WithOperation(op)
}

// Helper functions

// shortPath returns the last two path components
func shortPath(path string) string {
	// Find last two slashes
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			for j := i - 1; j >= 0; j-- {
				if path[j] == '/' {
					return path[j+1:]
				}
			}
			// Only one slash found, return original path (already short enough)
			return path
		}
	}
	// No slashes, return as-is
	return path
}

// Timing returns a function that logs the duration when called
// Usage: defer logging.Timing(logger, "operation")()
func Timing(l *Logger, operation string) func() {
	start := time.Now()
	return func() {
		l.WithDuration(time.Since(start)).Info(operation + " completed")
	}
}

// LogPanic recovers from a panic and logs it
// Usage: defer logging.LogPanic(logger)
func LogPanic(l *Logger) {
	if r := recover(); r != nil {
		stack := make([]byte, 4096)
		n := runtime.Stack(stack, false)
		l.Error("panic recovered",
			"panic", r,
			"stack", string(stack[:n]),
		)
	}
}

// Config holds logger configuration
type Config struct {
	Level      Level
	Output     io.Writer
	JSON       bool
	AddSource  bool
}

// DefaultConfig returns default logging configuration
func DefaultConfig() Config {
	return Config{
		Level:  LevelInfo,
		Output: os.Stderr,
		JSON:   false,
	}
}

// NewFromConfig creates a logger from configuration
func NewFromConfig(cfg Config) *Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stderr
	}

	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.AddSource,
	}

	var handler slog.Handler
	if cfg.JSON {
		handler = slog.NewJSONHandler(cfg.Output, opts)
	} else {
		handler = slog.NewTextHandler(cfg.Output, opts)
	}

	return &Logger{slog.New(handler)}
}

// ParseLevel parses a level string
func ParseLevel(s string) Level {
	switch s {
	case "debug", "DEBUG":
		return LevelDebug
	case "info", "INFO":
		return LevelInfo
	case "warn", "WARN", "warning", "WARNING":
		return LevelWarn
	case "error", "ERROR":
		return LevelError
	default:
		return LevelInfo
	}
}
