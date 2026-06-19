// Package logger provides a configured slog.Logger for the application.
package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// New creates a slog.Logger from a level ("debug", "info", "warn", "error")
// and format ("json" or "text"). Unknown levels fall back to info; unknown
// formats fall back to JSON.
func New(level, format string) *slog.Logger {
	handlerOpts := &slog.HandlerOptions{Level: parseLevel(level)}

	var handler slog.Handler
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "text":
		handler = slog.NewTextHandler(os.Stdout, handlerOpts)
	default:
		handler = slog.NewJSONHandler(os.Stdout, handlerOpts)
	}

	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
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

// Level returns the slog.Level for the given name, or info if unknown.
func Level(level string) slog.Level { return parseLevel(level) }

// ParseLevelFlag parses a level name and reports an error if it is not one of
// debug, info, warn, or error. Useful for config validation.
func ParseLevelFlag(level string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid log level %q", level)
	}
}
