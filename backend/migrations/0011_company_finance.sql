-- 0011_company_finance.sql
-- Per-company finance settings. Holds the monthly marketing budget that feeds
-- the finance engine's burn calculation (infrastructure scales with customers).

CREATE TABLE company_finance (
    company_id            UUID PRIMARY KEY REFERENCES companies(id) ON DELETE CASCADE,
    marketing_budget_cents BIGINT NOT NULL DEFAULT 0,
    infra_tier            INTEGER NOT NULL DEFAULT 1,
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT cf_marketing CHECK (marketing_budget_cents >= 0),
    CONSTRAINT cf_infra CHECK (infra_tier BETWEEN 1 AND 5)
);

CREATE TRIGGER company_finance_updated_at
    BEFORE UPDATE ON company_finance
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
