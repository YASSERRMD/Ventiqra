package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// TimelineEvent is a single chronological history entry for a company.
type TimelineEvent struct {
	ID          string
	CompanyID   string
	Kind        string
	Title       string
	Description string
	SimDay      int
	CreatedAt   time.Time
}

// TimelineRepo provides persistence for the unified company timeline.
type TimelineRepo struct {
	*Repository
}

// NewTimelineRepo creates a TimelineRepo over the shared pool.
func NewTimelineRepo(base *Repository) *TimelineRepo {
	return &TimelineRepo{Repository: base}
}

// Record appends a timeline entry and returns it.
func (r *TimelineRepo) Record(ctx context.Context, e *TimelineEvent) (*TimelineEvent, error) {
	const q = `INSERT INTO timeline_events (company_id, kind, title, description, sim_day)
	           VALUES ($1, $2, $3, $4, $5)
	           RETURNING id, company_id, kind, title, description, sim_day, created_at`
	return scanTimelineEvent(r.pool.QueryRow(ctx, q,
		e.CompanyID, e.Kind, e.Title, e.Description, e.SimDay))
}

// ListByCompany returns a company's timeline entries newest-first, capped at limit.
func (r *TimelineRepo) ListByCompany(ctx context.Context, companyID string, limit int) ([]*TimelineEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	const q = `SELECT id, company_id, kind, title, description, sim_day, created_at
	           FROM timeline_events WHERE company_id = $1 ORDER BY sim_day DESC, created_at DESC LIMIT $2`
	rows, err := r.pool.Query(ctx, q, companyID, limit)
	if err != nil {
		return nil, fmt.Errorf("list timeline: %w", err)
	}
	defer rows.Close()
	var out []*TimelineEvent
	for rows.Next() {
		e, err := scanTimelineEvent(rows)
		if err != nil {
			return nil, fmt.Errorf("scan timeline: %w", err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("timeline rows: %w", err)
	}
	return out, nil
}

// CountInDayRange returns how many timeline entries fall in [startDay, endDay].
func (r *TimelineRepo) CountInDayRange(ctx context.Context, companyID string, startDay, endDay int) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM timeline_events WHERE company_id = $1 AND sim_day >= $2 AND sim_day <= $3`,
		companyID, startDay, endDay).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count timeline: %w", err)
	}
	return n, nil
}

type timelineScanner interface {
	Scan(dest ...any) error
}

func scanTimelineEvent(row timelineScanner) (*TimelineEvent, error) {
	var e TimelineEvent
	if err := row.Scan(&e.ID, &e.CompanyID, &e.Kind, &e.Title, &e.Description, &e.SimDay, &e.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &e, nil
}
