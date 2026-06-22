package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// SimControl is a company's run-mode and speed state.
type SimControl struct {
	CompanyID string
	Mode      string // 'paused' | 'auto'
	Speed     int    // 1, 5, 30
	UpdatedAt time.Time
}

// SimControlRepo provides persistence for per-company speed control.
type SimControlRepo struct {
	*Repository
}

// NewSimControlRepo creates a SimControlRepo over the shared pool.
func NewSimControlRepo(base *Repository) *SimControlRepo {
	return &SimControlRepo{Repository: base}
}

// GetOrCreate returns the company's control row, creating a paused/1x default.
func (r *SimControlRepo) GetOrCreate(ctx context.Context, companyID string) (*SimControl, error) {
	const q = `INSERT INTO sim_control (company_id, mode, speed)
	           VALUES ($1, 'paused', 1)
	           ON CONFLICT (company_id) DO NOTHING`
	if _, err := r.pool.Exec(ctx, q, companyID); err != nil {
		return nil, fmt.Errorf("ensure sim control: %w", err)
	}
	return r.Get(ctx, companyID)
}

// Get returns the company's control row, or ErrNotFound.
func (r *SimControlRepo) Get(ctx context.Context, companyID string) (*SimControl, error) {
	const q = `SELECT company_id, mode, speed, updated_at FROM sim_control WHERE company_id = $1`
	var c SimControl
	err := r.pool.QueryRow(ctx, q, companyID).Scan(&c.CompanyID, &c.Mode, &c.Speed, &c.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get sim control: %w", err)
	}
	return &c, nil
}

// SetMode updates the run mode (paused/auto).
func (r *SimControlRepo) SetMode(ctx context.Context, companyID, mode string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE sim_control SET mode = $2, updated_at = NOW() WHERE company_id = $1`, companyID, mode)
	if err != nil {
		return fmt.Errorf("set sim mode: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// SetSpeed updates the tick speed.
func (r *SimControlRepo) SetSpeed(ctx context.Context, companyID string, speed int) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE sim_control SET speed = $2, updated_at = NOW() WHERE company_id = $1`, companyID, speed)
	if err != nil {
		return fmt.Errorf("set sim speed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
