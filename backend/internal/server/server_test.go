package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"log/slog"

	"github.com/YASSERRMD/Ventiqra/backend/internal/config"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	cfg := config.Config{Env: "test"}
	cfg.HTTP.Port = "0"
	return New(cfg, slog.Default())
}

func TestHealthEndpoint(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Status != "ok" {
		t.Errorf("status = %q, want ok", body.Status)
	}
	if body.Service != "ventiqra-api" {
		t.Errorf("service = %q, want ventiqra-api", body.Service)
	}
	if body.Version == "" {
		t.Error("version should not be empty")
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("content-type = %q, want application/json", ct)
	}
}

func TestHealthMethodNotAllowed(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestAddrUsesConfig(t *testing.T) {
	cfg := config.Config{Env: "test"}
	cfg.HTTP.Host = "127.0.0.1"
	cfg.HTTP.Port = "8081"
	srv := New(cfg, slog.Default())
	if got := srv.Addr(); got != "127.0.0.1:8081" {
		t.Errorf("Addr() = %q, want 127.0.0.1:8081", got)
	}
}
