package pgsql

import (
	"fmt"
	"strings"
)

var (
	queryCheckUserByEmailPhoneNumber = `SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE 
				(email = ? AND ? != '') 
				OR 
				(phone_number = ? AND ? != '')
		);`

	queryCheckUserByCommunityId = `SELECT EXISTS (SELECT 1 FROM users WHERE community_id = ?)`

	queryGetOneUserByEmailPhoneNumber = `SELECT *
	FROM users
	WHERE
		(email = ? AND ? != '')
	   OR
		(phone_number = ? AND ? != '');`

	queryGetUserNameByIdentifier = `SELECT name, community_id
	FROM users WHERE email = ? OR phone_number = ? LIMIT 1`

	queryGetUserNameByCommunityId = `SELECT name, community_id
	FROM users WHERE community_id = ? LIMIT 1`

	queryGetOneUserByIdentifier = `SELECT * FROM users WHERE email = ? OR phone_number = ? LIMIT 1`

	queryMultipleCheckUser = "SELECT COUNT(*) FROM users WHERE community_id = ANY(?)"

	baseQueryGetAllUser = `
	SELECT
		u.id AS id,
		u.community_id AS community_id,
		u.name AS name,
		u.phone_number AS phone_number,
		u.email AS email,
		u.user_types AS user_types,
		u.roles AS roles,
		u.status AS status,
		u.gender AS gender,
		u.address AS address,
		u.campus_code AS campus_code,
		u.cool_id AS cool_id,
		c.name AS cool_name,
		u.department AS department,
		u.date_of_birth AS date_of_birth,
		u.place_of_birth AS place_of_birth,
		u.marital_status AS marital_status,
		u.kkj_number AS kkj_number,
		u.jemaat_id AS jemaat_id,
		u.is_baptized AS is_baptized,
		u.is_kom100 AS is_kom100,
		u.created_at AS created_at,
		u.updated_at AS updated_at,
		u.deleted_at AS deleted_at
	FROM
		users u
	LEFT JOIN
		cools c ON u.cool_id = c.id
	WHERE
		1=1`

	queryCountAllUser = `
		SELECT COUNT(*)
		FROM users u
		WHERE 1=1
	`
)

func ConditionExistOrNot(email string, phoneNumber string) (condition string, args []interface{}) {
	if email != "" {
		condition = "email = ?"
		args = append(args, email)
	}

	if phoneNumber != "" {
		if condition != "" {
			condition += " OR "
		}
		condition += " phone_number = ?"
		args = append(args, phoneNumber)
	}

	return condition, args
}

func BuildQueryGetAllUser(baseQuery string, searchBy string, search string, campusCode string, coolId int, departmentCode string, cursor string, direction string, limit int) (string, []interface{}, error) {
	var conditions []string
	var params []interface{}

	// Apply filters
	if campusCode != "" {
		conditions = append(conditions, "u.campus_code = ?")
		params = append(params, campusCode)
	}
	if departmentCode != "" {
		conditions = append(conditions, "u.department = ?")
		params = append(params, departmentCode)
	}
	if coolId != 0 {
		conditions = append(conditions, "u.cool_id = ?")
		params = append(params, coolId)
	}

	// Apply search
	if searchBy != "" && search != "" {
		switch searchBy {
		case "name":
			conditions = append(conditions, "u.name ILIKE ?")
			params = append(params, "%"+search+"%")
		case "phoneNumber":
			conditions = append(conditions, "u.phone_number ILIKE ?")
			params = append(params, "%"+search+"%")
		case "email":
			conditions = append(conditions, "u.email ILIKE ?")
			params = append(params, "%"+search+"%")
		case "communityId":
			conditions = append(conditions, "u.community_id = ?")
			params = append(params, search)
		default:
			return "", nil, fmt.Errorf("invalid searchBy: %s, must be 'communityId', 'email', 'phoneNumber', or 'name'", searchBy)
		}
	}

	// Apply cursor for pagination
	if cursor != "" {
		if direction == "next" {
			conditions = append(conditions, "u.community_id > ?")
		} else if direction == "prev" {
			conditions = append(conditions, "u.community_id < ?")
		} else {
			return "", nil, fmt.Errorf("invalid direction: %s, must be 'next' or 'prev'", direction)
		}
		params = append(params, cursor)
	}

	// Build WHERE clause
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Add ordering based on direction
	if direction == "prev" {
		baseQuery += " ORDER BY u.community_id DESC"
	} else {
		baseQuery += " ORDER BY u.community_id ASC"
	}

	// Add LIMIT clause
	if limit > 0 {
		baseQuery += " LIMIT ?"
		params = append(params, limit)
	}

	return baseQuery, params, nil
}
