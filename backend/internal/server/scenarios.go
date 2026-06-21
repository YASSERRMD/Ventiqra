// Scenario listing and application handlers. This file exposes the predefined
// scenario catalog and an endpoint to apply a scenario to the owner's company,
// setting its cash, industry, and starting market configuration.
package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/YASSERRMD/Ventiqra/backend/internal/middleware"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
	"github.com/YASSERRMD/Ventiqra/backend/internal/scenarios"
)

type scenarioMarketResponse struct {
	TAM             int     `json:"tam"`
	GrowthRate      float64 `json:"growth_rate"`
	TrendMultiplier float64 `json:"trend_multiplier"`
}

type scenarioResponse struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	Category          string                 `json:"category"`
	Description       string                 `json:"description"`
	Difficulty        string                 `json:"difficulty"`
	Industry          string                 `json:"industry"`
	StartingCashCents int64                  `json:"starting_cash_cents"`
	StartingBurnCents int64                  `json:"starting_burn_cents"`
	Market            scenarioMarketResponse `json:"market"`
}

type applyScenarioResponse struct {
	Scenario  scenarioResponse `json:"scenario"`
	Company   companyResponse  `json:"company"`
	AppliedAt time.Time        `json:"applied_at"`
}

func toScenarioResponse(s scenarios.Scenario) scenarioResponse {
	return scenarioResponse{
		ID: s.ID, Name: s.Name, Category: s.Category, Description: s.Description,
		Difficulty: string(s.Difficulty), Industry: s.Industry,
		StartingCashCents: s.StartingCashCents, StartingBurnCents: s.StartingBurnCents,
		Market: scenarioMarketResponse{
			TAM: s.Market.TAM, GrowthRate: s.Market.GrowthRate, TrendMultiplier: s.Market.TrendMultiplier,
		},
	}
}

// handleListScenarios returns the predefined scenario catalog. The catalog is
// static so this is a simple read with no company required; it is still placed
// behind auth so only authenticated users see the playable scenarios.
func (s *Server) handleListScenarios(w http.ResponseWriter, r *http.Request) {
	out := make([]scenarioResponse, 0, len(scenarios.Catalog))
	for _, sc := range scenarios.Catalog {
		out = append(out, toScenarioResponse(sc))
	}
	writeJSON(w, http.StatusOK, out)
}

// handleApplyScenario applies a predefined scenario to the owner's latest
// company: sets its cash and industry and seeds its market configuration.
// Returns 404 if the owner has no company or the scenario id is unknown.
func (s *Server) handleApplyScenario(w http.ResponseWriter, r *http.Request) {
	if s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "company service not configured")
		return
	}
	scenarioID := r.PathValue("id")
	sc, ok := scenarios.Find(scenarioID)
	if !ok {
		writeError(w, http.StatusNotFound, "unknown scenario")
		return
	}

	ownerID := middleware.UserIDFrom(r.Context())
	company, err := s.companies.GetLatestCompanyForOwner(r.Context(), ownerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no company found; create a company first")
			return
		}
		s.log.Error("apply scenario: load company failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load company")
		return
	}
	if company.Status == repository.CompanyBankrupt {
		writeError(w, http.StatusConflict, "bankrupt companies cannot apply a scenario; restart first")
		return
	}

	// Apply cash + industry from the scenario.
	if err := s.companies.UpdateProfile(r.Context(), company.ID, sc.StartingCashCents, sc.Industry); err != nil {
		s.log.Error("apply scenario: update profile failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not apply scenario")
		return
	}

	// Seed the market configuration if the market service is configured.
	if s.market != nil {
		if _, err := s.market.GetOrCreate(r.Context(), company.ID); err == nil {
			if err := s.market.Save(r.Context(), company.ID, int64(sc.Market.TAM), sc.Market.GrowthRate, sc.Market.TrendMultiplier); err != nil {
				s.log.Error("apply scenario: seed market failed", "error", err)
			}
		}
	}

	updated, err := s.companies.GetCompany(r.Context(), company.ID)
	if err != nil {
		updated = company
	}
	writeJSON(w, http.StatusOK, applyScenarioResponse{
		Scenario:  toScenarioResponse(sc),
		Company:   toCompanyResponse(updated),
		AppliedAt: time.Now().UTC(),
	})
}
