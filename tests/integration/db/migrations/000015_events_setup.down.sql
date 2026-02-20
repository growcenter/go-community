-- Drop all constraints
ALTER TABLE events
DROP CONSTRAINT IF EXISTS chk_occurrences_status;

ALTER TABLE events
DROP CONSTRAINT IF EXISTS chk_occurrences_dates;

ALTER TABLE events
DROP CONSTRAINT IF EXISTS chk_occurrences_checkin_window;

ALTER TABLE events
DROP CONSTRAINT IF EXISTS chk_occurrences_counts;

-- Drop all indexes
DROP INDEX IF EXISTS idx_events_event;

DROP INDEX IF EXISTS idx_events_session;

DROP INDEX IF EXISTS idx_events_event_date;

DROP INDEX IF EXISTS idx_events_event_status;

DROP INDEX IF EXISTS idx_events_session_date;

DROP INDEX IF EXISTS idx_events_date_range;

DROP INDEX IF EXISTS idx_events_upcoming;

DROP INDEX IF EXISTS idx_events_modified;

DROP INDEX IF EXISTS idx_events_cancelled;

DROP INDEX IF EXISTS idx_events_checkin_window;

DROP INDEX IF EXISTS idx_events_unique;

DROP INDEX IF EXISTS idx_events_deleted_at;

DROP INDEX IF EXISTS idx_events_form;

DROP INDEX IF EXISTS idx_events_form_field_overrides;

-- Drop table
DROP TABLE IF EXISTS events;