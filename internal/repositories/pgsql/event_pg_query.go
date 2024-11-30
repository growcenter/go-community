package pgsql

var (
	queryCheckEventByCode = "SELECT EXISTS (SELECT 1 FROM events WHERE code = ?)"
)
