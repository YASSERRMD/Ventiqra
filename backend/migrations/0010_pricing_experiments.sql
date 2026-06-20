-- 0010_pricing_experiments.sql
-- Records each product price change as a "pricing experiment" so the impact of
-- pricing decisions can be reviewed over time. Money is BIGINT cents.

CREATE TABLE pricing_experiments (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id     UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    company_id     UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    old_price_cents BIGINT,
    new_price_cents BIGINT NOT NULL,
    sim_day        INTEGER NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pricing_new_nonneg CHECK (new_price_cents >= 0)
);

CREATE INDEX pricing_experiments_product_idx ON pricing_experiments(product_id);
CREATE INDEX pricing_experiments_company_idx ON pricing_experiments(company_id);
