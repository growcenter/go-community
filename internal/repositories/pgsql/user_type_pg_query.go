package pgsql

var (
	querySingleCheckUserType = "SELECT EXISTS (SELECT 1 FROM user_types WHERE type = ?)"
)
