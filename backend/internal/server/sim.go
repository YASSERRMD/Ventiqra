package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/YASSERRMD/Ventiqra/backend/internal/customers"
	"github.com/YASSERRMD/Ventiqra/backend/internal/develop"
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

	// Apply employee-driven effects for the day: payroll raises burn and builders
	// advance any in-progress products. Both are nil-guarded so the simulation
	// still works before the employee/product services are configured.
	roster, salaries := s.loadRoster(r.Context(), company.ID)
	if len(roster) > 0 {
		simState.MonthlyBurn = sim.BaseMonthlyBurnCents + develop.MonthlyPayroll(salaries)
		s.advanceBuildingProducts(r.Context(), company.ID, develop.DailyProgress(roster))
	}

	// Advance customer dynamics (acquisition/churn/MAU/satisfaction) for every
	// launched product, deterministically for the round. Pricing feeds demand and
	// the daily revenue total is returned to the engine.
	simState.Revenue = s.advanceCustomers(r.Context(), company.ID, state.Seed, state.Day+1)

	engine.Tick(simState)

	if err := s.sim.Save(r.Context(), company.ID, simState.Day, simState.Cash, simState.Revenue, simState.MonthlyBurn); err != nil {
		s.log.Error("sim tick: save state failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not save simulation state")
		return
	}

	if err := s.companies.UpdateCash(r.Context(), company.ID, simState.Cash); err != nil {
		s.log.Error("sim tick: update company cash failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not update company cash")
		return
	}

	writeJSON(w, http.StatusOK, simTickResponse{
		CompanyID: company.ID,
		Day:       simState.Day,
		Seed:      simState.Seed,
		CashCents: simState.Cash,
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
// Pricing drives per-product demand and the returned value is the total daily
// revenue (in cents) produced by the paying customer base.
func (s *Server) advanceCustomers(ctx context.Context, companyID string, seed int64, day int) int64 {
	if s.customers == nil || s.products == nil {
		return 0
	}
	list, err := s.customers.ListByCompany(ctx, companyID)
	if err != nil || len(list) == 0 {
		return 0
	}
	products, _ := s.products.ListProductsByCompany(ctx, companyID)
	launched := make(map[string]*repository.Product, len(products))
	for _, p := range products {
		if p.Stage == repository.ProductLaunched {
			launched[p.ID] = p
		}
	}

	var dailyRevenue int64
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
		next := customers.Advance(customers.Product{
			Total: c.Total, MAU: c.MAU, Churned: c.Churned, Satisfaction: c.Satisfaction,
		}, seed, day, demand)
		if err := s.customers.Save(ctx, c.ProductID, next.Total, next.MAU, next.Churned, next.Satisfaction); err != nil {
			s.log.Error("advance customers failed", "product", c.ProductID, "error", err)
		}
		dailyRevenue += pricing.DailyRevenueCents(next.Total, price)
	}
	return dailyRevenue
}
