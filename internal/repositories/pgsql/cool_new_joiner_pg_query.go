package pgsql

import (
	"fmt"
	"go-community/internal/models"
	"go-community/internal/pkg/cursor"
	"strings"
)

var (
	baseQueryGetAllCoolNewJoiner = `
	SELECT
		cnj.id AS id,
		cnj.name AS name,
		cnj.marital_status AS marital_status,
		cnj.gender AS gender,
		cnj.year_of_birth AS year_of_birth,
		cnj.phone_number AS phone_number,
		cnj.address AS address,
		cnj.community_of_interest AS community_of_interest,
		cnj.campus_code AS campus_code,
		cnj.location AS location,
		cnj.updated_by AS updated_by,
		cnj.status AS status,
		cnj.created_at AS created_at,
		cnj.updated_at AS updated_at,
		cnj.deleted_at AS deleted_at
	FROM
		cool_new_joiners cnj
	WHERE
		1=1`

	queryCountAllCoolNewJoiner = `
		SELECT COUNT(*)
		FROM cool_new_joiners cnj
	`
)

func BuildCountGetAllCoolNewJoiner(param models.GetAllCoolNewJoinerCursorParam) (string, []interface{}, error) {
	var queryBuilder strings.Builder
	var args []interface{}

	queryBuilder.WriteString(queryCountAllCoolNewJoiner)
	queryBuilder.WriteString(" WHERE cnj.deleted_at IS NULL")

	// Add conditions dynamically
	if param.Name != "" {
		queryBuilder.WriteString(" AND cnj.name ILIKE ?")
		args = append(args, "%"+param.Name+"%")
	}
	if param.PhoneNumber != "" {
		queryBuilder.WriteString(" AND cnj.phone_number ILIKE ?")
		args = append(args, "%"+param.PhoneNumber+"%")
	}
	if param.CampusCode != "" {
		queryBuilder.WriteString(" AND cnj.campus_code = ?")
		args = append(args, param.CampusCode)
	}
	if param.MaritalStatus != "" {
		queryBuilder.WriteString(" AND cnj.marital_status = ?")
		args = append(args, param.MaritalStatus)
	}
	if param.CommunityOfInterest != "" {
		queryBuilder.WriteString(" AND cnj.community_of_interest = ?")
		args = append(args, param.CommunityOfInterest)
	}
	if param.Status != "" {
		queryBuilder.WriteString(" AND cnj.status = ?")
		args = append(args, param.Status)
	}
	if param.Gender != "" {
		queryBuilder.WriteString(" AND cnj.gender = ?")
		args = append(args, param.Gender)
	}
	if param.Location != "" {
		queryBuilder.WriteString(" AND cnj.location = ?")
		args = append(args, param.Location)
	}

	return queryBuilder.String(), args, nil
}

func BuildQueryGetAllCoolNewJoiner(param models.GetAllCoolNewJoinerCursorParam) (string, []interface{}, error) {
	var queryBuilder strings.Builder
	var args []interface{}

	queryBuilder.WriteString(baseQueryGetAllCoolNewJoiner)

	// Add conditions dynamically
	if param.Name != "" {
		queryBuilder.WriteString(" AND cnj.name ILIKE ?")
		args = append(args, "%"+param.Name+"%")
	}
	if param.PhoneNumber != "" {
		queryBuilder.WriteString(" AND cnj.phone_number ILIKE ?")
		args = append(args, "%"+param.PhoneNumber+"%")
	}
	if param.CampusCode != "" {
		queryBuilder.WriteString(" AND cnj.campus_code = ?")
		args = append(args, param.CampusCode)
	}
	if param.MaritalStatus != "" {
		queryBuilder.WriteString(" AND cnj.marital_status = ?")
		args = append(args, param.MaritalStatus)
	}
	if param.CommunityOfInterest != "" {
		queryBuilder.WriteString(" AND cnj.community_of_interest = ?")
		args = append(args, param.CommunityOfInterest)
	}
	if param.Status != "" {
		queryBuilder.WriteString(" AND cnj.status = ?")
		args = append(args, param.Status)
	}
	if param.Gender != "" {
		queryBuilder.WriteString(" AND cnj.gender = ?")
		args = append(args, param.Gender)
	}
	if param.Location != "" {
		queryBuilder.WriteString(" AND cnj.location = ?")
		args = append(args, param.Location)
	}

	isForward := param.Direction != "prev"
	if param.Cursor != "" {
		createdCursor, err := cursor.DecryptCursorForGetAllUser(param.Cursor)
		if err != nil {
			return "", nil, err
		}

		operator := "<"
		if !isForward {
			operator = ">"
		}

		// Add cursor condition
		queryBuilder.WriteString(fmt.Sprintf(" AND (cnj.created_at, cnj.id) %s (?, ?)", operator))
		args = append(args, createdCursor.CreatedAt, createdCursor.ID)
	}

	// Add ordering - Note the direction changes based on pagination direction
	queryBuilder.WriteString(" ORDER BY cnj.created_at DESC, cnj.id DESC")

	// Add limit
	queryBuilder.WriteString(" LIMIT ?")
	args = append(args, param.Limit+1)

	return queryBuilder.String(), args, nil
}
