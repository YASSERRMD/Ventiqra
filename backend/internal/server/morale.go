package server

import (
	"context"
	"net/http"

	"github.com/YASSERRMD/Ventiqra/backend/internal/morale"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
)

type moraleSummaryResponse struct {
	Headcount     int     `json:"headcount"`
	AverageMorale int     `json:"average_morale"`
	AtRisk        int     `json:"at_risk"`
	BurntOut      int     `json:"burnt_out"`
}

// handleGetMorale returns a team morale summary.
func (s *Server) handleGetMorale(w http.ResponseWriter, r *http.Request) {
	if s.employees == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "morale service not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	team, err := s.employees.ListEmployeesByCompany(r.Context(), companyID)
	if err != nil {
		s.log.Error("morale: load team failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load morale")
		return
	}
	if team == nil {
		team = []*repository.Employee{}
	}
	var sum, atRisk, burnt int
	for _, e := range team {
		sum += e.Morale
		if e.Morale <= morale.BurnoutThreshold {
			burnt++
		} else if e.Morale < 50 {
			atRisk++
		}
	}
	avg := 0
	if len(team) > 0 {
		avg = sum / len(team)
	}
	writeJSON(w, http.StatusOK, moraleSummaryResponse{
		Headcount: len(team), AverageMorale: avg, AtRisk: atRisk, BurntOut: burnt,
	})
}

// advanceMorale applies daily decay and resignation risk to the team, returning
// the number of employees who resigned.
func (s *Server) advanceMorale(ctx context.Context, companyID string, seed int64, day int) int {
	if s.employees == nil {
		return 0
	}
	team, err := s.employees.ListEmployeesByCompany(ctx, companyID)
	if err != nil || len(team) == 0 {
		return 0
	}
	resigned := 0
	for i, e := range team {
		next := e.Morale + morale.DailyDecay(e.Morale)
		if next < 0 {
			next = 0
		}
		if next != e.Morale {
			_ = s.employees.UpdateMorale(ctx, e.ID, next)
		}
		// Resignation is decided on the post-decay morale.
		if morale.Resigns(next, seed, int64(day), i) {
			if err := s.employees.DeleteEmployee(ctx, e.ID); err == nil {
				resigned++
				s.recordReputationEvent(ctx, companyID, "employee resignation", -1, day)
			}
		}
	}
	return resigned
}

// boostTeam lifts every employee's morale by a milestone amount.
func (s *Server) boostTeam(ctx context.Context, companyID string, amount int) {
	if s.employees == nil || amount <= 0 {
		return
	}
	team, err := s.employees.ListEmployeesByCompany(ctx, companyID)
	if err != nil || len(team) == 0 {
		return
	}
	for _, e := range team {
		_ = s.employees.UpdateMorale(ctx, e.ID, morale.Boost(e.Morale, amount))
	}
}
