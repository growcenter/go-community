-- Drop trigger first (depends on function)
DROP TRIGGER IF EXISTS trg_events_updated_at ON events;

DROP FUNCTION IF EXISTS update_events_updated_at ();

-- Drop all indexes
DROP INDEX IF EXISTS idx_events_status_start;

DROP INDEX IF EXISTS idx_events_access_level_status;

DROP INDEX IF EXISTS idx_events_category_status;

DROP INDEX IF EXISTS idx_events_creator_status;

DROP INDEX IF EXISTS idx_events_creator;

DROP INDEX IF EXISTS idx_events_date_range;

DROP INDEX IF EXISTS idx_events_recurring;

DROP INDEX IF EXISTS idx_events_templates;

DROP INDEX IF EXISTS idx_events_deleted_at;

DROP INDEX IF EXISTS idx_events_allowed_user_types;

DROP INDEX IF EXISTS idx_events_allowed_roles;

DROP INDEX IF EXISTS idx_events_allowed_campuses;

DROP INDEX IF EXISTS idx_events_allowed_community_ids;

DROP INDEX IF EXISTS idx_events_organizer_community_ids;

DROP INDEX IF EXISTS idx_events_contact_community_ids;

DROP INDEX IF EXISTS idx_events_recurrence_pattern;

DROP INDEX IF EXISTS idx_events_reminder_config;

DROP INDEX IF EXISTS idx_events_search;

-- Drop table (constraints dropped automatically)
DROP TABLE IF EXISTS events;