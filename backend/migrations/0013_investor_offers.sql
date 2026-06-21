-- 0013_investor_offers.sql
-- Negotiable investor offers generated during fundraising. Each offer carries an
-- asked equity percentage; the player accepts, rejects, or negotiates it.

CREATE TABLE investor_offers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id      UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    investor_name   TEXT NOT NULL,
    amount_cents    BIGINT NOT NULL,
    equity_percent  NUMERIC(5,2) NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    round_seed      BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT io_amount_pos CHECK (amount_cents > 0),
    CONSTRAINT io_equity_range CHECK (equity_percent > 0 AND equity_percent <= 100),
    CONSTRAINT io_status CHECK (status IN ('pending','accepted','rejected','withdrawn'))
);

CREATE INDEX investor_offers_company_idx ON investor_offers(company_id);

CREATE TRIGGER investor_offers_updated_at
    BEFORE UPDATE ON investor_offers
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
