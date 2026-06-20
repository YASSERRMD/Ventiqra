package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// EmployeeRole describes a staff member's function within a company.
type EmployeeRole string

const (
	RoleEngineer   EmployeeRole = "engineer"
	RoleDesigner   EmployeeRole = "designer"
	RoleSales      EmployeeRole = "sales"
	RoleMarketing  EmployeeRole = "marketing"
	RoleSupport    EmployeeRole = "support"
	RoleOperations EmployeeRole = "operations"
)

// ValidEmployeeRoles is the set of accepted employee role values.
var ValidEmployeeRoles = map[EmployeeRole]bool{
	RoleEngineer: true, RoleDesigner: true, RoleSales: true,
	RoleMarketing: true, RoleSupport: true, RoleOperations: true,
}

// DefaultSalaryByRole is the default monthly salary in cents applied when a
// newly hired employee does not specify one. Centralizing the table here keeps
// the hiring defaults reproducible and easy to tune during balancing.
var DefaultSalaryByRole = map[EmployeeRole]int64{
	RoleEngineer:   12_000_00,
	RoleDesigner:   10_000_00,
	RoleSales:      9_000_00,
	RoleMarketing:  9_000_00,
	RoleSupport:    7_000_00,
	RoleOperations: 8_000_00,
}

// Employee is the application-level employee model. Skill and Morale are 0..100;
// SalaryCents is a monthly figure in cents.
type Employee struct {
	ID          string
	CompanyID   string
	Name        string
	Role        EmployeeRole
	SalaryCents int64
	Skill       int
	Morale      int
	HiredAt     time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// EmployeeRepo provides persistence for employees.
type EmployeeRepo struct {
	*Repository
}

// NewEmployeeRepo creates an EmployeeRepo over the shared pool.
func NewEmployeeRepo(base *Repository) *EmployeeRepo {
	return &EmployeeRepo{Repository: base}
}

// CreateEmployee inserts a new employee for the given company. When salary is
// zero it falls back to the role's default. Skill defaults to 50 and morale to
// 70 when unset (out of range).
func (r *EmployeeRepo) CreateEmployee(ctx context.Context, companyID, name string, role EmployeeRole, salaryCents int64, skill, morale int) (*Employee, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("employee name is required")
	}
	if companyID == "" {
		return nil, errors.New("company id is required")
	}
	if !ValidEmployeeRoles[role] {
		return nil, fmt.Errorf("invalid employee role: %q", role)
	}
	if salaryCents <= 0 {
		salaryCents = DefaultSalaryByRole[role]
	}
	if skill < 0 || skill > 100 {
		skill = 50
	}
	if morale < 0 || morale > 100 {
		morale = 70
	}

	const q = `INSERT INTO employees (company_id, name, role, salary_cents, skill, morale)
	           VALUES ($1, $2, $3, $4, $5, $6)
	           RETURNING id, company_id, name, role, salary_cents, skill, morale, hired_at, created_at, updated_at`

	e, err := scanEmployee(r.pool.QueryRow(ctx, q, companyID, name, role, salaryCents, skill, morale))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("create employee: %w", err)
	}
	return e, nil
}

// GetEmployee returns an employee by id or ErrNotFound.
func (r *EmployeeRepo) GetEmployee(ctx context.Context, id string) (*Employee, error) {
	const q = `SELECT id, company_id, name, role, salary_cents, skill, morale, hired_at, created_at, updated_at
	           FROM employees WHERE id = $1`
	e, err := scanEmployee(r.pool.QueryRow(ctx, q, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get employee: %w", err)
	}
	return e, nil
}

// ListEmployeesByCompany returns a company's employees ordered by hire date
// ascending.
func (r *EmployeeRepo) ListEmployeesByCompany(ctx context.Context, companyID string) ([]*Employee, error) {
	const q = `SELECT id, company_id, name, role, salary_cents, skill, morale, hired_at, created_at, updated_at
	           FROM employees WHERE company_id = $1 ORDER BY hired_at ASC, created_at ASC`
	rows, err := r.pool.Query(ctx, q, companyID)
	if err != nil {
		return nil, fmt.Errorf("list employees: %w", err)
	}
	defer rows.Close()

	var out []*Employee
	for rows.Next() {
		e, err := scanEmployee(rows)
		if err != nil {
			return nil, fmt.Errorf("scan employee: %w", err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list employees rows: %w", err)
	}
	return out, nil
}

// DeleteEmployee removes an employee (used when firing).
func (r *EmployeeRepo) DeleteEmployee(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM employees WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete employee: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateSalary sets the monthly salary in cents for an employee.
func (r *EmployeeRepo) UpdateSalary(ctx context.Context, id string, salaryCents int64) error {
	if salaryCents < 0 {
		return errors.New("salary must be non-negative")
	}
	return r.updateColumn(ctx, id, `UPDATE employees SET salary_cents = $2 WHERE id = $1`, salaryCents)
}

// UpdateMorale sets the morale (0..100), clamped to range.
func (r *EmployeeRepo) UpdateMorale(ctx context.Context, id string, morale int) error {
	if morale < 0 {
		morale = 0
	}
	if morale > 100 {
		morale = 100
	}
	return r.updateColumn(ctx, id, `UPDATE employees SET morale = $2 WHERE id = $1`, morale)
}

// updateColumn runs a single-column update keyed on id, returning ErrNotFound
// when no row matched.
func (r *EmployeeRepo) updateColumn(ctx context.Context, id, q string, arg any) error {
	tag, err := r.pool.Exec(ctx, q, id, arg)
	if err != nil {
		return fmt.Errorf("update employee: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type employeeScanner interface {
	Scan(dest ...any) error
}

func scanEmployee(row employeeScanner) (*Employee, error) {
	var e Employee
	if err := row.Scan(&e.ID, &e.CompanyID, &e.Name, &e.Role, &e.SalaryCents,
		&e.Skill, &e.Morale, &e.HiredAt, &e.CreatedAt, &e.UpdatedAt); err != nil {
		return nil, err
	}
	return &e, nil
}
