package server

import (
	"net/http"

	"github.com/YASSERRMD/Ventiqra/backend/internal/marketing"
)

type channelResponse struct {
	Name       string  `json:"name"`
	Weight     float64 `json:"weight"`
	Conversion float64 `json:"conversion"`
}

type marketingResponse struct {
	MonthlyBudgetCents int64             `json:"monthly_budget_cents"`
	DailyConversions   int               `json:"daily_conversions"`
	CACCents           int64             `json:"cac_cents"`
	ConversionRate     float64           `json:"conversion_rate"`
	Channels           []channelResponse `json:"channels"`
}

// handleGetMarketing returns the company's marketing performance (conversions,
// CAC, conversion rate, and the acquisition channel mix).
func (s *Server) handleGetMarketing(w http.ResponseWriter, r *http.Request) {
	if s.finance == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "marketing service not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	fin, err := s.finance.GetOrCreate(r.Context(), companyID)
	if err != nil {
		s.log.Error("marketing: load finance failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load marketing")
		return
	}

	seed, day := s.companyRoundKey(r, companyID)
	budget := fin.MarketingBudgetCents
	conv := marketing.Conversions(budget, seed, int64(day))
	channels := make([]channelResponse, 0, len(marketing.Channels))
	for _, c := range marketing.Channels {
		channels = append(channels, channelResponse{Name: c.Name, Weight: c.Weight, Conversion: c.Conversion})
	}
	writeJSON(w, http.StatusOK, marketingResponse{
		MonthlyBudgetCents: budget,
		DailyConversions:   conv,
		CACCents:           marketing.CAC(budget, conv*30),
		ConversionRate:     marketing.ConversionRate(budget),
		Channels:           channels,
	})
}
