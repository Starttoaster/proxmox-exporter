package proxmox

import (
	"testing"
	"time"
)

func TestGetBannedClientCount(t *testing.T) {
	tests := []struct {
		name     string
		setup    map[string]wrappedClient
		expected int
	}{
		{
			name:     "no clients",
			setup:    map[string]wrappedClient{},
			expected: 0,
		},
		{
			name: "all unbanned",
			setup: map[string]wrappedClient{
				"host1": {banned: false},
				"host2": {banned: false},
			},
			expected: 0,
		},
		{
			name: "all banned",
			setup: map[string]wrappedClient{
				"host1": {banned: true, bannedUntil: time.Now().Add(time.Minute)},
				"host2": {banned: true, bannedUntil: time.Now().Add(time.Minute)},
			},
			expected: 2,
		},
		{
			name: "mixed banned and unbanned",
			setup: map[string]wrappedClient{
				"host1": {banned: true, bannedUntil: time.Now().Add(time.Minute)},
				"host2": {banned: false},
				"host3": {banned: true, bannedUntil: time.Now().Add(time.Minute)},
			},
			expected: 2,
		},
		{
			name: "single client unbanned",
			setup: map[string]wrappedClient{
				"host1": {banned: false},
			},
			expected: 0,
		},
		{
			name: "single client banned",
			setup: map[string]wrappedClient{
				"host1": {banned: true, bannedUntil: time.Now().Add(time.Minute)},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients = tt.setup
			result := GetBannedClientCount()
			if result != tt.expected {
				t.Errorf("GetBannedClientCount() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestGetUnbannedClientCount(t *testing.T) {
	tests := []struct {
		name     string
		setup    map[string]wrappedClient
		expected int
	}{
		{
			name:     "no clients",
			setup:    map[string]wrappedClient{},
			expected: 0,
		},
		{
			name: "all unbanned",
			setup: map[string]wrappedClient{
				"host1": {banned: false},
				"host2": {banned: false},
			},
			expected: 2,
		},
		{
			name: "all banned",
			setup: map[string]wrappedClient{
				"host1": {banned: true},
				"host2": {banned: true},
			},
			expected: 0,
		},
		{
			name: "mixed banned and unbanned",
			setup: map[string]wrappedClient{
				"host1": {banned: true},
				"host2": {banned: false},
				"host3": {banned: false},
			},
			expected: 2,
		},
		{
			name: "single client unbanned",
			setup: map[string]wrappedClient{
				"host1": {banned: false},
			},
			expected: 1,
		},
		{
			name: "single client banned",
			setup: map[string]wrappedClient{
				"host1": {banned: true},
			},
			expected: 0,
		},
		{
			name: "many clients",
			setup: map[string]wrappedClient{
				"host1": {banned: false},
				"host2": {banned: false},
				"host3": {banned: true},
				"host4": {banned: false},
				"host5": {banned: true},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients = tt.setup
			result := GetUnbannedClientCount()
			if result != tt.expected {
				t.Errorf("GetUnbannedClientCount() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestBannedPlusUnbannedEqualsTotal(t *testing.T) {
	clients = map[string]wrappedClient{
		"host1": {banned: true},
		"host2": {banned: false},
		"host3": {banned: true},
		"host4": {banned: false},
		"host5": {banned: false},
	}

	banned := GetBannedClientCount()
	unbanned := GetUnbannedClientCount()
	total := len(clients)

	if banned+unbanned != total {
		t.Errorf("banned(%d) + unbanned(%d) != total(%d)", banned, unbanned, total)
	}
}
