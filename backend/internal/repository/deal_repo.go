package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Deal is a B2B sales opportunity in the pipeline.
type Deal struct {
	ID         string
	CompanyID  string
	Name       string
	Stage      string
	ValueCents int64
	Probability int
	ClosedWon  bool
	CreatedDay int
	ClosedDay  *int
	CreatedAt  time.Time
}

type DealRepo struct{ *Repository }

func NewDealRepo(base *Repository) *DealRepo { return &DealRepo{Repository: base} }

func (r *DealRepo) Create(ctx context.Context, d *Deal) (*Deal, error) {
	const q = `INSERT INTO deals (company_id, name, stage, value_cents, probability, created_day)
	           VALUES ($1, $2, $3, $4, $5, $6)
	           RETURNING id, company_id, name, stage, value_cents, probability, closed_won, created_day, closed_day, created_at`
	return scanDeal(r.pool.QueryRow(ctx, q, d.CompanyID, d.Name, d.Stage, d.ValueCents, d.Probability, d.CreatedDay))
}

func (r *DealRepo) ListByCompany(ctx context.Context, companyID string) ([]*Deal, error) {
	const q = `SELECT id, company_id, name, stage, value_cents, probability, closed_won, created_day, closed_day, created_at
	           FROM deals WHERE company_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, companyID)
	if err != nil {
		return nil, fmt.Errorf("list deals: %w", err)
	}
	defer rows.Close()
	var out []*Deal
	for rows.Next() {
		d, err := scanDeal(rows)
		if err != nil {
			return nil, fmt.Errorf("scan deal: %w", err)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *DealRepo) Get(ctx context.Context, id string) (*Deal, error) {
	const q = `SELECT id, company_id, name, stage, value_cents, probability, closed_won, created_day, closed_day, created_at
	           FROM deals WHERE id = $1`
	d, err := scanDeal(r.pool.QueryRow(ctx, q, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get deal: %w", err)
	}
	return d, nil
}

func (r *DealRepo) UpdateStage(ctx context.Context, id, stage string, probability int, closedWon bool, closedDay *int) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE deals SET stage = $2, probability = $3, closed_won = $4, closed_day = $5 WHERE id = $1`,
		id, stage, probability, closedWon, closedDay)
	if err != nil {
		return fmt.Errorf("update deal stage: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *DealRepo) ListOpen(ctx context.Context, companyID string) ([]*Deal, error) {
	const q = `SELECT id, company_id, name, stage, value_cents, probability, closed_won, created_day, closed_day, created_at
	           FROM deals WHERE company_id = $1 AND stage NOT IN ('closed_won','closed_lost') ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, q, companyID)
	if err != nil {
		return nil, fmt.Errorf("list open deals: %w", err)
	}
	defer rows.Close()
	var out []*Deal
	for rows.Next() {
		d, err := scanDeal(rows)
		if err != nil {
			return nil, fmt.Errorf("scan deal: %w", err)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

type dealScanner interface {
	Scan(dest ...any) error
}

func scanDeal(row dealScanner) (*Deal, error) {
	var d Deal
	if err := row.Scan(&d.ID, &d.CompanyID, &d.Name, &d.Stage, &d.ValueCents, &d.Probability, &d.ClosedWon, &d.CreatedDay, &d.ClosedDay, &d.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &d, nil
}
