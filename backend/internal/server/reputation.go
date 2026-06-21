package server

import (
	"context"
	"net/http"
	"time"

	"github.com/YASSERRMD/Ventiqra/backend/internal/reputation"
)

type reputationEventResponse struct {
	ID        string    `json:"id"`
	Event     string    `json:"event"`
	Delta     int       `json:"delta"`
	SimDay    int       `json:"sim_day"`
	CreatedAt time.Time `json:"created_at"`
}

type reputationResponse struct {
	Score  int                       `json:"score"`
	Growth float64                   `json:"growth_multiplier"`
	Events []reputationEventResponse `json:"events"`
}

// handleGetReputation returns the company's reputation score, growth multiplier,
// and recent reputation events.
func (s *Server) handleGetReputation(w http.ResponseWriter, r *http.Request) {
	if s.reputation == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "reputation service not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	score, err := s.reputation.GetOrCreate(r.Context(), companyID)
	if err != nil {
		s.log.Error("reputation load failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load reputation")
		return
	}
	events, _ := s.reputation.ListEvents(r.Context(), companyID, 20)
	resp := reputationResponse{
		Score:  score,
		Growth: reputation.GrowthMultiplier(score),
		Events: make([]reputationEventResponse, 0, len(events)),
	}
	for _, e := range events {
		resp.Events = append(resp.Events, reputationEventResponse{
			ID: e.ID, Event: e.Event, Delta: e.Delta, SimDay: e.SimDay, CreatedAt: e.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

// applyReputationDrift evolves the reputation score from satisfaction and health
// during a tick, recording events for non-zero changes. Returns the growth
// multiplier the customer model should apply.
func (s *Server) applyReputationDrift(ctx context.Context, companyID string, avgSatisfaction int, health string, day int) float64 {
	if s.reputation == nil {
		return 1.0
	}
	score, err := s.reputation.GetOrCreate(ctx, companyID)
	if err != nil {
		return 1.0
	}
	// Satisfaction drift (small, recorded only when it changes the score band).
	if drift := reputation.SatisfactionDrift(avgSatisfaction); drift != 0 {
		score, _ = s.reputation.Adjust(ctx, companyID, "customer satisfaction", drift, day)
	}
	// Health-driven damage (crises, bankruptcy).
	if hd := reputation.HealthDelta(health); hd != 0 {
		score, _ = s.reputation.Adjust(ctx, companyID, health+" health", hd, day)
	}
	return reputation.GrowthMultiplier(score)
}

// recordReputationEvent applies an explicit reputation change for a milestone.
func (s *Server) recordReputationEvent(ctx context.Context, companyID, event string, delta, day int) {
	if s.reputation == nil {
		return
	}
	if _, err := s.reputation.Adjust(ctx, companyID, event, delta, day); err != nil {
		s.log.Error("reputation event failed", "event", event, "error", err)
	}
}
