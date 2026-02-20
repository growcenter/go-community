-- Drop all constraints
ALTER TABLE event_occurrences
DROP CONSTRAINT IF EXISTS chk_occurrences_status;

ALTER TABLE event_occurrences
DROP CONSTRAINT IF EXISTS chk_occurrences_dates;

ALTER TABLE event_occurrences
DROP CONSTRAINT IF EXISTS chk_occurrences_checkin_window;

ALTER TABLE event_occurrences
DROP CONSTRAINT IF EXISTS chk_occurrences_counts;

-- Drop all indexes
DROP INDEX IF EXISTS idx_occurrences_event;

DROP INDEX IF EXISTS idx_occurrences_session;

DROP INDEX IF EXISTS idx_occurrences_event_date;

DROP INDEX IF EXISTS idx_occurrences_event_status;

DROP INDEX IF EXISTS idx_occurrences_session_date;

DROP INDEX IF EXISTS idx_occurrences_date_range;

DROP INDEX IF EXISTS idx_occurrences_upcoming;

DROP INDEX IF EXISTS idx_occurrences_modified;

DROP INDEX IF EXISTS idx_occurrences_cancelled;

DROP INDEX IF EXISTS idx_occurrences_checkin_window;

DROP INDEX IF EXISTS idx_occurrences_unique;

-- Drop table
DROP TABLE IF EXISTS event_occurrences;