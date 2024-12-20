package pgsql

var (
	queryCountRecordByIdentifierOriginAndStatus = `SELECT COUNT(*) FROM event_registration_records WHERE identifier_origin = ? AND status = ?`
	queryCountRecordByCommunityIdOrigin         = `SELECT COUNT(*) FROM event_registration_records WHERE community_id_origin = ?`
	queryCheckRecordByIdentifier                = `SELECT EXISTS (SELECT 1 FROM event_registration_records WHERE identifier = ?)`
	queryCheckRecordByName                      = `SELECT EXISTS (SELECT 1 FROM event_registration_records WHERE name = ?)`
	queryCheckRecordByCommunityId               = `SELECT EXISTS (SELECT 1 FROM event_registration_records WHERE community_id = ?)`
)
