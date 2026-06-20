package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ProductStage describes the lifecycle state of a product.
type ProductStage string

const (
	ProductIdea     ProductStage = "idea"
	ProductBuilding ProductStage = "building"
	ProductLaunched ProductStage = "launched"
	ProductRetired  ProductStage = "retired"
)

// ValidProductStages is the set of accepted product stage values.
var ValidProductStages = map[ProductStage]bool{
	ProductIdea: true, ProductBuilding: true, ProductLaunched: true, ProductRetired: true,
}

// Product is the application-level product model. DevProgress is a 0..100
// percentage and PriceCents is nil when no price has been set.
type Product struct {
	ID          string
	CompanyID   string
	Name        string
	Slug        string
	Stage       ProductStage
	DevProgress float64
	PriceCents  *int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ProductRepo provides persistence for products.
type ProductRepo struct {
	*Repository
}

// NewProductRepo creates a ProductRepo over the shared pool.
func NewProductRepo(base *Repository) *ProductRepo {
	return &ProductRepo{Repository: base}
}

// CreateProduct inserts a new product for the given company, generating a slug
// unique within the company from the name. The product starts in the 'idea'
// stage with zero development progress and no price.
func (r *ProductRepo) CreateProduct(ctx context.Context, companyID, name string) (*Product, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("product name is required")
	}
	if companyID == "" {
		return nil, errors.New("company id is required")
	}

	slug, err := r.uniqueProductSlug(ctx, companyID, name)
	if err != nil {
		return nil, err
	}

	const q = `INSERT INTO products (company_id, name, slug)
	           VALUES ($1, $2, $3)
	           RETURNING id, company_id, name, slug, stage, dev_progress, price_cents, created_at, updated_at`

	p, err := scanProduct(r.pool.QueryRow(ctx, q, companyID, name, slug))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("create product: %w", err)
	}
	return p, nil
}

// GetProduct returns a product by id or ErrNotFound.
func (r *ProductRepo) GetProduct(ctx context.Context, id string) (*Product, error) {
	const q = `SELECT id, company_id, name, slug, stage, dev_progress, price_cents, created_at, updated_at
	           FROM products WHERE id = $1`
	p, err := scanProduct(r.pool.QueryRow(ctx, q, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get product: %w", err)
	}
	return p, nil
}

// ListProductsByCompany returns the products for a company ordered by creation
// time ascending.
func (r *ProductRepo) ListProductsByCompany(ctx context.Context, companyID string) ([]*Product, error) {
	const q = `SELECT id, company_id, name, slug, stage, dev_progress, price_cents, created_at, updated_at
	           FROM products WHERE company_id = $1 ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, q, companyID)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	var out []*Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list products rows: %w", err)
	}
	return out, nil
}

// UpdateStage sets the lifecycle stage of a product.
func (r *ProductRepo) UpdateStage(ctx context.Context, id string, stage ProductStage) error {
	if !ValidProductStages[stage] {
		return fmt.Errorf("invalid product stage: %q", stage)
	}
	return r.updateColumn(ctx, id, `UPDATE products SET stage = $2 WHERE id = $1`, string(stage))
}

// UpdateProgress sets the development progress, clamped to the 0..100 range.
func (r *ProductRepo) UpdateProgress(ctx context.Context, id string, progress float64) error {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	return r.updateColumn(ctx, id, `UPDATE products SET dev_progress = $2 WHERE id = $1`, progress)
}

// SetPrice sets the price in cents for a product.
func (r *ProductRepo) SetPrice(ctx context.Context, id string, priceCents int64) error {
	return r.updateColumn(ctx, id, `UPDATE products SET price_cents = $2 WHERE id = $1`, priceCents)
}

// updateColumn runs a single-column update that keys on id, returning
// ErrNotFound when no row matched.
func (r *ProductRepo) updateColumn(ctx context.Context, id, q string, arg any) error {
	tag, err := r.pool.Exec(ctx, q, id, arg)
	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// uniqueProductSlug returns a slug guaranteed not to collide with another
// product slug within the same company, appending a short numeric suffix when
// necessary.
func (r *ProductRepo) uniqueProductSlug(ctx context.Context, companyID, name string) (string, error) {
	base := slugify(name)
	if base == "" {
		base = "product"
	}
	candidate := base
	for i := 1; i < 1000; i++ {
		var exists bool
		err := r.pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM products WHERE company_id = $1 AND slug = $2)`,
			companyID, candidate).Scan(&exists)
		if err != nil {
			return "", fmt.Errorf("check product slug: %w", err)
		}
		if !exists {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, i)
	}
	return "", errors.New("could not allocate unique product slug")
}

type productScanner interface {
	Scan(dest ...any) error
}

func scanProduct(row productScanner) (*Product, error) {
	var p Product
	if err := row.Scan(&p.ID, &p.CompanyID, &p.Name, &p.Slug, &p.Stage,
		&p.DevProgress, &p.PriceCents, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	return &p, nil
}
