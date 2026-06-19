package db

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func testPoolForMigrations(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("DATABASE_TEST_URL")
	if dsn == "" {
		dsn = "postgres://ventiqra:changeme@localhost:5432/ventiqra?sslmode=disable"
	}
	ctx := context.Background()
	pool, err := Connect(ctx, dsn, DefaultPoolConfig())
	if err != nil {
		t.Skipf("database not available, skipping: %v", err)
	}
	t.Cleanup(func() { Close(pool) })

	// Start from a clean slate for deterministic migration tests.
	if _, err := pool.Exec(ctx, `DROP TABLE IF EXISTS schema_migrations`); err != nil {
		t.Fatalf("drop schema_migrations: %v", err)
	}
	return pool
}

func TestParseMigrationName(t *testing.T) {
	cases := []struct {
		in       string
		wantV    int
		wantName string
		wantOK   bool
	}{
		{"0001_init.sql", 1, "init", true},
		{"0042_add_users_table.sql", 42, "add_users_table", true},
		{"nope.sql", 0, "", false},
		{"abc_init.sql", 0, "", false},
		{"0001.sql", 0, "", false},
	}
	for _, c := range cases {
		v, name, ok := parseMigrationName(c.in)
		if v != c.wantV || name != c.wantName || ok != c.wantOK {
			t.Errorf("parseMigrationName(%q) = (%d,%q,%v), want (%d,%q,%v)",
				c.in, v, name, ok, c.wantV, c.wantName, c.wantOK)
		}
	}
}

func TestLoadMigrationsIsSortedAndNonEmpty(t *testing.T) {
	ms, err := LoadMigrations()
	if err != nil {
		t.Fatalf("LoadMigrations: %v", err)
	}
	if len(ms) == 0 {
		t.Fatal("expected at least one migration")
	}
	for i := 1; i < len(ms); i++ {
		if ms[i].Version <= ms[i-1].Version {
			t.Errorf("migrations not sorted at %d", i)
		}
	}
}

func TestMigrateAppliesAndIsIdempotent(t *testing.T) {
	pool := testPoolForMigrations(t)
	ctx := context.Background()

	n, err := Migrate(ctx, pool)
	if err != nil {
		t.Fatalf("first migrate: %v", err)
	}
	if n == 0 {
		t.Fatal("expected at least one migration applied")
	}

	// Second run should apply nothing.
	n2, err := Migrate(ctx, pool)
	if err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	if n2 != 0 {
		t.Errorf("second migrate applied %d, want 0", n2)
	}

	version, err := CurrentVersion(ctx, pool)
	if err != nil {
		t.Fatalf("current version: %v", err)
	}
	if version == 0 {
		t.Error("expected non-zero current version after migrate")
	}

	// The init migration should have created the pgcrypto extension.
	var ext string
	err = pool.QueryRow(ctx,
		`SELECT extname FROM pg_extension WHERE extname = 'pgcrypto'`).Scan(&ext)
	if err != nil || ext != "pgcrypto" {
		t.Errorf("pgcrypto extension missing: %v", err)
	}
}
