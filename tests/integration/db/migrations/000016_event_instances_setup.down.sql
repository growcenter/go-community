-- Drop indexes first
DROP INDEX IF EXISTS idx_event_instances_type_status;

DROP INDEX IF EXISTS idx_event_instances_family;

DROP INDEX IF EXISTS idx_event_instances_age;

DROP INDEX IF EXISTS idx_event_instances_type;

-- Drop table
DROP TABLE IF EXISTS "event_instances";