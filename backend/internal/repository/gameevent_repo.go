package repository

import (
	"context"
	"fmt"
	"time"
)

// GameEvent is a persisted random event that fired for a company.
type GameEvent struct {
	ID              string
	CompanyID       string
	Kind            string
	Title           string
	Description     string
	CashDelta       int64
	ReputationDelta int
	MoraleDelta     int
	SimDay          int
	CreatedAt       time.Time
}

// GameEventRepo provides persistence for random events.
type GameEventRepo struct {
	*Repository
}

// NewGameEventRepo creates a GameEventRepo over the shared pool.
func NewGameEventRepo(base *Repository) *GameEventRepo {
	return &GameEventRepo{Repository: base}
}

// Record inserts a game event and returns it.
func (r *GameEventRepo) Record(ctx context.Context, e *GameEvent) (*GameEvent, error) {
	const q = `INSERT INTO game_events (company_id, kind, title, description, cash_delta, reputation_delta, morale_delta, sim_day)
	           VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	           RETURNING id, company_id, kind, title, description, cash_delta, reputation_delta, morale_delta, sim_day, created_at`
	var out GameEvent
	err := r.pool.QueryRow(ctx, q,
		e.CompanyID, e.Kind, e.Title, e.Description, e.CashDelta, e.ReputationDelta, e.MoraleDelta, e.SimDay).
		Scan(&out.ID, &out.CompanyID, &out.Kind, &out.Title, &out.Description, &out.CashDelta, &out.ReputationDelta, &out.MoraleDelta, &out.SimDay, &out.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("record game event: %w", err)
	}
	return &out, nil
}

// ListByCompany returns a company's recent events, newest first.
func (r *GameEventRepo) ListByCompany(ctx context.Context, companyID string, limit int) ([]*GameEvent, error) {
	if limit <= 0 {
		limit = 20
	}
	const q = `SELECT id, company_id, kind, title, description, cash_delta, reputation_delta, morale_delta, sim_day, created_at
	           FROM game_events WHERE company_id = $1 ORDER BY created_at DESC LIMIT $2`
	rows, err := r.pool.Query(ctx, q, companyID, limit)
	if err != nil {
		return nil, fmt.Errorf("list game events: %w", err)
	}
	defer rows.Close()
	var out []*GameEvent
	for rows.Next() {
		var e GameEvent
		if err := rows.Scan(&e.ID, &e.CompanyID, &e.Kind, &e.Title, &e.Description, &e.CashDelta, &e.ReputationDelta, &e.MoraleDelta, &e.SimDay, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan game event: %w", err)
		}
		out = append(out, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("game events rows: %w", err)
	}
	return out, nil
}
