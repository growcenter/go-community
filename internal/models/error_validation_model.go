package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

type (
	ErrorValidateResponse struct {
		Code    string `json:"code,omitempty" example:"accountNumber_required"`
		Field   string `json:"field,omitempty" example:"MISSING_FIELD"`
		Message string `json:"message,omitempty" example:"field is missing"`
	}
)

func (e ErrorValidateResponse) Error() string {
	return fmt.Sprintf("code: %s, field: %s, message: %s", e.Code, e.Field, e.Message)
}

func ErrorValidationMapping(validationError validator.FieldError) string {
	switch {
	case validationError.Tag() == "required":
		return fmt.Sprintf("%s is required", validationError.Field())
	case validationError.Tag() == "required_without_all":
		return fmt.Sprintf("%s is required when %s is not inputted", validationError.Field(), strings.ReplaceAll(validationError.Param(), " ", ", "))
	case validationError.Tag() == "oneof":
		return fmt.Sprintf("%s should be inputted either %s", validationError.Field(), validationError.Param())
	case validationError.Tag() == "min":
		return fmt.Sprintf("%s must be at least %s characters", validationError.Field(), validationError.Param())
	case validationError.Tag() == "max":
		return fmt.Sprintf("%s must be at most %s characters", validationError.Field(), validationError.Param())
	case validationError.Tag() == "email":
		return fmt.Sprintf("%s must be a valid email address", validationError.Field())
	case validationError.Tag() == "nospecial":
		return fmt.Sprintf("%s must not contain special characters", validationError.Field())
	case validationError.Tag() == "noStartEndSpaces":
		return fmt.Sprintf("%s must not start or end with spaces", validationError.Field())
	case validationError.Tag() == "date":
		return fmt.Sprintf("%s must be a valid date in YYYY-MM-DD format", validationError.Field())
	case validationError.Tag() == "datetime":
		return fmt.Sprintf("%s must be a valid date in YYYY-MM-DD HH:MM:SS format", validationError.Field())
	case validationError.Tag() == "communityId":
		return fmt.Sprintf("%s must be a valid community id", validationError.Field())
	case validationError.Tag() == "yyymmddFormat":
		return fmt.Sprintf("%s must be a valid date in YYYY-MM-DD format", validationError.Field())
	case validationError.Tag() == "phoneFormat":
		return fmt.Sprintf("%s must be a valid phone number in format +628123456789", validationError.Field())
	case validationError.Tag() == "emailFormat":
		return fmt.Sprintf("%s must be a valid email address", validationError.Field())
	case validationError.Tag() == "emailPhoneFormat":
		return fmt.Sprintf("%s must be a valid email or phone number", validationError.Field())
	case validationError.Tag() == "numeric":
		return fmt.Sprintf("%s must be a valid number", validationError.Field())
	case validationError.Tag() == "daterange":
		return fmt.Sprintf("Your date range should be between %s and %s", time.Now().AddDate(0, -6, 0).Format("2006-01-02"), time.Now().Format("2006-01-02"))
	case validationError.Tag() == "yyyymmddNoExceedToday":
		return fmt.Sprintf("%s must be a valid date in YYYY-MM-DD format and not exceed today", validationError.Field())
	case validationError.Tag() == "hhmmFormat":
		return fmt.Sprintf("%s must be a valid time in HH:MM format", validationError.Field())
	case validationError.Tag() == "phoneFormat0862":
		return fmt.Sprintf("%s must be a valid phone number in format +628123456789 or 08123456789 or 628123456789", validationError.Field())
	case validationError.Tag() == "nameIdentifierCommunityIdField":
		return fmt.Sprintf("%s cannot be empty", validationError.Field())
	case validationError.Tag() == "emailOrPhoneField":
		return fmt.Sprintf("Email or Phone Number cannot be empty")
	default:
		return fmt.Sprintf("invalid input on field %s: %s", validationError.Field(), validationError.Tag())
	}
}

func ErrorValidateResponseMapping(validationError validator.FieldError) ErrorValidateResponse {
	switch {
	default:
		return ErrorValidateResponse{
			Code:    "INVALID_REQUEST",
			Field:   validationError.Field(),
			Message: ErrorValidationMapping(validationError),
		}
	}
}
