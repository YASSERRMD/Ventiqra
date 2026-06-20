package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
)

type pricingExperimentResponse struct {
	ID            string    `json:"id"`
	ProductID     string    `json:"product_id"`
	ProductName   string    `json:"product_name"`
	OldPriceCents *int64    `json:"old_price_cents"`
	NewPriceCents int64     `json:"new_price_cents"`
	SimDay        int       `json:"sim_day"`
	CreatedAt     time.Time `json:"created_at"`
}

type setProductPriceRequest struct {
	PriceCents *int64 `json:"price_cents"`
}

// handleSetProductPrice sets a product's monthly price and records the change
// as a pricing experiment.
func (s *Server) handleSetProductPrice(w http.ResponseWriter, r *http.Request) {
	if s.pricing == nil || s.products == nil {
		writeError(w, http.StatusServiceUnavailable, "pricing service not configured")
		return
	}

	var req setProductPriceRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.PriceCents == nil || *req.PriceCents < 0 {
		writeError(w, http.StatusBadRequest, "price_cents must be a non-negative integer")
		return
	}

	product, ok := s.loadOwnedProduct(w, r)
	if !ok {
		return
	}

	oldPrice := product.PriceCents
	if err := s.products.SetPrice(r.Context(), product.ID, *req.PriceCents); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "product not found")
			return
		}
		s.log.Error("set price failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not set price")
		return
	}

	day := s.currentSimDay(r.Context(), product.CompanyID)
	if _, err := s.pricing.Record(r.Context(), product.ID, product.CompanyID, oldPrice, *req.PriceCents, day); err != nil {
		s.log.Error("record pricing experiment failed", "error", err)
	}

	updated, err := s.products.GetProduct(r.Context(), product.ID)
	if err != nil {
		updated = product
	}
	writeJSON(w, http.StatusOK, toProductResponse(updated))
}

// handleListPricingExperiments returns the owner's pricing experiment history.
func (s *Server) handleListPricingExperiments(w http.ResponseWriter, r *http.Request) {
	if s.pricing == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "pricing service not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	list, err := s.pricing.ListByCompany(r.Context(), companyID)
	if err != nil {
		s.log.Error("list pricing experiments failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load pricing experiments")
		return
	}

	names := map[string]string{}
	if s.products != nil {
		for _, e := range list {
			if _, ok := names[e.ProductID]; ok {
				continue
			}
			if p, err := s.products.GetProduct(r.Context(), e.ProductID); err == nil {
				names[e.ProductID] = p.Name
			}
		}
	}

	out := make([]pricingExperimentResponse, 0, len(list))
	for _, e := range list {
		name := names[e.ProductID]
		if name == "" {
			name = e.ProductID
		}
		out = append(out, pricingExperimentResponse{
			ID: e.ID, ProductID: e.ProductID, ProductName: name,
			OldPriceCents: e.OldPriceCents, NewPriceCents: e.NewPriceCents,
			SimDay: e.SimDay, CreatedAt: e.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// currentSimDay returns the company's current simulated day, or 0 if no state.
func (s *Server) currentSimDay(ctx context.Context, companyID string) int {
	if s.sim == nil {
		return 0
	}
	if state, err := s.sim.Get(ctx, companyID); err == nil {
		return state.Day
	}
	return 0
}
