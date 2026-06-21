package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/YASSERRMD/Ventiqra/backend/internal/funding"
	"github.com/YASSERRMD/Ventiqra/backend/internal/middleware"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
)

type fundingRoundResponse struct {
	ID            string    `json:"id"`
	RoundName     string    `json:"round_name"`
	AmountCents   int64     `json:"amount_cents"`
	PreMoneyCents int64     `json:"pre_money_cents"`
	EquityPercent float64   `json:"equity_percent"`
	SimDay        int       `json:"sim_day"`
	CreatedAt     time.Time `json:"created_at"`
}

type fundingSummaryResponse struct {
	PreMoneyCents   int64                  `json:"pre_money_cents"`
	FounderEquity   float64                `json:"founder_equity_percent"`
	InvestorInterest float64               `json:"investor_interest"`
	RoundsRaised    int                    `json:"rounds_raised"`
	Rounds          []fundingRoundResponse `json:"rounds"`
}

type raiseFundingRequest struct {
	AmountCents *int64 `json:"amount_cents"`
}

// handleListFunding returns the company's funding summary and round history.
func (s *Server) handleListFunding(w http.ResponseWriter, r *http.Request) {
	if s.funding == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "funding service not configured")
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

	summary := s.buildFundingSummary(r.Context(), company)
	writeJSON(w, http.StatusOK, summary)
}

// handleRaiseFunding closes a new funding round: computes the pre-money
// valuation from current traction, derives the equity dilution for the requested
// amount, records the round, and adds the cash to the company.
func (s *Server) handleRaiseFunding(w http.ResponseWriter, r *http.Request) {
	if s.funding == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "funding service not configured")
		return
	}
	var req raiseFundingRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.AmountCents == nil || *req.AmountCents <= 0 {
		writeError(w, http.StatusBadRequest, "amount_cents must be a positive integer")
		return
	}

	ownerID := middleware.UserIDFrom(r.Context())
	company, err := s.companies.GetLatestCompanyForOwner(r.Context(), ownerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no company found")
			return
		}
		s.log.Error("raise: load company failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not process raise")
		return
	}
	if company.Status == repository.CompanyBankrupt {
		writeError(w, http.StatusConflict, "bankrupt companies cannot raise funding")
		return
	}

	monthlyRevenue := s.companyMonthlyRevenue(r.Context(), company.ID)
	preMoney := funding.PreMoneyValuation(company.Cash, monthlyRevenue)
	equity := funding.EquityPercent(*req.AmountCents, preMoney)
	if equity > 90 {
		writeError(w, http.StatusConflict, "round would dilute founders below 10%; reduce the amount")
		return
	}

	priorRounds, _ := s.funding.CountByCompany(r.Context(), company.ID)
	roundName := funding.NextRoundName(priorRounds)
	day := s.currentSimDay(r.Context(), company.ID)

	recorded, err := s.funding.Record(r.Context(), &repository.FundingRound{
		CompanyID:     company.ID,
		RoundName:     roundName,
		AmountCents:   *req.AmountCents,
		PreMoneyCents: preMoney,
		EquityPercent: equity,
		SimDay:        day,
	})
	if err != nil {
		s.log.Error("record funding round failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not record round")
		return
	}

	// Cash in the raise.
	newCash := company.Cash + *req.AmountCents
	if err := s.companies.UpdateCash(r.Context(), company.ID, newCash); err != nil {
		s.log.Error("apply raise cash failed", "error", err)
	}

	// Closing a round is a reputation milestone and a team morale boost.
	s.recordReputationEvent(r.Context(), company.ID, "raised "+roundName, 2, day)
	s.boostTeam(r.Context(), company.ID, 5)

	// Record the funding round on the unified timeline.
	s.recordTimeline(r.Context(), company.ID, "funding", "Raised "+roundName,
		"Closed a "+roundName+" round.", day)

	updated, _ := s.companies.GetCompany(r.Context(), company.ID)
	summary := s.buildFundingSummary(r.Context(), updated)
	summary.Rounds = append(summary.Rounds, fundingRoundResponse{
		ID: recorded.ID, RoundName: recorded.RoundName, AmountCents: recorded.AmountCents,
		PreMoneyCents: recorded.PreMoneyCents, EquityPercent: recorded.EquityPercent,
		SimDay: recorded.SimDay, CreatedAt: recorded.CreatedAt,
	})
	writeJSON(w, http.StatusCreated, summary)
}

// buildFundingSummary assembles the funding dashboard summary for a company.
func (s *Server) buildFundingSummary(ctx context.Context, company *repository.Company) fundingSummaryResponse {
	rounds, _ := s.funding.ListByCompany(ctx, company.ID)
	monthlyRevenue := s.companyMonthlyRevenue(ctx, company.ID)
	preMoney := funding.PreMoneyValuation(company.Cash, monthlyRevenue)

	founder := 100.0
	// rounds are newest-first; iterate oldest-first for compounding.
	for i := len(rounds) - 1; i >= 0; i-- {
		founder = funding.FounderEquity(founder, rounds[i].EquityPercent)
	}

	resp := make([]fundingRoundResponse, 0, len(rounds))
	for _, fr := range rounds {
		resp = append(resp, fundingRoundResponse{
			ID: fr.ID, RoundName: fr.RoundName, AmountCents: fr.AmountCents,
			PreMoneyCents: fr.PreMoneyCents, EquityPercent: fr.EquityPercent,
			SimDay: fr.SimDay, CreatedAt: fr.CreatedAt,
		})
	}

	return fundingSummaryResponse{
		PreMoneyCents:   preMoney,
		FounderEquity:   founder,
		InvestorInterest: funding.InvestorInterest(monthlyRevenue, company.Cash),
		RoundsRaised:    len(rounds),
		Rounds:          resp,
	}
}

// companyMonthlyRevenue derives a monthly revenue estimate from the persisted
// daily revenue (scaled to 30 days).
func (s *Server) companyMonthlyRevenue(ctx context.Context, companyID string) int64 {
	if s.sim == nil {
		return 0
	}
	if state, err := s.sim.Get(ctx, companyID); err == nil {
		return state.Revenue * 30
	}
	return 0
}
