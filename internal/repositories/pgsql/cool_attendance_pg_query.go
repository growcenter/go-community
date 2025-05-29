package pgsql

var (
	queryCheckAttendanceOnMeetingId = `
		SELECT EXISTS (
			SELECT 1 FROM cool_attendances 
			WHERE cool_meeting_id = ? 
			AND deleted_at IS NULL
		)
	`
)
