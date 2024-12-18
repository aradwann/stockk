package config

import (
	"log/slog"
	"os"
)

// CreateLogger creates a structured logger
func CreateLogger(env string) *slog.Logger {
	var logger *slog.Logger

	switch env {
	case "production":
		// JSON logger for production (easier to parse in log management systems)
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	default:
		// Pretty-printed logger for development
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
			// Add source code location to logs
			AddSource: true,
		}))
	}

	return logger
}
