-- 0017_game_events.sql
-- Random event log. Each row records an event that fired during the simulation,
-- its kind (positive/negative/neutral), and the effect that was applied.

CREATE TABLE game_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id  UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    kind        TEXT NOT NULL,
    title       TEXT NOT NULL,
    description TEXT NOT NULL,
    cash_delta  BIGINT NOT NULL DEFAULT 0,
    reputation_delta INTEGER NOT NULL DEFAULT 0,
    morale_delta INTEGER NOT NULL DEFAULT 0,
    sim_day     INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ge_kind CHECK (kind IN ('positive','negative','neutral'))
);

CREATE INDEX game_events_company_idx ON game_events(company_id);
