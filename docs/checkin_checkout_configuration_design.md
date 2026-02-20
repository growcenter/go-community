# Check-in/Check-out Configuration Design

## Current Design (Your Model)

```go
// Check-in Config
CheckInRequired  bool `gorm:"type:boolean;default:true" json:"check_in_required"`
CheckOutRequired bool `gorm:"type:boolean;default:false" json:"check_out_required"`
```

**Problem:** Only two states (required/not required) - what about "optional but tracked"?

---

## Recommended Design: Three-State Configuration

### Option 1: Enum-Based (Recommended)

```go
// Check-in/Check-out Config
CheckInMode  string `gorm:"type:varchar(20);default:'optional'" json:"check_in_mode"`   // none, optional, required
CheckOutMode string `gorm:"type:varchar(20);default:'none'" json:"check_out_mode"`      // none, optional, required
```

**Values:**

- `"none"` - Feature disabled, no check-in/check-out tracking
- `"optional"` - Feature enabled, tracked but not enforced
- `"required"` - Feature enabled, enforced (must check-in/out)

**Example Configurations:**

```json
// Christmas Service - track check-ins but don't enforce
{
  "check_in_mode": "optional",
  "check_out_mode": "none"
}

// Sunday Service - require check-in for attendance tracking
{
  "check_in_mode": "required",
  "check_out_mode": "none"
}

// Kids Program - require both for safety
{
  "check_in_mode": "required",
  "check_out_mode": "required"
}

// Announcement Event - no tracking at all
{
  "check_in_mode": "none",
  "check_out_mode": "none"
}

// Conference Workshop - track both but don't enforce
{
  "check_in_mode": "optional",
  "check_out_mode": "optional"
}
```

---

### Option 2: Detailed Configuration (More Flexible)

```go
// Check-in/Check-out Config
CheckInConfig  AttendanceConfig `gorm:"type:jsonb" json:"check_in_config"`
CheckOutConfig AttendanceConfig `gorm:"type:jsonb" json:"check_out_config"`

type AttendanceConfig struct {
    Enabled       bool   `json:"enabled"`        // Is this feature enabled?
    Required      bool   `json:"required"`       // Is it mandatory?
    WindowBefore  int    `json:"window_before"`  // Minutes before event start
    WindowAfter   int    `json:"window_after"`   // Minutes after event start
    AllowLate     bool   `json:"allow_late"`     // Allow check-in after window?
    LateThreshold int    `json:"late_threshold"` // Minutes to mark as "late"
}
```

**Example:**

```json
{
  "check_in_config": {
    "enabled": true,
    "required": true,
    "window_before": 30, // Can check-in 30 min before
    "window_after": 15, // Can check-in up to 15 min after start
    "allow_late": true, // Allow late check-ins
    "late_threshold": 10 // Mark as "late" if >10 min after start
  },
  "check_out_config": {
    "enabled": false,
    "required": false
  }
}
```

---

## Comparison Table

| Feature               | Current (Boolean) | Option 1 (Enum)            | Option 2 (JSONB)  |
| --------------------- | ----------------- | -------------------------- | ----------------- |
| **Simplicity**        | ✅ Very simple    | ✅ Simple                  | ❌ Complex        |
| **Flexibility**       | ❌ Limited        | ✅ Good                    | ✅ Excellent      |
| **States**            | 2 (on/off)        | 3 (none/optional/required) | Unlimited         |
| **Time Windows**      | ❌ No             | ❌ No                      | ✅ Yes            |
| **Late Tracking**     | ❌ No             | ❌ No                      | ✅ Yes            |
| **Database Size**     | ✅ Small          | ✅ Small                   | ❌ Larger         |
| **Query Performance** | ✅ Fast           | ✅ Fast                    | ⚠️ Slower (JSONB) |
| **Validation**        | ✅ Easy           | ✅ Easy                    | ⚠️ Complex        |

---

## Real-World Use Cases

### Use Case 1: Christmas Service (Optional Check-in)

**Scenario:** Want to track who actually showed up, but don't block registration if they don't check-in.

**Current Design (Boolean):**

```go
CheckInRequired: false  // ❌ Problem: No tracking at all
```

**Option 1 (Enum):**

```go
CheckInMode: "optional"  // ✅ Perfect: Track but don't enforce
```

**Option 2 (JSONB):**

```json
{
  "check_in_config": {
    "enabled": true,
    "required": false,
    "window_before": 30,
    "window_after": 60
  }
}
```

---

### Use Case 2: Kids Program (Strict Check-in/out)

**Scenario:** Parents must check-in kids when dropping off and check-out when picking up (safety).

**Current Design (Boolean):**

```go
CheckInRequired: true
CheckOutRequired: true  // ✅ Works, but no time windows
```

**Option 1 (Enum):**

```go
CheckInMode: "required"
CheckOutMode: "required"  // ✅ Works well
```

**Option 2 (JSONB):**

```json
{
  "check_in_config": {
    "enabled": true,
    "required": true,
    "window_before": 15,
    "window_after": 0, // No late check-ins
    "allow_late": false
  },
  "check_out_config": {
    "enabled": true,
    "required": true
  }
}
```

---

### Use Case 3: Sunday Service (Required Check-in, Track Late)

**Scenario:** Track attendance, mark people as late if they arrive >10 min after start.

**Current Design (Boolean):**

```go
CheckInRequired: true  // ❌ Can't track "late" status
```

**Option 1 (Enum):**

```go
CheckInMode: "required"  // ⚠️ Can track, but late logic is elsewhere
```

**Option 2 (JSONB):**

```json
{
  "check_in_config": {
    "enabled": true,
    "required": true,
    "window_before": 30,
    "window_after": 120, // Can check-in up to 2 hours late
    "allow_late": true,
    "late_threshold": 10 // ✅ Mark as late if >10 min
  }
}
```

---

### Use Case 4: Conference (Track Both, Don't Enforce)

**Scenario:** Track check-in/out for session attendance reports, but don't block access.

**Current Design (Boolean):**

```go
CheckInRequired: false
CheckOutRequired: false  // ❌ No tracking at all
```

**Option 1 (Enum):**

```go
CheckInMode: "optional"
CheckOutMode: "optional"  // ✅ Perfect
```

**Option 2 (JSONB):**

```json
{
  "check_in_config": {
    "enabled": true,
    "required": false
  },
  "check_out_config": {
    "enabled": true,
    "required": false
  }
}
```

---

## Recommendation for Your Church Event System

### **Use Option 1 (Enum-Based)**

**Why:**

1. ✅ **Covers 95% of use cases**
   - None: Announcement events
   - Optional: Christmas service (track but don't enforce)
   - Required: Sunday service, Kids program

2. ✅ **Simple to implement and validate**

   ```go
   validate:"oneof=none optional required"
   ```

3. ✅ **Easy to query**

   ```sql
   SELECT * FROM event_sessions WHERE check_in_mode = 'required';
   ```

4. ✅ **Clear semantics**
   - Developers understand immediately
   - Non-technical users can grasp the concept

5. ✅ **Extensible**
   - Can add time windows later if needed
   - Can migrate to Option 2 (JSONB) if requirements grow

### **When to Use Option 2 (JSONB)**

Only if you need:

- ❌ Complex time window logic per session
- ❌ Different late thresholds per session
- ❌ Highly customizable check-in rules

**For most church events, Option 1 is sufficient.**

---

## Implementation

### Updated Model

```go
type EventSession struct {
    // ... other fields ...

    // Check-in/Check-out Config (Three-state)
    CheckInMode  string `gorm:"type:varchar(20);default:'optional'" json:"check_in_mode" validate:"required,oneof=none optional required"`
    CheckOutMode string `gorm:"type:varchar(20);default:'none'" json:"check_out_mode" validate:"required,oneof=none optional required"`

    // Check-in/Check-out Windows (optional, can be NULL)
    CheckInStartAt *time.Time `gorm:"type:timestamptz" json:"check_in_start_at"`
    CheckInEndAt   *time.Time `gorm:"type:timestamptz" json:"check_in_end_at"`
    CheckOutStartAt *time.Time `gorm:"type:timestamptz" json:"check_out_start_at"`
    CheckOutEndAt   *time.Time `gorm:"type:timestamptz" json:"check_out_end_at"`

    // ... other fields ...
}
```

### Migration

```sql
-- Update existing table
ALTER TABLE event_sessions
    DROP COLUMN check_in_required,
    DROP COLUMN check_out_required,
    ADD COLUMN check_in_mode VARCHAR(20) DEFAULT 'optional' CHECK (check_in_mode IN ('none', 'optional', 'required')),
    ADD COLUMN check_out_mode VARCHAR(20) DEFAULT 'none' CHECK (check_out_mode IN ('none', 'optional', 'required'));

-- Migrate existing data
UPDATE event_sessions
SET check_in_mode = CASE
    WHEN check_in_required = true THEN 'required'
    ELSE 'optional'
END;

UPDATE event_sessions
SET check_out_mode = CASE
    WHEN check_out_required = true THEN 'required'
    ELSE 'none'
END;
```

### Validation Logic

```go
func ValidateCheckInMode(session *EventSession) error {
    // If check-in is required, must have time windows
    if session.CheckInMode == "required" {
        if session.CheckInStartAt == nil || session.CheckInEndAt == nil {
            return errors.New("check-in time windows required when check_in_mode is 'required'")
        }
    }

    // Check-out can't be required if check-in is none
    if session.CheckInMode == "none" && session.CheckOutMode != "none" {
        return errors.New("cannot require check-out without check-in")
    }

    return nil
}
```

---

## Summary

**Answer:** ✅ **YES, make check-in/check-out configurable with three states**

**Recommended Design:**

```go
CheckInMode  string  // "none" | "optional" | "required"
CheckOutMode string  // "none" | "optional" | "required"
```

**Why:**

- Different event types have different needs
- "Optional" state is critical (track but don't enforce)
- Simple, clear, and covers 95% of use cases
- Easy to implement and validate
- Aligns with your spec's event type matrix

**Replace your current boolean fields with the enum approach!**
