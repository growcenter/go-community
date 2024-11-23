package common

import (
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
)

func CapitalizeFirstWord(str string) string {
	// Create a title case function for the English language
	title := cases.Title(language.English)
	// Apply the title case transformation
	return title.String(str)
}

func StringTrimSpaceAndLower(str string) string {
	return strings.TrimSpace(strings.ToLower(str))
}

func GetValueFromMapString(header string, value string) (string, bool) {
	headerMap := viper.GetStringMapString(header)
	result, exist := headerMap[value]
	if !exist {
		return "", false
	}

	return result, true
}

//func CombineMapStrings(userTypeRoles, additionalRoles []string) []string {
//	uniqueRoles := make(map[string]bool)
//
//	// Add roles from userTypeRoles
//	for _, role := range userTypeRoles {
//		uniqueRoles[role] = true
//	}
//
//	// Add roles from additionalRoles
//	for _, role := range additionalRoles {
//		uniqueRoles[role] = true
//	}
//
//	// Convert map keys back to a slice
//	var allRoles []string
//	for role := range uniqueRoles {
//		allRoles = append(allRoles, role)
//	}
//
//	return allRoles
//}

func CombineMapStrings(mappingA, mappingB []string) []string {
	uniqueStrings := make(map[string]bool)

	// Add roles from userTypeRoles
	for _, mappedString := range mappingA {
		uniqueStrings[mappedString] = true
	}

	// Add roles from additionalRoles
	for _, mappedString := range mappingB {
		uniqueStrings[mappedString] = true
	}

	// Convert map keys back to a slice
	var allMappedStrings []string
	for mappedString := range uniqueStrings {
		allMappedStrings = append(allMappedStrings, mappedString)
	}

	return allMappedStrings
}
