package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Achievement is an awarded milestone.
type Achievement struct {
	ID         string
	CompanyID  string
	Key        string
	AwardedDay int
	AwardedAt  time.Time
}

type AchievementRepo struct{ *Repository }

func NewAchievementRepo(base *Repository) *AchievementRepo {
	return &AchievementRepo{Repository: base}
}

// Award records an achievement, ignoring duplicates via ON CONFLICT.
func (r *AchievementRepo) Award(ctx context.Context, companyID, key string, day int) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO achievements (company_id, key, awarded_day) VALUES ($1, $2, $3)
		 ON CONFLICT (company_id, key) DO NOTHING`, companyID, key, day)
	if err != nil {
		return fmt.Errorf("award achievement: %w", err)
	}
	return nil
}

// ListByCompany returns awarded achievements newest-first.
func (r *AchievementRepo) ListByCompany(ctx context.Context, companyID string) ([]*Achievement, error) {
	const q = `SELECT id, company_id, key, awarded_day, awarded_at FROM achievements WHERE company_id = $1 ORDER BY awarded_day DESC`
	rows, err := r.pool.Query(ctx, q, companyID)
	if err != nil {
		return nil, fmt.Errorf("list achievements: %w", err)
	}
	defer rows.Close()
	var out []*Achievement
	for rows.Next() {
		var a Achievement
		if err := rows.Scan(&a.ID, &a.CompanyID, &a.Key, &a.AwardedDay, &a.AwardedAt); err != nil {
			if err == pgx.ErrNoRows {
				return nil, ErrNotFound
			}
			return nil, fmt.Errorf("scan achievement: %w", err)
		}
		out = append(out, &a)
	}
	return out, rows.Err()
}

// AwardedSet returns the set of awarded keys for dedup.
func (r *AchievementRepo) AwardedSet(ctx context.Context, companyID string) (map[string]bool, error) {
	list, err := r.ListByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}
	set := make(map[string]bool, len(list))
	for _, a := range list {
		set[a.Key] = true
	}
	return set, nil
}
