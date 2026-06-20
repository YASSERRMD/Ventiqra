package repository

import (
	"context"
	"fmt"
	"time"
)

// PricingExperiment records a single product price change.
type PricingExperiment struct {
	ID            string
	ProductID     string
	CompanyID     string
	OldPriceCents *int64
	NewPriceCents int64
	SimDay        int
	CreatedAt     time.Time
}

// PricingRepo provides persistence for pricing experiments.
type PricingRepo struct {
	*Repository
}

// NewPricingRepo creates a PricingRepo over the shared pool.
func NewPricingRepo(base *Repository) *PricingRepo {
	return &PricingRepo{Repository: base}
}

// Record inserts a pricing experiment row and returns it.
func (r *PricingRepo) Record(ctx context.Context, productID, companyID string, oldPrice *int64, newPrice int64, simDay int) (*PricingExperiment, error) {
	const q = `INSERT INTO pricing_experiments (product_id, company_id, old_price_cents, new_price_cents, sim_day)
	           VALUES ($1, $2, $3, $4, $5)
	           RETURNING id, product_id, company_id, old_price_cents, new_price_cents, sim_day, created_at`
	e, err := scanPricingExperiment(r.pool.QueryRow(ctx, q, productID, companyID, oldPrice, newPrice, simDay))
	if err != nil {
		return nil, fmt.Errorf("record pricing experiment: %w", err)
	}
	return e, nil
}

// ListByCompany returns a company's pricing experiments, newest first.
func (r *PricingRepo) ListByCompany(ctx context.Context, companyID string) ([]*PricingExperiment, error) {
	const q = `SELECT id, product_id, company_id, old_price_cents, new_price_cents, sim_day, created_at
	           FROM pricing_experiments WHERE company_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, companyID)
	if err != nil {
		return nil, fmt.Errorf("list pricing experiments: %w", err)
	}
	defer rows.Close()
	var out []*PricingExperiment
	for rows.Next() {
		e, err := scanPricingExperiment(rows)
		if err != nil {
			return nil, fmt.Errorf("scan pricing experiment: %w", err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("pricing experiments rows: %w", err)
	}
	return out, nil
}

type pricingScanner interface {
	Scan(dest ...any) error
}

func scanPricingExperiment(row pricingScanner) (*PricingExperiment, error) {
	var e PricingExperiment
	if err := row.Scan(&e.ID, &e.ProductID, &e.CompanyID, &e.OldPriceCents, &e.NewPriceCents, &e.SimDay, &e.CreatedAt); err != nil {
		return nil, err
	}
	return &e, nil
}
