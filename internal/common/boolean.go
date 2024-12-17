package common

import "regexp"

// Check in the array if one of them true == true, but if none is true == false
func CheckOneBoolIsTrue(values []bool) bool {
	for _, v := range values {
		if v {
			return true
		}
	}
	return false
}

func CountTrue(values []bool) int {
	count := 0
	for _, v := range values {
		if v {
			count++
		}
	}
	return count
}

func BoolToInt(value bool) int {
	if value {
		return 1
	} else {
		return 0
	}
}

func GetBooleanArrayFromStringArray(input []string) []bool {
	var boolValues []bool

	// Loop through each string in the input slice
	for _, element := range input {
		// Use regex to extract the last character 't' or 'f'
		re := regexp.MustCompile(`[a-z]$`)
		matches := re.FindStringSubmatch(element)

		// If there is a match, convert 't' to true and 'f' to false
		if len(matches) > 0 {
			if matches[0] == "t" {
				boolValues = append(boolValues, true)
			} else if matches[0] == "f" {
				boolValues = append(boolValues, false)
			}
		}
	}

	return boolValues
}
