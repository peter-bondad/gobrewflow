package logger

// Package logger provides one place to configure structured logging
// for the whole app. JSON in production (machine-readable, ships to
// log aggregators cleanly), human-readable text elsewhere.

import (
	"log/slog"
	"os"
	"strings"
)

// New builds a slog.Logger, sets it as the process-wide default, and
// returns it so it can also be injected explicitly where needed.
func New(env, level string) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     parseLevel(level),
		AddSource: env != "production",
	}

	var handler slog.Handler
	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	log := slog.New(handler).With(slog.String("env", env))
	slog.SetDefault(log)
	return log
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
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
