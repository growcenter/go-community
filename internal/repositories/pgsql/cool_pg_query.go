package pgsql

var (
	queryCheckCoolByCode     = "SELECT EXISTS (SELECT 1 FROM cools WHERE code = ?)"
	queryGetNameByCode       = "SELECT cools.code, cools.name FROM cools WHERE code = ?"
	queryGetCoolsOptions     = `SELECT code, name, campus_code, leader_community_ids, status FROM cools WHERE deleted_at IS NULL AND status = 'active'`
	queryGetCoolMemberByCode = `
	SELECT
		u.community_id AS community_id,
		u.name AS name,
		u.cool_code AS cool_code,
		(
			SELECT json_agg(json_build_object('type', ut.type, 'name', ut.name))
			FROM user_types ut
			WHERE ut.type = ANY(u.user_types)
		) AS user_types
	FROM
		users u
	WHERE
		u.cool_code = ?
	;`

	queryGetCoolFacilitatorByCode = `
	SELECT
		u.name AS name,
		u.community_id AS community_id,
		c.code AS cool_code,
		(
			SELECT json_agg(json_build_object('type', ut.type, 'name', ut.name))
				FROM user_types ut
				WHERE ut.type = ANY(u.user_types)
			) AS user_types
	FROM
		users u
	JOIN
		cools c ON u.community_id = ANY(c.facilitator_community_ids)
	WHERE
		u.status = 'active' AND c.code = ?
	ORDER BY
		u.name;`

	queryGetAllMembersByCoolCode = `
	WITH members AS (
			(SELECT 
				u.community_id,
				u.name,
				u.cool_code,
				(
                SELECT json_agg(json_build_object('type', 'userType', 'userType', ut.type, 'name', ut.name))
                    FROM user_types ut
                    WHERE ut.type = ANY(u.user_types)
                ) AS user_types,
				true as is_facilitator
			FROM
                users u
            JOIN
                cools c ON u.community_id = ANY(c.facilitator_community_ids)
			WHERE c.code = ? )
			UNION ALL
			(SELECT 
				u.community_id,
				u.name,
				u.cool_code,
				(
                SELECT json_agg(json_build_object('type', 'userType', 'userType', ut.type, 'name', ut.name))
                    FROM user_types ut
                    WHERE ut.type = ANY(u.user_types)
                ) AS user_types,
				false as is_facilitator
			FROM users u
			WHERE u.cool_code = ?)
		)
		SELECT * FROM members;
	`

	queryGetCoolByCommunityId = `
	SELECT 
		c.id,
		c.code,
		c.name,
		c.description,
		c.campus_code,
		c.facilitator_community_ids,
		c.leader_community_ids,
		c.core_community_ids,
		c.category,
		c.gender,
		c.recurrence,
		c.location_type,
		c.location_name,
		c.status
	FROM 
		cools c
	JOIN 
		users u ON u.cool_code = c.code
	WHERE 
		u.community_id = ?
		AND c.deleted_at IS NULL
	LIMIT 1
	;`
)
