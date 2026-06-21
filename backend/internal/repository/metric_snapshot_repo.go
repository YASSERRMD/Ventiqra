package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// MetricSnapshot is one day's headline metrics for a company.
type MetricSnapshot struct {
	ID              string
	CompanyID       string
	SimDay          int
	CashCents       int64
	RevenueCents    int64
	MonthlyBurn     int64
	Customers       int
	ValuationCents  int64
	CreatedAt       time.Time
}

// MetricSnapshotRepo provides persistence for daily metric snapshots.
type MetricSnapshotRepo struct {
	*Repository
}

// NewMetricSnapshotRepo creates a MetricSnapshotRepo over the shared pool.
func NewMetricSnapshotRepo(base *Repository) *MetricSnapshotRepo {
	return &MetricSnapshotRepo{Repository: base}
}

// Upsert records (or replaces) a day's metrics.
func (r *MetricSnapshotRepo) Upsert(ctx context.Context, m *MetricSnapshot) error {
	const q = `INSERT INTO metric_snapshots (company_id, sim_day, cash_cents, revenue_cents, monthly_burn, customers, valuation_cents)
	           VALUES ($1, $2, $3, $4, $5, $6, $7)
	           ON CONFLICT (company_id, sim_day) DO UPDATE
	             SET cash_cents = EXCLUDED.cash_cents, revenue_cents = EXCLUDED.revenue_cents,
	                 monthly_burn = EXCLUDED.monthly_burn, customers = EXCLUDED.customers,
	                 valuation_cents = EXCLUDED.valuation_cents`
	_, err := r.pool.Exec(ctx, q,
		m.CompanyID, m.SimDay, m.CashCents, m.RevenueCents, m.MonthlyBurn, m.Customers, m.ValuationCents)
	if err != nil {
		return fmt.Errorf("upsert metric snapshot: %w", err)
	}
	return nil
}

// ListByCompany returns a company's snapshots oldest-first, capped at limit.
func (r *MetricSnapshotRepo) ListByCompany(ctx context.Context, companyID string, limit int) ([]*MetricSnapshot, error) {
	if limit <= 0 {
		limit = 180
	}
	const q = `SELECT id, company_id, sim_day, cash_cents, revenue_cents, monthly_burn, customers, valuation_cents, created_at
	           FROM metric_snapshots WHERE company_id = $1 ORDER BY sim_day ASC LIMIT $2`
	rows, err := r.pool.Query(ctx, q, companyID, limit)
	if err != nil {
		return nil, fmt.Errorf("list metric snapshots: %w", err)
	}
	defer rows.Close()
	var out []*MetricSnapshot
	for rows.Next() {
		m, err := scanMetricSnapshot(rows)
		if err != nil {
			return nil, fmt.Errorf("scan metric snapshot: %w", err)
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("metric snapshots rows: %w", err)
	}
	return out, nil
}

type metricSnapshotScanner interface {
	Scan(dest ...any) error
}

func scanMetricSnapshot(row metricSnapshotScanner) (*MetricSnapshot, error) {
	var m MetricSnapshot
	if err := row.Scan(
		&m.ID, &m.CompanyID, &m.SimDay, &m.CashCents, &m.RevenueCents,
		&m.MonthlyBurn, &m.Customers, &m.ValuationCents, &m.CreatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}
