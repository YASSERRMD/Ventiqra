package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// LeaderboardEntry is a finalized run's score.
type LeaderboardEntry struct {
	ID            string
	OwnerID       string
	CompanyName   string
	Score         int64
	DaysSurvived  int
	PeakValuation int64
	Outcome       string
	CreatedAt     time.Time
}

type LeaderboardRepo struct{ *Repository }

func NewLeaderboardRepo(base *Repository) *LeaderboardRepo {
	return &LeaderboardRepo{Repository: base}
}

func (r *LeaderboardRepo) Record(ctx context.Context, e *LeaderboardEntry) (*LeaderboardEntry, error) {
	const q = `INSERT INTO leaderboard (owner_id, company_name, score, days_survived, peak_valuation, outcome)
	           VALUES ($1, $2, $3, $4, $5, $6)
	           RETURNING id, owner_id, company_name, score, days_survived, peak_valuation, outcome, created_at`
	row := r.pool.QueryRow(ctx, q, e.OwnerID, e.CompanyName, e.Score, e.DaysSurvived, e.PeakValuation, e.Outcome)
	var out LeaderboardEntry
	if err := row.Scan(&out.ID, &out.OwnerID, &out.CompanyName, &out.Score, &out.DaysSurvived, &out.PeakValuation, &out.Outcome, &out.CreatedAt); err != nil {
		return nil, fmt.Errorf("record leaderboard: %w", err)
	}
	return &out, nil
}

// Top returns the top N entries globally (local leaderboard), highest score first.
func (r *LeaderboardRepo) Top(ctx context.Context, limit int) ([]*LeaderboardEntry, error) {
	if limit <= 0 {
		limit = 20
	}
	const q = `SELECT id, owner_id, company_name, score, days_survived, peak_valuation, outcome, created_at
	           FROM leaderboard ORDER BY score DESC LIMIT $1`
	rows, err := r.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("leaderboard top: %w", err)
	}
	defer rows.Close()
	var out []*LeaderboardEntry
	for rows.Next() {
		var e LeaderboardEntry
		if err := rows.Scan(&e.ID, &e.OwnerID, &e.CompanyName, &e.Score, &e.DaysSurvived, &e.PeakValuation, &e.Outcome, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan leaderboard: %w", err)
		}
		out = append(out, &e)
	}
	return out, rows.Err()
}

// HasEntryForCompany reports whether a finalized entry already exists for a
// company name + owner (to avoid duplicate records). Uses a soft match.
func (r *LeaderboardRepo) HasEntryForCompany(ctx context.Context, ownerID, companyName string) (bool, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM leaderboard WHERE owner_id = $1 AND company_name = $2`, ownerID, companyName).Scan(&n)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("check leaderboard: %w", err)
	}
	return n > 0, nil
}
