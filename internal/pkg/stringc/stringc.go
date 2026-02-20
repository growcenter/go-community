// stringc (stringcustom) is a custom string helper package
package stringc

import (
	"slices"
	"strings"
	"unicode"
)

// ContainsAlphabet checks if a string contains any alphabet characters
// Example:
//
//	stringc.ContainsAlphabet("abc123") // returns true
//	stringc.ContainsAlphabet("123")    // returns false
func ContainsAlphabet(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

// SlicesToInterfaces converts a slice of strings to a slice of interfaces
// Example:
//
//	stringc.SlicesToInterfaces([]string{"abc", "def"}) // returns []interface{}{"abc", "def"}
func SlicesToInterfaces(args []string) []interface{} {
	result := make([]interface{}, len(args))
	for i, v := range args {
		result[i] = v
	}
	return result
}

// UpperAndTrimSpace trims whitespace and converts a string to uppercase
// Example:
//
//	stringc.UpperAndTrimSpace("  hello  ") // returns "HELLO"
func UpperAndTrimSpace(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}

// LowerAndTrimSpace trims whitespace and converts a string to lowercase
// Example:
//
//	stringc.LowerAndTrimSpace("  Hello  ") // returns "hello"
func LowerAndTrimSpace(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// Pointer returns a pointer to the given string
// Example:
//
//	stringc.Pointer("hello") // returns &"hello"
func Pointer(s string) *string {
	return &s
}

// ContainsAny checks if a string contains any of the given substrings
// Example:
//
//	stringc.ContainsAny("hello", []string{"world", "foo"}) // returns false
//	stringc.ContainsAny("hello", []string{"world", "hello"}) // returns true
func ContainsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// OneOf checks if a string is equal to any of the given strings
// How to read: One of values contains s
// Example:
//
//	stringc.OneOf([]string{"world", "foo"}, "hello") // returns false
//	stringc.OneOf([]string{"world", "hello"}, "hello") // returns true
func OneOf(values []string, s ...string) bool {
	for _, v := range s {
		if slices.Contains(values, v) {
			return true
		}
	}
	return false
}

// ContainsString checks whether a string slice contains the exact target value.
// For substring matching, use ContainsAny instead.
// Example:
//
//	stringc.ContainsString([]string{"parent", "child"}, "parent") // returns true
//	stringc.ContainsString([]string{"parent", "child"}, "admin")  // returns false
func ContainsString(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}

// SplitTrimmed splits a string by sep, trims whitespace from each part,
// and omits empty parts from the result.
// Example:
//
//	stringc.SplitTrimmed("a, b , , c", ",") // returns ["a", "b", "c"]
func SplitTrimmed(s, sep string) []string {
	parts := strings.Split(s, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
