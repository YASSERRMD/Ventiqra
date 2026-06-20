package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// CompanyFinance holds a company's finance settings (marketing budget, infra tier).
type CompanyFinance struct {
	CompanyID           string
	MarketingBudgetCents int64
	InfraTier           int
	UpdatedAt           time.Time
}

// FinanceRepo provides persistence for per-company finance settings.
type FinanceRepo struct {
	*Repository
}

// NewFinanceRepo creates a FinanceRepo over the shared pool.
func NewFinanceRepo(base *Repository) *FinanceRepo {
	return &FinanceRepo{Repository: base}
}

// Get returns the finance settings for a company, or ErrNotFound.
// GetOrCreate returns existing settings, initializing defaults if absent.
func (r *FinanceRepo) GetOrCreate(ctx context.Context, companyID string) (*CompanyFinance, error) {
	const q = `INSERT INTO company_finance (company_id) VALUES ($1)
	           ON CONFLICT (company_id) DO NOTHING`
	if _, err := r.pool.Exec(ctx, q, companyID); err != nil {
		return nil, fmt.Errorf("ensure finance: %w", err)
	}
	const sel = `SELECT company_id, marketing_budget_cents, infra_tier, updated_at
	             FROM company_finance WHERE company_id = $1`
	f, err := scanFinance(r.pool.QueryRow(ctx, sel, companyID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get finance: %w", err)
	}
	return f, nil
}

// SetMarketingBudget updates the monthly marketing budget for a company.
func (r *FinanceRepo) SetMarketingBudget(ctx context.Context, companyID string, budget int64) (*CompanyFinance, error) {
	if budget < 0 {
		budget = 0
	}
	// Ensure a row exists, then update.
	if _, err := r.GetOrCreate(ctx, companyID); err != nil {
		return nil, err
	}
	const q = `UPDATE company_finance SET marketing_budget_cents = $2 WHERE company_id = $1
	           RETURNING company_id, marketing_budget_cents, infra_tier, updated_at`
	f, err := scanFinance(r.pool.QueryRow(ctx, q, companyID, budget))
	if err != nil {
		return nil, fmt.Errorf("set marketing budget: %w", err)
	}
	return f, nil
}

type financeScanner interface {
	Scan(dest ...any) error
}

func scanFinance(row financeScanner) (*CompanyFinance, error) {
	var f CompanyFinance
	if err := row.Scan(&f.CompanyID, &f.MarketingBudgetCents, &f.InfraTier, &f.UpdatedAt); err != nil {
		return nil, err
	}
	return &f, nil
}
