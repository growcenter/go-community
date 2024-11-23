package pgsql

var (
	querySingleCheckRole   = "SELECT EXISTS (SELECT 1 FROM roles WHERE role = ?)"
	queryMultipleCheckRole = "SELECT COUNT(*) FROM roles WHERE role = ANY(?)"
)
