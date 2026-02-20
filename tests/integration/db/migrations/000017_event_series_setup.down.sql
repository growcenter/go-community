-- Drop all constraints
ALTER TABLE event_series
DROP CONSTRAINT IF EXISTS chk_series_status;

-- Drop all indexes
DROP INDEX IF EXISTS idx_series_creator;

DROP INDEX IF EXISTS idx_series_status;

DROP INDEX IF EXISTS idx_series_creator_status;

-- Drop table
DROP TABLE IF EXISTS event_series;