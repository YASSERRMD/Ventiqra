package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// DecisionStatus describes the lifecycle state of a strategic decision card.
type DecisionStatus string

const (
	DecisionPending  DecisionStatus = "pending"
	DecisionResolved DecisionStatus = "resolved"
)

// DecisionOutcome labels the risk-roll result of a resolved choice.
type DecisionOutcome string

const (
	DecisionSuccess DecisionOutcome = "success"
	DecisionFailure DecisionOutcome = "failure"
)

// StrategicDecision is an offered (or resolved) decision card.
type StrategicDecision struct {
	ID                 string
	CompanyID          string
	DecisionID         string
	Title              string
	Description        string
	SimDayOffered      int
	Status             DecisionStatus
	ChosenChoice       *string
	Outcome            *DecisionOutcome
	RecurringCashDelta int64
	RemainingDays      int
	ResolvedAt         *time.Time
	CreatedAt          time.Time
}

// DecisionRepo provides persistence for strategic decisions.
type DecisionRepo struct {
	*Repository
}

// NewDecisionRepo creates a DecisionRepo over the shared pool.
func NewDecisionRepo(base *Repository) *DecisionRepo {
	return &DecisionRepo{Repository: base}
}

// Offer inserts a pending decision card and returns it.
func (r *DecisionRepo) Offer(ctx context.Context, companyID, decisionID, title, description string, simDay int) (*StrategicDecision, error) {
	const q = `INSERT INTO strategic_decisions (company_id, decision_id, title, description, sim_day_offered, status)
	           VALUES ($1, $2, $3, $4, $5, 'pending')
	           RETURNING id, company_id, decision_id, title, description, sim_day_offered, status, chosen_choice, outcome, recurring_cash_delta, remaining_days, resolved_at, created_at`
	return scanDecision(r.pool.QueryRow(ctx, q, companyID, decisionID, title, description, simDay))
}

// HasPending reports whether the company has a pending decision card.
func (r *DecisionRepo) HasPending(ctx context.Context, companyID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM strategic_decisions WHERE company_id = $1 AND status = 'pending')`, companyID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check pending decision: %w", err)
	}
	return exists, nil
}

// GetPending returns the company's pending card, or ErrNotFound when none.
func (r *DecisionRepo) GetPending(ctx context.Context, companyID string) (*StrategicDecision, error) {
	const q = `SELECT id, company_id, decision_id, title, description, sim_day_offered, status, chosen_choice, outcome, recurring_cash_delta, remaining_days, resolved_at, created_at
	           FROM strategic_decisions WHERE company_id = $1 AND status = 'pending' ORDER BY sim_day_offered DESC LIMIT 1`
	d, err := scanDecision(r.pool.QueryRow(ctx, q, companyID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get pending decision: %w", err)
	}
	return d, nil
}

// Get returns a strategic decision by id, or ErrNotFound.
func (r *DecisionRepo) Get(ctx context.Context, id string) (*StrategicDecision, error) {
	const q = `SELECT id, company_id, decision_id, title, description, sim_day_offered, status, chosen_choice, outcome, recurring_cash_delta, remaining_days, resolved_at, created_at
	           FROM strategic_decisions WHERE id = $1`
	d, err := scanDecision(r.pool.QueryRow(ctx, q, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get decision: %w", err)
	}
	return d, nil
}

// Resolve marks a card resolved: records the chosen option, the risk outcome,
// and the long-term recurring commitment (cash delta per day for duration days).
func (r *DecisionRepo) Resolve(ctx context.Context, id, choiceID string, outcome DecisionOutcome, recurringCash int64, duration int) (*StrategicDecision, error) {
	const q = `UPDATE strategic_decisions
	           SET status = 'resolved', chosen_choice = $2, outcome = $3,
	               recurring_cash_delta = $4, remaining_days = $5, resolved_at = NOW()
	           WHERE id = $1 AND status = 'pending'
	           RETURNING id, company_id, decision_id, title, description, sim_day_offered, status, chosen_choice, outcome, recurring_cash_delta, remaining_days, resolved_at, created_at`
	d, err := scanDecision(r.pool.QueryRow(ctx, q, id, choiceID, string(outcome), recurringCash, duration))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("resolve decision: %w", err)
	}
	return d, nil
}

// ListActive returns the company's resolved cards that still have remaining
// recurring days (in oldest-offered order so long-term effects apply cleanly).
func (r *DecisionRepo) ListActive(ctx context.Context, companyID string) ([]*StrategicDecision, error) {
	const q = `SELECT id, company_id, decision_id, title, description, sim_day_offered, status, chosen_choice, outcome, recurring_cash_delta, remaining_days, resolved_at, created_at
	           FROM strategic_decisions WHERE company_id = $1 AND status = 'resolved' AND remaining_days > 0
	           ORDER BY resolved_at ASC`
	return queryDecisions(r.pool.Query(ctx, q, companyID))
}

// DecrementRemaining reduces a card's remaining days by one (clamped at zero).
func (r *DecisionRepo) DecrementRemaining(ctx context.Context, id string) error {
	const q = `UPDATE strategic_decisions SET remaining_days = GREATEST(remaining_days - 1, 0) WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("decrement decision remaining: %w", err)
	}
	return nil
}

// ListResolved returns the company's resolved decision history, newest first.
func (r *DecisionRepo) ListResolved(ctx context.Context, companyID string, limit int) ([]*StrategicDecision, error) {
	if limit <= 0 {
		limit = 20
	}
	const q = `SELECT id, company_id, decision_id, title, description, sim_day_offered, status, chosen_choice, outcome, recurring_cash_delta, remaining_days, resolved_at, created_at
	           FROM strategic_decisions WHERE company_id = $1 AND status = 'resolved'
	           ORDER BY resolved_at DESC LIMIT $2`
	return queryDecisions(r.pool.Query(ctx, q, companyID, limit))
}

func queryDecisions(rows pgx.Rows, err error) ([]*StrategicDecision, error) {
	if err != nil {
		return nil, fmt.Errorf("query decisions: %w", err)
	}
	defer rows.Close()
	var out []*StrategicDecision
	for rows.Next() {
		d, err := scanDecision(rows)
		if err != nil {
			return nil, fmt.Errorf("scan decision: %w", err)
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("decisions rows: %w", err)
	}
	return out, nil
}

type decisionScanner interface {
	Scan(dest ...any) error
}

func scanDecision(row decisionScanner) (*StrategicDecision, error) {
	var d StrategicDecision
	var chosenChoice *string
	var outcome *string
	if err := row.Scan(
		&d.ID, &d.CompanyID, &d.DecisionID, &d.Title, &d.Description, &d.SimDayOffered,
		&d.Status, &chosenChoice, &outcome, &d.RecurringCashDelta, &d.RemainingDays,
		&d.ResolvedAt, &d.CreatedAt,
	); err != nil {
		return nil, err
	}
	d.ChosenChoice = chosenChoice
	if outcome != nil {
		o := DecisionOutcome(*outcome)
		d.Outcome = &o
	}
	return &d, nil
}
