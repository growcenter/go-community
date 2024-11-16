package common

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func CapitalizeFirstWord(str string) string {
	// Create a title case function for the English language
	title := cases.Title(language.English)
	// Apply the title case transformation
	return title.String(str)
}
