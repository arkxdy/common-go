package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// Logger defines the minimal interface for structured logging.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
	WithGroup(name string) Logger
}

// Config holds settings for the logger.
type Config struct {
	Level  slog.Level
	Format string // "json" or "text"
	Output io.Writer
}

// DefaultConfig returns a sane default configuration.
func DefaultConfig() Config {
	return Config{
		Level:  slog.LevelInfo,
		Format: "json",
		Output: os.Stdout,
	}
}

// New creates a new Logger from the given configuration.
func New(cfg Config) Logger {
	opts := &slog.HandlerOptions{
		Level: cfg.Level,
	}
	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(cfg.Output, opts)
	} else {
		handler = slog.NewTextHandler(cfg.Output, opts)
	}
	return &slogLogger{
		logger: slog.New(handler),
	}
}

type slogLogger struct {
	logger *slog.Logger
}

func (l *slogLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *slogLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *slogLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *slogLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

func (l *slogLogger) With(args ...any) Logger {
	return &slogLogger{logger: l.logger.With(args...)}
}

func (l *slogLogger) WithGroup(name string) Logger {
	return &slogLogger{logger: l.logger.WithGroup(name)}
}

// Helper to extract a logger from context (optional)
type contextKey struct{}

func WithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(contextKey{}).(Logger); ok {
		return logger
	}
	return New(DefaultConfig())
}
