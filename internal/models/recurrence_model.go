package models

import (
	"errors"
	"fmt"
	"time"
)

// RecurrencePattern defines the structure for recurring events
// Follows iCalendar RFC 5545 principles but simplified for user-friendliness
type RecurrencePattern struct {
	// Basic recurrence
	Frequency string `json:"frequency" validate:"required,oneof=daily weekly monthly yearly"`
	Interval  int    `json:"interval" validate:"omitempty,min=1"` // Every N days/weeks/months (default: 1)

	// End condition (use either Count OR EndDate, not both)
	Count   *int       `json:"count" validate:"omitempty,min=1"` // Number of occurrences
	EndDate *time.Time `json:"endDate" validate:"omitempty"`     // End by this date

	// Weekly recurrence: which days of the week
	WeekDays []string `json:"weekDays" validate:"omitempty,dive,oneof=monday tuesday wednesday thursday friday saturday sunday"`

	// Monthly recurrence: choose ONE of these patterns
	MonthlyPattern string `json:"monthlyPattern" validate:"omitempty,oneof=day_of_month nth_weekday"`
	DayOfMonth     int    `json:"dayOfMonth" validate:"omitempty,min=1,max=31"`                                                // e.g., 15 = "15th of month"
	NthWeekday     string `json:"nthWeekday" validate:"omitempty,oneof=first second third fourth last"`                        // e.g., "first"
	Weekday        string `json:"weekday" validate:"omitempty,oneof=monday tuesday wednesday thursday friday saturday sunday"` // e.g., "sunday"

	// Optional: Exception and additional dates
	ExcludeDates    []time.Time `json:"excludeDates" validate:"omitempty"`    // Dates to skip (like iCalendar EXDATE)
	AdditionalDates []time.Time `json:"additionalDates" validate:"omitempty"` // Extra dates to include (like iCalendar RDATE)
}

// ValidateRecurrencePattern validates the RecurrencePattern struct
// ensuring that the fields are correctly configured based on the recurrence type
func ValidateRecurrencePattern(pattern *RecurrencePattern) error {
	if pattern == nil {
		return nil // Pattern is optional
	}

	// Validate end condition: cannot specify both Count and EndDate
	if pattern.Count != nil && pattern.EndDate != nil {
		return errors.New("cannot specify both count and endDate in recurrence pattern")
	}

	// Type-specific validation
	switch pattern.Frequency {
	case "weekly":
		return validateWeeklyPattern(pattern)
	case "monthly":
		return validateMonthlyPattern(pattern)
	case "daily", "yearly":
		// Daily and yearly don't have additional requirements beyond basic fields
		return nil
	default:
		return fmt.Errorf("invalid recurrence type: %s", pattern.Frequency)
	}
}

// validateWeeklyPattern validates weekly recurrence patterns
func validateWeeklyPattern(pattern *RecurrencePattern) error {
	// For weekly recurrence, at least one weekday should be specified
	if len(pattern.WeekDays) == 0 {
		return errors.New("at least one weekday required for weekly recurrence")
	}

	// Monthly-specific fields should not be set for weekly recurrence
	if pattern.MonthlyPattern != "" {
		return errors.New("monthlyPattern should not be set for weekly recurrence")
	}
	if pattern.DayOfMonth != 0 {
		return errors.New("dayOfMonth should not be set for weekly recurrence")
	}
	if pattern.NthWeekday != "" {
		return errors.New("nthWeekday should not be set for weekly recurrence")
	}
	if pattern.Weekday != "" {
		return errors.New("weekday should not be set for weekly recurrence")
	}

	return nil
}

// validateMonthlyPattern validates monthly recurrence patterns
func validateMonthlyPattern(pattern *RecurrencePattern) error {
	// For monthly recurrence, a pattern must be specified
	if pattern.MonthlyPattern == "" {
		return errors.New("monthlyPattern required for monthly recurrence")
	}

	switch pattern.MonthlyPattern {
	case "day_of_month":
		// Validate day of month pattern
		if pattern.DayOfMonth < 1 || pattern.DayOfMonth > 31 {
			return errors.New("dayOfMonth must be between 1 and 31")
		}
		// NthWeekday and Weekday should not be set
		if pattern.NthWeekday != "" {
			return errors.New("nthWeekday should not be set for day_of_month pattern")
		}
		if pattern.Weekday != "" {
			return errors.New("weekday should not be set for day_of_month pattern")
		}

	case "nth_weekday":
		// Validate nth weekday pattern
		if pattern.NthWeekday == "" {
			return errors.New("nthWeekday required for nth_weekday pattern")
		}
		if pattern.Weekday == "" {
			return errors.New("weekday required for nth_weekday pattern")
		}
		// DayOfMonth should not be set
		if pattern.DayOfMonth != 0 {
			return errors.New("dayOfMonth should not be set for nth_weekday pattern")
		}

	default:
		return fmt.Errorf("invalid monthlyPattern: %s", pattern.MonthlyPattern)
	}

	// WeekDays should not be set for monthly recurrence
	if len(pattern.WeekDays) > 0 {
		return errors.New("weekDays should not be set for monthly recurrence")
	}

	return nil
}

// GetRecurrenceDescription returns a human-readable description of the recurrence pattern
func (r *RecurrencePattern) GetRecurrenceDescription() string {
	if r == nil {
		return "No recurrence"
	}

	interval := r.Interval
	if interval == 0 {
		interval = 1
	}

	var desc string

	switch r.Frequency {
	case "daily":
		if interval == 1 {
			desc = "Every day"
		} else {
			desc = fmt.Sprintf("Every %d days", interval)
		}

	case "weekly":
		if interval == 1 {
			desc = fmt.Sprintf("Every week on %s", formatWeekDays(r.WeekDays))
		} else {
			desc = fmt.Sprintf("Every %d weeks on %s", interval, formatWeekDays(r.WeekDays))
		}

	case "monthly":
		var patternDesc string
		if r.MonthlyPattern == "day_of_month" {
			patternDesc = fmt.Sprintf("on day %d", r.DayOfMonth)
		} else if r.MonthlyPattern == "nth_weekday" {
			patternDesc = fmt.Sprintf("on the %s %s", r.NthWeekday, r.Weekday)
		}

		if interval == 1 {
			desc = fmt.Sprintf("Every month %s", patternDesc)
		} else {
			desc = fmt.Sprintf("Every %d months %s", interval, patternDesc)
		}

	case "yearly":
		if interval == 1 {
			desc = "Every year"
		} else {
			desc = fmt.Sprintf("Every %d years", interval)
		}
	}

	// Add end condition
	if r.Count != nil {
		desc += fmt.Sprintf(", %d times", *r.Count)
	} else if r.EndDate != nil {
		desc += fmt.Sprintf(", until %s", r.EndDate.Format("2006-01-02"))
	}

	// Add exception info
	if len(r.ExcludeDates) > 0 {
		desc += fmt.Sprintf(" (excluding %d date(s))", len(r.ExcludeDates))
	}

	// Add additional dates info
	if len(r.AdditionalDates) > 0 {
		desc += fmt.Sprintf(" (including %d additional date(s))", len(r.AdditionalDates))
	}

	return desc
}

// formatWeekDays formats the weekdays array into a human-readable string
func formatWeekDays(days []string) string {
	if len(days) == 0 {
		return ""
	}
	if len(days) == 1 {
		return capitalize(days[0])
	}
	if len(days) == 7 {
		return "every day"
	}

	result := ""
	for i, day := range days {
		if i > 0 && i == len(days)-1 {
			result += " and "
		} else if i > 0 {
			result += ", "
		}
		result += capitalize(day)
	}
	return result
}

// capitalize capitalizes the first letter of a string
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}
