# Event System Sequence Diagrams - Quick Reference

This document provides a quick overview of all sequence diagrams in the event management system specification.

---

## Overview of Diagrams

The event management system includes **7 comprehensive sequence diagrams** that cover all major workflows:

### Core Workflows (Individual Operations)

1. **Event Creation Flow** - Creating a single event
2. **Event Session Creation Flow** - Creating a single session
3. **Recurring Event with Sessions Flow** - Creating recurring event + multiple sessions
4. **Registration with Check-in Flow** - User registration and check-in scenarios

### Bulk Workflows (Batch Operations)

5. **Bulk Event + Sessions Creation Flow** - Creating event with sessions in one request
6. **Bulk Form + Questions Creation Flow** - Creating form with questions in one request
7. **Event + Sessions + Forms Creation Flow** - Complete setup in one workflow

---

## 1. Event Creation Flow

**Purpose:** Shows how to create a single event with recurrence pattern

**Key Steps:**

1. Client sends `POST /api/v1/events`
2. Validate request (fields, recurrence pattern, location)
3. Generate event code
4. Insert into `events` table
5. If recurring: Generate occurrences
6. Return event + occurrences

**Result:**

- 1 Event
- N Occurrences (if recurring)

**Use Case:** Creating a standalone event like "Christmas Service 2026"

---

## 2. Event Session Creation Flow

**Purpose:** Shows how to create a session within an existing event

**Key Steps:**

1. Client sends `POST /api/v1/events/:code/sessions`
2. Validate session data (type, capacity, check-in config)
3. Validate parent event exists
4. Process check-in/check-out configuration
5. Process form configuration
6. Insert into `event_sessions` table
7. Store JSONB configs (geolocation, identifier, form overrides)

**Result:**

- 1 Session linked to parent event

**Use Case:** Adding "Kids Session (9:00-10:00)" to Sunday Service

---

## 3. Recurring Event with Sessions Flow

**Purpose:** Shows creating a recurring event and then adding multiple sessions

**Key Steps:**

1. Create recurring event (e.g., 12 weekly occurrences)
2. Generate 12 occurrences
3. Create Kids Session
4. Create Youth Session
5. Each session applies to all occurrences

**Result:**

- 1 Event
- 12 Occurrences
- 2 Sessions
- 24 Session-Occurrence combinations

**Use Case:** Setting up "Sunday Service" with Kids and Youth sessions for 3 months

---

## 4. Registration with Check-in Flow

**Purpose:** Shows how check-in configuration affects registration and attendance

**Scenarios Covered:**

- **On-time check-in** (08:45, 15 min early) → Status: `on_time`
- **Late check-in** (09:12, 12 min late) → Status: `late`
- **Very late check-in** (09:20, 20 min late) → Status: `very_late` or rejected

**Key Validation:**

1. Calculate check-in window (start - 30 min to start + 15 min)
2. Check if current time is within window
3. Calculate lateness (current time - event start)
4. Compare with late threshold (10 min)
5. Determine status: `on_time`, `late`, `very_late`

**Use Case:** User checking in for Sunday Service with late tracking

---

## 5. Bulk Event + Sessions Creation Flow ⭐ NEW

**Purpose:** Create event with multiple sessions in a single API call

**Key Steps:**

1. Client sends `POST /api/v1/events/bulk` with event + sessions array
2. Validate event data
3. Validate all sessions (check overlaps, check-in config)
4. **BEGIN TRANSACTION** (all-or-nothing)
5. Create event
6. Generate occurrences (if recurring)
7. Loop: Create each session
8. **COMMIT TRANSACTION** (or rollback if any step fails)

**Result:**

- 1 Event
- 12 Occurrences (if recurring)
- 3 Sessions (Kids, Youth, Adult)
- 36 Session-Occurrence combinations

**Benefits:**

- ✅ Atomic operation (all succeed or all fail)
- ✅ Single API call
- ✅ Validates session time overlaps
- ✅ Faster than individual calls

**Use Case:** Admin setting up "Christmas Service" with Kids, Youth, and Adult sessions all at once

**Request Example:**

```json
{
  "event": {
    "title": "Christmas Service 2026",
    "recurrence": {
      "frequency": "weekly",
      "weekDays": ["sunday"],
      "count": 12
    }
  },
  "sessions": [
    {
      "title": "Kids Session",
      "startAt": "09:00",
      "endAt": "10:00",
      "checkInConfig": { "enabled": true, "required": true }
    },
    {
      "title": "Youth Session",
      "startAt": "10:00",
      "endAt": "11:00",
      "checkInConfig": { "enabled": true, "required": false }
    },
    {
      "title": "Adult Session",
      "startAt": "11:00",
      "endAt": "12:00",
      "checkInConfig": { "enabled": true, "required": false }
    }
  ]
}
```

---

## 6. Bulk Form + Questions Creation Flow ⭐ NEW

**Purpose:** Create a form with all questions in a single API call

**Key Steps:**

1. Client sends `POST /api/v1/forms/bulk` with form + fields array
2. Validate form metadata
3. Validate all fields (unique keys, types, display order)
4. Validate conditional logic (dependencies exist)
5. **BEGIN TRANSACTION** (all-or-nothing)
6. Create form
7. Loop: Create each field
8. Create field conditions (if any)
9. Create field validations (if any)
10. **COMMIT TRANSACTION** (or rollback if any step fails)

**Result:**

- 1 Form
- 10 Fields
- Field options (for select/multiselect)
- Conditional logic rules
- Validation rules

**Benefits:**

- ✅ Atomic operation
- ✅ Single API call
- ✅ Validates field dependencies
- ✅ Prevents orphaned fields

**Use Case:** Admin creating "Christmas Registration Form" with 10 custom questions

**Request Example:**

```json
{
  "form": {
    "code": "FORM-CHRISTMAS-2026",
    "title": "Christmas Registration",
    "formType": "event_registration"
  },
  "fields": [
    {
      "fieldKey": "dietary_preference",
      "fieldType": "multiselect",
      "label": "Dietary Restrictions",
      "displayOrder": 1,
      "isRequired": false,
      "options": ["vegetarian", "vegan", "halal", "kosher", "none"]
    },
    {
      "fieldKey": "tshirt_size",
      "fieldType": "select",
      "label": "T-shirt Size",
      "displayOrder": 2,
      "isRequired": true,
      "options": ["S", "M", "L", "XL", "XXL"]
    },
    {
      "fieldKey": "special_needs",
      "fieldType": "textarea",
      "label": "Any special needs?",
      "displayOrder": 3,
      "isRequired": false
    }
    // ... 7 more fields
  ]
}
```

---

## 7. Event + Sessions + Forms Creation Flow (Complete) ⭐ NEW

**Purpose:** Shows the complete workflow for setting up an event system from scratch

**Workflow:**

1. **Create Primary Form** (10 questions) → Form ID: 100
2. **Create Kids Form** (5 questions) → Form ID: 101
3. **Create Event with Sessions** (link to forms)
   - Event: Recurring, 12 weeks
   - Kids Session → Form 101
   - Youth Session → Form 100
   - Adult Session → Form 100

**Result:**

- 2 Forms (100, 101)
- 15 Form Fields (10 + 5)
- 1 Event
- 12 Occurrences
- 3 Sessions
- 36 Session-Occurrence combinations

**Benefits:**

- ✅ Shows complete setup workflow
- ✅ Demonstrates form reusability (Form 100 used by 2 sessions)
- ✅ Shows form specialization (Kids use different form)

**Use Case:** Admin setting up entire "Sunday Service" system from scratch

**Final State:**

```
Sunday Service Event
├── 12 Occurrences (every Sunday for 12 weeks)
├── Kids Session (9:00-10:00)
│   └── Uses Form 101 (Kids Form - 5 questions)
├── Youth Session (10:00-11:00)
│   └── Uses Form 100 (Primary Form - 10 questions)
└── Adult Session (11:00-12:00)
    └── Uses Form 100 (Primary Form - 10 questions)

Users can now:
1. Register for any session
2. Fill out appropriate form (Kids or Primary)
3. Check-in on event day
```

---

## Comparison: Individual vs Bulk Operations

### Individual Operations (Diagrams 1-4)

**Pros:**

- ✅ Simple, focused workflows
- ✅ Easy to understand
- ✅ Granular control

**Cons:**

- ❌ Multiple API calls needed
- ❌ No atomicity across calls
- ❌ Slower for bulk operations
- ❌ Manual rollback if something fails

**When to Use:**

- Creating/updating single items
- Exploratory/testing scenarios
- When you need to create items at different times

### Bulk Operations (Diagrams 5-7)

**Pros:**

- ✅ Single API call
- ✅ Atomic transactions (all-or-nothing)
- ✅ Faster performance
- ✅ Automatic rollback on failure
- ✅ Validates relationships (e.g., session overlaps, field dependencies)

**Cons:**

- ❌ More complex request structure
- ❌ Larger payload size
- ❌ All-or-nothing (can't partially succeed)

**When to Use:**

- Initial event setup
- Bulk imports
- Admin workflows
- When consistency is critical

---

## Transaction Management

### Individual Operations

```
Call 1: Create Event ✅
Call 2: Create Session 1 ✅
Call 3: Create Session 2 ❌ FAILS
Result: Event + Session 1 exist, Session 2 doesn't
        Manual cleanup needed!
```

### Bulk Operations

```
BEGIN TRANSACTION
  Create Event ✅
  Create Session 1 ✅
  Create Session 2 ❌ FAILS
ROLLBACK TRANSACTION
Result: Nothing created, database is clean
```

---

## API Endpoints Summary

### Individual Operations

```
POST   /api/v1/events                    # Create event
POST   /api/v1/events/:code/sessions     # Create session
POST   /api/v1/forms                     # Create form
POST   /api/v1/forms/:id/fields          # Create field
```

### Bulk Operations

```
POST   /api/v1/events/bulk               # Create event + sessions
POST   /api/v1/forms/bulk                # Create form + fields
```

---

## Use Case Decision Tree

```
Do you need to create multiple related items?
├─ NO → Use individual operations (Diagrams 1-2)
│   └─ Example: Adding one session to existing event
│
└─ YES → Do they need to be created atomically?
    ├─ NO → Use individual operations (Diagrams 1-2)
    │   └─ Example: Creating forms over time
    │
    └─ YES → Use bulk operations (Diagrams 5-7)
        ├─ Event + Sessions → Diagram 5
        ├─ Form + Questions → Diagram 6
        └─ Complete Setup → Diagram 7
```

---

## Error Handling

### Individual Operations

- Each call can fail independently
- Partial success is possible
- Manual cleanup may be needed

### Bulk Operations

- Single point of failure
- All-or-nothing guarantee
- Automatic rollback on any error
- No manual cleanup needed

**Example Errors:**

- ❌ Session time overlap detected → Rollback entire transaction
- ❌ Form field dependency missing → Rollback entire transaction
- ❌ Invalid check-in configuration → Rollback entire transaction

---

## Performance Considerations

### Individual Operations

```
Create Event:     200ms
Create Session 1: 150ms
Create Session 2: 150ms
Create Session 3: 150ms
Total:            650ms + network latency (4 round trips)
```

### Bulk Operations

```
Create Event + 3 Sessions: 400ms
Total:                     400ms + network latency (1 round trip)
```

**Savings:** ~40% faster + reduced network overhead

---

## Summary

The event management system provides **7 comprehensive sequence diagrams** covering:

1. ✅ **Individual operations** - For simple, focused workflows
2. ✅ **Bulk operations** - For atomic, efficient batch creation
3. ✅ **Complete workflows** - For end-to-end setup scenarios
4. ✅ **Check-in flows** - For attendance tracking with late detection

**Key Features:**

- Transaction management for data consistency
- Validation at every step
- Rollback on failure
- Optimized for both individual and bulk operations
- Clear error handling

**Best Practices:**

- Use bulk operations for initial setup
- Use individual operations for incremental changes
- Always validate relationships (overlaps, dependencies)
- Leverage transactions for atomicity
