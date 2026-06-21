package server

import (
	"errors"
	"net/http"

	"github.com/YASSERRMD/Ventiqra/backend/internal/middleware"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
	"github.com/YASSERRMD/Ventiqra/backend/internal/sim"
)

type restartResponse struct {
	CompanyID string `json:"company_id"`
	Day       int    `json:"day"`
	CashCents int64  `json:"cash_cents"`
	Status    string `json:"status"`
}

// handleRestart resets the owner's latest company to a fresh starting state:
// status back to active, cash restored to the default starting capital, and the
// simulation rewound to day zero. Employees, products, and customer state are
// preserved so the player keeps their built-up world.
func (s *Server) handleRestart(w http.ResponseWriter, r *http.Request) {
	if s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "company service not configured")
		return
	}
	ownerID := middleware.UserIDFrom(r.Context())
	company, err := s.companies.GetLatestCompanyForOwner(r.Context(), ownerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no company found")
			return
		}
		s.log.Error("restart: load company failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load company")
		return
	}

	if err := s.companies.UpdateStatus(r.Context(), company.ID, repository.CompanyActive); err != nil {
		s.log.Error("restart: set status failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not restart")
		return
	}
	if err := s.companies.UpdateCash(r.Context(), company.ID, defaultStartingCashCents); err != nil {
		s.log.Error("restart: set cash failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not restart")
		return
	}

	// Rewind the simulation state to day zero with the restored cash. Save
	// updates an existing row; if none exists yet, initialize it fresh.
	if s.sim != nil {
		if err := s.sim.Save(r.Context(), company.ID, 0, defaultStartingCashCents, 0, sim.BaseMonthlyBurnCents); err != nil {
			seed := sim.SeedFromCompanyID(company.ID)
			_, _ = s.sim.Init(r.Context(), company.ID, seed, defaultStartingCashCents, sim.BaseMonthlyBurnCents)
		}
	}

	writeJSON(w, http.StatusOK, restartResponse{
		CompanyID: company.ID,
		Day:       0,
		CashCents: defaultStartingCashCents,
		Status:    string(repository.CompanyActive),
	})
}
