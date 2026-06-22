-- 0028_support.sql
-- Customer support state: an open-ticket backlog that grows with customers and
-- shrinks as the support team resolves tickets. A large backlog drags down
-- customer satisfaction.

CREATE TABLE support_state (
    company_id      UUID PRIMARY KEY REFERENCES companies(id) ON DELETE CASCADE,
    open_tickets    INTEGER NOT NULL DEFAULT 0,
    resolved_total  INTEGER NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ss_tickets CHECK (open_tickets >= 0)
);
