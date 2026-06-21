package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Market is the persisted market model for a company.
type Market struct {
	CompanyID       string
	TAM             int64
	GrowthRate      float64
	TrendMultiplier float64
	UpdatedAt       time.Time
}

// MarketRepo provides persistence for the market model.
type MarketRepo struct {
	*Repository
}

// NewMarketRepo creates a MarketRepo over the shared pool.
func NewMarketRepo(base *Repository) *MarketRepo {
	return &MarketRepo{Repository: base}
}

// GetOrCreate returns the market row, initializing defaults if absent.
func (r *MarketRepo) GetOrCreate(ctx context.Context, companyID string) (*Market, error) {
	const ins = `INSERT INTO market (company_id) VALUES ($1) ON CONFLICT (company_id) DO NOTHING`
	if _, err := r.pool.Exec(ctx, ins, companyID); err != nil {
		return nil, fmt.Errorf("ensure market: %w", err)
	}
	const q = `SELECT company_id, tam, growth_rate, trend_multiplier, updated_at
	           FROM market WHERE company_id = $1`
	var m Market
	err := r.pool.QueryRow(ctx, q, companyID).Scan(&m.CompanyID, &m.TAM, &m.GrowthRate, &m.TrendMultiplier, &m.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get market: %w", err)
	}
	return &m, nil
}

// Save updates the market model for a company.
func (r *MarketRepo) Save(ctx context.Context, companyID string, tam int64, growth, trend float64) error {
	const q = `UPDATE market SET tam = $2, growth_rate = $3, trend_multiplier = $4 WHERE company_id = $1`
	_, err := r.pool.Exec(ctx, q, companyID, tam, growth, trend)
	if err != nil {
		return fmt.Errorf("save market: %w", err)
	}
	return nil
}
