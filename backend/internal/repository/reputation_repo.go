package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// ReputationEvent records a single reputation change.
type ReputationEvent struct {
	ID        string
	CompanyID string
	Event     string
	Delta     int
	SimDay    int
	CreatedAt time.Time
}

// ReputationRepo provides persistence for reputation score and events.
type ReputationRepo struct {
	*Repository
}

// NewReputationRepo creates a ReputationRepo over the shared pool.
func NewReputationRepo(base *Repository) *ReputationRepo {
	return &ReputationRepo{Repository: base}
}

// GetOrCreate returns the reputation score row, initializing at the neutral 50.
func (r *ReputationRepo) GetOrCreate(ctx context.Context, companyID string) (int, error) {
	const ins = `INSERT INTO reputation (company_id, score) VALUES ($1, 50) ON CONFLICT (company_id) DO NOTHING`
	if _, err := r.pool.Exec(ctx, ins, companyID); err != nil {
		return 0, fmt.Errorf("ensure reputation: %w", err)
	}
	var score int
	err := r.pool.QueryRow(ctx, `SELECT score FROM reputation WHERE company_id = $1`, companyID).Scan(&score)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, ErrNotFound
		}
		return 0, fmt.Errorf("get reputation: %w", err)
	}
	return score, nil
}

// Adjust applies a delta to the score (clamped 0..100) and records an event.
func (r *ReputationRepo) Adjust(ctx context.Context, companyID, event string, delta, simDay int) (int, error) {
	score, err := r.GetOrCreate(ctx, companyID)
	if err != nil {
		return 0, err
	}
	score += delta
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	if _, err := r.pool.Exec(ctx, `UPDATE reputation SET score = $2 WHERE company_id = $1`, companyID, score); err != nil {
		return 0, fmt.Errorf("update reputation: %w", err)
	}
	if _, err := r.pool.Exec(ctx,
		`INSERT INTO reputation_events (company_id, event, delta, sim_day) VALUES ($1, $2, $3, $4)`,
		companyID, event, delta, simDay); err != nil {
		return 0, fmt.Errorf("insert reputation event: %w", err)
	}
	return score, nil
}

// ListEvents returns a company's recent reputation events, newest first.
func (r *ReputationRepo) ListEvents(ctx context.Context, companyID string, limit int) ([]*ReputationEvent, error) {
	if limit <= 0 {
		limit = 20
	}
	const q = `SELECT id, company_id, event, delta, sim_day, created_at
	           FROM reputation_events WHERE company_id = $1 ORDER BY created_at DESC LIMIT $2`
	rows, err := r.pool.Query(ctx, q, companyID, limit)
	if err != nil {
		return nil, fmt.Errorf("list reputation events: %w", err)
	}
	defer rows.Close()
	var out []*ReputationEvent
	for rows.Next() {
		var e ReputationEvent
		if err := rows.Scan(&e.ID, &e.CompanyID, &e.Event, &e.Delta, &e.SimDay, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan reputation event: %w", err)
		}
		out = append(out, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reputation events rows: %w", err)
	}
	return out, nil
}
