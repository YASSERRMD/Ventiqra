// Strategic-decision handlers and sim integration. This file adds the HTTP
// endpoints for pending decision cards and their resolution, plus the helpers
// the simulation tick calls to apply long-term commitments and offer new cards.
package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/YASSERRMD/Ventiqra/backend/internal/decisions"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
)

type decisionChoiceResponse struct {
	ID                  string  `json:"id"`
	Label               string  `json:"label"`
	Description         string  `json:"description"`
	CashDelta           int64   `json:"cash_delta"`
	ReputationDelta     int     `json:"reputation_delta"`
	MoraleDelta         int     `json:"morale_delta"`
	SuccessChance       float64 `json:"success_chance"`
	FailCashDelta       int64   `json:"fail_cash_delta"`
	FailReputationDelta int     `json:"fail_reputation_delta"`
	FailMoraleDelta     int     `json:"fail_morale_delta"`
	RecurringCashDelta  int64   `json:"recurring_cash_delta"`
	DurationDays        int     `json:"duration_days"`
}

type pendingDecisionResponse struct {
	ID            string                   `json:"id"`
	DecisionID    string                   `json:"decision_id"`
	Title         string                   `json:"title"`
	Description   string                   `json:"description"`
	Category      string                   `json:"category"`
	SimDayOffered int                      `json:"sim_day_offered"`
	Choices       []decisionChoiceResponse `json:"choices"`
}

type resolveDecisionRequest struct {
	ChoiceID *string `json:"choice_id"`
}

type resolveDecisionResponse struct {
	ID                 string `json:"id"`
	Outcome            string `json:"outcome"`
	AppliedCash        int64  `json:"applied_cash_delta"`
	AppliedReputation  int    `json:"applied_reputation_delta"`
	AppliedMorale      int    `json:"applied_morale_delta"`
	RecurringCashDelta int64  `json:"recurring_cash_delta"`
	RemainingDays      int    `json:"remaining_days"`
}

// handleGetPendingDecision returns the company's currently pending decision
// card, or 404 when none is pending.
func (s *Server) handleGetPendingDecision(w http.ResponseWriter, r *http.Request) {
	if s.decisions == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "decision service not configured")
		return
	}
	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	card, err := s.decisions.GetPending(r.Context(), companyID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no pending decision")
			return
		}
		s.log.Error("load pending decision failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load decision")
		return
	}
	d, ok := decisions.FindDecision(card.DecisionID)
	if !ok {
		writeError(w, http.StatusNotFound, "no pending decision")
		return
	}
	writeJSON(w, http.StatusOK, toPendingDecisionResponse(card, d))
}

// handleResolveDecision resolves the pending card: validates the chosen option,
// rolls the risk, applies short-term cash/reputation/morale, and persists the
// long-term recurring commitment.
func (s *Server) handleResolveDecision(w http.ResponseWriter, r *http.Request) {
	if s.decisions == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "decision service not configured")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "decision id is required")
		return
	}
	var req resolveDecisionRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.ChoiceID == nil || *req.ChoiceID == "" {
		writeError(w, http.StatusBadRequest, "choice_id is required")
		return
	}

	card, err := s.decisions.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "decision not found")
			return
		}
		s.log.Error("load decision failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load decision")
		return
	}

	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}
	if card.CompanyID != companyID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	if card.Status != repository.DecisionPending {
		writeError(w, http.StatusConflict, "decision is no longer pending")
		return
	}

	company, err := s.companies.GetCompany(r.Context(), card.CompanyID)
	if err != nil {
		writeError(w, http.StatusNotFound, "company not found")
		return
	}
	if company.Status == repository.CompanyBankrupt {
		writeError(w, http.StatusConflict, "bankrupt companies cannot resolve decisions")
		return
	}

	d, ok := decisions.FindDecision(card.DecisionID)
	if !ok {
		writeError(w, http.StatusConflict, "decision catalog entry no longer exists")
		return
	}
	ch, ok := d.FindChoice(*req.ChoiceID)
	if !ok {
		writeError(w, http.StatusBadRequest, "unknown choice for this decision")
		return
	}

	seed := s.companySeed(r.Context(), card.CompanyID)
	day := card.SimDayOffered
	eff, outcome := decisions.ResolveOutcome(seed, d.ID, ch.ID, int64(day), ch)

	// Apply short-term effects.
	newCash := company.Cash + eff.CashDelta
	if eff.CashDelta != 0 {
		if err := s.companies.UpdateCash(r.Context(), card.CompanyID, newCash); err != nil {
			s.log.Error("apply decision cash failed", "error", err)
		}
	}
	if eff.ReputationDelta != 0 {
		s.recordReputationEvent(r.Context(), card.CompanyID, d.Title+": "+string(outcome), eff.ReputationDelta, day)
	}
	if eff.MoraleDelta != 0 {
		if eff.MoraleDelta > 0 {
			s.boostTeam(r.Context(), card.CompanyID, eff.MoraleDelta)
		} else {
			s.drainTeam(r.Context(), card.CompanyID, -eff.MoraleDelta)
		}
	}

	// Persist resolution including the long-term commitment.
	resolved, err := s.decisions.Resolve(r.Context(), card.ID, ch.ID, toRepoOutcome(outcome), ch.RecurringCashDelta, ch.DurationDays)
	if err != nil {
		s.log.Error("resolve decision failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not resolve decision")
		return
	}

	// Record the decision on the unified timeline.
	s.recordTimeline(r.Context(), card.CompanyID, "decision", d.Title+": "+string(outcome),
		"Chose: "+ch.Label, day)

	writeJSON(w, http.StatusOK, resolveDecisionResponse{
		ID:                 resolved.ID,
		Outcome:            string(outcome),
		AppliedCash:        eff.CashDelta,
		AppliedReputation:  eff.ReputationDelta,
		AppliedMorale:      eff.MoraleDelta,
		RecurringCashDelta: resolved.RecurringCashDelta,
		RemainingDays:      resolved.RemainingDays,
	})
}

func toPendingDecisionResponse(card *repository.StrategicDecision, d decisions.Decision) pendingDecisionResponse {
	choices := make([]decisionChoiceResponse, 0, len(d.Choices))
	for _, c := range d.Choices {
		choices = append(choices, decisionChoiceResponse{
			ID: c.ID, Label: c.Label, Description: c.Description,
			CashDelta: c.CashDelta, ReputationDelta: c.ReputationDelta, MoraleDelta: c.MoraleDelta,
			SuccessChance: c.SuccessChance,
			FailCashDelta: c.FailCashDelta, FailReputationDelta: c.FailReputationDelta, FailMoraleDelta: c.FailMoraleDelta,
			RecurringCashDelta: c.RecurringCashDelta, DurationDays: c.DurationDays,
		})
	}
	return pendingDecisionResponse{
		ID: card.ID, DecisionID: card.DecisionID, Title: card.Title, Description: card.Description,
		Category: string(d.Category), SimDayOffered: card.SimDayOffered, Choices: choices,
	}
}

func toRepoOutcome(o decisions.Outcome) repository.DecisionOutcome {
	if o == decisions.OutcomeFailure {
		return repository.DecisionFailure
	}
	return repository.DecisionSuccess
}

// companySeed returns the company's simulation seed, falling back to a stable
// id-derived seed when no simulation state exists yet.
func (s *Server) companySeed(ctx context.Context, companyID string) int64 {
	if s.sim != nil {
		if state, err := s.sim.Get(ctx, companyID); err == nil {
			return state.Seed
		}
	}
	return companyIDHashSeed(companyID)
}

// applyActiveDecisionEffects advances every active long-term commitment by one
// day, applying its recurring cash delta and decrementing the remaining days.
// Cash is mutated in place.
func (s *Server) applyActiveDecisionEffects(ctx context.Context, companyID string, cash *int64, day int) {
	if s.decisions == nil || cash == nil {
		return
	}
	active, err := s.decisions.ListActive(ctx, companyID)
	if err != nil || len(active) == 0 {
		return
	}
	mutated := false
	for _, d := range active {
		if d.RecurringCashDelta != 0 {
			*cash += d.RecurringCashDelta
			mutated = true
		}
		_ = s.decisions.DecrementRemaining(ctx, d.ID)
	}
	if mutated {
		if err := s.companies.UpdateCash(ctx, companyID, *cash); err != nil {
			s.log.Error("apply recurring decision cash failed", "error", err)
		}
	}
}

// maybeOfferDecision offers a fresh decision card on the cadence defined by the
// decisions package, only when no card is currently pending.
func (s *Server) maybeOfferDecision(ctx context.Context, companyID string, seed int64, day int) {
	if s.decisions == nil {
		return
	}
	pending, err := s.decisions.HasPending(ctx, companyID)
	if err != nil || pending {
		return
	}
	d, ok := decisions.MaybeOffer(seed, int64(day))
	if !ok {
		return
	}
	if _, err := s.decisions.Offer(ctx, companyID, d.ID, d.Title, d.Description, day); err != nil {
		s.log.Error("offer decision failed", "error", err)
	}
}

// companyIDHashSeed returns a stable non-zero seed from a company id.
func companyIDHashSeed(id string) int64 {
	// FNV-1a folded over the id bytes. Uses uint64 to avoid int64 overflow on
	// the large FNV offset basis, then narrows to a non-zero int64 seed.
	const offset uint64 = 14695981039346656037
	const prime uint64 = 1099511628211
	h := offset
	for _, b := range []byte(id) {
		h ^= uint64(b)
		h *= prime
	}
	if h == 0 {
		h = 1
	}
	return int64(h)
}

// decisionCreatedAt exposes the created-at timestamp for tests/history.
func decisionCreatedAt(d *repository.StrategicDecision) time.Time { return d.CreatedAt }
