-- 0029_deals.sql
-- B2B sales deals: each progresses through a pipeline (lead → qualified →
-- proposal → negotiation → closed-won / closed-lost). Sales employees advance
-- deals; won deals pay out their value.

CREATE TABLE deals (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id  UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    stage       TEXT NOT NULL DEFAULT 'lead',
    value_cents BIGINT NOT NULL DEFAULT 0,
    probability INTEGER NOT NULL DEFAULT 10,  -- 0..100 close probability
    closed_won  BOOLEAN NOT NULL DEFAULT FALSE,
    created_day INTEGER NOT NULL DEFAULT 0,
    closed_day  INTEGER,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT d_stage CHECK (stage IN ('lead','qualified','proposal','negotiation','closed_won','closed_lost')),
    CONSTRAINT d_prob CHECK (probability >= 0 AND probability <= 100)
);

CREATE INDEX deals_company_idx ON deals(company_id);
CREATE INDEX deals_company_stage_idx ON deals(company_id, stage);
