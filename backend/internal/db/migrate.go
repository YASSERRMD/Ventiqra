package db

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/YASSERRMD/Ventiqra/backend/migrations"
)

// Migration is a single parsed migration with its version, name, and SQL body.
type Migration struct {
	Version int
	Name    string
	SQL     string
}

const migrationsTableSQL = `CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    name    TEXT NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`

// LoadMigrations reads embedded migration SQL files and returns them sorted by
// version ascending.
func LoadMigrations() ([]Migration, error) {
	return loadFromFS(migrations.MigrationsFS, ".")
}

// LoadSeeds reads embedded seed SQL files and returns them sorted by version.
func LoadSeeds() ([]Migration, error) {
	return loadFromFS(migrations.SeedsFS, "seeds")
}

func loadFromFS(fsys fs.FS, root string) ([]Migration, error) {
	entries, err := fs.ReadDir(fsys, root)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	var out []Migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		version, name, ok := parseMigrationName(e.Name())
		if !ok {
			return nil, fmt.Errorf("unparseable migration filename %q", e.Name())
		}
		body, err := fs.ReadFile(fsys, path.Join(root, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}
		out = append(out, Migration{Version: version, Name: name, SQL: string(body)})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out, nil
}

// parseMigrationName extracts the leading integer version and the descriptive
// name from a filename like "0001_init_users.sql".
func parseMigrationName(filename string) (version int, name string, ok bool) {
	base := strings.TrimSuffix(filename, ".sql")
	parts := strings.SplitN(base, "_", 2)
	if len(parts) != 2 || parts[0] == "" {
		return 0, "", false
	}
	v, err := strconv.Atoi(parts[0])
	if err != nil || v < 0 {
		return 0, "", false
	}
	return v, parts[1], true
}

// CurrentVersion returns the highest applied migration version, or 0 if none.
func CurrentVersion(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	if _, err := pool.Exec(ctx, migrationsTableSQL); err != nil {
		return 0, fmt.Errorf("ensure schema_migrations: %w", err)
	}
	var version int
	err := pool.QueryRow(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("read current version: %w", err)
	}
	return version, nil
}

// Migrate applies all pending migrations in version order. Each migration runs
// in its own transaction and is recorded in schema_migrations.
func Migrate(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	all, err := LoadMigrations()
	if err != nil {
		return 0, err
	}
	current, err := CurrentVersion(ctx, pool)
	if err != nil {
		return 0, err
	}

	applied := 0
	for _, m := range all {
		if m.Version <= current {
			continue
		}
		if err := applyMigration(ctx, pool, m); err != nil {
			return applied, fmt.Errorf("migration %d (%s): %w", m.Version, m.Name, err)
		}
		applied++
	}
	return applied, nil
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, m Migration) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, m.SQL); err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)
		 ON CONFLICT (version) DO NOTHING`, m.Version, m.Name); err != nil {
		return fmt.Errorf("record: %w", err)
	}
	return tx.Commit(ctx)
}

// Seed applies all seed files in version order within a single transaction.
// It returns the number of seed files applied.
func Seed(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	all, err := LoadSeeds()
	if err != nil {
		return 0, err
	}
	if len(all) == 0 {
		return 0, nil
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin seed tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for i, s := range all {
		if _, err := tx.Exec(ctx, s.SQL); err != nil {
			return i, fmt.Errorf("seed %d (%s): %w", s.Version, s.Name, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit seeds: %w", err)
	}
	return len(all), nil
}

// WithTx is a helper that runs fn inside a transaction, committing on nil error.
func WithTx(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}
