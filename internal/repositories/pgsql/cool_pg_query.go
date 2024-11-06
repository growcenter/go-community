package pgsql

var (
	queryCheckCoolById = "SELECT EXISTS (SELECT 1 FROM cools WHERE id = ?)"
)
