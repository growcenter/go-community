package pgsql

var (
	queryCountEventInstanceByCode = `SELECT COUNT(*) FROM event_instances WHERE event_code = ?`

	queryGetSessionsByEventCode = `
		SELECT 
			code, title, location, description, event_code, register_start_at, register_end_at, instance_start_at, instance_end_at, max_register, total_seats, booked_seats, scanned_seats, is_required, status, COALESCE(SUM(total_seats - booked_seats), 0) AS total_remaining_seats
		FROM 
			event_instances
		WHERE 
			event_code = ? and deleted_at is null 
		GROUP BY code, title, location, description, event_code, register_start_at, register_end_at, instance_start_at, instance_end_at, max_register, total_seats, booked_seats, scanned_seats, is_required, status
`
)
