// Customer-support handlers and tick integration. The support endpoint reads
// the ticket backlog; each tick advances it (arrivals + resolutions) and the
// backlog feeds a satisfaction penalty into the customer model.
package server

import (
	"context"
	"net/http"

	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
	"github.com/YASSERRMD/Ventiqra/backend/internal/support"
)

type supportResponse struct {
	OpenTickets   int `json:"open_tickets"`
	ResolvedTotal int `json:"resolved_total"`
	Agents        int `json:"support_agents"`
	SatPenalty    int `json:"satisfaction_penalty"`
}

func (s *Server) toSupportResponse(ss *repository.SupportState, agents int) supportResponse {
	return supportResponse{
		OpenTickets: ss.OpenTickets, ResolvedTotal: ss.ResolvedTotal,
		Agents: agents, SatPenalty: support.SatisfactionPenaltyPerHundredOpen(ss.OpenTickets),
	}
}

// handleGetSupport returns the company's support backlog.
func (s *Server) handleGetSupport(w http.ResponseWriter, r *http.Request) {
	if s.support == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "support service not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	ss, err := s.support.GetOrCreate(r.Context(), companyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load support state")
		return
	}
	writeJSON(w, http.StatusOK, s.toSupportResponse(ss, s.supportAgentCount(r.Context(), companyID)))
}

// supportAgentCount returns the number of support-role employees, or 0.
func (s *Server) supportAgentCount(ctx context.Context, companyID string) int {
	if s.employees == nil {
		return 0
	}
	emps, err := s.employees.ListEmployeesByCompany(ctx, companyID)
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range emps {
		if e.Role == repository.RoleSupport {
			n++
		}
	}
	return n
}

// advanceSupport advances the ticket backlog by one day during a tick. Best-effort.
func (s *Server) advanceSupport(ctx context.Context, companyID string, customers int) {
	if s.support == nil {
		return
	}
	ss, err := s.support.GetOrCreate(ctx, companyID)
	if err != nil {
		return
	}
	agents := s.supportAgentCount(ctx, companyID)
	newOpen, resolved, _ := support.ApplyDay(ss.OpenTickets, customers, agents)
	_ = s.support.ApplyDay(ctx, companyID, newOpen, resolved)
}
