-- 0026_tech_debt.sql
-- Per-company technical-debt state: a debt score (0-100, higher = worse), a code
-- quality score (100 - debt), and a refactor cost. Debt accumulates as features
-- ship; refactoring pays it down. High debt raises outage risk.

CREATE TABLE tech_debt (
    company_id    UUID PRIMARY KEY REFERENCES companies(id) ON DELETE CASCADE,
    debt          INTEGER NOT NULL DEFAULT 0,   -- 0..100, higher = more debt
    refactors     INTEGER NOT NULL DEFAULT 0,   -- count of refactors performed
    last_refactor_day INTEGER,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT td_debt CHECK (debt >= 0 AND debt <= 100)
);
