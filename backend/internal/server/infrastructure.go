// Infrastructure handlers: read capacity/cost/load and perform a scale-up
// action that raises the tier (and recurring hosting cost) for cash.
package server

import (
	"context"
	"net/http"

	"github.com/YASSERRMD/Ventiqra/backend/internal/infrastructure"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
)

type infrastructureResponse struct {
	Tier        int     `json:"tier"`
	Capacity    int     `json:"capacity"`
	HostingCost int64   `json:"hosting_cost_cents"`
	Customers   int     `json:"customers"`
	LoadRatio   float64 `json:"load_ratio"`
	OutageRisk  float64 `json:"outage_risk"`
	ScaleUpCost int64   `json:"scale_up_cost_cents"`
}

func (s *Server) toInfrastructureResponse(inf *repository.Infrastructure, customers int) infrastructureResponse {
	load := infrastructure.LoadRatio(customers, inf.Capacity)
	return infrastructureResponse{
		Tier: inf.Tier, Capacity: inf.Capacity, HostingCost: inf.HostingCost,
		Customers: customers, LoadRatio: load,
		OutageRisk: infrastructure.OutageRisk(load),
		ScaleUpCost: infrastructure.ScaleUpCostCents,
	}
}

func (s *Server) handleGetInfrastructure(w http.ResponseWriter, r *http.Request) {
	if s.infrastructure == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "infrastructure service not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	inf, err := s.infrastructure.GetOrCreate(r.Context(), companyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load infrastructure")
		return
	}
	customers := s.currentCustomerCount(r.Context(), companyID)
	writeJSON(w, http.StatusOK, s.toInfrastructureResponse(inf, customers))
}

func (s *Server) handleScaleUp(w http.ResponseWriter, r *http.Request) {
	if s.infrastructure == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "infrastructure service not configured")
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
		writeError(w, http.StatusConflict, "bankrupt companies cannot scale")
		return
	}
	inf, err := s.infrastructure.GetOrCreate(r.Context(), companyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load infrastructure")
		return
	}
	if inf.Tier >= infrastructure.MaxTier {
		writeError(w, http.StatusConflict, "already at maximum tier")
		return
	}
	// Charge the scale-up cost.
	newCash := company.Cash - infrastructure.ScaleUpCostCents
	if err := s.companies.UpdateCash(r.Context(), companyID, newCash); err != nil {
		writeError(w, http.StatusInternalServerError, "could not charge scale-up")
		return
	}
	newTier := inf.Tier + 1
	newCapacity := infrastructure.CapacityForTier(newTier)
	newCost := infrastructure.HostingCostForTier(newTier)
	if err := s.infrastructure.SetTier(r.Context(), companyID, newTier, newCapacity, newCost); err != nil {
		writeError(w, http.StatusInternalServerError, "could not scale up")
		return
	}
	day := s.currentSimDay(r.Context(), companyID)
	s.recordTimeline(r.Context(), companyID, "milestone", "Scaled infrastructure to tier "+formatInt(newTier),
		"Capacity now "+formatInt(newCapacity), day)
	updated, _ := s.infrastructure.GetOrCreate(r.Context(), companyID)
	customers := s.currentCustomerCount(r.Context(), companyID)
	writeJSON(w, http.StatusOK, s.toInfrastructureResponse(updated, customers))
}

// currentCustomerCount returns the company's total customer count across all
// products, defaulting to 0 when the customer service is unavailable.
func (s *Server) currentCustomerCount(ctx context.Context, companyID string) int {
	if s.customers == nil {
		return 0
	}
	list, err := s.customers.ListByCompany(ctx, companyID)
	if err != nil {
		return 0
	}
	total := 0
	for _, c := range list {
		total += c.Total
	}
	return total
}
