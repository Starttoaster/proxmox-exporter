package prometheus

import (
	"testing"
	"time"
)

func TestFqAddPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple metric name", "node_up", "proxmox_node_up"},
		{"empty string", "", "proxmox_"},
		{"nested underscores", "node_cpu_total", "proxmox_node_cpu_total"},
		{"single word", "uptime", "proxmox_uptime"},
		{"already prefixed", "proxmox_test", "proxmox_proxmox_test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fqAddPrefix(tt.input)
			if result != tt.expected {
				t.Errorf("fqAddPrefix(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDaysUntilUnixTime(t *testing.T) {
	tests := []struct {
		name    string
		offset  time.Duration
		minDays int
		maxDays int
	}{
		{"30 days in future", 30 * 24 * time.Hour, 29, 30},
		{"1 day in future", 24 * time.Hour, 0, 1},
		{"1 day in past", -24 * time.Hour, -2, -1},
		{"now", 0, -1, 0},
		{"365 days in future", 365 * 24 * time.Hour, 364, 365},
		{"7 days in future", 7 * 24 * time.Hour, 6, 7},
		{"90 days in future", 90 * 24 * time.Hour, 89, 90},
		{"1 hour in future", 1 * time.Hour, 0, 0},
		{"30 days in past", -30 * 24 * time.Hour, -31, -30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notAfter := int(time.Now().Add(tt.offset).Unix())
			result := daysUntilUnixTime(notAfter)
			if result < tt.minDays || result > tt.maxDays {
				t.Errorf("daysUntilUnixTime() = %d, want between %d and %d", result, tt.minDays, tt.maxDays)
			}
		})
	}
}

func TestDaysUntilUnixTime_Epoch(t *testing.T) {
	result := daysUntilUnixTime(0)
	if result >= 0 {
		t.Errorf("days until epoch should be negative, got %d", result)
	}
}
