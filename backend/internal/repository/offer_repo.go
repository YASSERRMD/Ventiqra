package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// OfferStatus describes the lifecycle state of an investor offer.
type OfferStatus string

const (
	OfferPending   OfferStatus = "pending"
	OfferAccepted  OfferStatus = "accepted"
	OfferRejected  OfferStatus = "rejected"
	OfferWithdrawn OfferStatus = "withdrawn"
)

// InvestorOffer is a negotiable funding proposal.
type InvestorOffer struct {
	ID            string
	CompanyID     string
	InvestorName  string
	AmountCents   int64
	EquityPercent float64
	Status        OfferStatus
	RoundSeed     int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// OfferRepo provides persistence for investor offers.
type OfferRepo struct {
	*Repository
}

// NewOfferRepo creates an OfferRepo over the shared pool.
func NewOfferRepo(base *Repository) *OfferRepo {
	return &OfferRepo{Repository: base}
}

// ClearPending deletes all pending offers for a company (used when soliciting a
// fresh batch).
func (r *OfferRepo) ClearPending(ctx context.Context, companyID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM investor_offers WHERE company_id = $1 AND status = 'pending'`, companyID)
	if err != nil {
		return fmt.Errorf("clear pending offers: %w", err)
	}
	return nil
}

// Insert stores a new pending offer and returns it.
func (r *OfferRepo) Insert(ctx context.Context, o *InvestorOffer) (*InvestorOffer, error) {
	const q = `INSERT INTO investor_offers (company_id, investor_name, amount_cents, equity_percent, status, round_seed)
	           VALUES ($1, $2, $3, $4, 'pending', $5)
	           RETURNING id, company_id, investor_name, amount_cents, equity_percent, status, round_seed, created_at, updated_at`
	return scanOffer(r.pool.QueryRow(ctx, q, o.CompanyID, o.InvestorName, o.AmountCents, o.EquityPercent, o.RoundSeed))
}

// Get returns an offer by id or ErrNotFound.
func (r *OfferRepo) Get(ctx context.Context, id string) (*InvestorOffer, error) {
	const q = `SELECT id, company_id, investor_name, amount_cents, equity_percent, status, round_seed, created_at, updated_at
	           FROM investor_offers WHERE id = $1`
	o, err := scanOffer(r.pool.QueryRow(ctx, q, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get offer: %w", err)
	}
	return o, nil
}

// UpdateStatus sets an offer's status.
func (r *OfferRepo) UpdateStatus(ctx context.Context, id string, status OfferStatus) error {
	const q = `UPDATE investor_offers SET status = $2 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q, id, status)
	if err != nil {
		return fmt.Errorf("update offer status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateEquity sets an offer's asked equity (used when a negotiation succeeds).
func (r *OfferRepo) UpdateEquity(ctx context.Context, id string, equity float64) error {
	const q = `UPDATE investor_offers SET equity_percent = $2 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q, id, equity)
	if err != nil {
		return fmt.Errorf("update offer equity: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListPendingByCompany returns a company's pending offers.
func (r *OfferRepo) ListPendingByCompany(ctx context.Context, companyID string) ([]*InvestorOffer, error) {
	const q = `SELECT id, company_id, investor_name, amount_cents, equity_percent, status, round_seed, created_at, updated_at
	           FROM investor_offers WHERE company_id = $1 AND status = 'pending' ORDER BY created_at ASC`
	return queryOffers(r.pool.Query(ctx, q, companyID))
}

func queryOffers(rows pgx.Rows, err error) ([]*InvestorOffer, error) {
	if err != nil {
		return nil, fmt.Errorf("query offers: %w", err)
	}
	defer rows.Close()
	var out []*InvestorOffer
	for rows.Next() {
		o, err := scanOffer(rows)
		if err != nil {
			return nil, fmt.Errorf("scan offer: %w", err)
		}
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("offers rows: %w", err)
	}
	return out, nil
}

type offerScanner interface {
	Scan(dest ...any) error
}

func scanOffer(row offerScanner) (*InvestorOffer, error) {
	var o InvestorOffer
	if err := row.Scan(&o.ID, &o.CompanyID, &o.InvestorName, &o.AmountCents, &o.EquityPercent, &o.Status, &o.RoundSeed, &o.CreatedAt, &o.UpdatedAt); err != nil {
		return nil, err
	}
	return &o, nil
}
