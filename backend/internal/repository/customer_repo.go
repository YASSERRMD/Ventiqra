package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// ProductCustomers is the persisted per-product customer state.
type ProductCustomers struct {
	ProductID    string
	CompanyID    string
	Total        int
	MAU          int
	Churned      int
	Satisfaction int
	UpdatedAt    time.Time
}

// CustomerRepo provides persistence for per-product customer state.
type CustomerRepo struct {
	*Repository
}

// NewCustomerRepo creates a CustomerRepo over the shared pool.
func NewCustomerRepo(base *Repository) *CustomerRepo {
	return &CustomerRepo{Repository: base}
}

// InitForLaunch creates the initial customer state for a freshly launched
// product. It is idempotent: if a row already exists it is left untouched.
func (r *CustomerRepo) InitForLaunch(ctx context.Context, productID, companyID string, total, satisfaction int) error {
	if satisfaction < 0 {
		satisfaction = 0
	}
	if satisfaction > 100 {
		satisfaction = 100
	}
	const q = `INSERT INTO product_customers (product_id, company_id, total_customers, mau, churned, satisfaction)
	           VALUES ($1, $2, $3, 0, 0, $4)
	           ON CONFLICT (product_id) DO NOTHING`
	_, err := r.pool.Exec(ctx, q, productID, companyID, total, satisfaction)
	if err != nil {
		return fmt.Errorf("init customers: %w", err)
	}
	return nil
}

// Get returns the customer state for a product, or ErrNotFound.
func (r *CustomerRepo) Get(ctx context.Context, productID string) (*ProductCustomers, error) {
	const q = `SELECT product_id, company_id, total_customers, mau, churned, satisfaction, updated_at
	           FROM product_customers WHERE product_id = $1`
	c, err := scanCustomers(r.pool.QueryRow(ctx, q, productID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get customers: %w", err)
	}
	return c, nil
}

// Save updates the customer state for a product.
func (r *CustomerRepo) Save(ctx context.Context, productID string, total, mau, churned, satisfaction int) error {
	const q = `UPDATE product_customers
	           SET total_customers = $2, mau = $3, churned = $4, satisfaction = $5
	           WHERE product_id = $1`
	tag, err := r.pool.Exec(ctx, q, productID, total, mau, churned, satisfaction)
	if err != nil {
		return fmt.Errorf("save customers: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListByCompany returns customer state for all of a company's products.
func (r *CustomerRepo) ListByCompany(ctx context.Context, companyID string) ([]*ProductCustomers, error) {
	const q = `SELECT product_id, company_id, total_customers, mau, churned, satisfaction, updated_at
	           FROM product_customers WHERE company_id = $1 ORDER BY product_id`
	rows, err := r.pool.Query(ctx, q, companyID)
	if err != nil {
		return nil, fmt.Errorf("list customers: %w", err)
	}
	defer rows.Close()
	var out []*ProductCustomers
	for rows.Next() {
		c, err := scanCustomers(rows)
		if err != nil {
			return nil, fmt.Errorf("scan customers: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("customers rows: %w", err)
	}
	return out, nil
}

type customerScanner interface {
	Scan(dest ...any) error
}

func scanCustomers(row customerScanner) (*ProductCustomers, error) {
	var c ProductCustomers
	if err := row.Scan(&c.ProductID, &c.CompanyID, &c.Total, &c.MAU, &c.Churned, &c.Satisfaction, &c.UpdatedAt); err != nil {
		return nil, err
	}
	return &c, nil
}
