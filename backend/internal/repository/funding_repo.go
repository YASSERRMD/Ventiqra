package repository

import (
	"context"
	"fmt"
	"time"
)

// FundingRound records a closed funding round.
type FundingRound struct {
	ID            string
	CompanyID     string
	RoundName     string
	AmountCents   int64
	PreMoneyCents int64
	EquityPercent float64
	SimDay        int
	CreatedAt     time.Time
}

// FundingRepo provides persistence for funding rounds.
type FundingRepo struct {
	*Repository
}

// NewFundingRepo creates a FundingRepo over the shared pool.
func NewFundingRepo(base *Repository) *FundingRepo {
	return &FundingRepo{Repository: base}
}

// Record inserts a funding round and returns it.
func (r *FundingRepo) Record(ctx context.Context, fr *FundingRound) (*FundingRound, error) {
	const q = `INSERT INTO funding_rounds (company_id, round_name, amount_cents, pre_money_cents, equity_percent, sim_day)
	           VALUES ($1, $2, $3, $4, $5, $6)
	           RETURNING id, company_id, round_name, amount_cents, pre_money_cents, equity_percent, sim_day, created_at`
	out, err := scanFundingRound(r.pool.QueryRow(ctx, q,
		fr.CompanyID, fr.RoundName, fr.AmountCents, fr.PreMoneyCents, fr.EquityPercent, fr.SimDay))
	if err != nil {
		return nil, fmt.Errorf("record funding round: %w", err)
	}
	return out, nil
}

// CountByCompany returns the number of closed rounds for a company.
func (r *FundingRepo) CountByCompany(ctx context.Context, companyID string) (int, error) {
	const q = `SELECT COUNT(*) FROM funding_rounds WHERE company_id = $1`
	var n int
	if err := r.pool.QueryRow(ctx, q, companyID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count funding rounds: %w", err)
	}
	return n, nil
}

// ListByCompany returns a company's funding rounds, newest first.
func (r *FundingRepo) ListByCompany(ctx context.Context, companyID string) ([]*FundingRound, error) {
	const q = `SELECT id, company_id, round_name, amount_cents, pre_money_cents, equity_percent, sim_day, created_at
	           FROM funding_rounds WHERE company_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, companyID)
	if err != nil {
		return nil, fmt.Errorf("list funding rounds: %w", err)
	}
	defer rows.Close()
	var out []*FundingRound
	for rows.Next() {
		fr, err := scanFundingRound(rows)
		if err != nil {
			return nil, fmt.Errorf("scan funding round: %w", err)
		}
		out = append(out, fr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("funding rounds rows: %w", err)
	}
	return out, nil
}

type fundingScanner interface {
	Scan(dest ...any) error
}

func scanFundingRound(row fundingScanner) (*FundingRound, error) {
	var f FundingRound
	if err := row.Scan(&f.ID, &f.CompanyID, &f.RoundName, &f.AmountCents, &f.PreMoneyCents, &f.EquityPercent, &f.SimDay, &f.CreatedAt); err != nil {
		return nil, err
	}
	return &f, nil
}
