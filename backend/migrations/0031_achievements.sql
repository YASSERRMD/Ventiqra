-- 0031_achievements.sql
-- Awarded achievements per company. Each row records when a milestone was
-- first reached.

CREATE TABLE achievements (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id  UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    key         TEXT NOT NULL,
    awarded_day INTEGER NOT NULL,
    awarded_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (company_id, key)
);

CREATE INDEX achievements_company_idx ON achievements(company_id);
