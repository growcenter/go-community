package validator

import (
	"errors"
	"fmt"
	"go-community/internal/models"
	"reflect"
	"regexp"
	"strings"

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
			errs = multierror.Append(errs, ErrorValidateResponse{
				Message: err.Error(),
			})
			return errs.ErrorOrNil()
		}

		var validatorErrs v10.ValidationErrors
		if errors.As(err, &validatorErrs) {
			for _, validatorErr := range validatorErrs {
				// Construct the key for the validation error map
				key := fmt.Sprintf("%s_%s", validatorErr.Namespace(), validatorErr.Tag())

				// Map the validation error key to a corresponding error
				mappedError, found := models.ValidationErrorMapping[key]
				if !found {
					// If not found, use the field and tag for a less specific key
					key = fmt.Sprintf("%s_%s", validatorErr.Field(), validatorErr.Tag())
					mappedError, found = models.ValidationErrorMapping[key]
					if !found {
						// If still not found, create a generic error
						mappedError = fmt.Errorf("%s %s", validatorErr.Tag(), validatorErr.Param())
					}
				}

				// Use the ErrorMapping function to get the ErrorResponse
				errorResponse := models.ErrorMapping(mappedError)
				validateResponse := ErrorValidateResponse{
					Code:    errorResponse.Status,
					Field:   validatorErr.Field(),
					Message: errorResponse.Message,
				}

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
		input := fl.Field().String()
		pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		return regexp.MustCompile(pattern).MatchString(input)
	})
}

func registerPhoneFormat() {
	valid.RegisterValidation("phoneFormat", func(fl v10.FieldLevel) bool {
		input := fl.Field().String()
		pattern := `^\+?(\d{1,3})?[-.\s]?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}$`
		return regexp.MustCompile(pattern).MatchString(input)
	})
}
