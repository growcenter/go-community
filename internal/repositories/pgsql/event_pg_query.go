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
	queryGetEventAndInstancesByEventCode = `
		SELECT 
			e.code AS event_code,
			e.title AS event_title,
			e.location AS event_location,
			e.description AS event_description,
			e.campus_code AS event_campus_code,
			e.is_recurring AS event_is_recurring,
			e.event_start_at AS event_start_at,
			e.event_end_at AS event_end_at,
			e.register_start_at AS event_register_start_at,
			e.register_end_at AS event_register_end_at,
			e.status AS event_status,
			ei.code AS instance_code,
			ei.title AS instance_title,
			ei.location AS instance_location,
			ei.instance_start_at AS instance_start_at,
			ei.instance_end_at AS instance_end_at,
			ei.register_start_at AS instance_register_start_at,
			ei.register_end_at AS instance_register_end_at,
			ei.description AS instance_description,
			ei.max_register as instance_max_register,
			ei.total_seats AS instance_total_seats,
			ei.booked_seats as instance_booked_seats,
			ei.scanned_seats as instance_scanned_seats,
			ei.status AS instance_status,
			COALESCE(SUM(ei.total_seats - ei.booked_seats), 0) AS total_remaining_seats
		FROM 
			events e
		LEFT JOIN 
			event_instances ei ON e.code = ei.event_code
		WHERE 
			e.code = ?
		GROUP BY e.code, e.title, e.location, e.description, e.campus_code, e.is_recurring, e.event_start_at, e.event_end_at, e.register_start_at, e.register_end_at, e.status, ei.code, ei.title, ei.location, ei.instance_start_at, ei.instance_end_at, ei.register_start_at, ei.register_end_at, ei.description, ei.max_register, ei.total_seats, ei.booked_seats, ei.scanned_seats, ei.status;
	`
)
