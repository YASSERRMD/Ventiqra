// Package server provides the HTTP API server, routing, and handlers for the
// Ventiqra backend.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"log/slog"

	"github.com/YASSERRMD/Ventiqra/backend/internal/config"
	"github.com/YASSERRMD/Ventiqra/backend/internal/middleware"
)

// Server is the Ventiqra HTTP API server.
type Server struct {
	cfg    config.Config
	log    *slog.Logger
	mux    *http.ServeMux
	server *http.Server
	db     HealthChecker
}

// HealthChecker is anything that can report its own health via Ping.
// A *pgxpool.Pool satisfies this interface.
type HealthChecker interface {
	Ping(ctx context.Context) error
}

// Option configures a Server at construction time.
type Option func(*Server)

// WithDB attaches a database health checker to the server so the health
// endpoint can report database reachability.
func WithDB(db HealthChecker) Option {
	return func(s *Server) { s.db = db }
}

// New constructs a Server with routes registered.
func New(cfg config.Config, log *slog.Logger, opts ...Option) *Server {
	s := &Server{
		cfg: cfg,
		log: log,
		mux: http.NewServeMux(),
	}
	for _, opt := range opts {
		opt(s)
	}
	s.registerRoutes()
	s.server = &http.Server{
		Addr: s.cfg.HTTP.Host + ":" + s.cfg.HTTP.Port,
		Handler: middleware.Chain(
			s.mux,
			middleware.Recover(log),
			middleware.RequestID,
			middleware.Logger(log),
			middleware.CORS(cfg.CORSOrigins),
		),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	return s
}

// Addr returns the configured listen address.
func (s *Server) Addr() string { return s.server.Addr }

// Handler returns the root HTTP handler (useful for testing).
func (s *Server) Handler() http.Handler { return s.mux }

// registerRoutes wires all API routes.
func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealth)
}

// ListenAndServe starts the HTTP server. It blocks until the server is shut
// down or returns an error other than http.ErrServerClosed.
func (s *Server) ListenAndServe() error {
	s.log.Info("http server listening", "addr", s.Addr())
	err := s.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown gracefully stops the server, waiting for in-flight requests.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// healthResponse is the body returned by the health endpoint.
type healthResponse struct {
	Status    string         `json:"status"`
	Service   string         `json:"service"`
	Env       string         `json:"env"`
	Version   string         `json:"version,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	Checks    map[string]any `json:"checks,omitempty"`
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	httpStatus := http.StatusOK
	checks := map[string]any{}

	if s.db != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		start := time.Now()
		err := s.db.Ping(ctx)
		latency := time.Since(start)

		dbCheck := map[string]any{"latency_ms": latency.Milliseconds()}
		if err != nil {
			dbCheck["status"] = "down"
			dbCheck["error"] = err.Error()
			status = "degraded"
			httpStatus = http.StatusServiceUnavailable
			s.log.Warn("health check: database down", "error", err)
		} else {
			dbCheck["status"] = "up"
		}
		checks["database"] = dbCheck
	}

	writeJSON(w, httpStatus, healthResponse{
		Status:    status,
		Service:   "ventiqra-api",
		Env:       s.cfg.Env,
		Version:   APIVersion,
		Timestamp: time.Now().UTC(),
		Checks:    checks,
	})
}

// APIVersion is the current API version reported by health and other endpoints.
const APIVersion = "0.1.0"

// writeJSON encodes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
