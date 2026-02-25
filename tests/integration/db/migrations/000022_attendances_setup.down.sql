-- Drop all constraints
ALTER TABLE attendances
DROP CONSTRAINT IF EXISTS chk_attendances_status;

ALTER TABLE attendances
DROP CONSTRAINT IF EXISTS chk_attendances_check_in_method;

ALTER TABLE attendances
DROP CONSTRAINT IF EXISTS chk_attendances_check_out_method;

ALTER TABLE attendances
DROP CONSTRAINT IF EXISTS chk_attendances_check_out_after_check_in;

ALTER TABLE attendances
DROP CONSTRAINT IF EXISTS chk_attendances_late_minutes;

ALTER TABLE attendances
DROP CONSTRAINT IF EXISTS chk_attendances_check_in_geolocation_accuracy;

ALTER TABLE attendances
DROP CONSTRAINT IF EXISTS chk_attendances_check_out_geolocation_accuracy;

ALTER TABLE attendances
DROP CONSTRAINT IF EXISTS chk_attendances_geolocation_distance;

-- Drop all indexes
DROP INDEX IF EXISTS idx_attendances_registration;

DROP INDEX IF EXISTS idx_attendances_event;

DROP INDEX IF EXISTS idx_attendances_session;

DROP INDEX IF EXISTS idx_attendances_occurrence;

DROP INDEX IF EXISTS idx_attendances_attendee;

DROP INDEX IF EXISTS idx_attendances_event_status;

DROP INDEX IF EXISTS idx_attendances_session_status;

DROP INDEX IF EXISTS idx_attendances_attendee_event;

DROP INDEX IF EXISTS idx_attendances_check_in_by;

DROP INDEX IF EXISTS idx_attendances_check_out_by;

DROP INDEX IF EXISTS idx_attendances_check_in_date;

DROP INDEX IF EXISTS idx_attendances_duration;

DROP INDEX IF EXISTS idx_attendances_present;

DROP INDEX IF EXISTS idx_attendances_late;

DROP INDEX IF EXISTS idx_attendances_absent;

DROP INDEX IF EXISTS idx_attendances_walk_ins;

DROP INDEX IF EXISTS idx_attendances_device_info;

DROP INDEX IF EXISTS idx_attendances_geolocation_failed;

-- Drop table
DROP TABLE IF EXISTS attendances;