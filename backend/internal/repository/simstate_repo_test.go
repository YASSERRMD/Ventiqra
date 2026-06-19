package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/YASSERRMD/Ventiqra/backend/internal/db"
	"github.com/YASSERRMD/Ventiqra/backend/internal/testutil"
)

func simStateRepoForTest(t *testing.T) (*SimStateRepo, *CompanyRepo, *UserRepo) {
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

	ctx := context.Background()
	if _, err := pool.Exec(ctx, `DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;`); err != nil {
		t.Fatalf("reset schema: %v", err)
	}
	if _, err := db.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	base := New(pool)
	return NewSimStateRepo(base), NewCompanyRepo(base), NewUserRepo(base)
}

func TestSimStateGetMissing(t *testing.T) {
	repo, _, _ := simStateRepoForTest(t)
	_, err := repo.Get(context.Background(), "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSimStateInitAndGet(t *testing.T) {
	repo, companies, users := simStateRepoForTest(t)
	ctx := context.Background()

	owner := mustSeedOwner(t, users)
	c, err := companies.CreateCompany(ctx, &Company{
		OwnerID:   owner,
		Name:      "Sim Co",
		FoundedAt: time.Now(),
		Cash:      1_000_00,
	})
	if err != nil {
		t.Fatalf("create company: %v", err)
	}

	state, err := repo.Init(ctx, c.ID, 4242, 1_000_00, 500_000)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if state.CompanyID != c.ID || state.Day != 0 || state.Seed != 4242 || state.Cash != 1_000_00 {
		t.Errorf("unexpected state: %+v", state)
	}
	if state.Revenue != 0 {
		t.Errorf("Revenue = %d, want 0 on init", state.Revenue)
	}
	if state.MonthlyBurn != 500_000 {
		t.Errorf("MonthlyBurn = %d, want 500000", state.MonthlyBurn)
	}

	got, err := repo.Get(ctx, c.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Seed != 4242 {
		t.Errorf("seed = %d, want 4242", got.Seed)
	}
	if got.Revenue != 0 || got.MonthlyBurn != 500_000 {
		t.Errorf("Get metrics = rev:%d burn:%d, want rev:0 burn:500000", got.Revenue, got.MonthlyBurn)
	}
}

func TestSimStateInitIsIdempotent(t *testing.T) {
	repo, companies, users := simStateRepoForTest(t)
	ctx := context.Background()

	owner := mustSeedOwner(t, users)
	c, err := companies.CreateCompany(ctx, &Company{OwnerID: owner, Name: "Idem Co", FoundedAt: time.Now(), Cash: 500})
	if err != nil {
		t.Fatalf("create company: %v", err)
	}

	first, err := repo.Init(ctx, c.ID, 100, 500, 250_000)
	if err != nil {
		t.Fatalf("first Init: %v", err)
	}
	// A second init with different args must not overwrite the existing row.
	second, err := repo.Init(ctx, c.ID, 999, 9999, 999_999)
	if err != nil {
		t.Fatalf("second Init: %v", err)
	}
	if second.Seed != first.Seed || second.Cash != first.Cash || second.MonthlyBurn != first.MonthlyBurn {
		t.Errorf("Init not idempotent: first=%+v second=%+v", first, second)
	}
}

func TestSimStateSave(t *testing.T) {
	repo, companies, users := simStateRepoForTest(t)
	ctx := context.Background()

	owner := mustSeedOwner(t, users)
	c, err := companies.CreateCompany(ctx, &Company{OwnerID: owner, Name: "Save Co", FoundedAt: time.Now(), Cash: 200})
	if err != nil {
		t.Fatalf("create company: %v", err)
	}
	if _, err := repo.Init(ctx, c.ID, 7, 200, 300_000); err != nil {
		t.Fatalf("Init: %v", err)
	}

	if err := repo.Save(ctx, c.ID, 3, 12_34_56, 1_00, 450_000); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := repo.Get(ctx, c.ID)
	if err != nil {
		t.Fatalf("Get after Save: %v", err)
	}
	if got.Day != 3 || got.Cash != 12_34_56 {
		t.Errorf("after save: day=%d cash=%d, want day=3 cash=123456", got.Day, got.Cash)
	}
	if got.Revenue != 1_00 || got.MonthlyBurn != 450_000 {
		t.Errorf("after save: revenue=%d burn=%d, want revenue=100 burn=450000", got.Revenue, got.MonthlyBurn)
	}
}

func TestSimStateSaveMissingIsNotFound(t *testing.T) {
	repo, _, _ := simStateRepoForTest(t)
	err := repo.Save(context.Background(), "00000000-0000-0000-0000-000000000000", 1, 1, 0, 0)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSimStateCascadesWithCompanyDelete(t *testing.T) {
	repo, companies, users := simStateRepoForTest(t)
	ctx := context.Background()

	owner := mustSeedOwner(t, users)
	c, err := companies.CreateCompany(ctx, &Company{OwnerID: owner, Name: "Cascade Co", FoundedAt: time.Now(), Cash: 0})
	if err != nil {
		t.Fatalf("create company: %v", err)
	}
	if _, err := repo.Init(ctx, c.ID, 1, 0, 0); err != nil {
		t.Fatalf("Init: %v", err)
	}

	if _, err := repo.Pool().Exec(ctx, `DELETE FROM companies WHERE id = $1`, c.ID); err != nil {
		t.Fatalf("delete company: %v", err)
	}
	_, err = repo.Get(ctx, c.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected cascade delete to remove sim state, got %v", err)
	}
}
