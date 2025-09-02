package pgsql

import (
	"fmt"
	"go-community/internal/models"
	"go-community/internal/pkg/cursor"
	"strconv"
	"strings"
)

var (
	queryCountRecordByIdentifierOriginAndStatus        = `SELECT COUNT(*) FROM event_registration_records WHERE identifier_origin = ? AND status = ?`
	queryCountRecordByCommunityIdOrigin                = `SELECT COUNT(*) FROM event_registration_records WHERE community_id_origin = ?`
	queryCountRecordByCommunityIdOriginAndInstanceCode = `SELECT COUNT(*) FROM event_registration_records WHERE community_id_origin = ? AND instance_code = ?`
	queryCheckRecordByIdentifier                       = `SELECT EXISTS (SELECT 1 FROM event_registration_records WHERE identifier = ?)`
	queryCheckRecordByIdentifierAndInstanceCode        = `SELECT EXISTS (SELECT 1 FROM event_registration_records WHERE identifier = ? AND instance_code = ?)`
	queryCheckRecordByName                             = `SELECT EXISTS (SELECT 1 FROM event_registration_records WHERE name = ?)`
	queryCheckRecordByNameAndInstanceCode              = `SELECT EXISTS (SELECT 1 FROM event_registration_records WHERE name = ? AND instance_code = ?)`
	queryCheckRecordByCommunityId                      = `SELECT EXISTS (SELECT 1 FROM event_registration_records WHERE community_id = ?)`
	queryCheckRecordByCommunityIdAndInstanceCode       = `SELECT EXISTS (SELECT 1 FROM event_registration_records WHERE community_id = ? AND instance_code = ?)`
	queryGetEventAttendance                            = `
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
			coalesce(u.email, '') AS email,
			coalesce(u.phone_number, '') AS phone_number,
			coalesce(u.campus_code, '') AS campus_code,
			coalesce(u.department, '') AS department,
			coalesce(u.cool_code, 0) AS cool_code,
			coalesce(c.name, '') AS cool_name,
			er.event_code,
			er.instance_code,
			er.identifier_origin,
			er.community_id_origin,
			er.updated_by,
			er.status,
			er.reason,
			er.description,
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
			LEFT JOIN users u ON er.community_id = u.community_id
			LEFT JOIN cools c ON u.cool_code = c.code
        WHERE er.deleted_at IS NULL		
	`

	queryCountEventAllRegistered = `
		SELECT COUNT(*)
		FROM event_registration_records er
			LEFT JOIN events e ON er.event_code = e.code
			LEFT JOIN event_instances i ON er.instance_code = i.code
			LEFT JOIN users u ON er.community_id_origin = u.community_id
			LEFT JOIN cools c ON u.cool_code = c.code
	`

	baseQueryGetDownloadRegisteredRecordList = `
		SELECT 
            er.id,
			er.name,
			er.identifier,
			er.community_id,
			coalesce(u.email, '') AS email,
			coalesce(u.phone_number, '') AS phone_number,
			coalesce(u.campus_code, '') AS campus_code,
			coalesce(u.department, '') AS department,
			coalesce(u.cool_code, 0) AS cool_code,
			coalesce(c.name, '') AS cool_name,
			er.event_code,
			er.instance_code,
			er.identifier_origin,
			er.community_id_origin,
			er.updated_by,
			er.status,
			er.reason,
			er.description,
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
			LEFT JOIN users u ON er.community_id = u.community_id
			LEFT JOIN cools c ON u.cool_code = c.code
        WHERE er.deleted_at IS NULL	
	`
)

func BuildCountGetRegisteredQuery(param models.GetAllRegisteredCursorParam) (string, []interface{}, error) {
	var queryBuilder strings.Builder
	var args []interface{}

	queryBuilder.WriteString(queryCountEventAllRegistered)
	queryBuilder.WriteString(" WHERE er.deleted_at IS NULL")

	// Add conditions dynamically
	if param.EventCode != "" {
		queryBuilder.WriteString(" AND er.event_code = ?")
		args = append(args, param.EventCode)
	}
	if param.InstanceCode != "" {
		queryBuilder.WriteString(" AND er.instance_code = ?")
		args = append(args, param.InstanceCode)
	}
	if param.NameSearch != "" {
		//queryBuilder.WriteString(" AND er.name ILIKE ?")
		//args = append(args, param.NameSearch)
		queryBuilder.WriteString(" AND (er.name ILIKE ? OR er.description ILIKE ?)")
		args = append(args, param.NameSearch, param.NameSearch)
	}
	if param.CampusCode != "" {
		queryBuilder.WriteString(" AND u.campus_code = ?")
		args = append(args, param.CampusCode)
	}
	if param.DepartmentCode != "" {
		queryBuilder.WriteString(" AND u.department = ?")
		args = append(args, param.DepartmentCode)
	}
	if param.CoolId != "" {
		queryBuilder.WriteString(" AND u.cool_code = ?")
		intCool, _ := strconv.Atoi(param.CoolId)
		args = append(args, intCool)
	}

	return queryBuilder.String(), args, nil
}

func BuildGetRegisteredQuery(param models.GetAllRegisteredCursorParam) (string, []interface{}, error) {
	var queryBuilder strings.Builder
	var args []interface{}

	queryBuilder.WriteString(baseQueryGetRegisteredRecordList)

	// Add conditions dynamically
	if param.EventCode != "" {
		queryBuilder.WriteString(" AND er.event_code = ?")
		args = append(args, param.EventCode)
	}
	if param.InstanceCode != "" {
		queryBuilder.WriteString(" AND er.instance_code = ?")
		args = append(args, param.InstanceCode)
	}
	if param.NameSearch != "" {
		//queryBuilder.WriteString(" AND er.name ILIKE ?")
		queryBuilder.WriteString(" AND (er.name ILIKE ? OR er.description ILIKE ?)")
		args = append(args, "%"+param.NameSearch+"%", "%"+param.NameSearch+"%")
	}
	if param.CampusCode != "" {
		queryBuilder.WriteString(" AND u.campus_code = ?")
		args = append(args, param.CampusCode)
	}
	if param.DepartmentCode != "" {
		queryBuilder.WriteString(" AND u.department = ?")
		args = append(args, param.DepartmentCode)
	}
	if param.CoolId != "" {
		queryBuilder.WriteString(" AND u.cool_code = ?")
		intCool, _ := strconv.Atoi(param.CoolId)
		args = append(args, intCool)
	}

	isForward := param.Direction != "prev"
	if param.Cursor != "" {
		createdCursor, err := cursor.DecryptCursorForGetRegisteredRecord(param.Cursor)
		if err != nil {
			return "", nil, err
		}

		operator := "<"
		if !isForward {
			operator = ">"
		}

		// Add cursor condition
		queryBuilder.WriteString(fmt.Sprintf(" AND (er.created_at, er.id) %s (?, ?)", operator))
		args = append(args, createdCursor.CreatedAt, createdCursor.ID)
	}

	// Add ordering - Note the direction changes based on pagination direction
	// Add ordering
	queryBuilder.WriteString(" ORDER BY er.created_at DESC, er.id DESC")
	//if isForward {
	//	queryBuilder.WriteString(" ORDER BY er.created_at DESC, er.id DESC")
	//} else {
	//	queryBuilder.WriteString(" ORDER BY er.created_at ASC, er.id ASC")
	//}

	// Add limit
	queryBuilder.WriteString(" LIMIT ?")
	args = append(args, param.Limit+1)

	return queryBuilder.String(), args, nil
}

// Helper function to reverse records slice
func reverseRecords(records []any) {
	for i, j := 0, len(records)-1; i < j; i, j = i+1, j-1 {
		records[i], records[j] = records[j], records[i]
	}
}

func BuildDownloadGetRegisteredQuery(param models.GetDownloadAllRegisteredParam) (string, []interface{}, error) {
	var queryBuilder strings.Builder
	var args []interface{}

	queryBuilder.WriteString(baseQueryGetDownloadRegisteredRecordList)

	// Add conditions dynamically
	if param.EventCode != "" {
		queryBuilder.WriteString(" AND er.event_code = ?")
		args = append(args, param.EventCode)
	}
	if param.InstanceCode != "" {
		queryBuilder.WriteString(" AND er.instance_code = ?")
		args = append(args, param.InstanceCode)
	}
	if param.NameSearch != "" {
		//queryBuilder.WriteString(" AND er.name ILIKE ?")
		//args = append(args, param.NameSearch)
		queryBuilder.WriteString(" AND (er.name ILIKE ? OR er.description ILIKE ?)")
		args = append(args, param.NameSearch, param.NameSearch)
	}
	if param.CampusCode != "" {
		queryBuilder.WriteString(" AND u.campus_code = ?")
		args = append(args, param.CampusCode)
	}
	if param.DepartmentCode != "" {
		queryBuilder.WriteString(" AND u.department = ?")
		args = append(args, param.DepartmentCode)
	}
	if param.CoolId != "" {
		queryBuilder.WriteString(" AND u.cool_code = ?")
		intCool, _ := strconv.Atoi(param.CoolId)
		args = append(args, intCool)
	}

	return queryBuilder.String(), args, nil
}
