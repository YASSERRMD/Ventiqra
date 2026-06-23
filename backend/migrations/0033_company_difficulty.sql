-- 0033_company_difficulty.sql
-- Adds a difficulty level column to companies so the simulation can apply
-- economy multipliers per run.

ALTER TABLE companies ADD COLUMN IF NOT EXISTS difficulty TEXT NOT NULL DEFAULT 'normal';
ALTER TABLE companies ADD CONSTRAINT companies_difficulty_chk CHECK (difficulty IN ('easy','normal','hard','brutal','custom'));
