-- 0024_sim_control.sql
-- Per-company simulation speed control. A single row per company records the
-- run mode (paused/auto), speed (ticks per simulated day per real second), and
-- the last-driven sim day. The auto-runner reads this to decide whether and how
-- fast to advance the clock.

CREATE TABLE sim_control (
    company_id  UUID PRIMARY KEY REFERENCES companies(id) ON DELETE CASCADE,
    mode        TEXT NOT NULL DEFAULT 'paused',  -- 'paused' | 'auto'
    speed       INTEGER NOT NULL DEFAULT 1,      -- 1, 5, or 30 (ticks per real second)
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT sc_mode CHECK (mode IN ('paused','auto')),
    CONSTRAINT sc_speed CHECK (speed IN (1,5,30))
);
