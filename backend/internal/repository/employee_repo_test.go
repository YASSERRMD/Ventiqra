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

func employeeRepoForTest(t *testing.T) (*EmployeeRepo, *CompanyRepo, *UserRepo) {
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
	return NewEmployeeRepo(base), NewCompanyRepo(base), NewUserRepo(base)
}

func TestCreateAndGetEmployee(t *testing.T) {
	repo, companies, users := employeeRepoForTest(t)
	companyID := mustSeedCompany(t, companies, users)
	ctx := context.Background()

	e, err := repo.CreateEmployee(ctx, companyID, "Ada Lovelace", RoleEngineer, 0, -1, -1)
	if err != nil {
		t.Fatalf("CreateEmployee: %v", err)
	}
	if e.ID == "" || e.CompanyID != companyID {
		t.Errorf("unexpected employee: %+v", e)
	}
	if e.Role != RoleEngineer {
		t.Errorf("role = %q, want engineer", e.Role)
	}
	if e.SalaryCents != DefaultSalaryByRole[RoleEngineer] {
		t.Errorf("salary = %d, want default %d", e.SalaryCents, DefaultSalaryByRole[RoleEngineer])
	}
	if e.Skill != 50 || e.Morale != 70 {
		t.Errorf("skill/morale = %d/%d, want 50/70", e.Skill, e.Morale)
	}

	got, err := repo.GetEmployee(ctx, e.ID)
	if err != nil {
		t.Fatalf("GetEmployee: %v", err)
	}
	if got.Name != "Ada Lovelace" {
		t.Errorf("name = %q", got.Name)
	}
}

func TestCreateEmployeeExplicitSalarySkillMorale(t *testing.T) {
	repo, companies, users := employeeRepoForTest(t)
	companyID := mustSeedCompany(t, companies, users)
	ctx := context.Background()

	e, err := repo.CreateEmployee(ctx, companyID, "Grace Hopper", RoleDesigner, 11_000_00, 88, 95)
	if err != nil {
		t.Fatalf("CreateEmployee: %v", err)
	}
	if e.SalaryCents != 11_000_00 || e.Skill != 88 || e.Morale != 95 {
		t.Errorf("got salary/skill/morale = %d/%d/%d", e.SalaryCents, e.Skill, e.Morale)
	}
}

func TestCreateEmployeeInvalidRole(t *testing.T) {
	repo, companies, users := employeeRepoForTest(t)
	companyID := mustSeedCompany(t, companies, users)
	if _, err := repo.CreateEmployee(context.Background(), companyID, "X", EmployeeRole("ceo"), 0, 0, 0); err == nil {
		t.Errorf("expected error for invalid role")
	}
}

func TestCreateEmployeeEmptyName(t *testing.T) {
	repo, companies, users := employeeRepoForTest(t)
	companyID := mustSeedCompany(t, companies, users)
	if _, err := repo.CreateEmployee(context.Background(), companyID, "   ", RoleEngineer, 0, 0, 0); err == nil {
		t.Errorf("expected error for empty name")
	}
}

func TestListEmployeesByCompanyOrdered(t *testing.T) {
	repo, companies, users := employeeRepoForTest(t)
	companyID := mustSeedCompany(t, companies, users)
	ctx := context.Background()

	a, _ := repo.CreateEmployee(ctx, companyID, "First Hire", RoleEngineer, 0, 0, 0)
	b, _ := repo.CreateEmployee(ctx, companyID, "Second Hire", RoleSales, 0, 0, 0)

	list, err := repo.ListEmployeesByCompany(ctx, companyID)
	if err != nil {
		t.Fatalf("ListEmployeesByCompany: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len = %d, want 2", len(list))
	}
	if list[0].ID != a.ID || list[1].ID != b.ID {
		t.Errorf("order mismatch: %s,%s", list[0].ID, list[1].ID)
	}

	otherOwner, err := users.CreateUser(ctx, "other@example.com", "hashed", "Other")
	if err != nil {
		t.Fatalf("seed other owner: %v", err)
	}
	other, err := companies.CreateCompany(ctx, &Company{OwnerID: otherOwner.ID, Name: "Other Co", FoundedAt: time.Now()})
	if err != nil {
		t.Fatalf("create other company: %v", err)
	}
	otherList, err := repo.ListEmployeesByCompany(ctx, other.ID)
	if err != nil {
		t.Fatalf("list other: %v", err)
	}
	if len(otherList) != 0 {
		t.Errorf("other company employees = %d, want 0", len(otherList))
	}
}

func TestUpdateSalary(t *testing.T) {
	repo, companies, users := employeeRepoForTest(t)
	companyID := mustSeedCompany(t, companies, users)
	ctx := context.Background()

	e, _ := repo.CreateEmployee(ctx, companyID, "Raise", RoleEngineer, 0, 0, 0)
	if err := repo.UpdateSalary(ctx, e.ID, 15_000_00); err != nil {
		t.Fatalf("UpdateSalary: %v", err)
	}
	got, _ := repo.GetEmployee(ctx, e.ID)
	if got.SalaryCents != 15_000_00 {
		t.Errorf("salary = %d, want 1500000", got.SalaryCents)
	}
}

func TestUpdateMoraleClamps(t *testing.T) {
	repo, companies, users := employeeRepoForTest(t)
	companyID := mustSeedCompany(t, companies, users)
	ctx := context.Background()

	e, _ := repo.CreateEmployee(ctx, companyID, "Moody", RoleSupport, 0, 0, 0)
	if err := repo.UpdateMorale(ctx, e.ID, 200); err != nil {
		t.Fatalf("UpdateMorale over: %v", err)
	}
	got, _ := repo.GetEmployee(ctx, e.ID)
	if got.Morale != 100 {
		t.Errorf("morale = %d, want 100 (clamped)", got.Morale)
	}
	if err := repo.UpdateMorale(ctx, e.ID, -5); err != nil {
		t.Fatalf("UpdateMorale under: %v", err)
	}
	got, _ = repo.GetEmployee(ctx, e.ID)
	if got.Morale != 0 {
		t.Errorf("morale = %d, want 0 (clamped)", got.Morale)
	}
}

func TestDeleteEmployee(t *testing.T) {
	repo, companies, users := employeeRepoForTest(t)
	companyID := mustSeedCompany(t, companies, users)
	ctx := context.Background()

	e, _ := repo.CreateEmployee(ctx, companyID, "Let Go", RoleEngineer, 0, 0, 0)
	if err := repo.DeleteEmployee(ctx, e.ID); err != nil {
		t.Fatalf("DeleteEmployee: %v", err)
	}
	_, err := repo.GetEmployee(ctx, e.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestEmployeeMissingIsNotFound(t *testing.T) {
	repo, _, _ := employeeRepoForTest(t)
	missing := "00000000-0000-0000-0000-000000000000"
	ctx := context.Background()
	cases := []struct {
		name string
		fn   func() error
	}{
		{"get", func() error { _, err := repo.GetEmployee(ctx, missing); return err }},
		{"salary", func() error { return repo.UpdateSalary(ctx, missing, 1_000) }},
		{"morale", func() error { return repo.UpdateMorale(ctx, missing, 50) }},
		{"delete", func() error { return repo.DeleteEmployee(ctx, missing) }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.fn(); !errors.Is(err, ErrNotFound) {
				t.Errorf("%s: expected ErrNotFound, got %v", c.name, err)
			}
		})
	}
}

func TestEmployeeCascadesWithCompanyDelete(t *testing.T) {
	repo, companies, users := employeeRepoForTest(t)
	companyID := mustSeedCompany(t, companies, users)
	ctx := context.Background()

	e, _ := repo.CreateEmployee(ctx, companyID, "Doomed", RoleEngineer, 0, 0, 0)
	if _, err := repo.Pool().Exec(ctx, `DELETE FROM companies WHERE id = $1`, companyID); err != nil {
		t.Fatalf("delete company: %v", err)
	}
	_, err := repo.GetEmployee(ctx, e.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected cascade delete, got %v", err)
	}
}
