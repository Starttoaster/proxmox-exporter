package logger

import (
	"log/slog"
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected slog.Level
	}{
		{"debug lowercase", "debug", slog.LevelDebug},
		{"debug mixed case", "Debug", slog.LevelDebug},
		{"debug uppercase", "DEBUG", slog.LevelDebug},
		{"info lowercase", "info", slog.LevelInfo},
		{"info mixed case", "Info", slog.LevelInfo},
		{"warn lowercase", "warn", slog.LevelWarn},
		{"warning lowercase", "warning", slog.LevelWarn},
		{"warning mixed case", "Warning", slog.LevelWarn},
		{"error lowercase", "error", slog.LevelError},
		{"error uppercase", "ERROR", slog.LevelError},
		{"unknown defaults to info", "unknown", slog.LevelInfo},
		{"empty defaults to info", "", slog.LevelInfo},
		{"garbage defaults to info", "foobar", slog.LevelInfo},
		{"trace unsupported defaults to info", "trace", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestInit(t *testing.T) {
	tests := []struct {
		name  string
		level string
	}{
		{"info level", "info"},
		{"debug level", "debug"},
		{"warn level", "warn"},
		{"error level", "error"},
		{"unknown level", "unknown"},
		{"empty level", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init(tt.level)
			if Logger == nil {
				t.Fatal("Logger should not be nil after Init")
			}
		})
	}
}

func TestInit_LoggerIsUsable(t *testing.T) {
	Init("debug")

	// Should not panic
	Logger.Info("test message")
	Logger.Debug("debug message")
	Logger.Warn("warn message")
	Logger.Error("error message")
	Logger.Info("structured", "key", "value", "count", 42)
}
