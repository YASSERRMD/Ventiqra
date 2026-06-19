package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// SimState is the persisted simulation state for a single company. Cash,
// Revenue, and Seed mirror the in-memory sim.State; Day is the simulated day
// counter; MonthlyBurn is the recurring monthly operating cost in cents.
type SimState struct {
	CompanyID   string
	Day         int
	Seed        int64
	Cash        int64 // cents
	Revenue     int64 // cents
	MonthlyBurn int64 // cents per month
	UpdatedAt   time.Time
}

// SimStateRepo provides persistence for a company's simulation state.
type SimStateRepo struct {
	*Repository
}

// NewSimStateRepo creates a SimStateRepo over the shared pool.
func NewSimStateRepo(base *Repository) *SimStateRepo {
	return &SimStateRepo{Repository: base}
}

// Get returns the persisted simulation state for a company, or ErrNotFound if
// no state has been initialized yet.
func (r *SimStateRepo) Get(ctx context.Context, companyID string) (*SimState, error) {
	const q = `SELECT company_id, day, seed, cash, revenue, monthly_burn, updated_at
	           FROM simulation_state WHERE company_id = $1`
	s, err := scanSimState(r.pool.QueryRow(ctx, q, companyID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get sim state: %w", err)
	}
	return s, nil
}

// Init inserts an initial simulation state for a company if one does not yet
// exist, returning the current (possibly pre-existing) row. The insert is
// idempotent via ON CONFLICT DO NOTHING. revenue starts at 0 (no products yet);
// monthlyBurn is persisted so the engine's daily burn is reproducible.
func (r *SimStateRepo) Init(ctx context.Context, companyID string, seed, cash, monthlyBurn int64) (*SimState, error) {
	const q = `INSERT INTO simulation_state (company_id, day, seed, cash, revenue, monthly_burn)
	           VALUES ($1, 0, $2, $3, 0, $4)
	           ON CONFLICT (company_id) DO NOTHING
	           RETURNING company_id, day, seed, cash, revenue, monthly_burn, updated_at`

	s, err := scanSimState(r.pool.QueryRow(ctx, q, companyID, seed, cash, monthlyBurn))
	if err != nil {
		// ON CONFLICT DO NOTHING with RETURNING yields no rows when the row
		// already existed; fall back to reading the existing state.
		if errors.Is(err, pgx.ErrNoRows) {
			return r.Get(ctx, companyID)
		}
		return nil, fmt.Errorf("init sim state: %w", err)
	}
	return s, nil
}

// Save updates the day, cash, revenue, and monthly burn for an existing
// simulation state.
func (r *SimStateRepo) Save(ctx context.Context, companyID string, day int, cash, revenue, monthlyBurn int64) error {
	const q = `UPDATE simulation_state
	           SET day = $2, cash = $3, revenue = $4, monthly_burn = $5
	           WHERE company_id = $1`
	tag, err := r.pool.Exec(ctx, q, companyID, day, cash, revenue, monthlyBurn)
	if err != nil {
		return fmt.Errorf("save sim state: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type simStateScanner interface {
	Scan(dest ...any) error
}

func scanSimState(row simStateScanner) (*SimState, error) {
	var s SimState
	if err := row.Scan(&s.CompanyID, &s.Day, &s.Seed, &s.Cash, &s.Revenue, &s.MonthlyBurn, &s.UpdatedAt); err != nil {
		return nil, err
	}
	return &s, nil
}
