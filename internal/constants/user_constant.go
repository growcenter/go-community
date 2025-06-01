package constants

// GUEST
var (
	ROLE_GUEST = []string{"guest-general-view"}
	TYPE_GUEST = []string{"guest"}

	USER_TYPE_COOL_FACILITATOR = "cool-facilitator"
	USER_TYPE_COOL_MEMBER      = "cool-member"
	USER_TYPE_COOL_CORE        = "cool-core"
	USER_TYPE_COOL_LEADER      = "cool-leader"

	CoolUserType = Dictionary{
		USER_TYPE_COOL_CORE:        {"core", "cool-core"},
		USER_TYPE_COOL_FACILITATOR: {"facilitator", "cool-facilitator"},
		USER_TYPE_COOL_LEADER:      {"leader", "cool-leader"},
		USER_TYPE_COOL_MEMBER:      {"member", "cool-member"},
	}
)
