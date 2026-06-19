package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/YASSERRMD/Ventiqra/backend/internal/db"
)

func companyRepoForTest(t *testing.T) (*CompanyRepo, *UserRepo) {
	t.Helper()
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
	return NewCompanyRepo(base), NewUserRepo(base)
}

func mustSeedOwner(t *testing.T, users *UserRepo) string {
	t.Helper()
	u, err := users.CreateUser(context.Background(), "owner@example.com", "hashed", "Owner")
	if err != nil {
		t.Fatalf("seed owner: %v", err)
	}
	return u.ID
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Acme Rockets!": "acme-rockets",
		"  Hello   World  ": "hello-world",
		"Café & Co.":     "caf-co",
		"---":            "",
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCreateAndGetCompany(t *testing.T) {
	repo, users := companyRepoForTest(t)
	owner := mustSeedOwner(t, users)
	ctx := context.Background()

	c, err := repo.CreateCompany(ctx, &Company{
		OwnerID:   owner,
		Name:      "Acme Rockets",
		Industry:  "Aerospace",
		FoundedAt: time.Now(),
		Cash:      50_000_00,
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}
	if c.ID == "" || c.Slug != "acme-rockets" {
		t.Errorf("unexpected company: %+v", c)
	}
	if c.Status != CompanyActive {
		t.Errorf("status = %q, want active", c.Status)
	}

	got, err := repo.GetCompany(ctx, c.ID)
	if err != nil {
		t.Fatalf("GetCompany: %v", err)
	}
	if got.Name != "Acme Rockets" {
		t.Errorf("name = %q", got.Name)
	}

	latest, err := repo.GetLatestCompanyForOwner(ctx, owner)
	if err != nil {
		t.Fatalf("GetLatestCompanyForOwner: %v", err)
	}
	if latest.ID != c.ID {
		t.Errorf("latest id mismatch")
	}
}

func TestUniqueSlugSuffixed(t *testing.T) {
	repo, users := companyRepoForTest(t)
	owner := mustSeedOwner(t, users)
	ctx := context.Background()

	if _, err := repo.CreateCompany(ctx, &Company{OwnerID: owner, Name: "Same Name", FoundedAt: time.Now()}); err != nil {
		t.Fatalf("first create: %v", err)
	}
	slug2, err := repo.UniqueSlug(ctx, "same-name")
	if err != nil {
		t.Fatalf("UniqueSlug: %v", err)
	}
	if slug2 != "same-name-1" {
		t.Errorf("slug2 = %q, want same-name-1", slug2)
	}
}

func TestGetCompanyMissing(t *testing.T) {
	repo, _ := companyRepoForTest(t)
	_, err := repo.GetCompany(context.Background(), "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdateCash(t *testing.T) {
	repo, users := companyRepoForTest(t)
	owner := mustSeedOwner(t, users)
	ctx := context.Background()

	c, _ := repo.CreateCompany(ctx, &Company{OwnerID: owner, Name: "Cash Inc", FoundedAt: time.Now(), Cash: 1000})
	if err := repo.UpdateCash(ctx, c.ID, 42_00); err != nil {
		t.Fatalf("UpdateCash: %v", err)
	}
	got, _ := repo.GetCompany(ctx, c.ID)
	if got.Cash != 42_00 {
		t.Errorf("cash = %d, want 4200", got.Cash)
	}
}
