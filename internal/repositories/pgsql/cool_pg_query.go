package pgsql

var (
	queryCheckCoolById   = "SELECT EXISTS (SELECT 1 FROM cools WHERE id = ?)"
	queryGetNameById     = "SELECT cools.id, cools.name FROM cools WHERE id = ?"
	queryGetCoolsOptions = `SELECT id, name, campus_code, leader_community_ids, status FROM cools WHERE deleted_at IS NULL AND status = 'active'`
)
