-- 0016_reputation.sql
-- Per-company brand reputation score (0..100) and an event log of what moved it.

CREATE TABLE reputation (
    company_id UUID PRIMARY KEY REFERENCES companies(id) ON DELETE CASCADE,
    score       INTEGER NOT NULL DEFAULT 50,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT reputation_score CHECK (score BETWEEN 0 AND 100)
);

CREATE TRIGGER reputation_updated_at
    BEFORE UPDATE ON reputation
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE reputation_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id  UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    event       TEXT NOT NULL,
    delta       INTEGER NOT NULL,
    sim_day     INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX reputation_events_company_idx ON reputation_events(company_id);
