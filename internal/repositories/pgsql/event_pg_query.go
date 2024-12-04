package pgsql

var (
	queryCheckEventByCode             = "SELECT EXISTS (SELECT 1 FROM events WHERE code = ?)"
	queryGetAllEventsByRolesAndStatus = `
    SELECT 
        e.code AS event_code,
        e.title AS event_title,
        e.location AS event_location,
        e.campus_code AS event_campus_code,
        e.is_recurring AS event_is_recurring,
        e.event_start_at AS event_start_at,
        e.event_end_at AS event_end_at,
        e.register_start_at as event_register_start_at,
        e.register_end_at as event_register_end_at,
        e.status AS event_status,
        COALESCE(SUM(ei.total_seats - ei.booked_seats), 0) AS total_remaining_seats,
        ei.is_required AS instance_is_required
    FROM 
        events e
    LEFT JOIN 
        event_instances ei ON e.code = ei.event_code
    WHERE 
        e.allowed_roles && ?::text[] 
        AND e.status = ?
    GROUP BY 
        e.code, e.title, e.location, e.description, e.campus_code, e.is_recurring, e.event_start_at, e.event_end_at, e.register_start_at, e.register_end_at, e.status, ei.is_required
`
	queryGetEventInstancesByEventCode = `
		SELECT 
			e.code AS event_code,
			e.title AS event_title,
			e.location AS event_location,
			e.description AS event_description,
			e.campus_code AS event_campus_code,
			e.allowed_roles AS event_allowed_roles,
			e.is_recurring AS event_is_recurring,
			e.recurrence AS event_recurrence,
			e.event_start_at AS event_start_at,
			e.event_end_at AS event_end_at,
			e.register_start_at AS event_register_start_at,
			e.register_end_at AS event_register_end_at,
			e.status AS event_status
		FROM 
			events e
		WHERE 
			e.code = ? and e.deleted_at is null
	`
)
