-- 0003_companies.sql
-- Companies table. Money is stored as BIGINT cents to avoid floating point.

CREATE TABLE companies (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    slug        TEXT NOT NULL UNIQUE,
    industry    TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    founded_at  DATE NOT NULL DEFAULT CURRENT_DATE,
    cash        BIGINT NOT NULL DEFAULT 0,   -- cents
    status      TEXT NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT companies_name_len CHECK (char_length(name) BETWEEN 1 AND 120),
    CONSTRAINT companies_status CHECK (status IN ('active','bankrupt','closed'))
);

CREATE INDEX companies_owner_idx ON companies(owner_id);
CREATE UNIQUE INDEX companies_slug_idx ON companies(slug);

CREATE TRIGGER companies_updated_at
    BEFORE UPDATE ON companies
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
