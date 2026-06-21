-- 0022_timeline_events.sql
-- A unified, append-only timeline of milestone and notable events for a company.
-- Each row is a single chronologically-orderable entry (founding, launches,
-- funding, decisions, crises, custom milestones) so the dashboard can render one
-- coherent history without joining many tables.

CREATE TABLE timeline_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id  UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    kind        TEXT NOT NULL,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    sim_day     INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX timeline_events_company_day_idx ON timeline_events(company_id, sim_day DESC);
CREATE INDEX timeline_events_company_created_idx ON timeline_events(company_id, created_at DESC);
