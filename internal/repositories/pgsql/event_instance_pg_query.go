package pgsql

var (
	queryCountEventInstanceByCode = `SELECT COUNT(*) FROM event_instances WHERE event_code = ?`

	queryGetSessionsByEventCode = `
		SELECT 
			ei.code AS instance_code, 
			ei.title AS instance_title, 
			ei.description AS instance_description, 
			ei.instance_start_at, 
			ei.instance_end_at, 
			ei.register_start_at AS instance_register_start_at, 
			ei.register_end_at AS instance_register_end_at, 
			ei.location_type AS instance_location, 
			ei.location_name AS instance_location_name, 
			ei.max_per_transaction AS instance_max_per_transaction, 
			ei.is_one_per_account AS instance_is_one_per_account, 
			ei.is_one_per_ticket AS instance_is_one_per_ticket, 
			ei.register_flow AS instance_register_flow, 
			ei.check_type as instance_check_type, 
			ei.total_seats AS instance_total_seats, 
			ei.booked_seats AS instance_booked_seats, 
			ei.scanned_seats AS instance_scanned_seats, 
			ei.status AS instance_status, 
			COALESCE(SUM(ei.total_seats - ei.booked_seats), 0) AS total_remaining_seats, 
			e.allowed_for AS event_allowed_for
		FROM 
			event_instances ei
				LEFT JOIN events e 
				ON ei.event_code = e.code
		WHERE 
			ei.event_code = ? AND 
			ei.status = ? AND 
			ei.deleted_at IS NULL 
		GROUP BY 
			ei.code, ei.title, ei.description, ei.instance_start_at, ei.instance_end_at, ei.register_start_at, ei.register_end_at, ei.location_type, ei.location_name, ei.max_per_transaction, ei.is_one_per_account, ei.is_one_per_ticket, ei.register_flow, ei.check_type, ei.total_seats, ei.booked_seats, ei.scanned_seats, ei.status, e.allowed_for
`

	queryGetSessionByCode = `
		SELECT 
			ei.code AS instance_code, 
			ei.event_code AS instance_event_code,
			ei.title AS instance_title, 
			ei.description AS instance_description, 
			ei.instance_start_at, 
			ei.instance_end_at, 
			ei.register_start_at AS instance_register_start_at, 
			ei.register_end_at AS instance_register_end_at, 
			ei.location_type AS instance_location, 
			ei.location_name AS instance_location_name, 
			ei.max_per_transaction AS instance_max_per_transaction, 
			ei.is_one_per_account AS instance_is_one_per_account, 
			ei.is_one_per_ticket AS instance_is_one_per_ticket, 
			ei.register_flow AS instance_register_flow, 
			ei.check_type AS instance_check_type, 
			ei.total_seats AS instance_total_seats, 
			ei.booked_seats AS instance_booked_seats, 
			ei.scanned_seats AS instance_scanned_seats, 
			ei.status AS instance_status, 
			COALESCE(SUM(ei.total_seats - ei.booked_seats), 0) AS total_remaining_seats
-- 			e.allowed_for AS event_allowed_for
		FROM 
			event_instances ei
-- 				LEFT JOIN events e 
-- 				ON ei.event_code = e.code
		WHERE 
			ei.code = ? AND 
			ei.status = ? AND 
			ei.deleted_at IS NULL 
		GROUP BY 
			ei.code, ei.event_code, ei.title, ei.description, ei.instance_start_at, ei.instance_end_at, ei.register_start_at, ei.register_end_at, ei.location_type, ei.location_name, ei.max_per_transaction, ei.is_one_per_account, ei.is_one_per_ticket, ei.register_flow, ei.check_type, ei.total_seats, ei.booked_seats, ei.scanned_seats, ei.status
`
	queryGetSeatsByInstanceCode = `SELECT ei.total_seats as total_seats,
		   ei.booked_seats as booked_seats,
		   ei.scanned_seats as scanned_seats,
		   ei.title as event_instance_title,
		   e.title as event_title,
		   coalesce(sum(ei.total_seats - ei.booked_seats), 0) as total_remaining_seats
	from event_instances ei
		left join events e on ei.event_code = e.code
	where ei.code = ?
	group by ei.total_seats, ei.booked_seats, ei.scanned_seats, ei.title, e.title`
)
