package pgsql

var (
	queryCountEventInstanceByCode = `SELECT COUNT(*) FROM event_instances WHERE event_code = ?`
)
