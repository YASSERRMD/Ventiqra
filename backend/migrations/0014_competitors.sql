-- 0014_competitors.sql
-- Rival companies competing for the same market. Strength (0..100) and market
-- share (0..1) drive how much they erode the player's acquisition.

CREATE TABLE competitors (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id   UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    strength     INTEGER NOT NULL DEFAULT 20,
    market_share NUMERIC(5,4) NOT NULL DEFAULT 0.05,
    last_launch_day INTEGER NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT comp_strength CHECK (strength BETWEEN 0 AND 100),
    CONSTRAINT comp_share CHECK (market_share BETWEEN 0 AND 1)
);

CREATE INDEX competitors_company_idx ON competitors(company_id);

CREATE TRIGGER competitors_updated_at
    BEFORE UPDATE ON competitors
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
