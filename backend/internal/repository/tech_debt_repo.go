package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// TechDebt is a company's technical-debt state.
type TechDebt struct {
	CompanyID       string
	Debt            int
	Refactors       int
	LastRefactorDay *int
	UpdatedAt       time.Time
}

type TechDebtRepo struct{ *Repository }

func NewTechDebtRepo(base *Repository) *TechDebtRepo { return &TechDebtRepo{Repository: base} }

func (r *TechDebtRepo) GetOrCreate(ctx context.Context, companyID string) (*TechDebt, error) {
	if _, err := r.pool.Exec(ctx,
		`INSERT INTO tech_debt (company_id) VALUES ($1) ON CONFLICT (company_id) DO NOTHING`, companyID); err != nil {
		return nil, fmt.Errorf("ensure tech debt: %w", err)
	}
	const q = `SELECT company_id, debt, refactors, last_refactor_day, updated_at FROM tech_debt WHERE company_id = $1`
	var td TechDebt
	err := r.pool.QueryRow(ctx, q, companyID).Scan(&td.CompanyID, &td.Debt, &td.Refactors, &td.LastRefactorDay, &td.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get tech debt: %w", err)
	}
	return &td, nil
}

func (r *TechDebtRepo) SetDebt(ctx context.Context, companyID string, debt int) error {
	_, err := r.pool.Exec(ctx, `UPDATE tech_debt SET debt = $2, updated_at = NOW() WHERE company_id = $1`, companyID, debt)
	if err != nil {
		return fmt.Errorf("set debt: %w", err)
	}
	return nil
}

func (r *TechDebtRepo) RecordRefactor(ctx context.Context, companyID string, newDebt, day int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE tech_debt SET debt = $2, refactors = refactors + 1, last_refactor_day = $3, updated_at = NOW() WHERE company_id = $1`,
		companyID, newDebt, day)
	if err != nil {
		return fmt.Errorf("record refactor: %w", err)
	}
	return nil
}
