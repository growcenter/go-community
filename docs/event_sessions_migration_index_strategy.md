# Event Sessions Table - Index & Optimization Strategy

**Migration File:** `000008_event_sessions_setup.up.sql`  
**Date:** February 7, 2026  
**Based on:** Event Management System Specification v2.1

---

## 📊 Index Strategy Overview

This document explains the indexing strategy for the `event_sessions` table, which serves as the **unified session model** handling all sub-event types: Sunday Service sessions, Christmas service times, Conference tracks and breakouts, Workshop sessions, and Age-specific sessions.

### Index Categories

1. **Primary Lookup Indexes** - Fast single-column lookups
2. **Composite Indexes** - Multi-column queries (most common patterns)
3. **Hierarchical Indexes** - Parent-child session relationships
4. **Capacity Management Indexes** - Real-time availability queries
5. **Time-Window Indexes** - Registration and check-in windows
6. **Array/JSONB Indexes (GIN)** - Prerequisites and configuration
7. **Full-Text Search** - Session search functionality

---

## 🎯 Query Pattern Analysis

### Pattern 1: List Sessions for an Event (Most Common)

**Query:** Get all sessions for a specific event, ordered by time

```sql
SELECT * FROM event_sessions
WHERE event_id = 123 AND deleted_at IS NULL
ORDER BY start_at ASC;
```

**Index:** `idx_sessions_event_start`

- Composite: `(event_id, start_at ASC)`
- Partial: `WHERE deleted_at IS NULL`
- **Why:** Covers filtering + sorting in one index
- **Use Case:** Display all sessions on event detail page

---

### Pattern 2: List Published Sessions by Event

**Query:** Get published sessions for an event (hide drafts)

```sql
SELECT * FROM event_sessions
WHERE event_id = 123 AND status = 'published' AND deleted_at IS NULL
ORDER BY start_at ASC;
```

**Index:** `idx_sessions_event_status_start`

- Composite: `(event_id, status, start_at ASC)`
- Partial: `WHERE deleted_at IS NULL`
- **Why:** Common for public-facing event pages
- **Use Case:** Show available sessions to users

---

### Pattern 3: Filter by Session Type

**Query:** Get all published kids sessions across all events

```sql
SELECT * FROM event_sessions
WHERE session_type = 'kids' AND status = 'published' AND deleted_at IS NULL
ORDER BY start_at ASC;
```

**Index:** `idx_sessions_type_status_start`

- Composite: `(session_type, status, start_at ASC)`
- Partial: `WHERE deleted_at IS NULL`
- **Why:** Useful for filtering by age group or session category
- **Use Case:** "All Kids Sessions", "All Workshop Sessions"

---

### Pattern 4: Hierarchical Queries (Conference Tracks)

**Query:** Get all child sessions under a parent (e.g., all tracks under "Day 1")

```sql
SELECT * FROM event_sessions
WHERE parent_session_id = 456 AND deleted_at IS NULL
ORDER BY start_at ASC;
```

**Index:** `idx_sessions_parent_start`

- Composite: `(parent_session_id, start_at ASC)`
- Partial: `WHERE parent_session_id IS NOT NULL AND deleted_at IS NULL`
- **Why:** Optimized for hierarchical event structures
- **Use Case:** Conference with Day → Track → Breakout hierarchy

---

### Pattern 5: Find Sessions by Instructor

**Query:** Get all sessions taught by a specific instructor

```sql
SELECT * FROM event_sessions
WHERE instructor_community_id = 'CID123456' AND deleted_at IS NULL
ORDER BY start_at DESC;
```

**Index:** `idx_sessions_instructor_start`

- Composite: `(instructor_community_id, start_at DESC)`
- Partial: `WHERE instructor_community_id IS NOT NULL AND deleted_at IS NULL`
- **Why:** Instructor profile pages, speaker schedules
- **Use Case:** "Sessions by Pastor John", "Speaker Schedule"

---

### Pattern 6: Capacity Management (Critical for Registration)

**Query:** Find sessions with available capacity

```sql
SELECT * FROM event_sessions
WHERE event_id = 123
  AND capacity IS NOT NULL
  AND current_count < capacity
  AND deleted_at IS NULL;
```

**Index:** `idx_sessions_capacity_available`

- Composite: `(event_id, capacity, current_count)`
- Partial: `WHERE capacity IS NOT NULL AND current_count < capacity AND deleted_at IS NULL`
- **Why:** Real-time availability checks during registration
- **Use Case:** Show only sessions with available seats

---

### Pattern 7: Waitlist Management

**Query:** Get sessions with waitlist enabled

```sql
SELECT * FROM event_sessions
WHERE event_id = 123 AND waitlist_enabled = TRUE AND deleted_at IS NULL;
```

**Index:** `idx_sessions_waitlist`

- Composite: `(event_id, waitlist_enabled)`
- Partial: `WHERE waitlist_enabled = TRUE AND deleted_at IS NULL`
- **Why:** Small subset, optimized with partial index
- **Use Case:** Waitlist notification system

---

### Pattern 8: Registration Window Queries

**Query:** Find sessions currently accepting registrations

```sql
SELECT * FROM event_sessions
WHERE registration_start_at <= NOW()
  AND registration_end_at >= NOW()
  AND deleted_at IS NULL;
```

**Index:** `idx_sessions_registration_window`

- Composite: `(registration_start_at, registration_end_at)`
- Partial: `WHERE registration_start_at IS NOT NULL AND deleted_at IS NULL`
- **Why:** Efficient range queries for open registration periods
- **Use Case:** "Register Now" button visibility

---

### Pattern 9: Check-in Window Queries

**Query:** Find sessions currently accepting check-ins

```sql
SELECT * FROM event_sessions
WHERE check_in_start_at <= NOW()
  AND check_in_end_at >= NOW()
  AND deleted_at IS NULL;
```

**Index:** `idx_sessions_checkin_window`

- Composite: `(check_in_start_at, check_in_end_at)`
- Partial: `WHERE check_in_start_at IS NOT NULL AND deleted_at IS NULL`
- **Why:** Real-time check-in availability
- **Use Case:** QR scanner app - show which sessions accept check-ins now

---

### Pattern 10: Age-Restricted Sessions

**Query:** Find sessions suitable for a specific age

```sql
SELECT * FROM event_sessions
WHERE session_type IN ('kids', 'youth', 'teen')
  AND (min_age IS NULL OR min_age <= 10)
  AND (max_age IS NULL OR max_age >= 10)
  AND deleted_at IS NULL;
```

**Index:** `idx_sessions_age_restricted`

- Composite: `(session_type, min_age, max_age)`
- Partial: `WHERE min_age IS NOT NULL OR max_age IS NOT NULL`
- **Why:** Age eligibility checks during registration
- **Use Case:** "Sessions for 10-year-old", "Kids Sessions"

---

### Pattern 11: Prerequisites Check

**Query:** Find sessions with specific prerequisites

```sql
SELECT * FROM event_sessions
WHERE 'Level 1 Completed' = ANY(prerequisites) AND deleted_at IS NULL;
```

**Index:** `idx_sessions_prerequisites` (GIN)

- Type: GIN index on array column
- **Why:** Efficient array containment queries
- **Use Case:** Multi-level training programs, course sequences

---

### Pattern 12: Full-Text Search

**Query:** Search sessions by title/description

```sql
SELECT * FROM event_sessions
WHERE to_tsvector('indonesian', title || ' ' || description) @@ to_tsquery('indonesian', 'leadership')
  AND deleted_at IS NULL;
```

**Index:** `idx_sessions_search` (GIN)

- Type: GIN index on tsvector
- Language: Indonesian (configurable)
- **Why:** Fast full-text search
- **Use Case:** Session search functionality

---

## 🔒 Data Integrity Constraints

### Enum Validation Constraints

| Constraint                          | Validates              | Valid Values                                                                                     |
| ----------------------------------- | ---------------------- | ------------------------------------------------------------------------------------------------ |
| `chk_sessions_type`                 | Session type           | `service`, `class`, `track`, `breakout`, `workshop`, `general`, `kids`, `youth`, `teen`, `adult` |
| `chk_sessions_location_type`        | Location type override | `online`, `offline`, `hybrid`, `NULL`                                                            |
| `chk_sessions_virtual_platform`     | Virtual platform       | `youtube`, `zoom`, `meet`, `custom`, `NULL`                                                      |
| `chk_sessions_status`               | Session status         | `draft`, `published`, `cancelled`, `completed`, `full`                                           |
| `chk_sessions_registration_mode`    | Registration mode      | `self_only`, `self_and_registered`, `self_and_others`                                            |
| `chk_sessions_additional_form_mode` | Additional form mode   | `same_as_primary`, `name_only`, `custom`                                                         |

### Business Logic Constraints

| Constraint                           | Purpose                   | Logic                                            |
| ------------------------------------ | ------------------------- | ------------------------------------------------ |
| `chk_sessions_dates`                 | Date validation           | `end_at > start_at`                              |
| `chk_sessions_registration_window`   | Registration window logic | Both start and end must be set, end > start      |
| `chk_sessions_checkin_window`        | Check-in window logic     | Both start and end must be set, end > start      |
| `chk_sessions_capacity`              | Capacity validation       | Must be positive or NULL                         |
| `chk_sessions_current_count`         | Count validation          | Must be >= 0 and <= capacity + waitlist_capacity |
| `chk_sessions_waitlist_capacity`     | Waitlist validation       | Must be >= 0 or NULL                             |
| `chk_sessions_max_registrations`     | Max registrations         | Must be > 0                                      |
| `chk_sessions_age_range`             | Age range logic           | max_age > min_age, both >= 0                     |
| `chk_sessions_custom_form_reference` | Form reference            | If mode = 'custom', form_id must be provided     |
| `chk_sessions_self_reference`        | Self-reference prevention | Session cannot be its own parent                 |

---

## 📈 Performance Considerations

### Partial Indexes

Most indexes include `WHERE deleted_at IS NULL` to:

- **Reduce index size** (exclude deleted records)
- **Improve query performance** (smaller index to scan)
- **Match query patterns** (most queries filter out deleted records)

### Specialized Partial Indexes

1. **Capacity Available Index**
   - `WHERE capacity IS NOT NULL AND current_count < capacity AND deleted_at IS NULL`
   - Only indexes sessions with available capacity
   - Critical for registration performance

2. **Waitlist Index**
   - `WHERE waitlist_enabled = TRUE AND deleted_at IS NULL`
   - Very small subset (few sessions have waitlist)
   - Extremely fast queries

3. **Parent Session Index**
   - `WHERE parent_session_id IS NOT NULL AND deleted_at IS NULL`
   - Only indexes child sessions
   - Optimizes hierarchical queries

### Denormalized Columns

**`current_count`** - Denormalized for performance

- Updated via triggers or application logic when registrations change
- Avoids expensive COUNT(\*) queries on registrations table
- Enables fast capacity checks: `current_count < capacity`

### GIN Indexes

Used for:

- **Array columns** (`prerequisites`) - Fast `ANY()` and `@>` operations
- **JSONB columns** (`identifier_config`) - Fast key/value lookups
- **Full-text search** - Fast text search without external tools

---

## 🏗️ Hierarchical Structure Support

### Parent-Child Relationships

The `parent_session_id` enables complex hierarchical events:

```
Event: Leadership Conference 2026
├── Session: Day 1 (parent_session_id = NULL)
│   ├── Session: Track A - Leadership (parent_session_id = Day 1 ID)
│   │   ├── Session: Breakout 1A (parent_session_id = Track A ID)
│   │   └── Session: Breakout 1B (parent_session_id = Track A ID)
│   └── Session: Track B - Communication (parent_session_id = Day 1 ID)
└── Session: Day 2 (parent_session_id = NULL)
    └── Session: Track C - Worship (parent_session_id = Day 2 ID)
```

### Query Examples

**Get all top-level sessions:**

```sql
SELECT * FROM event_sessions
WHERE event_id = 123 AND parent_session_id IS NULL AND deleted_at IS NULL;
```

**Get all child sessions recursively:**

```sql
WITH RECURSIVE session_tree AS (
  -- Base case: top-level sessions
  SELECT * FROM event_sessions WHERE event_id = 123 AND parent_session_id IS NULL
  UNION ALL
  -- Recursive case: child sessions
  SELECT s.* FROM event_sessions s
  INNER JOIN session_tree st ON s.parent_session_id = st.id
)
SELECT * FROM session_tree WHERE deleted_at IS NULL;
```

---

## 📊 Expected Query Performance

| Query Type               | Without Index                        | With Index           | Improvement     |
| ------------------------ | ------------------------------------ | -------------------- | --------------- |
| List sessions by event   | Full table scan (~300ms for 5K rows) | Index scan (~3ms)    | **100x faster** |
| Filter by event + status | Full table scan (~300ms)             | Index scan (~2ms)    | **150x faster** |
| Hierarchical queries     | Full table scan (~300ms)             | Index scan (~1ms)    | **300x faster** |
| Capacity checks          | Full table scan (~300ms)             | Partial index (~1ms) | **300x faster** |
| Instructor lookup        | Full table scan (~300ms)             | Index scan (~2ms)    | **150x faster** |
| Full-text search         | Sequential scan (~500ms)             | GIN index (~5ms)     | **100x faster** |
| Prerequisites check      | Sequential scan (~400ms)             | GIN index (~4ms)     | **100x faster** |

_Performance estimates based on 5,000 sessions. Actual performance varies by hardware and data distribution._

---

## 🔄 Capacity Management Strategy

### Real-Time Capacity Tracking

The `current_count` column is denormalized for performance:

**Update on Registration:**

```sql
UPDATE event_sessions
SET current_count = current_count + 1,
    status = CASE
      WHEN current_count + 1 >= capacity THEN 'full'
      ELSE status
    END
WHERE id = :session_id;
```

**Update on Cancellation:**

```sql
UPDATE event_sessions
SET current_count = current_count - 1,
    status = CASE
      WHEN status = 'full' AND current_count - 1 < capacity THEN 'published'
      ELSE status
    END
WHERE id = :session_id;
```

### Waitlist Logic

When `capacity` is reached:

1. Check if `waitlist_enabled = TRUE`
2. If yes, allow registration with `status = 'waitlisted'`
3. Track waitlist position in `registrations.waitlist_position`
4. When someone cancels, promote first waitlisted registrant

---

## 📝 Registration Configuration Examples

### Example 1: Christmas Service (Family Registration)

```json
{
  "registration_mode": "self_and_others",
  "max_registrations_per_user": 5,
  "one_session_per_event": true,
  "additional_registrant_form_mode": "name_only",
  "identifier_config": {
    "primary": {
      "name": { "visible": true, "required": true },
      "email": { "visible": true, "required": true },
      "phone": { "visible": true, "required": false }
    },
    "additional": {
      "name": { "visible": true, "required": true }
    }
  }
}
```

### Example 2: Volunteer Meeting (Self-Only)

```json
{
  "registration_mode": "self_only",
  "max_registrations_per_user": 1,
  "one_session_per_event": true,
  "require_approval": true,
  "identifier_config": {
    "primary": {
      "name": { "visible": true, "required": true },
      "email": { "visible": true, "required": true },
      "phone": { "visible": true, "required": true }
    }
  }
}
```

### Example 3: Conference Track (Registered Users Only)

```json
{
  "registration_mode": "self_and_registered",
  "max_registrations_per_user": 3,
  "one_session_per_event": false,
  "additional_registrant_form_mode": "same_as_primary",
  "prerequisites": ["Completed Level 1", "Member for 6 months"],
  "identifier_config": {
    "primary": {
      "name": { "visible": true, "required": true },
      "email": { "visible": true, "required": true },
      "phone": { "visible": true, "required": true }
    },
    "additional": {
      "name": { "visible": true, "required": true },
      "email": { "visible": true, "required": true }
    }
  }
}
```

---

## 🔄 Migration Rollback

If you need to rollback this migration, create `000008_event_sessions_setup.down.sql`:

```sql
-- Drop all constraints
ALTER TABLE event_sessions DROP CONSTRAINT IF EXISTS chk_sessions_type;
ALTER TABLE event_sessions DROP CONSTRAINT IF EXISTS chk_sessions_location_type;
ALTER TABLE event_sessions DROP CONSTRAINT IF EXISTS chk_sessions_virtual_platform;
ALTER TABLE event_sessions DROP CONSTRAINT IF EXISTS chk_sessions_status;
ALTER TABLE event_sessions DROP CONSTRAINT IF EXISTS chk_sessions_registration_mode;
ALTER TABLE event_sessions DROP CONSTRAINT IF EXISTS chk_sessions_additional_form_mode;
ALTER TABLE event_sessions DROP CONSTRAINT IF EXISTS chk_sessions_dates;
ALTER TABLE event_sessions DROP CONSTRAINT IF EXISTS chk_sessions_registration_window;
ALTER TABLE event_sessions DROP CONSTRAINT IF EXISTS chk_sessions_checkin_window;
ALTER TABLE event_sessions DROP CONSTRAINT IF EXISTS chk_sessions_capacity;
ALTER TABLE event_sessions DROP CONSTRAINT IF EXISTS chk_sessions_current_count;
ALTER TABLE event_sessions DROP CONSTRAINT IF EXISTS chk_sessions_waitlist_capacity;
ALTER TABLE event_sessions DROP CONSTRAINT IF EXISTS chk_sessions_max_registrations;
ALTER TABLE event_sessions DROP CONSTRAINT IF EXISTS chk_sessions_age_range;
ALTER TABLE event_sessions DROP CONSTRAINT IF EXISTS chk_sessions_custom_form_reference;
ALTER TABLE event_sessions DROP CONSTRAINT IF EXISTS chk_sessions_self_reference;

-- Drop all indexes
DROP INDEX IF EXISTS idx_sessions_event;
DROP INDEX IF EXISTS idx_sessions_parent;
DROP INDEX IF EXISTS idx_sessions_instructor;
DROP INDEX IF EXISTS idx_sessions_registration_form;
DROP INDEX IF EXISTS idx_sessions_additional_form;
DROP INDEX IF EXISTS idx_sessions_event_start;
DROP INDEX IF EXISTS idx_sessions_event_status_start;
DROP INDEX IF EXISTS idx_sessions_type_status_start;
DROP INDEX IF EXISTS idx_sessions_parent_start;
DROP INDEX IF EXISTS idx_sessions_instructor_start;
DROP INDEX IF EXISTS idx_sessions_capacity_available;
DROP INDEX IF EXISTS idx_sessions_waitlist;
DROP INDEX IF EXISTS idx_sessions_date_range;
DROP INDEX IF EXISTS idx_sessions_registration_window;
DROP INDEX IF EXISTS idx_sessions_checkin_window;
DROP INDEX IF EXISTS idx_sessions_age_restricted;
DROP INDEX IF EXISTS idx_sessions_deleted_at;
DROP INDEX IF EXISTS idx_sessions_prerequisites;
DROP INDEX IF EXISTS idx_sessions_identifier_config;
DROP INDEX IF EXISTS idx_sessions_search;

-- Drop table
DROP TABLE IF EXISTS event_sessions;
```

---

## 📚 References

- [PostgreSQL Partial Indexes](https://www.postgresql.org/docs/current/indexes-partial.html)
- [PostgreSQL Recursive Queries](https://www.postgresql.org/docs/current/queries-with.html)
- [GIN Indexes for Arrays](https://www.postgresql.org/docs/current/gin.html)
- Event Management System Specification v2.1

---

**Last Updated:** February 7, 2026  
**Maintained By:** Development Team
