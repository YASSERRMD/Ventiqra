package repository

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/YASSERRMD/Ventiqra/backend/internal/db"
	"github.com/YASSERRMD/Ventiqra/backend/internal/testutil"
)

func baseRepo(t *testing.T) *Repository {
	t.Helper()
	t.Cleanup(testutil.LockDB())
	dsn := os.Getenv("DATABASE_TEST_URL")
	if dsn == "" {
		dsn = "postgres://ventiqra:changeme@localhost:5432/ventiqra?sslmode=disable"
	}
	pool, err := db.Connect(context.Background(), dsn, db.DefaultPoolConfig())
	if err != nil {
		t.Skipf("database not available, skipping: %v", err)
	}
	t.Cleanup(func() { db.Close(pool) })
	return New(pool)
}

func TestNewReturnsRepositoryWithPool(t *testing.T) {
	r := &Repository{pool: (*pgxpool.Pool)(nil)}
	if r.Pool() != nil {
		t.Error("expected nil pool")
	}
}

func TestErrNotFoundSentinel(t *testing.T) {
	if !IsNotFound(ErrNotFound) {
		t.Error("IsNotFound(ErrNotFound) = false")
	}
	if IsNotFound(errors.New("other")) {
		t.Error("IsNotFound(other) = true")
	}
}

func TestSystemRepoPostgresVersionIntegration(t *testing.T) {
	base := baseRepo(t)
	repo := NewSystemRepo(base)

	version, err := repo.PostgresVersion(context.Background())
	if err != nil {
		t.Fatalf("PostgresVersion: %v", err)
	}
	if version == "" {
		t.Error("expected non-empty version string")
	}
}

func TestSystemRepoMigrationCountIntegration(t *testing.T) {
	base := baseRepo(t)
	repo := NewSystemRepo(base)
	ctx := context.Background()

	// Ensure schema is migrated so schema_migrations exists.
	if _, err := db.Migrate(ctx, base.Pool()); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	count, err := repo.AppliedMigrationCount(ctx)
	if err != nil {
		t.Fatalf("AppliedMigrationCount: %v", err)
	}
	if count < 1 {
		t.Errorf("expected >=1 applied migration, got %d", count)
	}
}

func TestSystemRepoNowIntegration(t *testing.T) {
	base := baseRepo(t)
	repo := NewSystemRepo(base)

	now, err := repo.Now(context.Background())
	if err != nil {
		t.Fatalf("Now: %v", err)
	}
	if now.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}
