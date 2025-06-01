package constants

import (
	"go-community/internal/common"
	"strings"
)

// Define Dictionary as a map of string keys to slices of strings
type Dictionary map[string][]string

// LookupValue finds the dictionary key based on the value (case-insensitive match)
func (d Dictionary) LookupValue(value string) (*string, bool) {
	for key, values := range d {
		for _, v := range values {
			if strings.EqualFold(v, common.StringTrimSpaceAndLower(value)) {
				return &key, true
			}
		}
	}
	return nil, false
}

func (d Dictionary) LookupValuesArray(values []string) ([]string, bool) {
	var keys []string
	for _, value := range values {
		for key, dictValues := range d {
			for _, v := range dictValues {
				if strings.EqualFold(v, common.StringTrimSpaceAndLower(value)) {
					keys = append(keys, key)
					break
				}
			}
		}
	}
	return keys, len(keys) > 0
}

func (d Dictionary) GetAllKeys() []string {
	// Extract keys
	var keys []string
	for key := range CoolUserType {
		keys = append(keys, key)
	}

	return keys
}

// Define maritalStatus using the Dictionary type
var MaritalStatus = Dictionary{
	"DIVORCED": {"Divorced"},
	"MARRIED":  {"Married", "Ya", "Menikah"},
	"OTHER":    {"Other", "Lainnya"},
	"SINGLE":   {"Single", "Tidak", "Belum Menikah"},
	"WIDOWED":  {"Widowed", "Janda", "Duda"},
}
