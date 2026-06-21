// Custom-scenario CRUD handlers. Users can create, list, update, delete, and
// apply their own authored scenarios, mirroring the predefined catalog but
// persisted per owner.
package server

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/YASSERRMD/Ventiqra/backend/internal/middleware"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
	"github.com/YASSERRMD/Ventiqra/backend/internal/scenarios"
)

type customScenarioResponse struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	Difficulty        string                 `json:"difficulty"`
	Industry          string                 `json:"industry"`
	StartingCashCents int64                  `json:"starting_cash_cents"`
	StartingBurnCents int64                  `json:"starting_burn_cents"`
	Market            scenarioMarketResponse `json:"market"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

type customScenarioInput struct {
	Name              *string  `json:"name"`
	Description       *string  `json:"description"`
	Difficulty        *string  `json:"difficulty"`
	Industry          *string  `json:"industry"`
	StartingCashCents *int64   `json:"starting_cash_cents"`
	StartingBurnCents *int64   `json:"starting_burn_cents"`
	MarketTAM         *int     `json:"market_tam"`
	MarketGrowthRate  *float64 `json:"market_growth_rate"`
	MarketTrend       *float64 `json:"market_trend"`
}

func toCustomScenarioResponse(c *repository.CustomScenario) customScenarioResponse {
	return customScenarioResponse{
		ID: c.ID, Name: c.Name, Description: c.Description,
		Difficulty: c.Difficulty, Industry: c.Industry,
		StartingCashCents: c.StartingCashCents, StartingBurnCents: c.StartingBurnCents,
		Market: scenarioMarketResponse{TAM: c.MarketTAM, GrowthRate: c.MarketGrowthRate, TrendMultiplier: c.MarketTrend},
		CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}

func buildCustomInput(req customScenarioInput) (scenarios.CustomInput, error) {
	if req.Name == nil {
		return scenarios.CustomInput{}, scenarios.ErrNameLength
	}
	in := scenarios.CustomInput{Name: *req.Name}
	if req.Description != nil {
		in.Description = *req.Description
	}
	if req.Difficulty != nil {
		in.Difficulty = scenarios.Difficulty(strings.TrimSpace(*req.Difficulty))
	}
	if req.Industry != nil {
		in.Industry = *req.Industry
	}
	if req.StartingCashCents != nil {
		in.StartingCashCents = *req.StartingCashCents
	}
	if req.StartingBurnCents != nil {
		in.StartingBurnCents = *req.StartingBurnCents
	}
	if req.MarketTAM != nil {
		in.MarketTAM = *req.MarketTAM
	}
	if req.MarketGrowthRate != nil {
		in.MarketGrowthRate = *req.MarketGrowthRate
	}
	if req.MarketTrend != nil {
		in.MarketTrend = *req.MarketTrend
	}
	if err := in.Validate(); err != nil {
		return in, err
	}
	return in, nil
}

func (s *Server) handleListCustomScenarios(w http.ResponseWriter, r *http.Request) {
	if s.customScenarios == nil {
		writeError(w, http.StatusServiceUnavailable, "custom scenario service not configured")
		return
	}
	ownerID := middleware.UserIDFrom(r.Context())
	list, err := s.customScenarios.ListByOwner(r.Context(), ownerID)
	if err != nil {
		s.log.Error("list custom scenarios failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load custom scenarios")
		return
	}
	out := make([]customScenarioResponse, 0, len(list))
	for _, c := range list {
		out = append(out, toCustomScenarioResponse(c))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleCreateCustomScenario(w http.ResponseWriter, r *http.Request) {
	if s.customScenarios == nil {
		writeError(w, http.StatusServiceUnavailable, "custom scenario service not configured")
		return
	}
	var req customScenarioInput
	if !decodeJSON(w, r, &req) {
		return
	}
	in, err := buildCustomInput(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	ownerID := middleware.UserIDFrom(r.Context())
	created, err := s.customScenarios.Create(r.Context(), &repository.CustomScenario{
		OwnerID: ownerID,
		Name: in.Name, Description: in.Description, Difficulty: string(in.Difficulty),
		Industry: in.Industry, StartingCashCents: in.StartingCashCents, StartingBurnCents: in.StartingBurnCents,
		MarketTAM: in.MarketTAM, MarketGrowthRate: in.MarketGrowthRate, MarketTrend: in.MarketTrend,
	})
	if err != nil {
		s.log.Error("create custom scenario failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not save scenario")
		return
	}
	writeJSON(w, http.StatusCreated, toCustomScenarioResponse(created))
}

func (s *Server) handleUpdateCustomScenario(w http.ResponseWriter, r *http.Request) {
	if s.customScenarios == nil {
		writeError(w, http.StatusServiceUnavailable, "custom scenario service not configured")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "scenario id is required")
		return
	}
	var req customScenarioInput
	if !decodeJSON(w, r, &req) {
		return
	}
	in, err := buildCustomInput(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	ownerID := middleware.UserIDFrom(r.Context())
	updated, err := s.customScenarios.Update(r.Context(), id, ownerID, &repository.CustomScenario{
		Name: in.Name, Description: in.Description, Difficulty: string(in.Difficulty),
		Industry: in.Industry, StartingCashCents: in.StartingCashCents, StartingBurnCents: in.StartingBurnCents,
		MarketTAM: in.MarketTAM, MarketGrowthRate: in.MarketGrowthRate, MarketTrend: in.MarketTrend,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "scenario not found")
			return
		}
		s.log.Error("update custom scenario failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not update scenario")
		return
	}
	writeJSON(w, http.StatusOK, toCustomScenarioResponse(updated))
}

func (s *Server) handleDeleteCustomScenario(w http.ResponseWriter, r *http.Request) {
	if s.customScenarios == nil {
		writeError(w, http.StatusServiceUnavailable, "custom scenario service not configured")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "scenario id is required")
		return
	}
	ownerID := middleware.UserIDFrom(r.Context())
	if err := s.customScenarios.Delete(r.Context(), id, ownerID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "scenario not found")
			return
		}
		s.log.Error("delete custom scenario failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not delete scenario")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": id, "status": "deleted"})
}

func customScenarioToScenario(c *repository.CustomScenario) scenarios.Scenario {
	in := scenarios.CustomInput{
		Name: c.Name, Description: c.Description, Difficulty: scenarios.Difficulty(c.Difficulty),
		Industry: c.Industry, StartingCashCents: c.StartingCashCents, StartingBurnCents: c.StartingBurnCents,
		MarketTAM: c.MarketTAM, MarketGrowthRate: c.MarketGrowthRate, MarketTrend: c.MarketTrend,
	}
	return in.Normalize(c.ID)
}

func (s *Server) handleApplyCustomScenario(w http.ResponseWriter, r *http.Request) {
	if s.customScenarios == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "custom scenario service not configured")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "scenario id is required")
		return
	}
	ownerID := middleware.UserIDFrom(r.Context())
	custom, err := s.customScenarios.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "scenario not found")
			return
		}
		s.log.Error("load custom scenario failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load scenario")
		return
	}
	if custom.OwnerID != ownerID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	company, err := s.companies.GetLatestCompanyForOwner(r.Context(), ownerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no company found; create a company first")
			return
		}
		s.log.Error("apply custom scenario: load company failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load company")
		return
	}
	if company.Status == repository.CompanyBankrupt {
		writeError(w, http.StatusConflict, "bankrupt companies cannot apply a scenario; restart first")
		return
	}

	sc := customScenarioToScenario(custom)
	if err := s.companies.UpdateProfile(r.Context(), company.ID, sc.StartingCashCents, sc.Industry); err != nil {
		s.log.Error("apply custom scenario: update profile failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not apply scenario")
		return
	}
	if s.market != nil {
		if _, err := s.market.GetOrCreate(r.Context(), company.ID); err == nil {
			if err := s.market.Save(r.Context(), company.ID, int64(sc.Market.TAM), sc.Market.GrowthRate, sc.Market.TrendMultiplier); err != nil {
				s.log.Error("apply custom scenario: seed market failed", "error", err)
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
