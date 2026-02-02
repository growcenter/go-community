package pgsql

var (
	// queryGetAllEventsByRolesAndStatusAndRange1Year gets events within current year
	queryGetAllEventsByRolesAndStatusAndRange1Year = `
SELECT
    e.code AS event_code,
    e.title AS event_title,
    e.topics AS event_topics,
    e.location_type AS event_location_type,
    e.access_level AS event_access_level,
    e.allowed_roles AS event_allowed_roles,
    e.allowed_community_ids AS event_allowed_community_ids,
    e.allowed_campuses AS event_allowed_campuses,
    e.recurrence AS event_recurrence,
    e.start_at AS event_start_at,
    e.end_at AS event_end_at,
    COALESCE(e.image_links, ARRAY[]::TEXT[]) AS event_image_links,
    e.status AS event_status,
    COALESCE(SUM(COALESCE(ei.total_seats, 0) - COALESCE(ei.booked_seats, 0)), 0) AS total_remaining_seats,
    COALESCE(MAX(ei.total_seats), 0) AS instance_total_seats,
    ARRAY_AGG(
        ROW(ei.total_seats, ei.booked_seats, ei.register_flow)
    ) AS instances_data
FROM
    events e
    LEFT JOIN event_instances ei ON e.code = ei.event_code
WHERE
    e.deleted_at IS NULL
    AND (
        (e.allowed_roles && ?::text[] OR e.allowed_community_ids && ?::text[])
        OR e.access_level = 'public'
    )
    AND e.status = ?
    AND e.start_at >= DATE_TRUNC('year', CURRENT_DATE) 
    AND e.start_at < DATE_TRUNC('year', CURRENT_DATE) + INTERVAL '1 year'
GROUP BY
    e.code, e.title, e.topics, e.location_type, e.access_level, 
    e.allowed_roles, e.allowed_community_ids, e.allowed_campuses, 
    e.recurrence, e.start_at, e.end_at, e.status, e.image_links
ORDER BY
    e.start_at DESC;
`

	// queryGetAllEventsByRolesAndStatusAndRangeEventTime gets future events
	queryGetAllEventsByRolesAndStatusAndRangeEventTime = `
SELECT
    e.code AS event_code,
    e.title AS event_title,
    e.topics AS event_topics,
    e.location_type AS event_location_type,
    e.access_level AS event_access_level,
    e.allowed_roles AS event_allowed_roles,
    e.allowed_community_ids AS event_allowed_community_ids,
    e.allowed_campuses AS event_allowed_campuses,
    e.recurrence AS event_recurrence,
    e.start_at AS event_start_at,
    e.end_at AS event_end_at,
    COALESCE(e.image_links, ARRAY[]::TEXT[]) AS event_image_links,
    e.status AS event_status,
    COALESCE(SUM(COALESCE(ei.total_seats, 0) - COALESCE(ei.booked_seats, 0)), 0) AS total_remaining_seats,
    COALESCE(MAX(ei.total_seats), 0) AS instance_total_seats,
    ARRAY_AGG(
        ROW(ei.total_seats, ei.booked_seats, ei.register_flow)
    ) AS instances_data
FROM
    events e
    LEFT JOIN event_instances ei ON e.code = ei.event_code
WHERE
    e.deleted_at IS NULL
    AND (
        (e.allowed_roles && ?::text[] OR e.allowed_community_ids && ?::text[])
        OR e.access_level = 'public'
    )
    AND e.status = ?
    AND e.start_at > CURRENT_DATE
GROUP BY
    e.code, e.title, e.topics, e.location_type, e.access_level, 
    e.allowed_roles, e.allowed_community_ids, e.allowed_campuses, 
    e.recurrence, e.start_at, e.end_at, e.status, e.image_links
ORDER BY
    e.start_at DESC;
`

	// queryGetEventInstancesByEventCode gets event details with instances
	queryGetEventInstancesByEventCode = `
SELECT
    e.code AS event_code,
    e.title AS event_title,
    COALESCE(e.topics, ARRAY[]::TEXT[]) AS event_topics,
    COALESCE(e.description, '') AS event_description,
    COALESCE(e.terms_and_conditions, '') AS event_terms_and_conditions,
    e.access_level AS event_access_level,
    COALESCE(e.allowed_roles, ARRAY[]::TEXT[]) AS event_allowed_roles,
    COALESCE(e.allowed_community_ids, ARRAY[]::TEXT[]) AS event_allowed_community_ids,
    COALESCE(e.allowed_campuses, ARRAY[]::TEXT[]) AS event_allowed_campuses,
    COALESCE(e.recurrence, '') AS event_recurrence,
    e.start_at AS event_start_at,
    e.end_at AS event_end_at,
    e.location_type AS event_location_type,
    COALESCE(e.physical_address, '') AS event_physical_address,
    COALESCE(e.virtual_link, '') AS event_virtual_link,
    COALESCE(e.image_links, ARRAY[]::TEXT[]) AS event_image_links,
    e.status AS event_status,
    COALESCE(ei.total_seats, 0) AS instance_total_seats,
    COALESCE(SUM(COALESCE(ei.total_seats, 0) - COALESCE(ei.booked_seats, 0)), 0) AS total_remaining_seats,
    ARRAY_AGG(
        ROW(COALESCE(ei.total_seats, 0), COALESCE(ei.booked_seats, 0), ei.register_flow)
    ) AS instances_data
FROM
    events e
    LEFT JOIN event_instances ei ON e.code = ei.event_code
WHERE
    e.code = ?
    AND e.deleted_at IS NULL
GROUP BY
    e.code, e.title, e.topics, e.description, e.terms_and_conditions, 
    e.access_level, e.allowed_roles, e.allowed_community_ids, e.allowed_campuses, 
    e.recurrence, e.start_at, e.end_at, e.location_type, e.physical_address, 
    e.virtual_link, e.status, ei.total_seats, e.image_links
`

	// queryGetRegisteredUserByCommunityIdOrigin gets registered events for a user
	queryGetRegisteredUserByCommunityIdOrigin = `
SELECT DISTINCT 
    e.code AS event_code,
    e.title AS event_title,
    e.description AS event_description,
    e.terms_and_conditions AS event_terms_and_conditions,
    e.start_at AS event_start_at,
    e.end_at AS event_end_at,
    e.location_type AS event_location_type,
    COALESCE(e.physical_address, '') AS event_physical_address,
    COALESCE(e.virtual_link, '') AS event_virtual_link,
    COALESCE(e.image_links, ARRAY[]::TEXT[]) AS event_image_links,
    e.status AS event_status,
    ei.code AS instance_code,
    ei.title AS instance_title,
    ei.description AS instance_description,
    ei.instance_start_at AS instance_start_at,
    ei.instance_end_at AS instance_end_at,
    ei.location_type AS instance_location_type,
    COALESCE(ei.physical_address, '') AS instance_physical_address,
    COALESCE(ei.virtual_link, '') AS instance_virtual_link,
    ei.status AS instance_status,
    rr.id AS registration_record_id,
    rr.name AS registration_record_name,
    COALESCE(rr.identifier, '') AS registration_record_identifier,
    COALESCE(rr.community_id, '') AS registration_record_community_id,
    COALESCE(rr.updated_by, '') AS registration_record_updated_by,
    rr.registered_at AS registration_record_registered_at,
    rr.verified_at AS registration_record_verified_at,
    rr.status AS registration_record_status
FROM
    events e
    JOIN event_instances ei ON e.code = ei.event_code
    JOIN event_registration_records rr ON rr.instance_code = ei.code
WHERE
    e.deleted_at IS NULL
    AND rr.community_id_origin = ?
`

	// queryGetEventTitles gets all event codes and titles
	queryGetEventTitles = `
SELECT 
    code, 
    title 
FROM 
    events 
WHERE 
    deleted_at IS NULL
ORDER BY 
    created_at DESC
`

	// queryGetEventSummary gets event summary with registration stats
	queryGetEventSummary = `
SELECT
    e.code AS event_code,
    e.title AS event_title,
    e.access_level AS event_access_level,
    COALESCE(e.allowed_roles, ARRAY[]::TEXT[]) AS event_allowed_roles,
    COALESCE(e.allowed_community_ids, ARRAY[]::TEXT[]) AS event_allowed_community_ids,
    COALESCE(e.allowed_campuses, ARRAY[]::TEXT[]) AS event_allowed_campuses,
    COALESCE(SUM(ei.booked_seats), 0) AS total_booked_seats,
    COALESCE(SUM(ei.scanned_seats), 0) AS total_scanned_seats,
    e.status AS event_status
FROM
    events e
    LEFT JOIN event_instances ei ON e.code = ei.event_code
WHERE
    e.code = ?
    AND e.deleted_at IS NULL
GROUP BY
    e.code, e.title, e.access_level, e.allowed_roles, 
    e.allowed_community_ids, e.allowed_campuses, e.status
`
)

// BuildQueryGetAllEvents returns the appropriate query based on user type
func BuildQueryGetAllEvents(isNotGeneral bool) string {
	if isNotGeneral {
		return queryGetAllEventsByRolesAndStatusAndRange1Year
	}
	return queryGetAllEventsByRolesAndStatusAndRangeEventTime
}
