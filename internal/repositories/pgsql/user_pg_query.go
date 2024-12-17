package pgsql

var (
	queryCheckUserByEmailPhoneNumber = `SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE 
				(email = ? AND ? != '') 
				OR 
				(phone_number = ? AND ? != '')
		);`

	queryGetOneUserByEmailPhoneNumber = `SELECT *
	FROM users
	WHERE
		(email = ? AND ? != '')
	   OR
		(phone_number = ? AND ? != '');`

	queryGetUserNameByIdentifier = `SELECT name, community_id
	FROM users WHERE email = ? OR phone_number = ? LIMIT 1`

	queryGetUserNameByCommunityId = `SELECT name, community_id
	FROM users WHERE community_id = ? LIMIT 1`

	queryGetOneUserByIdentifier = `SELECT * FROM users WHERE email = ? OR phone_number = ? LIMIT 1`
)

func ConditionExistOrNot(email string, phoneNumber string) (condition string, args []interface{}) {
	if email != "" {
		condition = "email = ?"
		args = append(args, email)
	}

	if phoneNumber != "" {
		if condition != "" {
			condition += " OR "
		}
		condition += " phone_number = ?"
		args = append(args, phoneNumber)
	}

	return condition, args
}
