package pgsql

import (
	"fmt"
	"strings"
)

var (
	queryCountAttendanceByInstanceCode = `SELECT COUNT(*) FROM event_attendances WHERE instance_code = ? AND deleted_at IS NULL`
	queryCountAttendanceByStatus       = `
		SELECT
			COALESCE(SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END), 0) AS pending,
			COALESCE(SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END), 0) AS success,
			COALESCE(SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END), 0) AS cancelled
		FROM
			event_attendances
		WHERE
			instance_code = ? AND deleted_at IS NULL
	`
	queryCheckAttendanceByCode             = `SELECT EXISTS (SELECT 1 FROM event_attendances WHERE code = ?)`
	queryCountByCommunityIdAndInstanceCode = `
		SELECT
			COUNT(ea.code)
		FROM
			event_attendances ea
		LEFT JOIN
			event_registrations er ON ea.registration_code = er.code
		WHERE
			er.community_id = ?
		AND
			ea.instance_code = ?
		AND
			ea.status IN ('success', 'pending')
		AND
			ea.deleted_at IS NULL
		AND
			er.deleted_at IS NULL
	`
	baseQueryCheckBulkIdentifier = `
		SELECT EXISTS (
			SELECT 1
			FROM event_attendances
			WHERE instance_code = ? AND (%s)
		)
	`
)

// QueryCheckByIdentifiersAndInstanceCode dynamically builds a query to check for the existence of multiple identifiers.
func QueryCheckByIdentifiersAndInstanceCode(identifiers map[string][]string) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	for idType, values := range identifiers {
		if len(values) == 0 {
			continue
		}
		var column string
		switch idType {
		case "email":
			column = "email"
		case "phoneNumber":
			column = "phone_number"
		case "legalId":
			column = "legal_id"
		case "communityId":
			column = "community_id"
		default:
			continue
		}
		conditions = append(conditions, fmt.Sprintf("%s IN (?)", column))
		args = append(args, values)
	}

	if len(conditions) == 0 {
		return "", nil
	}

	return fmt.Sprintf(baseQueryCheckBulkIdentifier, strings.Join(conditions, " OR ")), args
}
