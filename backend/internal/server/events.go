package server

import (
	"context"
	"net/http"
	"time"

	"github.com/YASSERRMD/Ventiqra/backend/internal/events"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
)

type gameEventResponse struct {
	ID              string    `json:"id"`
	Kind            string    `json:"kind"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	CashDelta       int64     `json:"cash_delta"`
	ReputationDelta int       `json:"reputation_delta"`
	MoraleDelta     int       `json:"morale_delta"`
	SimDay          int       `json:"sim_day"`
	CreatedAt       time.Time `json:"created_at"`
}

// handleListEvents returns the company's recent random events.
func (s *Server) handleListEvents(w http.ResponseWriter, r *http.Request) {
	if s.gameEvents == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "event service not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	list, err := s.gameEvents.ListByCompany(r.Context(), companyID, 25)
	if err != nil {
		s.log.Error("list events failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load events")
		return
	}
	out := make([]gameEventResponse, 0, len(list))
	for _, e := range list {
		out = append(out, gameEventResponse{
			ID: e.ID, Kind: e.Kind, Title: e.Title, Description: e.Description,
			CashDelta: e.CashDelta, ReputationDelta: e.ReputationDelta, MoraleDelta: e.MoraleDelta,
			SimDay: e.SimDay, CreatedAt: e.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// rollAndApplyEvent deterministically rolls a daily event; if one fires it
// applies its effects (cash, reputation, morale) and records it.
func (s *Server) rollAndApplyEvent(ctx context.Context, companyID string, cash int64, seed int64, day int) int64 {
	if s.gameEvents == nil {
		return cash
	}
	ev, ok := events.MaybeRoll(seed, int64(day))
	if !ok {
		return cash
	}
	newCash := cash + ev.CashDelta
	if ev.CashDelta != 0 {
		_ = s.companies.UpdateCash(ctx, companyID, newCash)
	}
	if ev.ReputationDelta != 0 {
		s.recordReputationEvent(ctx, companyID, ev.Title, ev.ReputationDelta, day)
	}
	if ev.MoraleDelta != 0 {
		if ev.MoraleDelta > 0 {
			s.boostTeam(ctx, companyID, ev.MoraleDelta)
		} else {
			s.drainTeam(ctx, companyID, -ev.MoraleDelta)
		}
	}
	_, _ = s.gameEvents.Record(ctx, &repository.GameEvent{
		CompanyID: companyID, Kind: string(ev.Kind), Title: ev.Title, Description: ev.Description,
		CashDelta: ev.CashDelta, ReputationDelta: ev.ReputationDelta, MoraleDelta: ev.MoraleDelta, SimDay: day,
	})
	return newCash
}

// drainTeam reduces every employee's morale by a milestone penalty.
func (s *Server) drainTeam(ctx context.Context, companyID string, amount int) {
	if s.employees == nil || amount <= 0 {
		return
	}
	team, err := s.employees.ListEmployeesByCompany(ctx, companyID)
	if err != nil || len(team) == 0 {
		return
	}
	for _, e := range team {
		_ = s.employees.UpdateMorale(ctx, e.ID, e.Morale-amount)
	}
}
