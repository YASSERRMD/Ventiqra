package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Feature is a roadmap backlog item.
type Feature struct {
	ID          string
	CompanyID   string
	ProductID   *string
	Name        string
	Description string
	Priority    int
	Status      string
	Progress    int
	ValuePoints int
	StartedDay  *int
	ShippedDay  *int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// FeatureRepo provides persistence for roadmap features.
type FeatureRepo struct {
	*Repository
}

func NewFeatureRepo(base *Repository) *FeatureRepo { return &FeatureRepo{Repository: base} }

func (r *FeatureRepo) Create(ctx context.Context, f *Feature) (*Feature, error) {
	const q = `INSERT INTO features (company_id, product_id, name, description, priority, status, progress, value_points)
	           VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	           RETURNING id, company_id, product_id, name, description, priority, status, progress, value_points, started_day, shipped_day, created_at, updated_at`
	return scanFeature(r.pool.QueryRow(ctx, q,
		f.CompanyID, f.ProductID, f.Name, f.Description, f.Priority, f.Status, f.Progress, f.ValuePoints))
}

func (r *FeatureRepo) ListByCompany(ctx context.Context, companyID string) ([]*Feature, error) {
	const q = `SELECT id, company_id, product_id, name, description, priority, status, progress, value_points, started_day, shipped_day, created_at, updated_at
	           FROM features WHERE company_id = $1 ORDER BY priority DESC, created_at ASC`
	rows, err := r.pool.Query(ctx, q, companyID)
	if err != nil {
		return nil, fmt.Errorf("list features: %w", err)
	}
	defer rows.Close()
	var out []*Feature
	for rows.Next() {
		f, err := scanFeature(rows)
		if err != nil {
			return nil, fmt.Errorf("scan feature: %w", err)
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (r *FeatureRepo) Get(ctx context.Context, id string) (*Feature, error) {
	const q = `SELECT id, company_id, product_id, name, description, priority, status, progress, value_points, started_day, shipped_day, created_at, updated_at
	           FROM features WHERE id = $1`
	f, err := scanFeature(r.pool.QueryRow(ctx, q, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get feature: %w", err)
	}
	return f, nil
}

func (r *FeatureRepo) UpdateProgress(ctx context.Context, id string, progress int, status string, shippedDay *int) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE features SET progress = $2, status = $3, shipped_day = COALESCE($4, shipped_day), updated_at = NOW() WHERE id = $1`,
		id, progress, status, shippedDay)
	if err != nil {
		return fmt.Errorf("update feature progress: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *FeatureRepo) Delete(ctx context.Context, id, companyID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM features WHERE id = $1 AND company_id = $2`, id, companyID)
	if err != nil {
		return fmt.Errorf("delete feature: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type featureScanner interface {
	Scan(dest ...any) error
}

func scanFeature(row featureScanner) (*Feature, error) {
	var f Feature
	if err := row.Scan(
		&f.ID, &f.CompanyID, &f.ProductID, &f.Name, &f.Description, &f.Priority,
		&f.Status, &f.Progress, &f.ValuePoints, &f.StartedDay, &f.ShippedDay,
		&f.CreatedAt, &f.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &f, nil
}
