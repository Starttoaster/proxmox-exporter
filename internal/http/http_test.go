package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/starttoaster/proxmox-exporter/internal/logger"
)

func init() {
	log.Init("error")
}

func TestNewServer(t *testing.T) {
	tests := []struct {
		name         string
		addr         string
		port         uint16
		expectedAddr string
		expectedPort uint16
	}{
		{"default values", "0.0.0.0", 8080, "0.0.0.0", 8080},
		{"localhost", "127.0.0.1", 9090, "127.0.0.1", 9090},
		{"empty addr", "", 8080, "", 8080},
		{"high port", "0.0.0.0", 65535, "0.0.0.0", 65535},
		{"low port", "0.0.0.0", 1, "0.0.0.0", 1},
		{"ipv6 localhost", "::1", 8080, "::1", 8080},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewServer(tt.addr, tt.port)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if s == nil {
				t.Fatal("expected non-nil server")
			}
			if s.addr != tt.expectedAddr {
				t.Errorf("expected addr=%s, got %s", tt.expectedAddr, s.addr)
			}
			if s.port != tt.expectedPort {
				t.Errorf("expected port=%d, got %d", tt.expectedPort, s.port)
			}
		})
	}
}

func TestNewServer_ReturnsNoError(t *testing.T) {
	s, err := NewServer("0.0.0.0", 8080)
	if err != nil {
		t.Fatalf("NewServer should not return error, got: %v", err)
	}
	if s == nil {
		t.Fatal("NewServer should return a non-nil server")
	}
}

func TestHealthcheck(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	healthcheck(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %q", w.Body.String())
	}
}

func TestHealthcheck_ResponseHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	healthcheck(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHealthcheck_MultipleRequests(t *testing.T) {
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		w := httptest.NewRecorder()

		healthcheck(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected status 200, got %d", i, w.Code)
		}
		if w.Body.String() != "ok" {
			t.Errorf("request %d: expected body 'ok', got %q", i, w.Body.String())
		}
	}
}
