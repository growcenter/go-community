DROP TRIGGER IF EXISTS trg_event_sessions_updated_at ON event_sessions;

DROP FUNCTION IF EXISTS update_sessions_updated_at ();

DROP TABLE IF EXISTS event_sessions;