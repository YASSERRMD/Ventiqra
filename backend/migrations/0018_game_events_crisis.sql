-- 0018_game_events_crisis.sql
-- Extend the game_events kind to include 'crisis' for severe events.

ALTER TABLE game_events DROP CONSTRAINT ge_kind;
ALTER TABLE game_events ADD CONSTRAINT ge_kind CHECK (kind IN ('positive','negative','neutral','crisis'));
