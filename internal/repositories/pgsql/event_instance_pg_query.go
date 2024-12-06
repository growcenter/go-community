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
			ei.max_per_transaction AS instance_max_register, 
			ei.is_required AS instance_is_required, 
			ei.is_one_per_account AS instance_is_one_per_account, 
			ei.is_one_per_ticket AS instance_is_one_per_ticket, 
			ei.allow_personal_qr, 
			ei.attendance_type, 
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
			ei.code, 
			ei.title, 
			ei.description, 
			ei.instance_start_at, 
			ei.instance_end_at, 
			ei.register_start_at, 
			ei.register_end_at, 
			ei.location_type, 
			ei.location_name, 
			ei.max_per_transaction, 
			ei.is_required, 
			ei.is_one_per_account, 
			ei.is_one_per_ticket, 
			ei.allow_personal_qr, 
			ei.attendance_type, 
			ei.total_seats, 
			ei.booked_seats, 
			ei.scanned_seats, 
			ei.status, 
			e.allowed_for
`
)
