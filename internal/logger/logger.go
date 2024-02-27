package logger

import (
	"log"
	"log/slog"
	"os"
	"strings"
)

// Logger is a custom logger from the stdlib slog package
var Logger *slog.Logger

// Init custom init function that accepts the log level for the application
func Init(level string) {
	Logger = slog.New(
		slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: parseLogLevel(level),
			},
		),
	)
}

// Function to convert log level string to slog.Level
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		log.Printf("unknown log level specified \"%s\", defaulting to info level", level)
		return slog.LevelInfo
	}
}
