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
	"github.com/YASSERRMD/Ventiqra/backend/internal/realtime"
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
	products  *repository.ProductRepo
	employees *repository.EmployeeRepo
	launches  *repository.LaunchRepo
	customers *repository.CustomerRepo
	pricing   *repository.PricingRepo
	finance   *repository.FinanceRepo
	funding   *repository.FundingRepo
	offers    *repository.OfferRepo
	competitors *repository.CompetitorRepo
	market      *repository.MarketRepo
	reputation  *repository.ReputationRepo
	gameEvents  *repository.GameEventRepo
	decisions   *repository.DecisionRepo
	customScenarios *repository.CustomScenarioRepo
	saveSlots   *repository.SaveSlotRepo
	timeline    *repository.TimelineRepo
	snapshots   *repository.MetricSnapshotRepo
	hub         *realtime.Hub
	simControl  *repository.SimControlRepo
	features    *repository.FeatureRepo
	techDebt    *repository.TechDebtRepo
	infrastructure *repository.InfrastructureRepo
	support     *repository.SupportRepo
	deals       *repository.DealRepo
	contracts   *repository.ContractRepo
	achievements *repository.AchievementRepo
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

// WithProducts enables the product service by providing a ProductRepo.
func WithProducts(products *repository.ProductRepo) Option {
	return func(s *Server) { s.products = products }
}

// WithEmployees enables the employee service by providing an EmployeeRepo.
func WithEmployees(employees *repository.EmployeeRepo) Option {
	return func(s *Server) { s.employees = employees }
}

// WithLaunches enables the launch service by providing a LaunchRepo.
func WithLaunches(launches *repository.LaunchRepo) Option {
	return func(s *Server) { s.launches = launches }
}

// WithCustomers enables the customer service by providing a CustomerRepo.
func WithCustomers(customers *repository.CustomerRepo) Option {
	return func(s *Server) { s.customers = customers }
}

// WithPricing enables the pricing service by providing a PricingRepo.
func WithPricing(pricing *repository.PricingRepo) Option {
	return func(s *Server) { s.pricing = pricing }
}

// WithFinance enables the finance service by providing a FinanceRepo.
func WithFinance(finance *repository.FinanceRepo) Option {
	return func(s *Server) { s.finance = finance }
}

// WithFunding enables the funding service by providing a FundingRepo.
func WithFunding(funding *repository.FundingRepo) Option {
	return func(s *Server) { s.funding = funding }
}

// WithOffers enables the investor-offer service by providing an OfferRepo.
func WithOffers(offers *repository.OfferRepo) Option {
	return func(s *Server) { s.offers = offers }
}

// WithCompetitors enables the competitor service by providing a CompetitorRepo.
func WithCompetitors(competitors *repository.CompetitorRepo) Option {
	return func(s *Server) { s.competitors = competitors }
}

// WithMarket enables the market service by providing a MarketRepo.
func WithMarket(market *repository.MarketRepo) Option {
	return func(s *Server) { s.market = market }
}

// WithReputation enables the reputation service by providing a ReputationRepo.
func WithReputation(reputation *repository.ReputationRepo) Option {
	return func(s *Server) { s.reputation = reputation }
}

// WithGameEvents enables the random-event service by providing a GameEventRepo.
func WithGameEvents(gameEvents *repository.GameEventRepo) Option {
	return func(s *Server) { s.gameEvents = gameEvents }
}

// WithDecisions enables the strategic-decision service by providing a
// DecisionRepo.
func WithDecisions(decisions *repository.DecisionRepo) Option {
	return func(s *Server) { s.decisions = decisions }
}

// WithCustomScenarios enables the custom-scenario editor by providing a
// CustomScenarioRepo.
func WithCustomScenarios(repo *repository.CustomScenarioRepo) Option {
	return func(s *Server) { s.customScenarios = repo }
}

// WithSaveSlots enables the save/load simulation service by providing a
// SaveSlotRepo.
func WithSaveSlots(repo *repository.SaveSlotRepo) Option {
	return func(s *Server) { s.saveSlots = repo }
}

// WithTimeline enables the timeline service by providing a TimelineRepo.
func WithTimeline(repo *repository.TimelineRepo) Option {
	return func(s *Server) { s.timeline = repo }
}

// WithSnapshots enables the analytics service by providing a MetricSnapshotRepo.
func WithSnapshots(repo *repository.MetricSnapshotRepo) Option {
	return func(s *Server) { s.snapshots = repo }
}

// WithHub enables WebSocket realtime updates by providing a realtime hub.
func WithHub(hub *realtime.Hub) Option {
	return func(s *Server) { s.hub = hub }
}

// WithSimControl enables simulation speed control by providing a SimControlRepo.
func WithSimControl(repo *repository.SimControlRepo) Option {
	return func(s *Server) { s.simControl = repo }
}

// WithFeatures enables the product roadmap by providing a FeatureRepo.
func WithFeatures(repo *repository.FeatureRepo) Option {
	return func(s *Server) { s.features = repo }
}

// WithTechDebt enables the technical-debt service by providing a TechDebtRepo.
func WithTechDebt(repo *repository.TechDebtRepo) Option {
	return func(s *Server) { s.techDebt = repo }
}

// WithInfrastructure enables the infrastructure scaling service.
func WithInfrastructure(repo *repository.InfrastructureRepo) Option {
	return func(s *Server) { s.infrastructure = repo }
}

// WithSupport enables the customer-support service.
func WithSupport(repo *repository.SupportRepo) Option {
	return func(s *Server) { s.support = repo }
}

// WithDeals enables the B2B sales pipeline.
func WithDeals(repo *repository.DealRepo) Option {
	return func(s *Server) { s.deals = repo }
}

// WithContracts enables enterprise recurring-revenue contracts.
func WithContracts(repo *repository.ContractRepo) Option {
	return func(s *Server) { s.contracts = repo }
}

// WithAchievements enables the achievement engine.
func WithAchievements(repo *repository.AchievementRepo) Option {
	return func(s *Server) { s.achievements = repo }
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
	if s.tokens != nil && s.companies != nil {
		s.mux.Handle("POST /api/v1/companies/me/restart", s.protected(http.HandlerFunc(s.handleRestart)))
	}

	if s.tokens != nil && s.companies != nil && s.products != nil {
		s.mux.Handle("POST /api/v1/companies/me/products", s.protected(http.HandlerFunc(s.handleCreateProduct)))
		s.mux.Handle("GET /api/v1/companies/me/products", s.protected(http.HandlerFunc(s.handleListProducts)))
		s.mux.Handle("PATCH /api/v1/products/{id}/stage", s.protected(http.HandlerFunc(s.handleUpdateProductStage)))
		s.mux.Handle("PATCH /api/v1/products/{id}/progress", s.protected(http.HandlerFunc(s.handleUpdateProductProgress)))
	}

	if s.tokens != nil && s.companies != nil && s.products != nil && s.launches != nil {
		s.mux.Handle("POST /api/v1/products/{id}/launch", s.protected(http.HandlerFunc(s.handleLaunchProduct)))
		s.mux.Handle("GET /api/v1/companies/me/launches", s.protected(http.HandlerFunc(s.handleListLaunches)))
	}

	if s.tokens != nil && s.companies != nil && s.employees != nil {
		s.mux.Handle("POST /api/v1/companies/me/employees", s.protected(http.HandlerFunc(s.handleCreateEmployee)))
		s.mux.Handle("GET /api/v1/companies/me/employees", s.protected(http.HandlerFunc(s.handleListEmployees)))
		s.mux.Handle("GET /api/v1/companies/me/morale", s.protected(http.HandlerFunc(s.handleGetMorale)))
		s.mux.Handle("PATCH /api/v1/employees/{id}/salary", s.protected(http.HandlerFunc(s.handleUpdateEmployeeSalary)))
		s.mux.Handle("PATCH /api/v1/employees/{id}/morale", s.protected(http.HandlerFunc(s.handleUpdateEmployeeMorale)))
		s.mux.Handle("DELETE /api/v1/employees/{id}", s.protected(http.HandlerFunc(s.handleDeleteEmployee)))
	}

	if s.tokens != nil && s.companies != nil && s.employees != nil {
		s.mux.Handle("GET /api/v1/companies/me/candidates", s.protected(http.HandlerFunc(s.handleListCandidates)))
		s.mux.Handle("POST /api/v1/companies/me/candidates/{index}/hire", s.protected(http.HandlerFunc(s.handleHireCandidate)))
	}

	if s.tokens != nil && s.companies != nil && s.customers != nil {
		s.mux.Handle("GET /api/v1/companies/me/customers", s.protected(http.HandlerFunc(s.handleListCustomers)))
	}

	if s.tokens != nil && s.companies != nil && s.products != nil && s.pricing != nil {
		s.mux.Handle("PATCH /api/v1/products/{id}/price", s.protected(http.HandlerFunc(s.handleSetProductPrice)))
		s.mux.Handle("GET /api/v1/companies/me/pricing-experiments", s.protected(http.HandlerFunc(s.handleListPricingExperiments)))
	}

	if s.tokens != nil && s.companies != nil && s.finance != nil {
		s.mux.Handle("GET /api/v1/companies/me/finance", s.protected(http.HandlerFunc(s.handleGetFinance)))
		s.mux.Handle("PATCH /api/v1/companies/me/finance", s.protected(http.HandlerFunc(s.handleUpdateFinance)))
		s.mux.Handle("GET /api/v1/companies/me/marketing", s.protected(http.HandlerFunc(s.handleGetMarketing)))
	}

	if s.tokens != nil && s.companies != nil && s.funding != nil {
		s.mux.Handle("GET /api/v1/companies/me/funding", s.protected(http.HandlerFunc(s.handleListFunding)))
		s.mux.Handle("POST /api/v1/companies/me/funding/raise", s.protected(http.HandlerFunc(s.handleRaiseFunding)))
	}

	if s.tokens != nil && s.companies != nil && s.offers != nil {
		s.mux.Handle("GET /api/v1/companies/me/funding/offers", s.protected(http.HandlerFunc(s.handleListOffers)))
		s.mux.Handle("POST /api/v1/companies/me/funding/offers/solicit", s.protected(http.HandlerFunc(s.handleSolicitOffers)))
		s.mux.Handle("POST /api/v1/companies/me/funding/offers/{id}/accept", s.protected(http.HandlerFunc(s.handleAcceptOffer)))
		s.mux.Handle("POST /api/v1/companies/me/funding/offers/{id}/reject", s.protected(http.HandlerFunc(s.handleRejectOffer)))
		s.mux.Handle("POST /api/v1/companies/me/funding/offers/{id}/negotiate", s.protected(http.HandlerFunc(s.handleNegotiateOffer)))
	}

	if s.tokens != nil && s.companies != nil && s.competitors != nil {
		s.mux.Handle("GET /api/v1/companies/me/competitors", s.protected(http.HandlerFunc(s.handleListCompetitors)))
	}

	if s.tokens != nil && s.companies != nil && s.market != nil {
		s.mux.Handle("GET /api/v1/companies/me/market", s.protected(http.HandlerFunc(s.handleGetMarket)))
	}

	if s.tokens != nil && s.companies != nil && s.reputation != nil {
		s.mux.Handle("GET /api/v1/companies/me/reputation", s.protected(http.HandlerFunc(s.handleGetReputation)))
	}

	if s.tokens != nil && s.companies != nil && s.gameEvents != nil {
		s.mux.Handle("GET /api/v1/companies/me/events", s.protected(http.HandlerFunc(s.handleListEvents)))
	}

	if s.tokens != nil && s.companies != nil && s.decisions != nil {
		s.mux.Handle("GET /api/v1/companies/me/decisions/pending", s.protected(http.HandlerFunc(s.handleGetPendingDecision)))
		s.mux.Handle("POST /api/v1/companies/me/decisions/{id}/resolve", s.protected(http.HandlerFunc(s.handleResolveDecision)))
	}

	if s.tokens != nil {
		s.mux.Handle("GET /api/v1/scenarios", s.protected(http.HandlerFunc(s.handleListScenarios)))
	}
	if s.tokens != nil && s.companies != nil {
		s.mux.Handle("POST /api/v1/scenarios/{id}/apply", s.protected(http.HandlerFunc(s.handleApplyScenario)))
	}

	if s.tokens != nil && s.customScenarios != nil {
		s.mux.Handle("GET /api/v1/scenarios/custom", s.protected(http.HandlerFunc(s.handleListCustomScenarios)))
		s.mux.Handle("POST /api/v1/scenarios/custom", s.protected(http.HandlerFunc(s.handleCreateCustomScenario)))
		s.mux.Handle("PATCH /api/v1/scenarios/custom/{id}", s.protected(http.HandlerFunc(s.handleUpdateCustomScenario)))
		s.mux.Handle("DELETE /api/v1/scenarios/custom/{id}", s.protected(http.HandlerFunc(s.handleDeleteCustomScenario)))
	}
	if s.tokens != nil && s.companies != nil && s.customScenarios != nil {
		s.mux.Handle("POST /api/v1/scenarios/custom/{id}/apply", s.protected(http.HandlerFunc(s.handleApplyCustomScenario)))
	}

	if s.tokens != nil && s.companies != nil && s.saveSlots != nil {
		s.mux.Handle("GET /api/v1/saves", s.protected(http.HandlerFunc(s.handleListSaveSlots)))
		s.mux.Handle("POST /api/v1/saves", s.protected(http.HandlerFunc(s.handleSaveSlot)))
		s.mux.Handle("POST /api/v1/saves/{slot}/load", s.protected(http.HandlerFunc(s.handleLoadSlot)))
		s.mux.Handle("DELETE /api/v1/saves/{slot}", s.protected(http.HandlerFunc(s.handleDeleteSlot)))
	}

	if s.tokens != nil && s.companies != nil && s.timeline != nil {
		s.mux.Handle("GET /api/v1/companies/me/timeline", s.protected(http.HandlerFunc(s.handleGetTimeline)))
	}

	if s.tokens != nil && s.companies != nil && s.snapshots != nil {
		s.mux.Handle("GET /api/v1/companies/me/analytics", s.protected(http.HandlerFunc(s.handleGetAnalytics)))
	}

	if s.tokens != nil && s.companies != nil && s.hub != nil {
		s.mux.Handle("GET /api/v1/realtime", http.HandlerFunc(s.handleWebSocket))
	}

	if s.tokens != nil && s.companies != nil && s.simControl != nil {
		s.mux.Handle("GET /api/v1/companies/me/sim/control", s.protected(http.HandlerFunc(s.handleGetSimControl)))
		s.mux.Handle("POST /api/v1/companies/me/sim/pause", s.protected(http.HandlerFunc(s.handlePauseSim)))
		s.mux.Handle("POST /api/v1/companies/me/sim/resume", s.protected(http.HandlerFunc(s.handleResumeSim)))
		s.mux.Handle("POST /api/v1/companies/me/sim/speed", s.protected(http.HandlerFunc(s.handleSetSimSpeed)))
	}

	if s.tokens != nil && s.companies != nil && s.features != nil {
		s.mux.Handle("GET /api/v1/companies/me/features", s.protected(http.HandlerFunc(s.handleListFeatures)))
		s.mux.Handle("POST /api/v1/companies/me/features", s.protected(http.HandlerFunc(s.handleCreateFeature)))
		s.mux.Handle("DELETE /api/v1/companies/me/features/{id}", s.protected(http.HandlerFunc(s.handleDeleteFeature)))
		s.mux.Handle("POST /api/v1/companies/me/features/{id}/develop", s.protected(http.HandlerFunc(s.handleDevelopFeature)))
	}

	if s.tokens != nil && s.companies != nil && s.techDebt != nil {
		s.mux.Handle("GET /api/v1/companies/me/tech-debt", s.protected(http.HandlerFunc(s.handleGetTechDebt)))
		s.mux.Handle("POST /api/v1/companies/me/tech-debt/refactor", s.protected(http.HandlerFunc(s.handleRefactor)))
	}

	if s.tokens != nil && s.companies != nil && s.infrastructure != nil {
		s.mux.Handle("GET /api/v1/companies/me/infrastructure", s.protected(http.HandlerFunc(s.handleGetInfrastructure)))
		s.mux.Handle("POST /api/v1/companies/me/infrastructure/scale", s.protected(http.HandlerFunc(s.handleScaleUp)))
	}

	if s.tokens != nil && s.companies != nil && s.support != nil {
		s.mux.Handle("GET /api/v1/companies/me/support", s.protected(http.HandlerFunc(s.handleGetSupport)))
	}

	if s.tokens != nil && s.companies != nil && s.deals != nil {
		s.mux.Handle("GET /api/v1/companies/me/deals", s.protected(http.HandlerFunc(s.handleListDeals)))
		s.mux.Handle("POST /api/v1/companies/me/deals", s.protected(http.HandlerFunc(s.handleCreateDeal)))
		s.mux.Handle("POST /api/v1/companies/me/deals/{id}/advance", s.protected(http.HandlerFunc(s.handleAdvanceDeal)))
	}

	if s.tokens != nil && s.companies != nil && s.contracts != nil {
		s.mux.Handle("GET /api/v1/companies/me/contracts", s.protected(http.HandlerFunc(s.handleListContracts)))
		s.mux.Handle("POST /api/v1/companies/me/contracts", s.protected(http.HandlerFunc(s.handleSignContract)))
	}

	if s.tokens != nil && s.companies != nil && s.achievements != nil {
		s.mux.Handle("GET /api/v1/companies/me/achievements", s.protected(http.HandlerFunc(s.handleListAchievements)))
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
