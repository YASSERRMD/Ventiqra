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

// CompanyStatus describes the lifecycle state of a company.
type CompanyStatus string

const (
	CompanyActive   CompanyStatus = "active"
	CompanyBankrupt CompanyStatus = "bankrupt"
	CompanyClosed   CompanyStatus = "closed"
)

// Company is the application-level company model. Cash is in cents.
type Company struct {
	ID          string
	OwnerID     string
	Name        string
	Slug        string
	Industry    string
	Description string
	FoundedAt   time.Time
	Cash        int64
	Status      CompanyStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CompanyRepo provides persistence for companies.
type CompanyRepo struct {
	*Repository
}

// NewCompanyRepo creates a CompanyRepo over the shared pool.
func NewCompanyRepo(base *Repository) *CompanyRepo {
	return &CompanyRepo{Repository: base}
}

// CreateCompany inserts a company, generating a unique slug from the name.
func (r *CompanyRepo) CreateCompany(ctx context.Context, c *Company) (*Company, error) {
	if c.Name == "" {
		return nil, errors.New("company name is required")
	}
	c.Slug = strings.TrimSpace(c.Slug)
	if c.Slug == "" {
		c.Slug = slugify(c.Name)
	}
	if c.Status == "" {
		c.Status = CompanyActive
	}

	slug, err := r.UniqueSlug(ctx, c.Slug)
	if err != nil {
		return nil, err
	}
	c.Slug = slug

	const q = `INSERT INTO companies (owner_id, name, slug, industry, description, founded_at, cash, status)
	           VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	           RETURNING id, owner_id, name, slug, industry, description, founded_at, cash, status, created_at, updated_at`

	row := r.pool.QueryRow(ctx, q,
		c.OwnerID, c.Name, c.Slug, c.Industry, c.Description, c.FoundedAt, c.Cash, c.Status)

	created, err := scanCompany(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("create company: %w", err)
	}
	return created, nil
}

// GetCompany returns a company by id or ErrNotFound.
func (r *CompanyRepo) GetCompany(ctx context.Context, id string) (*Company, error) {
	const q = `SELECT id, owner_id, name, slug, industry, description, founded_at, cash, status, created_at, updated_at
	           FROM companies WHERE id = $1`
	c, err := scanCompany(r.pool.QueryRow(ctx, q, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get company: %w", err)
	}
	return c, nil
}

// GetLatestCompanyForOwner returns the owner's most recently created company.
func (r *CompanyRepo) GetLatestCompanyForOwner(ctx context.Context, ownerID string) (*Company, error) {
	const q = `SELECT id, owner_id, name, slug, industry, description, founded_at, cash, status, created_at, updated_at
	           FROM companies WHERE owner_id = $1 ORDER BY created_at DESC LIMIT 1`
	c, err := scanCompany(r.pool.QueryRow(ctx, q, ownerID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get latest company: %w", err)
	}
	return c, nil
}

// UniqueSlug returns a slug guaranteed not to collide with an existing company
// slug, appending a short numeric suffix when necessary.
func (r *CompanyRepo) UniqueSlug(ctx context.Context, base string) (string, error) {
	base = slugify(base)
	if base == "" {
		base = "company"
	}
	candidate := base
	for i := 1; i < 1000; i++ {
		var exists bool
		err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM companies WHERE slug = $1)`, candidate).Scan(&exists)
		if err != nil {
			return "", fmt.Errorf("check slug: %w", err)
		}
		if !exists {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, i)
	}
	return "", errors.New("could not allocate unique slug")
}

// UpdateCash sets the company's cash balance.
func (r *CompanyRepo) UpdateCash(ctx context.Context, id string, cash int64) error {
	const q = `UPDATE companies SET cash = $2 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q, id, cash)
	if err != nil {
		return fmt.Errorf("update cash: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateStatus sets the company's lifecycle status.
func (r *CompanyRepo) UpdateStatus(ctx context.Context, id string, status CompanyStatus) error {
	const q = `UPDATE companies SET status = $2 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q, id, status)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type companyScanner interface {
	Scan(dest ...any) error
}

func scanCompany(row companyScanner) (*Company, error) {
	var c Company
	err := row.Scan(&c.ID, &c.OwnerID, &c.Name, &c.Slug, &c.Industry, &c.Description,
		&c.FoundedAt, &c.Cash, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// slugify converts a name into a URL-safe slug.
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevDash := true
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash {
				b.WriteRune('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
