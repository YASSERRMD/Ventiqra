-- 0020_custom_scenarios.sql
-- User-authored custom scenarios. Each row is a saved starting configuration a
-- user can create in the scenario editor and apply to their company, mirroring
-- the predefined catalog but with user-chosen market, cash, and difficulty.

CREATE TABLE custom_scenarios (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id            UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name                TEXT NOT NULL,
    description         TEXT NOT NULL DEFAULT '',
    difficulty          TEXT NOT NULL DEFAULT 'normal',
    industry            TEXT NOT NULL DEFAULT '',
    starting_cash_cents BIGINT NOT NULL,
    starting_burn_cents BIGINT NOT NULL,
    market_tam          INTEGER NOT NULL,
    market_growth_rate  DOUBLE PRECISION NOT NULL,
    market_trend        DOUBLE PRECISION NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT cs_difficulty CHECK (difficulty IN ('easy','normal','hard','brutal')),
    CONSTRAINT cs_cash CHECK (starting_cash_cents > 0),
    CONSTRAINT cs_burn CHECK (starting_burn_cents > 0),
    CONSTRAINT cs_tam CHECK (market_tam > 0)
);

CREATE INDEX custom_scenarios_owner_idx ON custom_scenarios(owner_id);
