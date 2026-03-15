-- Drop all constraints
ALTER TABLE event_volunteers
DROP CONSTRAINT IF EXISTS chk_volunteers_role;

-- Drop all indexes
DROP INDEX IF EXISTS idx_volunteers_event;

DROP INDEX IF EXISTS idx_volunteers_session;

DROP INDEX IF EXISTS idx_volunteers_community;

DROP INDEX IF EXISTS idx_volunteers_event_role;

DROP INDEX IF EXISTS idx_volunteers_community_role;

DROP INDEX IF EXISTS idx_volunteers_assigned_by;

-- Drop table
DROP TABLE IF EXISTS event_volunteers;