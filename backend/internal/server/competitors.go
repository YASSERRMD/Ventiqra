package server

import (
	"context"
	"net/http"
	"time"

	"github.com/YASSERRMD/Ventiqra/backend/internal/competitors"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
	"github.com/YASSERRMD/Ventiqra/backend/internal/sim"
)

type competitorResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Strength      int       `json:"strength"`
	MarketShare   float64   `json:"market_share"`
	LastLaunchDay int       `json:"last_launch_day"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// handleListCompetitors returns the company's rivals, seeding them lazily.
func (s *Server) handleListCompetitors(w http.ResponseWriter, r *http.Request) {
	if s.competitors == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "competitor service not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	s.ensureCompetitors(r.Context(), companyID)

	list, err := s.competitors.ListByCompany(r.Context(), companyID)
	if err != nil {
		s.log.Error("list competitors failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load competitors")
		return
	}
	out := make([]competitorResponse, 0, len(list))
	for _, c := range list {
		out = append(out, competitorResponse{
			ID: c.ID, Name: c.Name, Strength: c.Strength, MarketShare: c.MarketShare,
			LastLaunchDay: c.LastLaunchDay, UpdatedAt: c.UpdatedAt,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// ensureCompetitors seeds rivals deterministically if the company has none.
func (s *Server) ensureCompetitors(ctx context.Context, companyID string) {
	if s.competitors == nil {
		return
	}
	seed := sim.SeedFromCompanyID(companyID)
	if s.sim != nil {
		if state, err := s.sim.Get(ctx, companyID); err == nil {
			seed = state.Seed
		}
	}
	pure := competitors.Generate(seed)
	seeds := make([]repository.Competitor, 0, len(pure))
	for _, c := range pure {
		seeds = append(seeds, repository.Competitor{
			Name: c.Name, Strength: c.Strength, MarketShare: c.MarketShare,
		})
	}
	if err := s.competitors.EnsureSeeded(ctx, companyID, seeds); err != nil {
		s.log.Error("seed competitors failed", "error", err)
	}
}

// advanceCompetitors evolves each rival by one day and persists the result,
// returning the aggregate market pressure they exert on the player.
func (s *Server) advanceCompetitors(ctx context.Context, companyID string, seed int64, day int) float64 {
	if s.competitors == nil {
		return 0
	}
	list, err := s.competitors.ListByCompany(ctx, companyID)
	if err != nil || len(list) == 0 {
		return 0
	}
	var pure []competitors.Competitor
	for _, c := range list {
		advanced := competitors.Advance(competitors.Competitor{
			Name: c.Name, Strength: c.Strength, MarketShare: c.MarketShare, LastLaunchDay: c.LastLaunchDay,
		}, seed, int64(day))
		pure = append(pure, advanced)
		_ = s.competitors.Update(ctx, c.ID, advanced.Strength, advanced.MarketShare, advanced.LastLaunchDay)
	}
	return competitors.Pressure(pure)
}
