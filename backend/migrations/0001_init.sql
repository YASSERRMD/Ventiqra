-- 0001_init.sql
-- Foundation schema: extensions and shared helpers used by later phases.

-- pgcrypto provides gen_random_uuid() for UUID primary keys.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- citext gives case-insensitive text columns (emails, slugs).
CREATE EXTENSION IF NOT EXISTS citext;

-- updated_at() bump helper for trigger-based updated_at columns.
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
