package validator

import (
	"errors"
	"go-community/internal/common"
	"go-community/internal/models"
	"reflect"
	"regexp"
	"strings"
	"time"

	v10 "github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-multierror"
)

var valid = v10.New()

func init() {
	registerNoSpecialCharacters()
	registerNoSpacesAtStartOrEnd()
	registerDate()
	registerDatetime()
	registerEmailFormat()
	registerPhoneFormat()
	registerEmailPhoneFormat()
	registeryyyymmddFormat()
	registerCommunityId()
	registerEmailOrPhoneField()
	registerNameIdentifierCommunityIdFields()
	registerPhoneStandardize()
	registerhhmmField()
	registeryyyymmddNoExceedTodayFormat()
}

func Validate(request interface{}) error {
	valid.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	var errs *multierror.Error
	if err := valid.Struct(request); err != nil {
		// This check is only needed when your code could produce
		// an invalid value for validation such as an interface with a nil
		// value. Most including myself do not usually have code like this.
		if _, ok := err.(*v10.InvalidValidationError); ok {
			errs = multierror.Append(errs, models.ErrorValidateResponse{
				Message: err.Error(),
			})
			return errs.ErrorOrNil()
		}

		var validatorErrs v10.ValidationErrors
		if errors.As(err, &validatorErrs) {
			for _, validatorErr := range validatorErrs {
				validateResponse := models.ErrorValidateResponseMapping(validatorErr)
				errs = multierror.Append(errs, validateResponse)
			}
		}
	}

	return errs.ErrorOrNil()
}

func registerNoSpecialCharacters() {
	valid.RegisterValidation("nospecial", func(fl v10.FieldLevel) bool {
		input := fl.Field().String()
		// Define a regular expression pattern that allows only letters and digits.
		// Allow space
		pattern := "^[a-zA-Z0-9 ]*$"
		return regexp.MustCompile(pattern).MatchString(input)
	})
}

func registerNoSpacesAtStartOrEnd() {
	valid.RegisterValidation("noStartEndSpaces", func(fl v10.FieldLevel) bool {
		str := fl.Field().String()
		return str == "" || (str[0] != ' ' && str[len(str)-1] != ' ')
	})
}

func registerDate() {
	valid.RegisterValidation("date", func(fl v10.FieldLevel) bool {
		input := fl.Field().String()
		pattern := `\d{4}-\d{2}-\d{2}`
		return regexp.MustCompile(pattern).MatchString(input)
	})
}

func registerDatetime() {
	valid.RegisterValidation("datetime", func(fl v10.FieldLevel) bool {
		input := fl.Field().String()
		pattern := `\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`
		return regexp.MustCompile(pattern).MatchString(input)
	})
}

func registerEmailFormat() {
	valid.RegisterValidation("emailFormat", func(fl v10.FieldLevel) bool {
		//input := fl.Field().String()
		//pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		//return regexp.MustCompile(pattern).MatchString(input)

		input := fl.Field().Interface()

		inputString := input.(string)
		if inputString == "" {
			return true
		}

		pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		return regexp.MustCompile(pattern).MatchString(inputString)
	})
}

func registerPhoneFormat() {
	valid.RegisterValidation("phoneFormat", func(fl v10.FieldLevel) bool {
		//input := fl.Field().String()
		// Minimum of 10 digits
		// pattern := `^\+?(\d{1,3})?[-.\s]?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}$`
		// Minimum of 8 digits
		//pattern := `^\+?(\d{1,3})?[-.\s]?\(?\d{1,3}\)?[-.\s]?\d{1,4}[-.\s]?\d{1,4}$`
		// Minimum of 7 digits, Maximum 14 digits, with pattern of 081,082,083,085,087,088,089
		// ini yang bener
		//pattern := `^(081|082|083|085|087|088|089)\d+$`
		//return regexp.MustCompile(pattern).MatchString(input)

		input := fl.Field().Interface()

		inputString := input.(string)
		if inputString == "" {
			return true
		}

		pattern := `^(081|082|083|085|087|088|089)\d+$`
		return regexp.MustCompile(pattern).MatchString(inputString)
	})
}

func registerEmailPhoneFormat() {
	valid.RegisterValidation("emailPhoneFormat", func(fl v10.FieldLevel) bool {
		input := fl.Field().String()
		// Minimum of 10 digits
		// pattern := `^\+62\d{10,}$|^0\d{10,}$|^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		// Minimum of 8 digits
		//pattern := `^\+62\d{8,}$|^0\d{8,}$|^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		// New Way
		pattern := `^(081|082|083|085|087|088|089)\d{4,11}$`
		isEmail := v10.New().Var(input, "email") == nil

		return regexp.MustCompile(pattern).MatchString(input) || isEmail
	})
}

func registeryyyymmddFormat() {
	valid.RegisterValidation("yyymmddFormat", func(fl v10.FieldLevel) bool {
		//date := fl.Field().String()
		//layout := "2006-01-02" // This layout corresponds to yyyy-mm-dd
		//
		//_, err := time.Parse(layout, date)
		//return err == nil // Returns true if date is valid

		input := fl.Field().Interface()

		inputString := input.(string)
		if inputString == "" {
			return true
		}

		_, err := time.Parse("2006-01-02", inputString)
		return err == nil // Returns true if date is valid
	})
}

func registerCommunityId() {
	valid.RegisterValidation("communityId", func(fl v10.FieldLevel) bool {
		communityId := fl.Field().String()

		return LuhnAccountNumber(communityId) // Returns true if date is valid
	})
}

func registerEmailOrPhoneField() {
	valid.RegisterValidation("emailOrPhoneField", func(fl v10.FieldLevel) bool {
		email := fl.Parent().FieldByName("Email").String()
		phone := fl.Parent().FieldByName("PhoneNumber").String()

		if (email != "" && phone == "") || (email == "" && phone != "") {
			return true
		}

		// If neither is filled, it's invalid
		if email == "" && phone == "" {
			return false
		}

		return true
	})
}

func registerNameIdentifierCommunityIdFields() {
	valid.RegisterValidation("nameIdentifierCommunityIdField", func(fl v10.FieldLevel) bool {
		name := fl.Parent().FieldByName("Name").String()
		identifier := fl.Parent().FieldByName("Identifier").String()
		communityId := fl.Parent().FieldByName("CommunityId").String()

		if (communityId != "" && name == "" && identifier == "") || (communityId == "" && name != "" && identifier != "") {
			return true
		}

		// If neither is filled, it's invalid
		if (name == "" && identifier == "" && communityId == "") || (name != "" && identifier == "" && communityId == "") || (name == "" && identifier != "" && communityId == "") {
			return false
		}

		return true
	})
}

func registerPhoneStandardize() {
	valid.RegisterValidation("phoneFormat0862", func(fl v10.FieldLevel) bool {
		input := fl.Field().Interface()

		inputString := input.(string)
		if inputString == "" {
			return true
		}

		_, err := PhoneNumber("", inputString)
		if err != nil {
			return false
		}

		return true
	})
}

func registerhhmmField() {
	valid.RegisterValidation("hhmmFormat", func(fl v10.FieldLevel) bool {
		// The "15:04" layout is the reference time format for "HH:MM".
		// time.Parse requires the input string to exactly match the layout.
		// If parsing is successful and the formatted output matches the input,
		// it confirms the input was in the correct "HH:MM" format.
		t, err := time.Parse("15:04", fl.Field().String())
		if err != nil {
			return false // Parsing failed, not in HH:MM format
		}

		// Check if formatting the parsed time back to "15:04" matches the original input.
		// This handles cases like "25:00" which time.Parse might partially parse but isn't valid HH:MM.
		return t.Format("15:04") == fl.Field().String()
	})
}

func registeryyyymmddNoExceedTodayFormat() {
	valid.RegisterValidation("yyyymmddNoExceedToday", func(fl v10.FieldLevel) bool {
		date := fl.Field().String()
		layout := "2006-01-02" // This layout corresponds to yyyy-mm-dd
		parsedDate, err := time.Parse(layout, date)
		if err != nil {
			return false
		}
		return parsedDate.Before(common.Now())
	})
}
