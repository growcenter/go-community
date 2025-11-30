package pgsql

var (
	queryCheckByCommunityIdAndInstanceCode = `SELECT EXISTS (SELECT 1 FROM event_registrations WHERE community_id = ? AND instance_code = ?)`
)
