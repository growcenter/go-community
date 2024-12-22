package pgsql

import (
	"fmt"
	"strings"
	"time"
)

var (
	queryCountRecordByIdentifierOriginAndStatus = `SELECT COUNT(*) FROM event_registration_records WHERE identifier_origin = ? AND status = ?`
	queryCountRecordByCommunityIdOrigin         = `SELECT COUNT(*) FROM event_registration_records WHERE community_id_origin = ?`
	queryCheckRecordByIdentifier                = `SELECT EXISTS (SELECT 1 FROM event_registration_records WHERE identifier = ?)`
	queryCheckRecordByName                      = `SELECT EXISTS (SELECT 1 FROM event_registration_records WHERE name = ?)`
	queryCheckRecordByCommunityId               = `SELECT EXISTS (SELECT 1 FROM event_registration_records WHERE community_id = ?)`
	queryGetEventAttendance                     = `
		SELECT 
			er.community_id,
			er.event_code,
			e.title,
			COUNT(CASE WHEN er.status = 'success' THEN 1 END) AS success_count,
			COUNT(CASE WHEN er.status = 'permit' AND (er.reason IS NOT NULL AND er.reason != '') THEN 1 END) AS permit_with_reason_count,
			COUNT(CASE WHEN er.status = 'permit' AND (er.reason IS NULL OR er.reason = '') THEN 1 END) AS permit_without_reason_count,
			COUNT(CASE WHEN er.status NOT IN ('success', 'permit') THEN 1 END) AS other_status_count,
			COUNT(*) AS total_instances
		FROM 
			event_registration_records AS er
		INNER JOIN 
			events AS e ON er.event_code = e.code
		WHERE 
			er.community_id = ?
			AND e.is_recurring = TRUE
			AND er.deleted_at IS NULL
			AND er.registered_at BETWEEN ? AND ?
		GROUP BY 
			e.title, er.event_code, er.community_id
	`

	queryPrevGetAllRegistrationRecord = `
			SELECT 
				er.id,
				er.name,
				er.identifier,
				er.community_id,
				er.event_code,
				er.instance_code,
				er.identifier_origin,
				er.community_id_origin,
				er.updated_by,
				er.status,
				er.reason,
				er.registered_at,
				er.verified_at,
				er.created_at,
				er.updated_at,
				er.deleted_at,
				e.title AS event_name,
				i.title AS instance_name
			FROM event_registration_records er
			LEFT JOIN events e ON er.event_code = e.code
			LEFT JOIN event_instances i ON er.instance_code = i.code
			WHERE er.updated_at > ? AND er.event_code = ?
			ORDER BY er.updated_at ASC
			LIMIT ?
		`

	queryNextGetAllRegistrationRecord = `
			SELECT 
				er.id,
				er.name,
				er.identifier,
				er.community_id,
				er.event_code,
				er.instance_code,
				er.identifier_origin,
				er.community_id_origin,
				er.updated_by,
				er.status,
				er.reason,
				er.registered_at,
				er.verified_at,
				er.created_at,
				er.updated_at,
				er.deleted_at,
				e.title AS event_name,
				i.title AS instance_name
			FROM event_registration_records er
			LEFT JOIN events e ON er.event_code = e.code
			LEFT JOIN event_instances i ON er.instance_code = i.code
			WHERE er.updated_at < ? AND er.event_code = ?
			ORDER BY er.updated_at DESC
			LIMIT ?
		`

	baseQueryGetRegisteredRecordList = `
		SELECT 
			er.id,
			er.name,
			er.identifier,
			er.community_id,
			er.event_code,
			er.instance_code,
			er.identifier_origin,
			er.community_id_origin,
			er.updated_by,
			er.status,
			er.reason,
			er.registered_at,
			er.verified_at,
			er.created_at,
			er.updated_at,
			er.deleted_at,
			e.title AS event_name,
			i.title AS instance_name,
			u.name AS registered_by
		FROM event_registration_records er
		LEFT JOIN events e ON er.event_code = e.code
		LEFT JOIN event_instances i ON er.instance_code = i.code
		LEFT JOIN users u ON er.community_id_origin = u.community_id
		WHERE 1=1
	`

	queryCountEventAllRegistered = `
		SELECT COUNT(*)
		FROM event_registration_records er
		WHERE 1=1
	`
)

func BuildEventRegistrationQuery(baseQuery string, eventCode string, nameSearch string, cursor time.Time, direction string) (string, []interface{}, error) {
	var conditions []string
	var params []interface{}

	// Add conditions dynamically
	if eventCode != "" {
		conditions = append(conditions, "er.event_code = ?")
		params = append(params, eventCode)
	}
	if nameSearch != "" {
		conditions = append(conditions, "er.name ILIKE ?")
		params = append(params, "%"+nameSearch+"%")
	}
	if !cursor.IsZero() {
		if direction == "next" {
			conditions = append(conditions, "er.updated_at > ?")
		} else if direction == "prev" {
			conditions = append(conditions, "er.updated_at < ?")
		} else {
			return "", nil, fmt.Errorf("invalid direction: %s, must be 'next' or 'prev'", direction)
		}
		params = append(params, cursor)
	}

	// Build WHERE clause
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Add ordering
	if direction == "next" {
		baseQuery += " ORDER BY er.updated_at ASC"
	} else if direction == "prev" {
		baseQuery += " ORDER BY er.updated_at DESC"
	}

	// Add limit placeholder
	baseQuery += " LIMIT ?"
	params = append(params, 100) // Default limit for now, can be adjusted

	return baseQuery, params, nil
}
