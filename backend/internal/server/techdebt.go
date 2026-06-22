// Technical-debt handlers: read the debt/quality state and perform a refactor
// action that pays debt down for cash. Debt accumulates when features ship.
package server

import (
	"context"
	"net/http"

	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
	"github.com/YASSERRMD/Ventiqra/backend/internal/techdebt"
)

type techDebtResponse struct {
	Debt     int `json:"debt"`
	Quality  int `json:"quality"`
	Risk     float64 `json:"outage_risk"`
	Refactors int `json:"refactors"`
}

func toTechDebtResponse(td *repository.TechDebt) techDebtResponse {
	return techDebtResponse{
		Debt: td.Debt, Quality: techdebt.Quality(td.Debt),
		Risk: techdebt.OutageRisk(td.Debt), Refactors: td.Refactors,
	}
}

func (s *Server) handleGetTechDebt(w http.ResponseWriter, r *http.Request) {
	if s.techDebt == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "tech debt service not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	td, err := s.techDebt.GetOrCreate(r.Context(), companyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load tech debt")
		return
	}
	writeJSON(w, http.StatusOK, toTechDebtResponse(td))
}

func (s *Server) handleRefactor(w http.ResponseWriter, r *http.Request) {
	if s.techDebt == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "tech debt service not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	company, err := s.companies.GetCompany(r.Context(), companyID)
	if err != nil {
		writeError(w, http.StatusNotFound, "company not found")
		return
	}
	if company.Status == repository.CompanyBankrupt {
		writeError(w, http.StatusConflict, "bankrupt companies cannot refactor")
		return
	}
	td, err := s.techDebt.GetOrCreate(r.Context(), companyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load tech debt")
		return
	}
	if td.Debt == 0 {
		writeError(w, http.StatusConflict, "no debt to refactor")
		return
	}
	// Pay the refactor cost.
	newCash := company.Cash - techdebt.RefactorCostCents
	if err := s.companies.UpdateCash(r.Context(), companyID, newCash); err != nil {
		writeError(w, http.StatusInternalServerError, "could not charge refactor")
		return
	}
	newDebt := techdebt.Refactor(td.Debt)
	day := s.currentSimDay(r.Context(), companyID)
	if err := s.techDebt.RecordRefactor(r.Context(), companyID, newDebt, day); err != nil {
		writeError(w, http.StatusInternalServerError, "could not record refactor")
		return
	}
	s.recordTimeline(r.Context(), companyID, "milestone", "Refactored codebase",
		"Debt reduced to "+formatInt(newDebt), day)
	updated, _ := s.techDebt.GetOrCreate(r.Context(), companyID)
	writeJSON(w, http.StatusOK, toTechDebtResponse(updated))
}

// accumulateDebtOnShip adds debt when a feature ships. Called from the roadmap
// develop handler when a feature crosses 100%.
func (s *Server) accumulateDebtOnShip(ctx context.Context, companyID string) {
	if s.techDebt == nil {
		return
	}
	td, err := s.techDebt.GetOrCreate(ctx, companyID)
	if err != nil {
		return
	}
	newDebt := techdebt.Accumulate(td.Debt)
	_ = s.techDebt.SetDebt(ctx, companyID, newDebt)
}
