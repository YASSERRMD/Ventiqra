-- 0021_save_slots.sql
-- Named save slots for the simulation. Each slot captures a restorable snapshot
-- of the owner's company and simulation state so a run can be saved and later
-- resumed. The snapshot is stored as JSONB so evolving the schema stays cheap.

CREATE TABLE save_slots (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    slot        TEXT NOT NULL,
    label       TEXT NOT NULL DEFAULT '',
    company_id  UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    -- Denormalized snapshot fields so the slot list renders without joins.
    day         INTEGER NOT NULL DEFAULT 0,
    cash_cents  BIGINT NOT NULL DEFAULT 0,
    status      TEXT NOT NULL DEFAULT 'active',
    snapshot    JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (owner_id, slot)
);

CREATE INDEX save_slots_owner_idx ON save_slots(owner_id);
