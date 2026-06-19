package server

import (
	"errors"
	"net/http"

	"github.com/YASSERRMD/Ventiqra/backend/internal/metrics"
	"github.com/YASSERRMD/Ventiqra/backend/internal/middleware"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
	"github.com/YASSERRMD/Ventiqra/backend/internal/sim"
)

type metricsResponse struct {
	CashCents         int64   `json:"cash_cents"`
	RevenueCents      int64   `json:"revenue_cents"`
	BurnCentsPerMonth int64   `json:"burn_cents_per_month"`
	ValuationCents    int64   `json:"valuation_cents"`
	RunwayMonths      float64 `json:"runway_months"`
	Day               int     `json:"day"`
}

// handleMetrics returns the owner's latest company metrics. If no simulation
// state exists yet, it is initialized deterministically (like the tick
// endpoint) so the dashboard has meaningful numbers from day zero.
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if s.sim == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "metrics service not configured")
		return
	}

	ownerID := middleware.UserIDFrom(r.Context())
	company, err := s.companies.GetLatestCompanyForOwner(r.Context(), ownerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no company found")
			return
		}
		s.log.Error("metrics: load company failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load company")
		return
	}

	// Load existing state or initialize it deterministically from the company.
	state, err := s.sim.Get(r.Context(), company.ID)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			s.log.Error("metrics: load state failed", "error", err)
			writeError(w, http.StatusInternalServerError, "could not load simulation state")
			return
		}
		seed := sim.SeedFromCompanyID(company.ID)
		state, err = s.sim.Init(r.Context(), company.ID, seed, company.Cash, sim.BaseMonthlyBurnCents)
		if err != nil {
			s.log.Error("metrics: init state failed", "error", err)
			writeError(w, http.StatusInternalServerError, "could not initialize simulation state")
			return
		}
	}

	m := metrics.Compute(state.Cash, state.Revenue, state.MonthlyBurn, 0, state.Day)
	writeJSON(w, http.StatusOK, metricsResponse{
		CashCents:         m.CashCents,
		RevenueCents:      m.RevenueCents,
		BurnCentsPerMonth: m.BurnCentsPerMonth,
		ValuationCents:    m.ValuationCents,
		RunwayMonths:      m.RunwayMonths,
		Day:               state.Day,
	})
}
