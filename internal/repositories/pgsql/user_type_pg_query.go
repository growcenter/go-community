package pgsql

var (
	querySingleCheckUserType = "SELECT EXISTS (SELECT 1 FROM user_types WHERE type = ?)"
	queryGetUserTypesByArray = "SELECT user_types.type, user_types.category, user_types.roles FROM user_types WHERE type = ANY(?)"
)
