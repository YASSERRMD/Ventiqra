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

	"github.com/YASSERRMD/Ventiqra/backend/internal/auth"
	"github.com/YASSERRMD/Ventiqra/backend/internal/config"
	"github.com/YASSERRMD/Ventiqra/backend/internal/middleware"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
)

// Server is the Ventiqra HTTP API server.
type Server struct {
	cfg       config.Config
	log       *slog.Logger
	mux       *http.ServeMux
	server    *http.Server
	db        HealthChecker
	users     *repository.UserRepo
	tokens    *auth.TokenManager
	companies *repository.CompanyRepo
	sim       *repository.SimStateRepo
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

// WithAuth enables authentication by providing a user repository and token
// manager. When set, the auth and protected routes are registered.
func WithAuth(users *repository.UserRepo, tokens *auth.TokenManager) Option {
	return func(s *Server) {
		s.users = users
		s.tokens = tokens
	}
}

// WithCompany enables the company service by providing a CompanyRepo.
func WithCompany(companies *repository.CompanyRepo) Option {
	return func(s *Server) { s.companies = companies }
}

// WithSim enables the simulation service by providing a SimStateRepo.
func WithSim(repo *repository.SimStateRepo) Option {
	return func(s *Server) { s.sim = repo }
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

	if s.users != nil && s.tokens != nil {
		s.mux.HandleFunc("POST /api/v1/auth/register", s.handleRegister)
		s.mux.HandleFunc("POST /api/v1/auth/login", s.handleLogin)
		s.mux.Handle("GET /api/v1/me", s.protected(http.HandlerFunc(s.handleMe)))
	}

	if s.tokens != nil && s.companies != nil {
		s.mux.Handle("POST /api/v1/companies", s.protected(http.HandlerFunc(s.handleCreateCompany)))
		s.mux.Handle("GET /api/v1/companies/me", s.protected(http.HandlerFunc(s.handleMyCompany)))
		s.mux.Handle("GET /api/v1/companies/{id}", s.protected(http.HandlerFunc(s.handleGetCompany)))
	}

	if s.tokens != nil && s.companies != nil && s.sim != nil {
		s.mux.Handle("POST /api/v1/companies/me/sim/tick", s.protected(http.HandlerFunc(s.handleSimTick)))
		s.mux.Handle("GET /api/v1/companies/me/metrics", s.protected(http.HandlerFunc(s.handleMetrics)))
	}
}

// protected wraps a handler with the JWT AuthRequired middleware.
func (s *Server) protected(h http.Handler) http.Handler {
	return middleware.AuthRequired(s.tokenParser(), s.log)(h)
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
