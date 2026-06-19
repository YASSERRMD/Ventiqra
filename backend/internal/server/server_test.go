package server

import (
	"context"
	"encoding/json"
	"errors"
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

type stubChecker struct{ err error }

func (s stubChecker) Ping(ctx context.Context) error { return s.err }

func TestHealthWithoutDB(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body healthResponse
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body.Status != "ok" {
		t.Errorf("status = %q, want ok", body.Status)
	}
	if len(body.Checks) != 0 {
		t.Errorf("expected no checks without DB, got %v", body.Checks)
	}
}

func TestHealthWithDBUp(t *testing.T) {
	cfg := config.Config{Env: "test"}
	srv := New(cfg, slog.Default(), WithDB(stubChecker{err: nil}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body healthResponse
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body.Status != "ok" {
		t.Errorf("status = %q, want ok", body.Status)
	}
	db, ok := body.Checks["database"].(map[string]any)
	if !ok {
		t.Fatalf("missing database check: %v", body.Checks)
	}
	if db["status"] != "up" {
		t.Errorf("db status = %v, want up", db["status"])
	}
}

func TestHealthWithDBDown(t *testing.T) {
	cfg := config.Config{Env: "test"}
	srv := New(cfg, slog.Default(), WithDB(stubChecker{err: errors.New("no connection")}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	var body healthResponse
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body.Status != "degraded" {
		t.Errorf("status = %q, want degraded", body.Status)
	}
	db, ok := body.Checks["database"].(map[string]any)
	if !ok || db["status"] != "down" {
		t.Errorf("expected db down, got %v", body.Checks)
	}
}
