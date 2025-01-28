package pgsql

import (
	"fmt"
	"go-community/internal/models"
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

	queryGetProfileByCommunityId = `
			select u.community_id as community_id,
				   u.name as name,
				   u.phone_number as phone_number,
				   u.email as email,
				   u.roles as roles,
					u.status as status,
					u.gender as gender,
					coalesce(u.address, '') as address,
					u.campus_code as campus_code,
					coalesce(u.cool_id, 0) as cool_id,
					coalesce(c.name, '') as cool_name,
					u.department as department,
					u.date_of_birth as date_of_birth,
					coalesce(u.place_of_birth, '') as place_of_birth,
					u.marital_status as marital_status,
					u.date_of_marriage as date_of_marriage,
					coalesce(u.employment_status, '') as employment_status,
					coalesce(u.education_level, '') as education_level,
					coalesce(u.kkj_number, '') as kkj_number,
					coalesce(u.jemaat_id, '') as jemaat_id,
					u.is_baptized as is_baptized,
					u.is_kom100 as is_kom100,
					u.created_at as created_at,
					u.updated_at as updated_at,
					u.user_types as user_types,
				   ru.community_id as related_community_id,
				   ru.name as related_name,
				   ur.relationship_type as relationship_type
			from users u
				left join user_relations ur on ur.community_id = u.community_id
				left join users ru on ru.community_id = ur.related_community_id
				left join cools c on c.id = u.cool_id
			WHERE ur.community_id = ?
			group by u.community_id, u.name, u.phone_number, u.email, u.roles, u.status, u.gender, coalesce(u.address, ''), u.campus_code, u.cool_id, c.name, u.department, u.date_of_birth, coalesce(u.place_of_birth, ''), u.marital_status, u.date_of_marriage, coalesce(u.employment_status, ''), coalesce(u.education_level, ''), coalesce(u.kkj_number, ''), coalesce(u.jemaat_id, ''), u.is_baptized, u.is_kom100, u.created_at, u.updated_at, u.user_types, ru.community_id, ru.name, ur.relationship_type
	`

	queryGetCommunityIdByName = `SELECT name, community_id
	FROM users WHERE name ILIKE ? LIMIT 1`

	queryGetCommunityIdsByParams = `SELECT name, community_id, email, phone_number
	FROM users`

	queryCountUserByUserTypeCategory = `SELECT COUNT(*)
	FROM users u
			 INNER JOIN user_types ut ON ut.type = ANY(u.user_types)
	WHERE ut.category = ANY(?)
	  AND u.status = 'active'
	  AND u.deleted_at IS NULL
	  AND ut.deleted_at IS NULL;`
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

func BuildQueryGetCommunityIdByParams(parameter models.GetCommunityIdsByParameter) (string, []interface{}, error) {
	var conditions []string
	var params []interface{}
	base := queryGetCommunityIdsByParams

	switch {
	case parameter.Name != "":
		conditions = append(conditions, "name ILIKE ?")
		params = append(params, "%"+parameter.Name+"%")
	case parameter.Email != "":
		conditions = append(conditions, "email ILIKE ?")
		params = append(params, "%"+parameter.Email+"%")
	case parameter.PhoneNumber != "":
		conditions = append(conditions, "phone_number ILIKE ?")
		params = append(params, "%"+parameter.PhoneNumber+"%")
	default:
		return "", nil, fmt.Errorf("invalid parameter: must be 'email', 'phoneNumber', or 'name'")
	}

	// Build WHERE clause
	if len(conditions) > 0 {
		base += " WHERE " + strings.Join(conditions, " AND ")
	}

	return base, params, nil
}
