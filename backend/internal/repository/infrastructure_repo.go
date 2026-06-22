package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Infrastructure is a company's hosting capacity and cost state.
type Infrastructure struct {
	CompanyID   string
	Tier        int
	Capacity    int
	HostingCost int64
	UpdatedAt   time.Time
}

type InfrastructureRepo struct{ *Repository }

func NewInfrastructureRepo(base *Repository) *InfrastructureRepo {
	return &InfrastructureRepo{Repository: base}
}

func (r *InfrastructureRepo) GetOrCreate(ctx context.Context, companyID string) (*Infrastructure, error) {
	if _, err := r.pool.Exec(ctx,
		`INSERT INTO infrastructure (company_id) VALUES ($1) ON CONFLICT (company_id) DO NOTHING`, companyID); err != nil {
		return nil, fmt.Errorf("ensure infrastructure: %w", err)
	}
	const q = `SELECT company_id, tier, capacity, hosting_cost, updated_at FROM infrastructure WHERE company_id = $1`
	var inf Infrastructure
	err := r.pool.QueryRow(ctx, q, companyID).Scan(&inf.CompanyID, &inf.Tier, &inf.Capacity, &inf.HostingCost, &inf.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get infrastructure: %w", err)
	}
	return &inf, nil
}

func (r *InfrastructureRepo) SetTier(ctx context.Context, companyID string, tier, capacity int, cost int64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE infrastructure SET tier = $2, capacity = $3, hosting_cost = $4, updated_at = NOW() WHERE company_id = $1`,
		companyID, tier, capacity, cost)
	if err != nil {
		return fmt.Errorf("set infrastructure tier: %w", err)
	}
	return nil
}
