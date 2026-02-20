# Recurrence Pattern Recommendation

## Current Implementation Analysis

Your current `RecurrencePattern` struct:

```go
RecurrencePattern struct {
    Type     string   `json:"type" validate:"required,oneof=daily weekly monthly yearly"`
    Days     []string `json:"days" validate:"omitempty,dive,oneof=monday tuesday wednesday thursday friday saturday sunday"`
    Interval int      `json:"interval" validate:"omitempty,min=1"`
}
```

## Industry Standards Comparison

### iCalendar RFC 5545 (RRULE)

- **Used by:** Google Calendar, Apple Calendar, Outlook, most calendar applications
- **Format:** `RRULE:FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,WE,FR;COUNT=10`
- **Pros:** Industry standard, widely supported, comprehensive
- **Cons:** Complex syntax, can be overwhelming for simple use cases

### Quartz Cron Expressions

- **Used by:** Job schedulers, background tasks, system automation
- **Format:** `0 15 10 ? * MON-FRI` (At 10:15 AM, Monday through Friday)
- **Pros:** Very powerful, precise to the second
- **Cons:** Cryptic syntax, overkill for user-facing event scheduling

## Gap Analysis

### ✅ What You Have (Good!)

1. **Basic frequencies**: daily, weekly, monthly, yearly
2. **Interval support**: Every N days/weeks/months
3. **Day of week selection**: For weekly recurrence
4. **End date**: `RecurrenceEndDate` field

### ❌ What You're Missing (Important for Church Events)

1. **Count-based recurrence**
   - Example: "Repeat 10 times" instead of setting an end date
   - Use case: "New members class - 6 sessions"

2. **Nth weekday of month**
   - Example: "2nd Sunday of every month" or "Last Friday"
   - Use case: "Monthly leadership meeting - 1st Tuesday"

3. **Day of month**
   - Example: "15th of every month"
   - Use case: "Monthly giving report - 15th of each month"

4. **Exception dates**
   - Example: Skip specific dates (holidays, special events)
   - Use case: "Weekly service except Christmas and Easter"

5. **Week start configuration**
   - Example: Define whether week starts on Sunday or Monday
   - Use case: Important for "weekly" calculations

## Recommended Enhanced Structure

### Option 1: Simplified (Recommended for MVP)

```go
// RecurrencePattern defines the structure for recurring events
// Follows iCalendar RFC 5545 principles but simplified for user-friendliness
RecurrencePattern struct {
    // Basic recurrence
    Type     string `json:"type" validate:"required,oneof=daily weekly monthly yearly"`
    Interval int    `json:"interval" validate:"omitempty,min=1"` // Every N days/weeks/months (default: 1)

    // End condition (use either Count OR EndDate, not both)
    Count   *int       `json:"count" validate:"omitempty,min=1"`         // Number of occurrences
    EndDate *time.Time `json:"endDate" validate:"omitempty"`             // End by this date

    // Weekly recurrence: which days of the week
    WeekDays []string `json:"weekDays" validate:"omitempty,dive,oneof=monday tuesday wednesday thursday friday saturday sunday"`

    // Monthly recurrence: choose ONE of these patterns
    MonthlyPattern string `json:"monthlyPattern" validate:"omitempty,oneof=day_of_month nth_weekday"`
    DayOfMonth     int    `json:"dayOfMonth" validate:"omitempty,min=1,max=31"`        // e.g., 15 = "15th of month"
    NthWeekday     string `json:"nthWeekday" validate:"omitempty,oneof=first second third fourth last"` // e.g., "first"
    Weekday        string `json:"weekday" validate:"omitempty,oneof=monday tuesday wednesday thursday friday saturday sunday"` // e.g., "sunday"

    // Optional: Exception and additional dates
    ExcludeDates    []time.Time `json:"excludeDates" validate:"omitempty"`    // Dates to skip (like iCalendar EXDATE)
    AdditionalDates []time.Time `json:"additionalDates" validate:"omitempty"` // Extra dates to include (like iCalendar RDATE)
}
```

**Examples:**

```json
// Weekly service every Sunday for 52 weeks
{
  "type": "weekly",
  "interval": 1,
  "count": 52,
  "weekDays": ["sunday"]
}

// Monthly leadership meeting - 1st Tuesday
{
  "type": "monthly",
  "interval": 1,
  "monthlyPattern": "nth_weekday",
  "nthWeekday": "first",
  "weekday": "tuesday",
  "endDate": "2026-12-31T00:00:00Z"
}

// Monthly giving report - 15th of each month
{
  "type": "monthly",
  "interval": 1,
  "monthlyPattern": "day_of_month",
  "dayOfMonth": 15,
  "count": 12
}

// Bi-weekly small group - every other Wednesday, 10 sessions
{
  "type": "weekly",
  "interval": 2,
  "weekDays": ["wednesday"],
  "count": 10
}
```

### Option 2: Full iCalendar Compatibility (Advanced)

```go
// RecurrencePattern follows iCalendar RFC 5545 RRULE specification
RecurrencePattern struct {
    // Core fields
    Freq     string `json:"freq" validate:"required,oneof=daily weekly monthly yearly"`
    Interval int    `json:"interval" validate:"omitempty,min=1"`

    // End conditions (mutually exclusive)
    Count *int       `json:"count" validate:"omitempty,min=1"`
    Until *time.Time `json:"until" validate:"omitempty"`

    // By-rules (filters)
    ByDay       []string `json:"byDay" validate:"omitempty"`       // e.g., ["MO", "WE", "FR"] or ["1SU", "-1FR"]
    ByMonthDay  []int    `json:"byMonthDay" validate:"omitempty"`  // e.g., [1, 15, -1] (1st, 15th, last day)
    ByMonth     []int    `json:"byMonth" validate:"omitempty"`     // e.g., [1, 6, 12] (Jan, Jun, Dec)
    BySetPos    []int    `json:"bySetPos" validate:"omitempty"`    // e.g., [1, -1] (first and last)

    // Configuration
    WeekStart string `json:"weekStart" validate:"omitempty,oneof=MO TU WE TH FR SA SU"`

    // Exceptions and additions
    ExDates []time.Time `json:"exDates" validate:"omitempty"` // Exception dates
    RDates  []time.Time `json:"rDates" validate:"omitempty"`  // Additional dates
}
```

**Pros:**

- Full compatibility with iCalendar standard
- Can import/export to Google Calendar, Outlook, etc.
- Maximum flexibility

**Cons:**

- More complex for users to understand
- Requires validation logic to prevent invalid combinations
- May be overkill for simple church event scheduling

## Recommendation

### For Your Church Event System: Use **Option 1 (Simplified)**

**Reasons:**

1. ✅ **Covers 95% of church event use cases**
   - Weekly services
   - Monthly meetings (1st Tuesday, last Friday, etc.)
   - Bi-weekly small groups
   - Limited-session classes (count-based)

2. ✅ **User-friendly**
   - Clear, intuitive field names
   - Easy to build UI forms for
   - Validation is straightforward

3. ✅ **Extensible**
   - Can add more fields later if needed
   - Can convert to RRULE format for export if required

4. ✅ **Production-ready**
   - Handles exception dates for holidays
   - Supports both count and date-based endings
   - Covers monthly patterns (both day-of-month and nth-weekday)

### Migration Path

1. **Phase 1 (Now):** Implement Option 1 structure
2. **Phase 2 (Later):** Add RRULE import/export helpers if integration with external calendars is needed
3. **Phase 3 (Future):** Consider full iCalendar compatibility if advanced users request it

## Implementation Notes

### Validation Rules

```go
// Ensure only one end condition is specified
if pattern.Count != nil && pattern.EndDate != nil {
    return errors.New("cannot specify both count and endDate")
}

// Monthly pattern validation
if pattern.Type == "monthly" && pattern.MonthlyPattern == "nth_weekday" {
    if pattern.NthWeekday == "" || pattern.Weekday == "" {
        return errors.New("nthWeekday and weekday required for nth_weekday pattern")
    }
}

if pattern.Type == "monthly" && pattern.MonthlyPattern == "day_of_month" {
    if pattern.DayOfMonth < 1 || pattern.DayOfMonth > 31 {
        return errors.New("dayOfMonth must be between 1 and 31")
    }
}

// Weekly pattern validation
if pattern.Type == "weekly" && len(pattern.WeekDays) == 0 {
    return errors.New("at least one weekday required for weekly recurrence")
}
```

### Database Storage

Store as JSONB in PostgreSQL (as you're already doing):

```sql
recurrence_pattern JSONB
```

### UI Considerations

Create a stepped form:

1. **Step 1:** Choose frequency (daily/weekly/monthly/yearly)
2. **Step 2:** Configure pattern based on frequency
   - Weekly: Select days
   - Monthly: Choose pattern type, then configure
3. **Step 3:** Set end condition (count or date)
4. **Step 4:** (Optional) Add exception dates

## Real-World Church Event Examples

### Example 1: Sunday Service

```json
{
  "type": "weekly",
  "interval": 1,
  "weekDays": ["sunday"],
  "excludeDates": ["2026-12-25", "2027-04-04"] // Christmas, Easter
}
```

### Example 2: Monthly Prayer Meeting (1st Friday)

```json
{
  "type": "monthly",
  "interval": 1,
  "monthlyPattern": "nth_weekday",
  "nthWeekday": "first",
  "weekday": "friday",
  "endDate": "2026-12-31T23:59:59Z"
}
```

### Example 3: New Members Class (6 sessions, bi-weekly)

```json
{
  "type": "weekly",
  "interval": 2,
  "weekDays": ["saturday"],
  "count": 6
}
```

### Example 4: Youth Camp Planning (Last Saturday of month)

```json
{
  "type": "monthly",
  "interval": 1,
  "monthlyPattern": "nth_weekday",
  "nthWeekday": "last",
  "weekday": "saturday",
  "count": 12
}
```

### Example 5: Weekly Service with Special Dates

```json
{
  "type": "weekly",
  "interval": 1,
  "weekDays": ["sunday"],
  "excludeDates": ["2026-12-25T00:00:00Z"], // Skip Christmas
  "additionalDates": [
    "2026-12-24T18:00:00Z", // Add Christmas Eve service
    "2026-12-31T22:00:00Z" // Add New Year's Eve service
  ]
}
```

**Use case:** Regular Sunday services, but skip Christmas Sunday and add special services on Christmas Eve and New Year's Eve.

## Conclusion

Your current implementation is a **good start** but **missing critical features** for a production church event system. The recommended Option 1 structure:

- ✅ Aligns with industry standards (iCalendar principles)
- ✅ Remains user-friendly and intuitive
- ✅ Covers all common church event patterns
- ✅ Is extensible for future needs
- ✅ Easier to implement and maintain than full RRULE

**Next Steps:**

1. Review this recommendation with your team
2. Decide on Option 1 vs Option 2
3. Update the `RecurrencePattern` struct
4. Add validation logic
5. Update database migrations
6. Build UI forms for pattern configuration
