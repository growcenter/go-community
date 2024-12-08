package pgsql

var (
	queryCheckCoolById = "SELECT EXISTS (SELECT 1 FROM cools WHERE id = ?)"
	queryGetNameById   = "SELECT cools.id, cools.name FROM cools WHERE id = ?"
)
