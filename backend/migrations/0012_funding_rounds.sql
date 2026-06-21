-- 0012_funding_rounds.sql
-- Funding round history. Each row records a closed round with the amount
-- raised, the pre-money valuation used, the equity granted, and the day.

CREATE TABLE funding_rounds (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id        UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    round_name        TEXT NOT NULL,
    amount_cents      BIGINT NOT NULL,
    pre_money_cents   BIGINT NOT NULL,
    equity_percent    NUMERIC(5,2) NOT NULL,
    sim_day           INTEGER NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fr_amount_pos CHECK (amount_cents > 0),
    CONSTRAINT fr_premoney_pos CHECK (pre_money_cents >= 0),
    CONSTRAINT fr_equity_range CHECK (equity_percent > 0 AND equity_percent <= 100),
    CONSTRAINT fr_name CHECK (round_name IN ('seed','pre-seed','series-a','series-b','series-c','growth'))
);

CREATE INDEX funding_rounds_company_idx ON funding_rounds(company_id);
