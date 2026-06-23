-- 0030_contracts.sql
-- Enterprise contracts: multi-year recurring-revenue agreements with a term
-- (days), annual value, and renewal/churn on expiry.

CREATE TABLE contracts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id      UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    customer_name   TEXT NOT NULL,
    annual_value    BIGINT NOT NULL,           -- cents per year
    term_days       INTEGER NOT NULL,          -- contract length in sim days
    remaining_days  INTEGER NOT NULL,          -- counts down each tick
    status          TEXT NOT NULL DEFAULT 'active', -- active | renewed | churned
    discount_pct    INTEGER NOT NULL DEFAULT 0,-- negotiated discount (0..50)
    signed_day      INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT c_status CHECK (status IN ('active','renewed','churned')),
    CONSTRAINT c_remaining CHECK (remaining_days >= 0),
    CONSTRAINT c_discount CHECK (discount_pct >= 0 AND discount_pct <= 50)
);

CREATE INDEX contracts_company_idx ON contracts(company_id);
CREATE INDEX contracts_company_status_idx ON contracts(company_id, status);
