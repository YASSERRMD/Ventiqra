-- 0008_product_launches.sql
-- Launch event history. Each row records a product launch with the computed
-- readiness score and the initial customer count granted at launch time.

CREATE TABLE product_launches (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id        UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    company_id        UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    readiness         NUMERIC(5,2) NOT NULL,
    initial_customers INTEGER NOT NULL,
    launched_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT launches_readiness CHECK (readiness BETWEEN 0 AND 100),
    CONSTRAINT launches_customers CHECK (initial_customers >= 0)
);

CREATE INDEX product_launches_product_idx ON product_launches(product_id);
CREATE INDEX product_launches_company_idx ON product_launches(company_id);
