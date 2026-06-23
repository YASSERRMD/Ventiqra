// Enterprise-contract handlers and tick integration. Players sign recurring-
// revenue contracts with a negotiated discount/term; each tick accrues daily
// revenue and counts down the term, rolling renewal at expiry.
package server

import (
	"context"
	"net/http"

	"github.com/YASSERRMD/Ventiqra/backend/internal/contracts"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
)

type contractResponse struct {
	ID            string `json:"id"`
	CustomerName  string `json:"customer_name"`
	AnnualValue   int64  `json:"annual_value"`
	TermDays      int    `json:"term_days"`
	RemainingDays int    `json:"remaining_days"`
	Status        string `json:"status"`
	DiscountPct   int    `json:"discount_pct"`
	SignedDay     int    `json:"signed_day"`
}

type contractInput struct {
	CustomerName *string `json:"customer_name"`
	AnnualValue  *int64  `json:"annual_value"`
	DiscountPct  *int    `json:"discount_pct"`
	TermYears    *int    `json:"term_years"`
}

func toContractResponse(c *repository.Contract) contractResponse {
	return contractResponse{
		ID: c.ID, CustomerName: c.CustomerName, AnnualValue: c.AnnualValue,
		TermDays: c.TermDays, RemainingDays: c.RemainingDays, Status: c.Status,
		DiscountPct: c.DiscountPct, SignedDay: c.SignedDay,
	}
}

func (s *Server) handleListContracts(w http.ResponseWriter, r *http.Request) {
	if s.contracts == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "contracts service not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	list, err := s.contracts.ListByCompany(r.Context(), companyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load contracts")
		return
	}
	out := make([]contractResponse, 0, len(list))
	for _, c := range list {
		out = append(out, toContractResponse(c))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleSignContract(w http.ResponseWriter, r *http.Request) {
	if s.contracts == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "contracts service not configured")
		return
	}
	var req contractInput
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.CustomerName == nil || *req.CustomerName == "" {
		writeError(w, http.StatusBadRequest, "customer_name is required")
		return
	}
	if req.AnnualValue == nil || *req.AnnualValue <= 0 {
		writeError(w, http.StatusBadRequest, "annual_value must be positive")
		return
	}
	discount := 0
	if req.DiscountPct != nil {
		discount = *req.DiscountPct
		if discount < 0 || discount > 50 {
			writeError(w, http.StatusBadRequest, "discount_pct must be 0-50")
			return
		}
	}
	years := 1
	if req.TermYears != nil && *req.TermYears >= 1 {
		years = *req.TermYears
	}
	term := contracts.TermForYears(years)
	effectiveAnnual := contracts.AnnualValueFor(*req.AnnualValue, discount)
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	day := s.currentSimDay(r.Context(), companyID)
	created, err := s.contracts.Create(r.Context(), &repository.Contract{
		CompanyID: companyID, CustomerName: *req.CustomerName, AnnualValue: effectiveAnnual,
		TermDays: term, RemainingDays: term, Status: "active", DiscountPct: discount, SignedDay: day,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not sign contract")
		return
	}
	s.recordTimeline(r.Context(), companyID, "milestone", "Signed enterprise contract: "+*req.CustomerName,
		"Annual value "+formatInt(int(effectiveAnnual/100))+" dollars", day)
	writeJSON(w, http.StatusCreated, toContractResponse(created))
}

// advanceContracts accrues daily revenue and counts down active contracts,
// rolling renewal at expiry. Best-effort during ticks.
func (s *Server) advanceContracts(ctx context.Context, companyID string, cash *int64) {
	if s.contracts == nil {
		return
	}
	active, err := s.contracts.ListActive(ctx, companyID)
	if err != nil || len(active) == 0 {
		return
	}
	seed := s.companySeed(ctx, companyID)
	sat := s.companySatisfaction(ctx, companyID)
	mutated := false
	for _, c := range active {
		*cash += contracts.DailyRevenue(c.AnnualValue)
		mutated = true
		_ = s.contracts.DecrementRemaining(ctx, c.ID)
		if c.RemainingDays-1 <= 0 {
			if contracts.RenewalRoll(seed^int64(c.SignedDay), sat) {
				_ = s.contracts.SetStatus(ctx, c.ID, "renewed", 0)
			} else {
				_ = s.contracts.SetStatus(ctx, c.ID, "churned", 0)
			}
		}
	}
	if mutated {
		if err := s.companies.UpdateCash(ctx, companyID, *cash); err != nil {
			s.log.Error("accrue contract revenue failed", "error", err)
		}
	}
}

// companySatisfaction returns an approximate satisfaction score for renewal
// rolls, defaulting to 50 when unavailable.
func (s *Server) companySatisfaction(ctx context.Context, companyID string) int {
	if s.reputation != nil {
		if score, err := s.reputation.GetOrCreate(ctx, companyID); err == nil {
			return score
		}
	}
	return 50
}
