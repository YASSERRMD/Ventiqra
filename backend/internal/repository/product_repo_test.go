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

func productRepoForTest(t *testing.T) (*ProductRepo, *CompanyRepo, *UserRepo) {
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
	return NewProductRepo(base), NewCompanyRepo(base), NewUserRepo(base)
}

func mustSeedCompany(t *testing.T, companies *CompanyRepo, users *UserRepo) string {
	t.Helper()
	owner := mustSeedOwner(t, users)
	c, err := companies.CreateCompany(context.Background(), &Company{
		OwnerID:   owner,
		Name:      "Product Co",
		FoundedAt: time.Now(),
		Cash:      100_000_00,
	})
	if err != nil {
		t.Fatalf("seed company: %v", err)
	}
	return c.ID
}

func TestCreateAndGetProduct(t *testing.T) {
	repo, companies, users := productRepoForTest(t)
	companyID := mustSeedCompany(t, companies, users)
	ctx := context.Background()

	p, err := repo.CreateProduct(ctx, companyID, "Acme App")
	if err != nil {
		t.Fatalf("CreateProduct: %v", err)
	}
	if p.ID == "" || p.CompanyID != companyID {
		t.Errorf("unexpected product: %+v", p)
	}
	if p.Name != "Acme App" || p.Slug != "acme-app" {
		t.Errorf("name/slug = %q/%q", p.Name, p.Slug)
	}
	if p.Stage != ProductIdea {
		t.Errorf("stage = %q, want idea", p.Stage)
	}
	if p.DevProgress != 0 {
		t.Errorf("dev_progress = %v, want 0", p.DevProgress)
	}
	if p.PriceCents != nil {
		t.Errorf("price_cents = %v, want nil", *p.PriceCents)
	}

	got, err := repo.GetProduct(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetProduct: %v", err)
	}
	if got.Name != "Acme App" {
		t.Errorf("got name = %q", got.Name)
	}
}

func TestCreateProductUniqueSlugSuffixed(t *testing.T) {
	repo, companies, users := productRepoForTest(t)
	companyID := mustSeedCompany(t, companies, users)
	ctx := context.Background()

	first, err := repo.CreateProduct(ctx, companyID, "Same Thing")
	if err != nil {
		t.Fatalf("first create: %v", err)
	}
	if first.Slug != "same-thing" {
		t.Fatalf("first slug = %q, want same-thing", first.Slug)
	}
	second, err := repo.CreateProduct(ctx, companyID, "Same Thing")
	if err != nil {
		t.Fatalf("second create: %v", err)
	}
	if second.Slug != "same-thing-1" {
		t.Errorf("second slug = %q, want same-thing-1", second.Slug)
	}
}

func TestCreateProductSlugScopedPerCompany(t *testing.T) {
	repo, companies, users := productRepoForTest(t)
	ctx := context.Background()

	owner := mustSeedOwner(t, users)
	c1, err := companies.CreateCompany(ctx, &Company{OwnerID: owner, Name: "Co A", FoundedAt: time.Now()})
	if err != nil {
		t.Fatalf("create company A: %v", err)
	}
	c2, err := companies.CreateCompany(ctx, &Company{OwnerID: owner, Name: "Co B", FoundedAt: time.Now()})
	if err != nil {
		t.Fatalf("create company B: %v", err)
	}

	p1, err := repo.CreateProduct(ctx, c1.ID, "Shared Name")
	if err != nil {
		t.Fatalf("create in c1: %v", err)
	}
	p2, err := repo.CreateProduct(ctx, c2.ID, "Shared Name")
	if err != nil {
		t.Fatalf("create in c2: %v", err)
	}
	if p1.Slug != "shared-name" || p2.Slug != "shared-name" {
		t.Errorf("expected same slug per company, got %q and %q", p1.Slug, p2.Slug)
	}
}

func TestListProductsByCompanyOrdered(t *testing.T) {
	repo, companies, users := productRepoForTest(t)
	companyID := mustSeedCompany(t, companies, users)
	ctx := context.Background()

	p1, _ := repo.CreateProduct(ctx, companyID, "First Product")
	p2, _ := repo.CreateProduct(ctx, companyID, "Second Product")
	p3, _ := repo.CreateProduct(ctx, companyID, "Third Product")

	list, err := repo.ListProductsByCompany(ctx, companyID)
	if err != nil {
		t.Fatalf("ListProductsByCompany: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("len = %d, want 3", len(list))
	}
	if list[0].ID != p1.ID || list[1].ID != p2.ID || list[2].ID != p3.ID {
		t.Errorf("order = %s,%s,%s; want %s,%s,%s",
			list[0].ID, list[1].ID, list[2].ID, p1.ID, p2.ID, p3.ID)
	}

	otherOwner, err := users.CreateUser(ctx, "other-owner@example.com", "hashed", "Other")
	if err != nil {
		t.Fatalf("seed other owner: %v", err)
	}
	other, err := companies.CreateCompany(ctx, &Company{OwnerID: otherOwner.ID, Name: "Other Co", FoundedAt: time.Now()})
	if err != nil {
		t.Fatalf("create other company: %v", err)
	}
	otherList, err := repo.ListProductsByCompany(ctx, other.ID)
	if err != nil {
		t.Fatalf("list other: %v", err)
	}
	if len(otherList) != 0 {
		t.Errorf("other company products = %d, want 0", len(otherList))
	}
}

func TestUpdateStage(t *testing.T) {
	repo, companies, users := productRepoForTest(t)
	companyID := mustSeedCompany(t, companies, users)
	ctx := context.Background()

	p, _ := repo.CreateProduct(ctx, companyID, "Staged")
	if err := repo.UpdateStage(ctx, p.ID, ProductBuilding); err != nil {
		t.Fatalf("UpdateStage: %v", err)
	}
	got, _ := repo.GetProduct(ctx, p.ID)
	if got.Stage != ProductBuilding {
		t.Errorf("stage = %q, want building", got.Stage)
	}
	if got.UpdatedAt.Equal(got.CreatedAt) {
		t.Errorf("updated_at should advance via trigger")
	}
}

func TestUpdateStageInvalid(t *testing.T) {
	repo, companies, users := productRepoForTest(t)
	companyID := mustSeedCompany(t, companies, users)
	p, _ := repo.CreateProduct(context.Background(), companyID, "X")
	if err := repo.UpdateStage(context.Background(), p.ID, ProductStage("nope")); err == nil {
		t.Errorf("expected error for invalid stage")
	}
}

func TestUpdateProgressClamps(t *testing.T) {
	repo, companies, users := productRepoForTest(t)
	companyID := mustSeedCompany(t, companies, users)
	ctx := context.Background()

	p, _ := repo.CreateProduct(ctx, companyID, "Progress App")
	if err := repo.UpdateProgress(ctx, p.ID, 150); err != nil {
		t.Fatalf("UpdateProgress over: %v", err)
	}
	got, _ := repo.GetProduct(ctx, p.ID)
	if got.DevProgress != 100 {
		t.Errorf("progress = %v, want 100 (clamped)", got.DevProgress)
	}
	if err := repo.UpdateProgress(ctx, p.ID, -20); err != nil {
		t.Fatalf("UpdateProgress under: %v", err)
	}
	got, _ = repo.GetProduct(ctx, p.ID)
	if got.DevProgress != 0 {
		t.Errorf("progress = %v, want 0 (clamped)", got.DevProgress)
	}
	if err := repo.UpdateProgress(ctx, p.ID, 42.5); err != nil {
		t.Fatalf("UpdateProgress mid: %v", err)
	}
	got, _ = repo.GetProduct(ctx, p.ID)
	if got.DevProgress < 42.4 || got.DevProgress > 42.6 {
		t.Errorf("progress = %v, want ~42.5", got.DevProgress)
	}
}

func TestSetPrice(t *testing.T) {
	repo, companies, users := productRepoForTest(t)
	companyID := mustSeedCompany(t, companies, users)
	ctx := context.Background()

	p, _ := repo.CreateProduct(ctx, companyID, "Paid App")
	if err := repo.SetPrice(ctx, p.ID, 19_99); err != nil {
		t.Fatalf("SetPrice: %v", err)
	}
	got, _ := repo.GetProduct(ctx, p.ID)
	if got.PriceCents == nil || *got.PriceCents != 19_99 {
		t.Errorf("price_cents = %v, want 1999", got.PriceCents)
	}
}

func TestProductUpdatesMissingIsNotFound(t *testing.T) {
	repo, _, _ := productRepoForTest(t)
	ctx := context.Background()
	missing := "00000000-0000-0000-0000-000000000000"
	cases := []struct {
		name string
		fn   func() error
	}{
		{"stage", func() error { return repo.UpdateStage(ctx, missing, ProductLaunched) }},
		{"progress", func() error { return repo.UpdateProgress(ctx, missing, 10) }},
		{"price", func() error { return repo.SetPrice(ctx, missing, 5_00) }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.fn(); !errors.Is(err, ErrNotFound) {
				t.Errorf("%s: expected ErrNotFound, got %v", c.name, err)
			}
		})
	}
}

func TestGetProductMissing(t *testing.T) {
	repo, _, _ := productRepoForTest(t)
	_, err := repo.GetProduct(context.Background(), "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestProductCascadesWithCompanyDelete(t *testing.T) {
	repo, companies, users := productRepoForTest(t)
	companyID := mustSeedCompany(t, companies, users)
	ctx := context.Background()

	p, _ := repo.CreateProduct(ctx, companyID, "Doomed")
	if _, err := repo.Pool().Exec(ctx, `DELETE FROM companies WHERE id = $1`, companyID); err != nil {
		t.Fatalf("delete company: %v", err)
	}
	_, err := repo.GetProduct(ctx, p.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected cascade delete to remove product, got %v", err)
	}
}
