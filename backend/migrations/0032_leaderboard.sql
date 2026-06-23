-- 0032_leaderboard.sql
-- Local leaderboard entries: one row per finalized company run with a computed
-- outcome score. Rows are immutable once written.

CREATE TABLE leaderboard (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_name   TEXT NOT NULL,
    score          BIGINT NOT NULL,
    days_survived  INTEGER NOT NULL,
    peak_valuation BIGINT NOT NULL,
    outcome        TEXT NOT NULL, -- 'bankrupt' | 'thriving' | 'acquired'
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX leaderboard_score_idx ON leaderboard(score DESC);
CREATE INDEX leaderboard_owner_idx ON leaderboard(owner_id);
