package server

import (
	"errors"
	"net/http"

	"github.com/YASSERRMD/Ventiqra/backend/internal/middleware"
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
