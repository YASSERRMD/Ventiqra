package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// SupportState is a company's support-ticket backlog state.
type SupportState struct {
	CompanyID     string
	OpenTickets   int
	ResolvedTotal int
	UpdatedAt     time.Time
}

type SupportRepo struct{ *Repository }

func NewSupportRepo(base *Repository) *SupportRepo { return &SupportRepo{Repository: base} }

func (r *SupportRepo) GetOrCreate(ctx context.Context, companyID string) (*SupportState, error) {
	if _, err := r.pool.Exec(ctx,
		`INSERT INTO support_state (company_id) VALUES ($1) ON CONFLICT (company_id) DO NOTHING`, companyID); err != nil {
		return nil, fmt.Errorf("ensure support state: %w", err)
	}
	const q = `SELECT company_id, open_tickets, resolved_total, updated_at FROM support_state WHERE company_id = $1`
	var ss SupportState
	err := r.pool.QueryRow(ctx, q, companyID).Scan(&ss.CompanyID, &ss.OpenTickets, &ss.ResolvedTotal, &ss.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get support state: %w", err)
	}
	return &ss, nil
}

// ApplyDay persists the day's resolution: sets open tickets and adds to the
// resolved total.
func (r *SupportRepo) ApplyDay(ctx context.Context, companyID string, newOpen, resolvedToday int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE support_state SET open_tickets = $2, resolved_total = resolved_total + $3, updated_at = NOW() WHERE company_id = $1`,
		companyID, newOpen, resolvedToday)
	if err != nil {
		return fmt.Errorf("apply support day: %w", err)
	}
	return nil
}
