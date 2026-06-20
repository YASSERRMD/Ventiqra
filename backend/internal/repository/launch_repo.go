package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// ProductLaunch records a single product launch event.
type ProductLaunch struct {
	ID               string
	ProductID        string
	CompanyID        string
	Readiness        float64
	InitialCustomers int
	LaunchedAt       time.Time
}

// LaunchRepo provides persistence for product launch events.
type LaunchRepo struct {
	*Repository
}

// NewLaunchRepo creates a LaunchRepo over the shared pool.
func NewLaunchRepo(base *Repository) *LaunchRepo {
	return &LaunchRepo{Repository: base}
}

// Record inserts a launch event and returns the persisted row.
func (r *LaunchRepo) Record(ctx context.Context, productID, companyID string, readiness float64, initialCustomers int) (*ProductLaunch, error) {
	const q = `INSERT INTO product_launches (product_id, company_id, readiness, initial_customers)
	           VALUES ($1, $2, $3, $4)
	           RETURNING id, product_id, company_id, readiness, initial_customers, launched_at`
	l, err := scanLaunch(r.pool.QueryRow(ctx, q, productID, companyID, readiness, initialCustomers))
	if err != nil {
		return nil, fmt.Errorf("record launch: %w", err)
	}
	return l, nil
}

// ListByCompany returns a company's launch events, newest first.
func (r *LaunchRepo) ListByCompany(ctx context.Context, companyID string) ([]*ProductLaunch, error) {
	const q = `SELECT id, product_id, company_id, readiness, initial_customers, launched_at
	           FROM product_launches WHERE company_id = $1 ORDER BY launched_at DESC`
	return queryLaunches(r.pool.Query(ctx, q, companyID))
}

// ListByProduct returns a product's launch events, newest first.
func (r *LaunchRepo) ListByProduct(ctx context.Context, productID string) ([]*ProductLaunch, error) {
	const q = `SELECT id, product_id, company_id, readiness, initial_customers, launched_at
	           FROM product_launches WHERE product_id = $1 ORDER BY launched_at DESC`
	return queryLaunches(r.pool.Query(ctx, q, productID))
}

func queryLaunches(rows pgx.Rows, err error) ([]*ProductLaunch, error) {
	if err != nil {
		return nil, fmt.Errorf("query launches: %w", err)
	}
	defer rows.Close()
	var out []*ProductLaunch
	for rows.Next() {
		l, err := scanLaunch(rows)
		if err != nil {
			return nil, fmt.Errorf("scan launch: %w", err)
		}
		out = append(out, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("launches rows: %w", err)
	}
	return out, nil
}

type launchScanner interface {
	Scan(dest ...any) error
}

func scanLaunch(row launchScanner) (*ProductLaunch, error) {
	var l ProductLaunch
	if err := row.Scan(&l.ID, &l.ProductID, &l.CompanyID, &l.Readiness, &l.InitialCustomers, &l.LaunchedAt); err != nil {
		return nil, err
	}
	return &l, nil
}
