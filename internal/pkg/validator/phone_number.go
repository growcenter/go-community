package validator

import (
	"errors"
	"go-community/internal/common"
	"regexp"
	"strings"

	"github.com/nyaruka/phonenumbers"
)

const (
	DefaultCountryCode        = "ID"
	IDPhonePrefix             = "08"
	IDE164PhonePrefix         = "62"
	IDE164PhonePrefixWithPlus = "+62"
)

var (
	ErrInvalidPhoneNumber = errors.New("invalid phone number")
)

func PhoneNumber(countryCode, phoneNumber string) (standardize *string, err error) {
	// if alphabet in number, then return invalid immediately
	// special handling because library ignores alphabets and still formats
	if common.ContainsAlphabet(phoneNumber) {
		return nil, ErrInvalidPhoneNumber
	}

	// if countryCode is empty, set to default "ID"
	if countryCode == "" {
		countryCode = DefaultCountryCode
	}

	// parse to phone struct format
	phone, err := phonenumbers.Parse(phoneNumber, countryCode)
	if err != nil {
		return nil, err
	}

	if phonenumbers.IsValidNumber(phone) {
		national := phonenumbers.Format(phone, phonenumbers.NATIONAL) // => "0858xxxx"

		// Remove all non-digit characters
		re := regexp.MustCompile(`\D`)
		clean := re.ReplaceAllString(national, "")
		return &clean, nil
	} else {
		return nil, ErrInvalidPhoneNumber
	}
}

func ValidateAndStandardize(countryCode, phoneNumber string) (standardize string, err error) {
	// if alphabet in number, then return invalid immediately
	// special handling because library ignores alphabets and still formats
	if common.ContainsAlphabet(phoneNumber) {
		return "", ErrInvalidPhoneNumber
	}

	// if countryCode is empty, set to default "ID"
	if countryCode == "" {
		countryCode = DefaultCountryCode
	}

	// parse to phone struct format
	phone, err := phonenumbers.Parse(phoneNumber, countryCode)
	if err != nil {
		return "", err
	}

	// check if phone number valid
	if !phonenumbers.IsValidNumberForRegion(phone, countryCode) {
		return "", ErrInvalidPhoneNumber
	}

	return phonenumbers.Format(phone, phonenumbers.E164), nil
}

func GetRegionCode(phoneNumber string) (string, error) {
	parsed, err := phonenumbers.Parse(phoneNumber, "")
	if err != nil {
		return "", err
	}

	regionCode := phonenumbers.GetRegionCodeForNumber(parsed)
	return regionCode, nil
}

func IsValidE164PhoneNumber(phone string) bool {
	formatted, err := phonenumbers.Parse(phone, "")
	if err != nil {
		return false
	}

	return phonenumbers.IsValidNumber(formatted)
}

// IsConsideredAsPhoneNumber value with prefix 08, 62 or +62
// but somehow failed to be parsed as phone number
func IsConsideredAsPhoneNumber(s string) bool {
	return strings.HasPrefix(s, IDPhonePrefix) ||
		strings.HasPrefix(s, IDE164PhonePrefix) ||
		strings.HasPrefix(s, IDE164PhonePrefixWithPlus)
}

// ForceConvertToE164PhoneNumber for handling looks like valid but actually not a valid
// phone number, this was loose phone formatting by changing the prefix to follow
// E164 format.
func ForceConvertToE164PhoneNumber(s string) (phoneNumber string, isConverted bool) {
	// from : 089900000000
	// to   : +6289900000000
	if strings.HasPrefix(s, IDPhonePrefix) {
		return IDE164PhonePrefixWithPlus + strings.TrimPrefix(s, "0"), true
	}

	// from : 6289900000000
	// to   : +6289900000000
	if strings.HasPrefix(s, IDE164PhonePrefix) {
		return "+" + s, true
	}

	// immediately return +6289900000000
	if strings.HasPrefix(s, IDE164PhonePrefixWithPlus) {
		return s, true
	}

	// not match any case, return as is
	return s, false
}
