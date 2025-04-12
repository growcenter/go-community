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

// Define maritalStatus using the Dictionary type
var MaritalStatus = Dictionary{
	"DIVORCED": {"Divorced"},
	"MARRIED":  {"Married", "Ya", "Menikah"},
	"OTHER":    {"Other", "Lainnya"},
	"SINGLE":   {"Single", "Tidak", "Belum Menikah"},
	"WIDOWED":  {"Widowed", "Janda", "Duda"},
}
