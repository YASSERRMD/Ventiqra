package repository

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/YASSERRMD/Ventiqra/backend/internal/db"
)

func userRepoForTest(t *testing.T) (*UserRepo, *Repository) {
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
	return NewUserRepo(base), base
}

func TestCreateUserAndGet(t *testing.T) {
	repo, _ := userRepoForTest(t)
	ctx := context.Background()

	u, err := repo.CreateUser(ctx, "alice@example.com", "hashed-1", "Alice")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if u.ID == "" || u.Email != "alice@example.com" {
		t.Errorf("unexpected user: %+v", u)
	}

	byEmail, err := repo.GetUserByEmail(ctx, "ALICE@example.com") // CITEXT case-insensitive
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if byEmail.ID != u.ID {
		t.Errorf("email lookup mismatch %s != %s", byEmail.ID, u.ID)
	}

	byID, err := repo.GetUserByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if byID.Email != u.Email {
		t.Errorf("id lookup mismatch %s != %s", byID.Email, u.Email)
	}
}

func TestCreateUserDuplicateEmail(t *testing.T) {
	repo, _ := userRepoForTest(t)
	ctx := context.Background()

	if _, err := repo.CreateUser(ctx, "bob@example.com", "h", "Bob"); err != nil {
		t.Fatalf("first create: %v", err)
	}
	_, err := repo.CreateUser(ctx, "bob@example.com", "h", "Bob 2")
	if !errors.Is(err, ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

func TestGetUserMissingReturnsNotFound(t *testing.T) {
	repo, _ := userRepoForTest(t)
	ctx := context.Background()

	if _, err := repo.GetUserByEmail(ctx, "missing@example.com"); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
	if _, err := repo.GetUserByID(ctx, "00000000-0000-0000-0000-000000000000"); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
