package logging

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func init() {
	SetLevel(slog.LevelInfo)
}

func FromEnv() {
	levelStr := os.Getenv("LOG_LEVEL")
	var level slog.Level
	switch levelStr {
	case "debug", "DEBUG":
		level = slog.LevelDebug
	case "warn", "WARN":
		level = slog.LevelWarn
	case "error", "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	SetLevel(level)
}

func SetLevel(level slog.Level) {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}

// SetID updates the Logger's prefix
func SetID(id string) {
	Logger = Logger.With("id", id)
	slog.SetDefault(Logger)
}
