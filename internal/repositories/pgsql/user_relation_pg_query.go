package pgsql

var (
	queryGetSpouseByCommunityId = `SELECT u.community_id, u.name, ur.relationship_type
	FROM
		users u
	JOIN user_relations ur ON u.community_id = ur.related_community_id
	WHERE
		ur.community_id = ?
		AND ur.relationship_type = 'spouse'
	LIMIT 1;`

	queryGetParentsByCommunityId = `SELECT u.community_id, u.name, ur.relationship_type
	FROM
		users u
			JOIN user_relations ur ON u.community_id = ur.related_community_id
	WHERE
		ur.community_id = ?
	  AND ur.relationship_type = 'parent';`

	queryGetChildByCommunityId = `SELECT u.community_id, u.name, ur.relationship_type
	FROM
		users u
			JOIN user_relations ur ON u.community_id = ur.related_community_id
	WHERE
		ur.community_id = ?
	  AND ur.relationship_type = 'child';`

	queryCountUserRelation = `
		SELECT COUNT(*)
		FROM user_relations
		WHERE (community_id = ? AND related_community_id = ?) OR (community_id = ? AND related_community_id = ?)
	`

	queryCountUserRelationSpecific = `
		SELECT COUNT(*)
		FROM user_relations
		WHERE (community_id = ? AND related_community_id = ?) AND relationship_type = ?
	`

	queryCountUserRelationMany = `
		SELECT COUNT(*)
		FROM user_relations
		WHERE community_id = ? AND relationship_type = ?
	`
)
