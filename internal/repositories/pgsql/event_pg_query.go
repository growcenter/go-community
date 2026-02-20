package pgsql

var (
	// QueryCheckEventByCode checks if an event exists with the given code
	// Returns true if an event with the given code exists
	QueryCheckEventByCode = `
        SELECT EXISTS(
            SELECT 1 FROM events 
            WHERE code = ? AND deleted_at IS NULL
        )
	`

	// QueryCheckEventBySlug checks if an event exists with the given slug
	// Returns true if an event with the given slug exists
	QueryCheckEventBySlug = `
        SELECT EXISTS(
            SELECT 1 FROM events 
            WHERE slug = ? AND deleted_at IS NULL
        )
	`

	// QueryCheckEventByCodeOrSlug checks if an event exists with the given code or slug
	// Returns true if an event with the given code or slug exists
	QueryCheckEventByCodeOrSlug = `
        SELECT EXISTS(
            SELECT 1 FROM events 
            WHERE (code = ? OR slug = ?) AND deleted_at IS NULL
        )
    `
)
