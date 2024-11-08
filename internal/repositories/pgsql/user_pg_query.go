package pgsql

var (
	queryCheckUserByEmailPhoneNumber = "SELECT EXISTS (SELECT 1 FROM users WHERE email = ? OR phone_number = ?)"
)
