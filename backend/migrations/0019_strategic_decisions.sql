-- 0019_strategic_decisions.sql
-- Strategic decision cards offered periodically during the simulation. Each row
-- records a card offered to the company and, once resolved, the chosen option
-- together with any long-term commitment that recurs daily for a duration.

CREATE TABLE strategic_decisions (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id           UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    decision_id          TEXT NOT NULL,
    title                TEXT NOT NULL,
    description          TEXT NOT NULL,
    sim_day_offered      INTEGER NOT NULL,
    status               TEXT NOT NULL DEFAULT 'pending',
    chosen_choice        TEXT,
    outcome              TEXT,
    recurring_cash_delta BIGINT NOT NULL DEFAULT 0,
    remaining_days       INTEGER NOT NULL DEFAULT 0,
    resolved_at          TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT sd_status CHECK (status IN ('pending','resolved')),
    CONSTRAINT sd_outcome CHECK (outcome IS NULL OR outcome IN ('success','failure')),
    CONSTRAINT sd_remaining CHECK (remaining_days >= 0)
);

CREATE INDEX strategic_decisions_company_idx ON strategic_decisions(company_id);
CREATE INDEX strategic_decisions_pending_idx ON strategic_decisions(company_id) WHERE status = 'pending';
