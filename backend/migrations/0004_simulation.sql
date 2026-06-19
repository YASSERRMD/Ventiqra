-- 0004_simulation.sql
-- Persistent simulation state for a company. Each company has at most one row.
-- Money (cash) is stored as BIGINT cents, consistent with companies.cash.

CREATE TABLE simulation_state (
    company_id  UUID PRIMARY KEY REFERENCES companies(id) ON DELETE CASCADE,
    day         INT     NOT NULL DEFAULT 0,
    seed        BIGINT  NOT NULL,
    cash        BIGINT  NOT NULL DEFAULT 0,   -- cents
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER simulation_state_updated_at
    BEFORE UPDATE ON simulation_state
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
