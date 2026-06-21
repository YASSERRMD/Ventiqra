package repository

import (
	"context"
	"fmt"
	"time"
)

// Competitor is a persisted rival company.
type Competitor struct {
	ID            string
	CompanyID     string
	Name          string
	Strength      int
	MarketShare   float64
	LastLaunchDay int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// CompetitorRepo provides persistence for rivals.
type CompetitorRepo struct {
	*Repository
}

// NewCompetitorRepo creates a CompetitorRepo over the shared pool.
func NewCompetitorRepo(base *Repository) *CompetitorRepo {
	return &CompetitorRepo{Repository: base}
}

// EnsureSeeded inserts rivals if the company has none yet.
func (r *CompetitorRepo) EnsureSeeded(ctx context.Context, companyID string, comps []Competitor) error {
	existing, err := r.ListByCompany(ctx, companyID)
	if err != nil {
		return err
	}
	if len(existing) > 0 || len(comps) == 0 {
		return nil
	}
	const q = `INSERT INTO competitors (company_id, name, strength, market_share, last_launch_day)
	           VALUES ($1, $2, $3, $4, $5)`
	for _, c := range comps {
		if _, err := r.pool.Exec(ctx, q, companyID, c.Name, c.Strength, c.MarketShare, c.LastLaunchDay); err != nil {
			return fmt.Errorf("seed competitor: %w", err)
		}
	}
	return nil
}

// ListByCompany returns a company's rivals.
func (r *CompetitorRepo) ListByCompany(ctx context.Context, companyID string) ([]*Competitor, error) {
	const q = `SELECT id, company_id, name, strength, market_share, last_launch_day, created_at, updated_at
	           FROM competitors WHERE company_id = $1 ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, q, companyID)
	if err != nil {
		return nil, fmt.Errorf("list competitors: %w", err)
	}
	defer rows.Close()
	var out []*Competitor
	for rows.Next() {
		var c Competitor
		if err := rows.Scan(&c.ID, &c.CompanyID, &c.Name, &c.Strength, &c.MarketShare, &c.LastLaunchDay, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan competitor: %w", err)
		}
		out = append(out, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("competitors rows: %w", err)
	}
	return out, nil
}

// Update advances a rival's strength, market share, and last launch day.
func (r *CompetitorRepo) Update(ctx context.Context, id string, strength int, marketShare float64, lastLaunchDay int) error {
	const q = `UPDATE competitors SET strength = $2, market_share = $3, last_launch_day = $4 WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id, strength, marketShare, lastLaunchDay)
	if err != nil {
		return fmt.Errorf("update competitor: %w", err)
	}
	return nil
}
