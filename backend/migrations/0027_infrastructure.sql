-- 0027_infrastructure.sql
-- Per-company infrastructure state: capacity (max concurrent customers),
-- hosting cost (monthly), and tier. Scaling raises capacity and cost.

CREATE TABLE infrastructure (
    company_id   UUID PRIMARY KEY REFERENCES companies(id) ON DELETE CASCADE,
    tier         INTEGER NOT NULL DEFAULT 1,    -- 1..10, higher = bigger
    capacity     INTEGER NOT NULL DEFAULT 1000, -- max customers before outages
    hosting_cost BIGINT NOT NULL DEFAULT 500_00,-- monthly cost in cents
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT inf_tier CHECK (tier >= 1 AND tier <= 10),
    CONSTRAINT inf_capacity CHECK (capacity > 0)
);
