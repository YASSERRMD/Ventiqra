package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// CustomScenario is a user-authored saved scenario.
type CustomScenario struct {
	ID                string
	OwnerID           string
	Name              string
	Description       string
	Difficulty        string
	Industry          string
	StartingCashCents int64
	StartingBurnCents int64
	MarketTAM         int
	MarketGrowthRate  float64
	MarketTrend       float64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// CustomScenarioRepo provides persistence for user-authored scenarios.
type CustomScenarioRepo struct {
	*Repository
}

// NewCustomScenarioRepo creates a CustomScenarioRepo over the shared pool.
func NewCustomScenarioRepo(base *Repository) *CustomScenarioRepo {
	return &CustomScenarioRepo{Repository: base}
}

// Create inserts a new custom scenario and returns it.
func (r *CustomScenarioRepo) Create(ctx context.Context, c *CustomScenario) (*CustomScenario, error) {
	const q = `INSERT INTO custom_scenarios (owner_id, name, description, difficulty, industry, starting_cash_cents, starting_burn_cents, market_tam, market_growth_rate, market_trend)
	           VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	           RETURNING id, owner_id, name, description, difficulty, industry, starting_cash_cents, starting_burn_cents, market_tam, market_growth_rate, market_trend, created_at, updated_at`
	return scanCustomScenario(r.pool.QueryRow(ctx, q,
		c.OwnerID, c.Name, c.Description, c.Difficulty, c.Industry,
		c.StartingCashCents, c.StartingBurnCents, c.MarketTAM, c.MarketGrowthRate, c.MarketTrend))
}

// Get returns a custom scenario by id, or ErrNotFound.
func (r *CustomScenarioRepo) Get(ctx context.Context, id string) (*CustomScenario, error) {
	const q = `SELECT id, owner_id, name, description, difficulty, industry, starting_cash_cents, starting_burn_cents, market_tam, market_growth_rate, market_trend, created_at, updated_at
	           FROM custom_scenarios WHERE id = $1`
	c, err := scanCustomScenario(r.pool.QueryRow(ctx, q, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get custom scenario: %w", err)
	}
	return c, nil
}

// ListByOwner returns a user's custom scenarios, newest first.
func (r *CustomScenarioRepo) ListByOwner(ctx context.Context, ownerID string) ([]*CustomScenario, error) {
	const q = `SELECT id, owner_id, name, description, difficulty, industry, starting_cash_cents, starting_burn_cents, market_tam, market_growth_rate, market_trend, created_at, updated_at
	           FROM custom_scenarios WHERE owner_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list custom scenarios: %w", err)
	}
	defer rows.Close()
	var out []*CustomScenario
	for rows.Next() {
		c, err := scanCustomScenario(rows)
		if err != nil {
			return nil, fmt.Errorf("scan custom scenario: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("custom scenarios rows: %w", err)
	}
	return out, nil
}

// Update modifies an existing custom scenario owned by ownerID. Returns
// ErrNotFound when the scenario does not exist or belongs to another user.
func (r *CustomScenarioRepo) Update(ctx context.Context, id, ownerID string, c *CustomScenario) (*CustomScenario, error) {
	const q = `UPDATE custom_scenarios
	           SET name = $3, description = $4, difficulty = $5, industry = $6,
	               starting_cash_cents = $7, starting_burn_cents = $8,
	               market_tam = $9, market_growth_rate = $10, market_trend = $11,
	               updated_at = NOW()
	           WHERE id = $1 AND owner_id = $2
	           RETURNING id, owner_id, name, description, difficulty, industry, starting_cash_cents, starting_burn_cents, market_tam, market_growth_rate, market_trend, created_at, updated_at`
	updated, err := scanCustomScenario(r.pool.QueryRow(ctx, q,
		id, ownerID, c.Name, c.Description, c.Difficulty, c.Industry,
		c.StartingCashCents, c.StartingBurnCents, c.MarketTAM, c.MarketGrowthRate, c.MarketTrend))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update custom scenario: %w", err)
	}
	return updated, nil
}

// Delete removes a custom scenario owned by ownerID.
func (r *CustomScenarioRepo) Delete(ctx context.Context, id, ownerID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM custom_scenarios WHERE id = $1 AND owner_id = $2`, id, ownerID)
	if err != nil {
		return fmt.Errorf("delete custom scenario: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type customScenarioScanner interface {
	Scan(dest ...any) error
}

func scanCustomScenario(row customScenarioScanner) (*CustomScenario, error) {
	var c CustomScenario
	if err := row.Scan(
		&c.ID, &c.OwnerID, &c.Name, &c.Description, &c.Difficulty, &c.Industry,
		&c.StartingCashCents, &c.StartingBurnCents, &c.MarketTAM, &c.MarketGrowthRate, &c.MarketTrend,
		&c.CreatedAt, &c.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &c, nil
}
