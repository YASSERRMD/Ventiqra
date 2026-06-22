-- 0025_features.sql
-- Product roadmap features: backlog items with priority, development progress,
-- and a value contribution once shipped. Employees develop the active feature;
-- shipped features lift product value (revenue/retention).

CREATE TABLE features (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id   UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    product_id   UUID REFERENCES products(id) ON DELETE SET NULL,
    name         TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    priority     INTEGER NOT NULL DEFAULT 0,   -- higher = sooner
    status       TEXT NOT NULL DEFAULT 'backlog', -- backlog | developing | shipped
    progress     INTEGER NOT NULL DEFAULT 0,    -- 0..100
    value_points INTEGER NOT NULL DEFAULT 0,    -- product-value contribution when shipped
    started_day  INTEGER,
    shipped_day  INTEGER,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT f_status CHECK (status IN ('backlog','developing','shipped')),
    CONSTRAINT f_progress CHECK (progress >= 0 AND progress <= 100)
);

CREATE INDEX features_company_idx ON features(company_id);
CREATE INDEX features_company_priority_idx ON features(company_id, priority DESC);
