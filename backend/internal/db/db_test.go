package db

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/YASSERRMD/Ventiqra/backend/internal/testutil"
)

// testPool returns a pool connected to the DATABASE_TEST_URL (or DATABASE_URL)
// when a database is reachable, or skips the test otherwise.
func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	t.Cleanup(testutil.LockDB())
	dsn := os.Getenv("DATABASE_TEST_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		dsn = "postgres://ventiqra:changeme@localhost:5432/ventiqra?sslmode=connect_timeout=2"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := Connect(ctx, dsn, DefaultPoolConfig())
	if err != nil {
		t.Skipf("database not available, skipping integration test: %v", err)
	}
	t.Cleanup(func() { Close(pool) })
	return pool
}

func TestParseConfigRejectsInvalidDSN(t *testing.T) {
	ctx := context.Background()
	_, err := Connect(ctx, "not a valid dsn", DefaultPoolConfig())
	if err == nil {
		t.Fatal("expected error for invalid DSN")
	}
}

func TestConnectAndPingIntegration(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	if err := Ping(ctx, pool); err != nil {
		t.Fatalf("ping failed: %v", err)
	}
}

func TestCloseOnNilIsSafe(t *testing.T) {
	Close(nil) // must not panic
}
