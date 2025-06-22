package common

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"reflect"
	"strings"
	"unicode"
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

func StringTrimSpaceAndUpper(str string) string {
	return strings.TrimSpace(strings.ToUpper(str))
}

func GetValueFromMapString(header string, value string) (string, bool) {
	headerMap := viper.GetStringMapString(header)
	result, exist := headerMap[value]
	if !exist {
		return "", false
	}

	return result, true
}

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

// all data inputted should be matched
func CheckAllDataMapStructure(mapstructure map[string]string, input []string) bool {
	for _, val := range input {
		if _, exists := mapstructure[val]; !exists {
			// If any value doesn't exist, return false immediately
			return false
		}
	}
	// Return true if all values exist
	return true
}

// only need one data to get true
func CheckOneDataInList(list []string, input []string) bool {
	set := make(map[string]struct{})
	for _, item := range list {
		set[item] = struct{}{}
	}

	for _, val := range input {
		if _, exists := set[val]; exists {
			// Return true immediately if any value is found in the list
			return true
		}
	}
	// Return false if none of the values exist in the list
	return false
}

func ContainsValueInModel[T any](items []T, predicate func(T) bool) bool {
	for _, item := range items {
		if predicate(item) {
			return true
		}
	}
	return false
}

// Utility function to get unique values from a slice of structs
func GetUniqueFieldValuesFromModel(data interface{}, fieldName string) ([]string, error) {
	// Ensure that the input is a slice
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Slice {
		return nil, fmt.Errorf("expected a slice, got %s", val.Kind())
	}

	// Create a map to store unique values
	uniqueValues := make(map[string]struct{})

	// Iterate through each element in the slice
	for i := 0; i < val.Len(); i++ {
		item := val.Index(i)

		// Ensure the item is a struct
		if item.Kind() != reflect.Struct {
			return nil, fmt.Errorf("expected a struct, got %s", item.Kind())
		}

		// Get the field by name
		fieldVal := item.FieldByName(fieldName)
		if !fieldVal.IsValid() {
			return nil, fmt.Errorf("field %s not found in struct", fieldName)
		}

		// Ensure the field is of the expected type (a slice of strings)
		if fieldVal.Kind() != reflect.Slice {
			return nil, fmt.Errorf("expected field %s to be a slice, got %s", fieldName, fieldVal.Kind())
		}

		// Iterate through the slice and add unique values
		for j := 0; j < fieldVal.Len(); j++ {
			role := fieldVal.Index(j).String()
			uniqueValues[role] = struct{}{}
		}
	}

	// Collect the unique values into a slice
	var result []string
	for value := range uniqueValues {
		result = append(result, value)
	}

	return result, nil
}

func IsValidUUID(input string) bool {
	_, err := uuid.Parse(input)
	return err == nil
}

func ContainsAlphabet(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func RemoveSliceIfExact(input []string, toRemove []string) []string {
	// Build a lookup map for faster checking
	removalMap := make(map[string]struct{})
	for _, val := range toRemove {
		removalMap[val] = struct{}{}
	}

	var result []string
	for _, item := range input {
		if _, found := removalMap[item]; !found {
			result = append(result, item)
		}
	}

	return result
}

// removeIfContains removes strings that contain any of the blocked substrings
func RemoveSliceIfContains(input []string, toRemove []string) []string {
	var result []string

	for _, item := range input {
		shouldRemove := false
		for _, rem := range toRemove {
			if strings.Contains(item, rem) {
				shouldRemove = true
				break
			}
		}
		if !shouldRemove {
			result = append(result, item)
		}
	}

	return result
}
