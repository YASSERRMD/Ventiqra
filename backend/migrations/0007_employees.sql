-- 0007_employees.sql
-- Employees belong to a company. Each tracks role, monthly salary, skill
-- (productivity proxy, 0..100), and morale (0..100). Money is BIGINT cents.

CREATE TABLE employees (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id   UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    role         TEXT NOT NULL,
    salary_cents BIGINT NOT NULL,
    skill        INTEGER NOT NULL DEFAULT 50,
    morale       INTEGER NOT NULL DEFAULT 70,
    hired_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT employees_name_len CHECK (char_length(name) BETWEEN 1 AND 120),
    CONSTRAINT employees_role CHECK (role IN ('engineer','designer','sales','marketing','support','operations')),
    CONSTRAINT employees_salary CHECK (salary_cents >= 0),
    CONSTRAINT employees_skill CHECK (skill BETWEEN 0 AND 100),
    CONSTRAINT employees_morale CHECK (morale BETWEEN 0 AND 100)
);

CREATE INDEX employees_company_idx ON employees(company_id);

CREATE TRIGGER employees_updated_at
    BEFORE UPDATE ON employees
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
