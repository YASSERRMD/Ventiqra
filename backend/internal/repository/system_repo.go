package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// SystemRepo exposes low-level database introspection queries. It exists in the
// foundation phase to validate the repository pattern end-to-end.
type SystemRepo struct {
	*Repository
}

// NewSystemRepo creates a SystemRepo over the shared pool.
func NewSystemRepo(base *Repository) *SystemRepo {
	return &SystemRepo{Repository: base}
}

// PostgresVersion returns the server's version string.
func (s *SystemRepo) PostgresVersion(ctx context.Context) (string, error) {
	var version string
	err := s.pool.QueryRow(ctx, `SELECT version()`).Scan(&version)
	if err != nil {
		return "", fmt.Errorf("system version: %w", err)
	}
	return version, nil
}

// AppliedMigrationCount returns the number of rows in schema_migrations. It
// assumes migrations have been applied (i.e. the table exists).
func (s *SystemRepo) AppliedMigrationCount(ctx context.Context) (int, error) {
	const q = `SELECT count(*) FROM schema_migrations`
	var count int
	err := s.pool.QueryRow(ctx, q).Scan(&count)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("migration count: %w", err)
	}
	return count, nil
}

// Now returns the database server's current timestamp.
func (s *SystemRepo) Now(ctx context.Context) (time.Time, error) {
	var t time.Time
	err := s.pool.QueryRow(ctx, `SELECT NOW()`).Scan(&t)
	return t, err
}
