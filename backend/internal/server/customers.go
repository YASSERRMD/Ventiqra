package server

import (
	"net/http"
	"time"
)

type customerStateResponse struct {
	ProductID    string    `json:"product_id"`
	ProductName  string    `json:"product_name"`
	Total        int       `json:"total_customers"`
	MAU          int       `json:"mau"`
	Churned      int       `json:"churned"`
	Satisfaction int       `json:"satisfaction"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// handleListCustomers returns customer state for the owner's launched products.
func (s *Server) handleListCustomers(w http.ResponseWriter, r *http.Request) {
	if s.customers == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "customer service not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}

	list, err := s.customers.ListByCompany(r.Context(), companyID)
	if err != nil {
		s.log.Error("list customers failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load customers")
		return
	}

	names := map[string]string{}
	if s.products != nil {
		for _, c := range list {
			if _, ok := names[c.ProductID]; ok {
				continue
			}
			if p, err := s.products.GetProduct(r.Context(), c.ProductID); err == nil {
				names[c.ProductID] = p.Name
			}
		}
	}

	out := make([]customerStateResponse, 0, len(list))
	for _, c := range list {
		name := names[c.ProductID]
		if name == "" {
			name = c.ProductID
		}
		out = append(out, customerStateResponse{
			ProductID: c.ProductID, ProductName: name,
			Total: c.Total, MAU: c.MAU, Churned: c.Churned,
			Satisfaction: c.Satisfaction, UpdatedAt: c.UpdatedAt,
		})
	}
	writeJSON(w, http.StatusOK, out)
}
