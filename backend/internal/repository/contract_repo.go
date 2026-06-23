package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Contract is an enterprise recurring-revenue agreement.
type Contract struct {
	ID            string
	CompanyID     string
	CustomerName  string
	AnnualValue   int64
	TermDays      int
	RemainingDays int
	Status        string
	DiscountPct   int
	SignedDay     int
	CreatedAt     time.Time
}

type ContractRepo struct{ *Repository }

func NewContractRepo(base *Repository) *ContractRepo { return &ContractRepo{Repository: base} }

func (r *ContractRepo) Create(ctx context.Context, c *Contract) (*Contract, error) {
	const q = `INSERT INTO contracts (company_id, customer_name, annual_value, term_days, remaining_days, status, discount_pct, signed_day)
	           VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	           RETURNING id, company_id, customer_name, annual_value, term_days, remaining_days, status, discount_pct, signed_day, created_at`
	return scanContract(r.pool.QueryRow(ctx, q,
		c.CompanyID, c.CustomerName, c.AnnualValue, c.TermDays, c.RemainingDays, c.Status, c.DiscountPct, c.SignedDay))
}

func (r *ContractRepo) ListByCompany(ctx context.Context, companyID string) ([]*Contract, error) {
	const q = `SELECT id, company_id, customer_name, annual_value, term_days, remaining_days, status, discount_pct, signed_day, created_at
	           FROM contracts WHERE company_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, companyID)
	if err != nil {
		return nil, fmt.Errorf("list contracts: %w", err)
	}
	defer rows.Close()
	var out []*Contract
	for rows.Next() {
		c, err := scanContract(rows)
		if err != nil {
			return nil, fmt.Errorf("scan contract: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *ContractRepo) ListActive(ctx context.Context, companyID string) ([]*Contract, error) {
	const q = `SELECT id, company_id, customer_name, annual_value, term_days, remaining_days, status, discount_pct, signed_day, created_at
	           FROM contracts WHERE company_id = $1 AND status = 'active' ORDER BY signed_day ASC`
	rows, err := r.pool.Query(ctx, q, companyID)
	if err != nil {
		return nil, fmt.Errorf("list active contracts: %w", err)
	}
	defer rows.Close()
	var out []*Contract
	for rows.Next() {
		c, err := scanContract(rows)
		if err != nil {
			return nil, fmt.Errorf("scan contract: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *ContractRepo) SetStatus(ctx context.Context, id, status string, remaining int) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE contracts SET status = $2, remaining_days = $3 WHERE id = $1`, id, status, remaining)
	if err != nil {
		return fmt.Errorf("set contract status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ContractRepo) DecrementRemaining(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE contracts SET remaining_days = GREATEST(remaining_days - 1, 0) WHERE id = $1 AND status = 'active'`, id)
	return err
}

type contractScanner interface {
	Scan(dest ...any) error
}

func scanContract(row contractScanner) (*Contract, error) {
	var c Contract
	if err := row.Scan(&c.ID, &c.CompanyID, &c.CustomerName, &c.AnnualValue, &c.TermDays, &c.RemainingDays, &c.Status, &c.DiscountPct, &c.SignedDay, &c.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}
