package pgsql

var (
	queryCheckAttendanceOnMeetingId = `
		SELECT EXISTS (
			SELECT 1 FROM cool_attendances 
			WHERE cool_meeting_id = ? 
			AND deleted_at IS NULL
		)
	`

	queryGetAllAttendanceOnMeetingId = `
	SELECT
		u.community_id as community_id,
		u.name as name,
		ca.id as attendance_id,
		ca.cool_meeting_id as cool_meeting_id,
		ca.is_present as is_present,
		ca.remarks as remarks
	FROM
		cool_meetings cm
			JOIN users u ON u.cool_code = cm.cool_code
			LEFT JOIN cool_attendances ca ON ca.cool_meeting_id = cm.id
			AND ca.community_id IS NOT DISTINCT FROM u.community_id
	WHERE
		cm.id = ?;
	`

	queryGetSummaryAttendanceOnMeetingId = `
	SELECT
		COUNT(ca.id) as total,
		COUNT(CASE WHEN ca.is_present = true THEN 1 ELSE NULL END) as present,
		COUNT(CASE WHEN ca.is_present = false THEN 1 ELSE NULL END) as absent
	FROM
		cool_meetings cm
			JOIN users u ON u.cool_code = cm.cool_code
			LEFT JOIN cool_attendances ca ON ca.cool_meeting_id = cm.id
			AND ca.community_id IS NOT DISTINCT FROM u.community_id
	WHERE
		cm.id =?;
	`

	queryGetSummaryAttendanceByCoolCode = `
	SELECT
		u.name,
		u.community_id,
		COUNT(CASE WHEN ca.is_present = true THEN 1 END) as present_count,
		COUNT(CASE WHEN ca.is_present = false THEN 1 END) as absent_count,
		COUNT(DISTINCT cm.id) as total_meeting_count
	FROM
		users u
		LEFT JOIN cool_meetings cm ON u.cool_code = cm.cool_code
		LEFT JOIN cool_attendances ca ON ca.cool_meeting_id = cm.id AND ca.community_id = u.community_id
	WHERE
		u.cool_code = ?
		AND cm.meeting_date BETWEEN ? AND ?
		AND cm.deleted_at IS NULL
		AND u.deleted_at IS NULL
	GROUP BY
		u.community_id, u.name
	ORDER BY
		u.name;
	`
)
