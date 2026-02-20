package timec

import (
	"go-community/internal/pkg/stringc"
	"strings"
	"time"
)

// ParseWeekday converts a weekday string to time.Weekday

func ParseWeekday(day string) time.Weekday {
	switch stringc.LowerAndTrimSpace(day) {
	case "sunday":
		return time.Sunday
	case "monday":
		return time.Monday
	case "tuesday":
		return time.Tuesday
	case "wednesday":
		return time.Wednesday
	case "thursday":
		return time.Thursday
	case "friday":
		return time.Friday
	case "saturday":
		return time.Saturday
	default:
		return time.Sunday
	}
}

// ParseDateBoundary interprets a date boundary string using the provided layout and timezone.
// Supports the special keyword "today" (case-insensitive), which resolves to
// midnight of the current day in loc. Any other value is parsed literally in loc.
// If loc is nil, time.UTC is used.
// Example:
//
//	wib, _ := time.LoadLocation("Asia/Jakarta")
//	timec.ParseDateBoundary("today", "2006-01-02", wib)       // today at 00:00 WIB
//	timec.ParseDateBoundary("2025-01-01", "2006-01-02", wib)  // 2025-01-01 00:00 WIB
func ParseDateBoundary(boundary, layout string, loc *time.Location) (time.Time, error) {
	if loc == nil {
		loc = time.UTC
	}
	if strings.EqualFold(boundary, "today") {
		now := time.Now().In(loc)
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc), nil
	}
	return time.ParseInLocation(layout, boundary, loc)
}
