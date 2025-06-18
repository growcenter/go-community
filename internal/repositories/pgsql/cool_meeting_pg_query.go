package pgsql

var (
	queryCheckMeetingOnDateExistByCoolCode = `
		SELECT EXISTS (
			SELECT 1 FROM cool_meetings 
			WHERE cool_code = ? 
			AND meeting_date = ? 
			AND deleted_at IS NULL
		)
	`

	queryGetMeetingsByCoolCodeAndDate = `
		SELECT
			id,
			cool_code,
			name,
			topic,
			description,
			location_type,
			location_name,
			meeting_date,
			meeting_start_at,
			meeting_end_at
		FROM
			cool_meetings
		WHERE
			cool_code = ? AND meeting_date >= ? AND deleted_at IS NULL
	`

	queryGetMeetingsWithAttendanceByCoolCodeAndDate = `
		SELECT
			cm.id,
			cm.cool_code,
			cm.name,
			cm.topic,
			cm.description,
			cm.location_type,
			cm.location_name,
			cm.meeting_date,
			cm.meeting_start_at,
			cm.meeting_end_at,
			ca.id as attendance_id,
			ca.community_id,
			ca.is_present,
			ca.remarks
		FROM
			cool_meetings cm
		LEFT JOIN
			cool_attendances ca ON cm.id = ca.cool_meeting_id AND ca.deleted_at IS NULL
		WHERE
			cm.cool_code = ? 
			AND cm.meeting_date BETWEEN ? AND ? 
			AND cm.deleted_at IS NULL
			AND (ca.community_id = ? OR ca.community_id IS NULL)
	`
)
