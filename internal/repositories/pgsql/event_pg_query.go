package pgsql

var (
	queryCheckEventByCode = "SELECT EXISTS (SELECT 1 FROM events WHERE code = ?)"
	//queryGetAllEventsByRolesAndStatus = `
	//   SELECT
	//       e.code AS event_code,
	//       e.title AS event_title,
	//       e.topics AS event_topics,
	//       e.location_type AS event_location_type,
	//       e.allowed_for AS event_allowed_for,
	//       e.allowed_roles AS event_allowed_roles,
	//       e.allowed_users AS event_allowed_users,
	//       e.allowed_campuses AS event_allowed_campuses,
	//       e.is_recurring AS event_is_recurring,
	//       e.recurrence AS event_recurrence,
	//       e.event_start_at AS event_start_at,
	//       e.event_end_at AS event_end_at,
	//       e.register_start_at as event_register_start_at,
	//       e.register_end_at as event_register_end_at,
	//       e.status AS event_status,
	//       COALESCE(SUM(COALESCE(ei.total_seats, 0) - COALESCE(ei.booked_seats, 0)), 0) AS total_remaining_seats,
	//       COALESCE(ei.total_seats, 0) AS instance_total_seats,
	//       ARRAY_AGG(ei.is_required) AS instance_is_required
	//   FROM
	//       events e
	//   LEFT JOIN
	//       event_instances ei ON e.code = ei.event_code
	//   WHERE
	//      (e.allowed_roles && ?::text[] OR e.allowed_users && ?::text[])
	//       AND e.status = ?
	//   GROUP BY
	//       e.code, e.title, e.topics, e.location_type, e.allowed_for, e.allowed_roles, e.allowed_users, e.allowed_campuses, e.is_recurring, e.recurrence, e.event_start_at, e.event_end_at, e.register_start_at, e.register_end_at, e.status, ei.is_required, ei.total_seats
	//`

	queryGetAllEventsByRolesAndStatus = `
	SELECT
		e.code AS event_code,
		e.title AS event_title,
		e.topics AS event_topics,
		e.location_type AS event_location_type,
		e.allowed_for AS event_allowed_for,
		e.allowed_roles AS event_allowed_roles,
		e.allowed_users AS event_allowed_users,
		e.allowed_campuses AS event_allowed_campuses,
		e.is_recurring AS event_is_recurring,
		e.recurrence AS event_recurrence,
		e.event_start_at AS event_start_at,
		e.event_end_at AS event_end_at,
		e.register_start_at AS event_register_start_at,
		e.register_end_at AS event_register_end_at,
		e.status AS event_status,
		COALESCE(SUM(COALESCE(ei.total_seats, 0) - COALESCE(ei.booked_seats, 0)), 0) AS total_remaining_seats,
		COALESCE(MAX(ei.total_seats), 0) AS instance_total_seats, -- Ensure no duplicates here
		ARRAY_AGG(
				ROW(ei.total_seats, ei.booked_seats, ei.register_flow)
		) AS instances_data -- Combine all instance data into a JSON array
	FROM
		events e
			LEFT JOIN
		event_instances ei ON e.code = ei.event_code
	WHERE
		(
			(e.allowed_roles && ?::text[] OR e.allowed_users && ?::text[])
				OR
			(
				e.allowed_for = 'public'
				)
			)
	  AND e.status = ?
	GROUP BY
		e.code, e.title, e.topics, e.location_type, e.allowed_for, e.allowed_roles, e.allowed_users, e.allowed_campuses, e.is_recurring, e.recurrence, e.event_start_at, e.event_end_at, e.register_start_at, e.register_end_at, e.status;
`

	queryGetEventInstancesByEventCode = `
		SELECT
			e.code AS event_code,
			e.title AS event_title,
			COALESCE(e.topics, ARRAY[]::TEXT[]) AS event_topics, -- Nullable text
			COALESCE(e.description, '') AS event_description, -- Nullable text
			COALESCE(e.terms_and_conditions, '') AS event_terms_and_conditions, -- Nullable text
			e.allowed_for AS event_allowed_for, -- Non-nullable
			COALESCE(e.allowed_roles, ARRAY[]::TEXT[]) AS event_allowed_roles, -- Default to empty array
			COALESCE(e.allowed_users, ARRAY[]::TEXT[]) AS event_allowed_users, -- Default to empty array
			COALESCE(e.allowed_campuses, ARRAY[]::TEXT[]) AS event_allowed_campuses, -- Default to empty array
			e.is_recurring AS event_is_recurring, -- Nullable boolean
			COALESCE(e.recurrence, '') AS event_recurrence, -- Nullable text
			e.event_start_at AS event_start_at, -- Non-nullable
			e.event_end_at AS event_end_at, -- Non-nullable
			e.register_start_at AS event_register_start_at, -- Non-nullable
			e.register_end_at AS event_register_end_at, -- Non-nullable
			e.location_type AS event_location_type, -- Non-nullable
			e.location_name AS event_location_name, -- Non-nullable
			e.status AS event_status, -- Non-nullable
			COALESCE(ei.total_seats, 0) AS instance_total_seats,
			COALESCE(SUM(COALESCE(ei.total_seats, 0) - COALESCE(ei.booked_seats, 0)), 0) AS total_remaining_seats,
			ARRAY_AGG(
				ROW(COALESCE(ei.total_seats, 0), COALESCE(ei.booked_seats, 0), ei.register_flow)
			) AS instances_data  -- Nullable boolean
		FROM
			events e
				LEFT JOIN
			event_instances ei ON e.code = ei.event_code
		WHERE
			e.code = ?
		  AND e.deleted_at IS NULL
		GROUP BY
			e.code, e.title, COALESCE(e.topics, ARRAY[]::TEXT[]), COALESCE(e.description, ''), COALESCE(e.terms_and_conditions, ''), e.allowed_for, COALESCE(e.allowed_roles, ARRAY[]::TEXT[]), COALESCE(e.allowed_users, ARRAY[]::TEXT[]), COALESCE(e.allowed_campuses, ARRAY[]::TEXT[]), e.is_recurring, COALESCE(e.recurrence, ''), e.event_start_at, e.event_end_at, e.register_start_at, e.register_end_at, e.location_type, e.location_name, e.status, ei.total_seats
`
	queryGetRegisteredUserByCommunityIdOrigin = `
	SELECT DISTINCT 
		e.code AS event_code,
		e.title AS event_title,
		e.description AS event_description,
		e.terms_and_conditions AS event_terms_and_conditions,
		e.event_start_at AS event_start_at,
		e.event_end_at AS event_end_at,
		e.location_type AS event_location_type,
		e.location_name AS event_location_name,
		e.status AS event_status,
		ei.code AS instance_code,
		ei.title AS instance_title,
		ei.description AS instance_description,
		ei.instance_start_at AS instance_start_at,
		ei.instance_end_at AS instance_end_at,
		ei.location_type AS instance_location_type,
		ei.location_name AS instance_location_name,
		ei.status AS instance_status,
		rr.id AS registration_record_id,
		rr.name AS registration_record_name,
		coalesce(rr.identifier, '') AS registration_record_identifier,
		coalesce(rr.community_id, '') AS registration_record_community_id,
		coalesce(rr.updated_by, '') AS registration_record_updated_by,
		rr.registered_at AS registration_record_registered_at,
		rr.verified_at as registration_record_verified_at,
		rr.status AS registration_record_status
	FROM
		events e
		JOIN
			event_instances ei ON e.code = ei.event_code
		JOIN
			event_registration_records rr ON rr.instance_code = ei.code
	WHERE
		rr.community_id_origin = ?
`
	queryGetEventTitles = `SELECT code, title FROM events WHERE deleted_at IS NULL`
)
