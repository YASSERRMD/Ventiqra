package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/YASSERRMD/Ventiqra/backend/internal/customers"
	"github.com/YASSERRMD/Ventiqra/backend/internal/develop"
	"github.com/YASSERRMD/Ventiqra/backend/internal/finance"
	"github.com/YASSERRMD/Ventiqra/backend/internal/leaderboard"
	"github.com/YASSERRMD/Ventiqra/backend/internal/marketing"
	metricsModule "github.com/YASSERRMD/Ventiqra/backend/internal/metrics"
	"github.com/YASSERRMD/Ventiqra/backend/internal/middleware"
	"github.com/YASSERRMD/Ventiqra/backend/internal/pricing"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
	"github.com/YASSERRMD/Ventiqra/backend/internal/sim"
)

type simTickResponse struct {
	CompanyID string `json:"company_id"`
	Day       int    `json:"day"`
	Seed      int64  `json:"seed"`
	CashCents int64  `json:"cash_cents"`
	Status    string `json:"status"`
	Health    string `json:"health"`
}

// handleSimTick advances the owner's latest company simulation by exactly one
// day, persisting the resulting state and mirroring cash back onto the company.
func (s *Server) handleSimTick(w http.ResponseWriter, r *http.Request) {
	if s.sim == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "simulation service not configured")
		return
	}

	ownerID := middleware.UserIDFrom(r.Context())
	company, err := s.companies.GetLatestCompanyForOwner(r.Context(), ownerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no company found")
			return
		}
		s.log.Error("sim tick: load company failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load company")
		return
	}

	// Game over: a bankrupt company cannot advance the simulation.
	if company.Status == repository.CompanyBankrupt {
		writeError(w, http.StatusConflict, "company is bankrupt; restart to play again")
		return
	}

	// Load existing state or initialize it deterministically from the company.
	state, err := s.sim.Get(r.Context(), company.ID)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			s.log.Error("sim tick: load state failed", "error", err)
			writeError(w, http.StatusInternalServerError, "could not load simulation state")
			return
		}
		seed := sim.SeedFromCompanyID(company.ID)
		state, err = s.sim.Init(r.Context(), company.ID, seed, company.Cash, sim.BaseMonthlyBurnCents)
		if err != nil {
			s.log.Error("sim tick: init state failed", "error", err)
			writeError(w, http.StatusInternalServerError, "could not initialize simulation state")
			return
		}
	}

	// Build the in-memory sim state from persisted values and advance one day.
	engine := sim.NewEngine(state.Seed)
	simState := &sim.State{
		CompanyID:   state.CompanyID,
		Day:         state.Day,
		Cash:        state.Cash,
		Revenue:     state.Revenue,
		MonthlyBurn: state.MonthlyBurn,
		Seed:        state.Seed,
		Rand:        sim.NewRand(state.Seed, state.Day),
	}

	// Apply employee-driven development for the day.
	roster, salaries := s.loadRoster(r.Context(), company.ID)
	if len(roster) > 0 {
		s.advanceBuildingProducts(r.Context(), company.ID, develop.DailyProgress(roster))
	}

	// Morale decays and burnt-out employees may resign.
	s.advanceMorale(r.Context(), company.ID, state.Seed, state.Day+1)

	// Rivals evolve and exert market pressure that dampens acquisition.
	s.ensureCompetitors(r.Context(), company.ID)
	pressure := s.advanceCompetitors(r.Context(), company.ID, state.Seed, state.Day+1)

	// The market grows and its trend multiplier amplifies or dampens demand.
	trend := s.advanceMarket(r.Context(), company.ID, state.Seed, state.Day+1)

	// Marketing budget produces deterministic customer conversions.
	mktBudget := s.marketingBudget(r.Context(), company.ID)
	mktConversions := marketing.Conversions(mktBudget, state.Seed, int64(state.Day+1))

	// Reputation drifts with customer satisfaction and company health; its growth
	// multiplier amplifies or dampens acquisition.
	avgSat := s.averageSatisfaction(r.Context(), company.ID)
	repGrowth := s.applyReputationDrift(r.Context(), company.ID, avgSat, metricsModule.Health(simState.Cash, metricsModule.Compute(simState.Cash, simState.Revenue, simState.MonthlyBurn, 0, simState.Day).RunwayMonths), state.Day+1)

	// Advance customer dynamics (acquisition/churn/MAU/satisfaction) for every
	// launched product, deterministically for the round. Pricing feeds demand and
	// the daily revenue total feeds the engine.
	dailyRevenue, totalCustomers := s.advanceCustomers(r.Context(), company.ID, state.Seed, state.Day+1, pressure, trend, repGrowth, mktConversions)

	// Compute the full monthly burn from the finance breakdown: base overhead,
	// payroll, infrastructure (scaled by customers), and marketing budget.
	payroll := develop.MonthlyPayroll(salaries)
	marketing := s.marketingBudget(r.Context(), company.ID)
	simState.MonthlyBurn = finance.MonthlyBreakdown(payroll, marketing, totalCustomers).Total()
	simState.Revenue = dailyRevenue

	engine.Tick(simState)

	// Roll for a daily random event; apply its effects if one fires.
	simState.Cash = s.rollAndApplyEvent(r.Context(), company.ID, simState.Cash, state.Seed, simState.Day)

	// Apply any active long-term decision commitments for the day, then offer a
	// fresh decision card on the cadence if none is pending.
	s.applyActiveDecisionEffects(r.Context(), company.ID, &simState.Cash, simState.Day)
	s.maybeOfferDecision(r.Context(), company.ID, state.Seed, simState.Day)

	if err := s.sim.Save(r.Context(), company.ID, simState.Day, simState.Cash, simState.Revenue, simState.MonthlyBurn); err != nil {
		s.log.Error("sim tick: save state failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not save simulation state")
		return
	}

	// Capture the day's headline metrics for the analytics dashboard.
	s.recordSnapshot(r.Context(), company.ID, simState.Day, simState.Cash, simState.Revenue, simState.MonthlyBurn, totalCustomers)

	// Advance the customer-support ticket backlog for the day.
	s.advanceSupport(r.Context(), company.ID, totalCustomers)

	// Accrue enterprise-contract revenue and roll renewals.
	s.advanceContracts(r.Context(), company.ID, &simState.Cash)

	// Evaluate and award achievements against the new state.
	s.evaluateAchievements(r.Context(), company.ID, simState.Day)

	// Broadcast the latest state to live dashboard subscribers.
	s.broadcastTick(company.ID, simState.Day, simState.Cash, simState.Revenue, simState.MonthlyBurn, totalCustomers)

	if err := s.companies.UpdateCash(r.Context(), company.ID, simState.Cash); err != nil {
		s.log.Error("sim tick: update company cash failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not update company cash")
		return
	}

	// Bankruptcy detection: once cash goes negative the game is over for this
	// company until the owner restarts.
	status := company.Status
	if simState.Cash < 0 {
		status = repository.CompanyBankrupt
		if err := s.companies.UpdateStatus(r.Context(), company.ID, status); err != nil {
			s.log.Error("sim tick: mark bankrupt failed", "error", err)
		}
		// Finalize the run on the local leaderboard.
		bankruptCompany := company
		bankruptCompany.Status = status
		s.finalizeRun(r.Context(), bankruptCompany, leaderboard.OutcomeBankrupt)
	}

	runway := metricsModule.Compute(simState.Cash, simState.Revenue, simState.MonthlyBurn, 0, simState.Day).RunwayMonths
	writeJSON(w, http.StatusOK, simTickResponse{
		CompanyID: company.ID,
		Day:       simState.Day,
		Seed:      simState.Seed,
		CashCents: simState.Cash,
		Status:    string(status),
		Health:    metricsModule.Health(simState.Cash, runway),
	})
}

// loadRoster returns the company's employees as develop inputs together with
// their monthly salaries (cents). Both slices are nil when the employee service
// is not configured or the company has no team yet.
func (s *Server) loadRoster(ctx context.Context, companyID string) ([]develop.Employee, []int64) {
	if s.employees == nil {
		return nil, nil
	}
	team, err := s.employees.ListEmployeesByCompany(ctx, companyID)
	if err != nil || len(team) == 0 {
		return nil, nil
	}
	roster := make([]develop.Employee, 0, len(team))
	salaries := make([]int64, 0, len(team))
	for _, e := range team {
		roster = append(roster, develop.Employee{Role: string(e.Role), Skill: e.Skill, Morale: e.Morale})
		salaries = append(salaries, e.SalaryCents)
	}
	return roster, salaries
}

// advanceBuildingProducts adds one day's worth of development progress to every
// product currently in the 'building' stage, persisting the clamped result.
func (s *Server) advanceBuildingProducts(ctx context.Context, companyID string, daily float64) {
	if s.products == nil || daily <= 0 {
		return
	}
	products, err := s.products.ListProductsByCompany(ctx, companyID)
	if err != nil {
		return
	}
	for _, p := range products {
		if p.Stage != repository.ProductBuilding {
			continue
		}
		next := p.DevProgress + daily
		if next > 100 {
			next = 100
		}
		if err := s.products.UpdateProgress(ctx, p.ID, next); err != nil {
			s.log.Error("advance product progress failed", "product", p.ID, "error", err)
		}
	}
}

// advanceCustomers advances acquisition/churn/MAU/satisfaction for every
// launched product that has customer state, deterministically for the round.
// `pressure` (0..0.5) from competitors dampens acquisition demand. Returns the
// total daily revenue (in cents) and the total customer count across launched
// products, so the finance engine can compute burn/P&L.
func (s *Server) advanceCustomers(ctx context.Context, companyID string, seed int64, day int, pressure, trend, reputationMul float64, marketingConversions int) (int64, int) {
	if s.customers == nil || s.products == nil {
		return 0, 0
	}
	list, err := s.customers.ListByCompany(ctx, companyID)
	if err != nil || len(list) == 0 {
		return 0, 0
	}
	products, _ := s.products.ListProductsByCompany(ctx, companyID)
	launched := make(map[string]*repository.Product, len(products))
	for _, p := range products {
		if p.Stage == repository.ProductLaunched {
			launched[p.ID] = p
		}
	}

	if len(launched) == 0 {
		return 0, 0
	}
	// Marketing conversions are company-wide; split them across launched products.
	perProduct := marketingConversions / len(launched)

	var dailyRevenue int64
	var totalCustomers int
	for _, c := range list {
		p, ok := launched[c.ProductID]
		if !ok {
			continue
		}
		var price int64
		if p.PriceCents != nil {
			price = *p.PriceCents
		}
		demand := pricing.DemandMultiplier(price, pricing.BaselineMonthlyCents, pricing.DefaultElasticity)
		if pressure > 0 {
			demand *= (1 - pressure)
		}
		if trend > 0 {
			demand *= trend
		}
		if reputationMul > 0 {
			demand *= reputationMul
		}
		next := customers.Advance(customers.Product{
			Total: c.Total, MAU: c.MAU, Churned: c.Churned, Satisfaction: c.Satisfaction,
		}, seed, day, demand)
		// Top up with marketing-driven acquisitions.
		if perProduct > 0 {
			next.Total += perProduct
			next.MAU = int(float64(next.Total) * customers.MauRatio(next.Satisfaction))
		}
		if err := s.customers.Save(ctx, c.ProductID, next.Total, next.MAU, next.Churned, next.Satisfaction); err != nil {
			s.log.Error("advance customers failed", "product", c.ProductID, "error", err)
		}
		dailyRevenue += pricing.DailyRevenueCents(next.Total, price)
		totalCustomers += next.Total
	}
	return dailyRevenue, totalCustomers
}

// marketingBudget returns the company's monthly marketing budget, or 0 when the
// finance service is unavailable.
func (s *Server) marketingBudget(ctx context.Context, companyID string) int64 {
	if s.finance == nil {
		return 0
	}
	if fin, err := s.finance.GetOrCreate(ctx, companyID); err == nil {
		return fin.MarketingBudgetCents
	}
	return 0
}

// averageSatisfaction returns the mean customer satisfaction across the
// company's launched-product customer states, or 70 when none exist.
func (s *Server) averageSatisfaction(ctx context.Context, companyID string) int {
	if s.customers == nil {
		return 70
	}
	list, err := s.customers.ListByCompany(ctx, companyID)
	if err != nil || len(list) == 0 {
		return 70
	}
	var sum int
	for _, c := range list {
		sum += c.Satisfaction
	}
	return sum / len(list)
}
