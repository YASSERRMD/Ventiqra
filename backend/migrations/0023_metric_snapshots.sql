-- 0023_metric_snapshots.sql
-- Daily metric snapshots captured during each simulation tick. Each row holds
-- the company's headline metrics at the end of a simulated day, so the analytics
-- dashboard can plot revenue, cash, customers, burn, and valuation over time.

CREATE TABLE metric_snapshots (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id    UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    sim_day       INTEGER NOT NULL,
    cash_cents    BIGINT NOT NULL,
    revenue_cents BIGINT NOT NULL,
    monthly_burn  BIGINT NOT NULL,
    customers     INTEGER NOT NULL DEFAULT 0,
    valuation_cents BIGINT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (company_id, sim_day)
);

CREATE INDEX metric_snapshots_company_day_idx ON metric_snapshots(company_id, sim_day);
