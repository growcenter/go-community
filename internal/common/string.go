package common

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

// AreStringFieldsUnique checks if string values are unique across specified fields in a slice of structs.
// It returns false if a duplicate is found or if the input is not a slice of structs.
func AreStringFieldsUnique(slice interface{}, fieldNames ...string) bool {
	val := reflect.ValueOf(slice)
	if val.Kind() != reflect.Slice {
		// Invalid input, so not unique in a sense.
		return false
	}

	// Create a map of sets to track seen values for each field.
	seenValues := make(map[string]map[string]struct{})
	for _, fieldName := range fieldNames {
		seenValues[fieldName] = make(map[string]struct{})
	}

	for i := 0; i < val.Len(); i++ {
		item := val.Index(i)
		if item.Kind() != reflect.Struct {
			return false // Slice items are not structs.
		}

		for _, fieldName := range fieldNames {
			field := item.FieldByName(fieldName)
			if !field.IsValid() || field.Kind() != reflect.String {
				continue // Skip if field doesn't exist or is not a string.
			}

			value := field.String()
			if value != "" {
				if _, exists := seenValues[fieldName][value]; exists {
					return false // Duplicate found.
				}
				seenValues[fieldName][value] = struct{}{}
			}
		}
	}
	return true // All values are unique.
}

// CheckUniqueStringFields checks for duplicate non-empty string values across multiple fields in a slice of structs.
// It uses reflection to be generic and flexible.
//
// Parameters:
//   - slice: The slice of structs to check (passed as an interface{}).
//   - fieldNames: A variadic list of string field names to check for uniqueness.
//
// Returns:
//   - An error if a duplicate value is found in any of the specified fields, or if reflection fails.
func CheckUniqueStringFields(slice interface{}, fieldNames ...string) error {
	val := reflect.ValueOf(slice)
	if val.Kind() != reflect.Slice {
		return fmt.Errorf("expected a slice, but got %s", val.Kind())
	}

	// Create a map of sets to track seen values for each field.
	seenValues := make(map[string]map[string]struct{})
	for _, fieldName := range fieldNames {
		seenValues[fieldName] = make(map[string]struct{})
	}

	for i := 0; i < val.Len(); i++ {
		item := val.Index(i)
		if item.Kind() != reflect.Struct {
			return fmt.Errorf("expected slice items to be structs, but got %s", item.Kind())
		}

		for _, fieldName := range fieldNames {
			field := item.FieldByName(fieldName)
			if !field.IsValid() || field.Kind() != reflect.String {
				continue // Skip if field doesn't exist or is not a string.
			}

			value := field.String()
			if value != "" {
				if _, exists := seenValues[fieldName][value]; exists {
					return fmt.Errorf("duplicate value found for field '%s': %s", fieldName, value)
				}
				seenValues[fieldName][value] = struct{}{}
			}
		}
	}
	return nil
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

// CheckPresenceOfValue checks for the presence of specific strings in a slice
// and returns a map indicating if each string was found.
func CheckPresenceOfValue(slice []string, values ...string) map[string]bool {
	presenceMap := make(map[string]bool)
	for _, v := range values {
		presenceMap[v] = false
	}

	for _, item := range slice {
		if _, ok := presenceMap[item]; ok {
			presenceMap[item] = true
		}
	}

	return presenceMap
}

// Convert []string to []interface{}
func SlicesToInterfaces(args []string) []interface{} {
	result := make([]interface{}, len(args))
	for i, v := range args {
		result[i] = v
	}
	return result
}

func CombineMapUUID(mappingA, mappingB []uuid.UUID) []uuid.UUID {
	uniqueUUIDs := make(map[uuid.UUID]bool)

	// Add roles from userTypeRoles
	for _, mappedUUID := range mappingA {
		uniqueUUIDs[mappedUUID] = true
	}

	// Add roles from additionalRoles
	for _, mappedUUID := range mappingB {
		uniqueUUIDs[mappedUUID] = true
	}

	// Convert map keys back to a slice
	var allMappedUUIDs []uuid.UUID
	for mappedUUID := range uniqueUUIDs {
		allMappedUUIDs = append(allMappedUUIDs, mappedUUID)
	}

	return allMappedUUIDs
}

// Utility function to get unique values from a slice of structs
func GetUniqueFieldValuesFromModelUUID(data interface{}, fieldName string) ([]uuid.UUID, error) {
	// Ensure that the input is a slice
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Slice {
		return nil, fmt.Errorf("expected a slice, got %s", val.Kind())
	}

	// Create a map to store unique values
	uniqueValues := make(map[uuid.UUID]struct{})

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
			role := fieldVal.Index(j).Interface().(uuid.UUID)
			uniqueValues[role] = struct{}{}
		}
	}

	// Collect the unique values into a slice
	var result []uuid.UUID
	for value := range uniqueValues {
		result = append(result, value)
	}

	return result, nil
}

func UUIDsToStrings(uuids []uuid.UUID) []string {
	strings := make([]string, len(uuids))
	for i, u := range uuids {
		strings[i] = u.String()
	}
	return strings
}
