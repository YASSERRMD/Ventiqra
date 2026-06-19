-- 0006_products.sql
-- Products table. A product belongs to exactly one company and tracks its
-- lifecycle stage and development progress. Money (price) is BIGINT cents.

CREATE TABLE products (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id   UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    slug         TEXT NOT NULL,
    stage        TEXT NOT NULL DEFAULT 'idea',
    dev_progress NUMERIC(5,2) NOT NULL DEFAULT 0,
    price_cents  BIGINT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT products_name_len CHECK (char_length(name) BETWEEN 1 AND 120),
    CONSTRAINT products_stage CHECK (stage IN ('idea','building','launched','retired')),
    CONSTRAINT products_dev_progress CHECK (dev_progress BETWEEN 0 AND 100)
);

CREATE UNIQUE INDEX products_company_slug_idx ON products(company_id, slug);
CREATE INDEX products_company_idx ON products(company_id);

CREATE TRIGGER products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
