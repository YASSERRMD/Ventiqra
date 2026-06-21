-- 0015_market.sql
-- Per-company market model: total addressable market (customer count), monthly
-- growth rate, and a trend multiplier capturing demand cycles.

CREATE TABLE market (
    company_id        UUID PRIMARY KEY REFERENCES companies(id) ON DELETE CASCADE,
    tam               BIGINT NOT NULL DEFAULT 100000,
    growth_rate       NUMERIC(6,5) NOT NULL DEFAULT 0.01000,
    trend_multiplier  NUMERIC(4,3) NOT NULL DEFAULT 1.000,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT market_tam CHECK (tam >= 0),
    CONSTRAINT market_growth CHECK (growth_rate BETWEEN 0 AND 1),
    CONSTRAINT market_trend CHECK (trend_multiplier BETWEEN 0 AND 5)
);

CREATE TRIGGER market_updated_at
    BEFORE UPDATE ON market
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
