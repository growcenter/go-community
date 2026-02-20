-- Drop all constraints
ALTER TABLE registrations
DROP CONSTRAINT IF EXISTS chk_registrations_status;

ALTER TABLE registrations
DROP CONSTRAINT IF EXISTS chk_registrations_method;

ALTER TABLE registrations
DROP CONSTRAINT IF EXISTS chk_registrations_channel;

ALTER TABLE registrations
DROP CONSTRAINT IF EXISTS chk_registrations_waitlist_position;

ALTER TABLE registrations
DROP CONSTRAINT IF EXISTS chk_registrations_confirmed_at;

ALTER TABLE registrations
DROP CONSTRAINT IF EXISTS chk_registrations_geolocation_accuracy;

ALTER TABLE registrations
DROP CONSTRAINT IF EXISTS chk_registrations_geolocation_distance;

-- Drop all indexes
DROP INDEX IF EXISTS idx_registrations_event;

DROP INDEX IF EXISTS idx_registrations_session;

DROP INDEX IF EXISTS idx_registrations_occurrence;

DROP INDEX IF EXISTS idx_registrations_registrant;

DROP INDEX IF EXISTS idx_registrations_registered_by;

DROP INDEX IF EXISTS idx_registrations_group;

DROP INDEX IF EXISTS idx_registrations_event_status;

DROP INDEX IF EXISTS idx_registrations_session_status;

DROP INDEX IF EXISTS idx_registrations_registrant_status;

DROP INDEX IF EXISTS idx_registrations_waitlist;

DROP INDEX IF EXISTS idx_registrations_group_members;

DROP INDEX IF EXISTS idx_registrations_method;

DROP INDEX IF EXISTS idx_registrations_registered_at;

DROP INDEX IF EXISTS idx_registrations_confirmed_at;

DROP INDEX IF EXISTS idx_registrations_pending;

DROP INDEX IF EXISTS idx_registrations_confirmed;

DROP INDEX IF EXISTS idx_registrations_deleted_at;

DROP INDEX IF EXISTS idx_registrations_form_answers;

DROP INDEX IF EXISTS idx_registrations_geolocation_failed;

-- Drop table
DROP TABLE IF EXISTS registrations;