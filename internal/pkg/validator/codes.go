package validator

import "strings"

// IdentifierCode checks if an event code follows the expected format
func IdentifierCode(code string) bool {
	parts := strings.Split(code, "-")
	if len(parts) != 3 {
		return false
	}

	// Check prefix (2-5 uppercase letters)
	prefix := parts[0]
	if len(prefix) < 2 || len(prefix) > 5 {
		return false
	}
	if prefix != strings.ToUpper(prefix) {
		return false
	}

	// Check date component (6 digits for YYYYMM or 4 digits for YYYY)
	dateComponent := parts[1]
	if len(dateComponent) != 6 && len(dateComponent) != 4 {
		return false
	}

	// Check random/sequence component (3-5 characters)
	randomComponent := parts[2]
	if len(randomComponent) < 3 || len(randomComponent) > 5 {
		return false
	}

	return true
}
