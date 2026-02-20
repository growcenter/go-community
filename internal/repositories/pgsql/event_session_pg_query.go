package pgsql

var (
	// QueryCheckSessionByCode checks if a session exists with the given code
	// Returns true if a session with the given code exists
	QueryCheckSessionByCode = `
        SELECT EXISTS(
            SELECT 1 FROM event_sessions 
            WHERE code = ? AND deleted_at IS NULL
        )
	`

	// QueryCheckSessionByEventCode checks if a session exists with the given event code
	// Returns true if a session with the given event code exists
	QueryCheckSessionByEventCode = `
        SELECT EXISTS(
            SELECT 1 FROM event_sessions 
            WHERE event_code = ? AND deleted_at IS NULL
        )
	`
)
