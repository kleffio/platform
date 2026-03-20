package logging

import (
	"log/slog"
	"os"
)

// NewLogger creates a production JSON logger with the given minimum level.
// level accepts "debug", "info", "warn", "error". Unrecognised values default
// to "info".
func NewLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}
