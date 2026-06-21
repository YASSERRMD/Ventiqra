package server

import (
	"context"
	"net/http"
	"time"

	"github.com/YASSERRMD/Ventiqra/backend/internal/market"
)

type marketResponse struct {
	TAM             int64     `json:"tam"`
	GrowthRate      float64   `json:"growth_rate"`
	TrendMultiplier float64   `json:"trend_multiplier"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// handleGetMarket returns the company's market model, initializing defaults.
func (s *Server) handleGetMarket(w http.ResponseWriter, r *http.Request) {
	if s.market == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "market service not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	m, err := s.market.GetOrCreate(r.Context(), companyID)
	if err != nil {
		s.log.Error("market load failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load market")
		return
	}
	writeJSON(w, http.StatusOK, marketResponse{
		TAM: m.TAM, GrowthRate: m.GrowthRate, TrendMultiplier: m.TrendMultiplier, UpdatedAt: m.UpdatedAt,
	})
}

// advanceMarket evolves the market one day (growth + trend drift), persists it,
// and returns the current trend multiplier so the tick can scale demand.
func (s *Server) advanceMarket(ctx context.Context, companyID string, seed int64, day int) float64 {
	if s.market == nil {
		return 1.0
	}
	m, err := s.market.GetOrCreate(ctx, companyID)
	if err != nil {
		return 1.0
	}
	advanced := market.Advance(market.Model{
		TAM: m.TAM, GrowthRate: m.GrowthRate, TrendMultiplier: m.TrendMultiplier,
	}, seed, int64(day))
	_ = s.market.Save(ctx, companyID, advanced.TAM, advanced.GrowthRate, advanced.TrendMultiplier)
	return advanced.TrendMultiplier
}
