-- 0009_product_customers.sql
-- Per-product customer state for launched products. Tracks total customers,
-- monthly active users (MAU), cumulative churned, and satisfaction (0..100).

CREATE TABLE product_customers (
    product_id    UUID PRIMARY KEY REFERENCES products(id) ON DELETE CASCADE,
    company_id    UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    total_customers INTEGER NOT NULL DEFAULT 0,
    mau           INTEGER NOT NULL DEFAULT 0,
    churned       INTEGER NOT NULL DEFAULT 0,
    satisfaction  INTEGER NOT NULL DEFAULT 70,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pc_satisfaction CHECK (satisfaction BETWEEN 0 AND 100),
    CONSTRAINT pc_nonneg CHECK (total_customers >= 0 AND mau >= 0 AND churned >= 0)
);

CREATE INDEX product_customers_company_idx ON product_customers(company_id);

CREATE TRIGGER product_customers_updated_at
    BEFORE UPDATE ON product_customers
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
