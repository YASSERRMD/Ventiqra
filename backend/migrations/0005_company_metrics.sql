-- 0005_company_metrics.sql
-- Track per-company financial metrics alongside simulation state. Both columns
-- are BIGINT cents, consistent with simulation_state.cash. revenue is 0 until a
-- company has products (Phase 9); monthly_burn captures the recurring monthly
-- operating cost used by the simulation engine and the metrics endpoint.

ALTER TABLE simulation_state
    ADD COLUMN revenue      BIGINT NOT NULL DEFAULT 0,   -- cents; accrued per period
    ADD COLUMN monthly_burn BIGINT NOT NULL DEFAULT 0;   -- cents per month
