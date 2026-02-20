# Events Table - Index & Optimization Strategy

**Migration File:** `000015_events_setup.up.sql`  
**Date:** February 7, 2026  
**Based on:** Event Management System Specification v2.1

---

## 📊 Index Strategy Overview

This document explains the indexing strategy for the `events` table, designed to optimize common query patterns identified in the event management system specification.

### Index Categories

1. **Primary Lookup Indexes** - Fast single-column lookups
2. **Composite Indexes** - Multi-column queries (most common patterns)
3. **Array Indexes (GIN)** - Access control filtering
4. **JSONB Indexes (GIN)** - Configuration queries
5. **Full-Text Search** - Event search functionality
6. **Partial Indexes** - Optimized for specific conditions

---

## 🎯 Query Pattern Analysis

### Pattern 1: List Published Events (Most Common)

**Query:** Get all published events ordered by date

```sql
SELECT * FROM events
WHERE status = 'published' AND deleted_at IS NULL
ORDER BY start_date DESC;
```

**Index:** `idx_events_status_start_date`

- Composite: `(status, start_date DESC)`
- Partial: `WHERE deleted_at IS NULL`
- **Why:** Covers filtering + sorting in one index

---

### Pattern 2: Filter by Visibility + Status

**Query:** Get public published events

```sql
SELECT * FROM events
WHERE visibility = 'public' AND status = 'published' AND deleted_at IS NULL
ORDER BY start_date DESC;
```

**Index:** `idx_events_visibility_status_start`

- Composite: `(visibility, status, start_date DESC)`
- Partial: `WHERE deleted_at IS NULL`
- **Why:** Common for public event listings

---

### Pattern 3: Filter by Category + Status

**Query:** Get all attendance-type events (Sunday Services)

```sql
SELECT * FROM events
WHERE category = 'attendance' AND status = 'published' AND deleted_at IS NULL
ORDER BY start_date DESC;
```

**Index:** `idx_events_category_status_start`

- Composite: `(category, status, start_date DESC)`
- Partial: `WHERE deleted_at IS NULL`
- **Why:** Category filtering is common (attendance vs registration events)

---

### Pattern 4: Organizer's Event Management

**Query:** Get all events created by a specific organizer

```sql
SELECT * FROM events
WHERE creator_community_id = 'CID123456' AND deleted_at IS NULL
ORDER BY status, start_date DESC;
```

**Index:** `idx_events_creator_status_start`

- Composite: `(creator_community_id, status, start_date DESC)`
- Partial: `WHERE deleted_at IS NULL`
- **Why:** Organizers need to see their events filtered by status

---

### Pattern 5: Date Range Queries

**Query:** Get upcoming events

```sql
SELECT * FROM events
WHERE start_date >= NOW() AND end_date <= '2026-12-31'
  AND status != 'draft' AND deleted_at IS NULL;
```

**Index:** `idx_events_date_range`

- Composite: `(start_date, end_date)`
- Partial: `WHERE deleted_at IS NULL AND status != 'draft'`
- **Why:** Efficient for "upcoming events" and date range filtering

---

### Pattern 6: Recurring Events Management

**Query:** Get all recurring events (Sunday Services)

```sql
SELECT * FROM events
WHERE is_recurring = TRUE AND deleted_at IS NULL
ORDER BY start_date;
```

**Index:** `idx_events_recurring`

- Composite: `(is_recurring, start_date)`
- Partial: `WHERE is_recurring = TRUE AND deleted_at IS NULL`
- **Why:** Small subset, optimized with partial index

---

### Pattern 7: Access Control Filtering

**Query:** Find events allowed for specific user type

```sql
SELECT * FROM events
WHERE 'member' = ANY(allowed_user_types) AND deleted_at IS NULL;
```

**Index:** `idx_events_allowed_user_types` (GIN)

- Type: GIN index on array column
- **Why:** Efficient array containment queries

**Similar indexes for:**

- `allowed_roles` - Role-based filtering
- `allowed_campuses` - Campus-specific events
- `allowed_community_ids` - Private event invitations
- `organizer_community_ids` - Multi-organizer queries

---

### Pattern 8: Full-Text Search

**Query:** Search events by title/description

```sql
SELECT * FROM events
WHERE to_tsvector('indonesian', title || ' ' || description) @@ to_tsquery('indonesian', 'christmas')
  AND deleted_at IS NULL;
```

**Index:** `idx_events_search` (GIN)

- Type: GIN index on tsvector
- Language: Indonesian (configurable)
- **Why:** Fast full-text search without external search engine

---

## 🔒 Data Integrity Constraints

### Enum Validation Constraints

| Constraint                    | Validates        | Valid Values                                                             |
| ----------------------------- | ---------------- | ------------------------------------------------------------------------ |
| `chk_events_category`         | Event category   | `registration`, `attendance`, `announcement`, `volunteer`, `hybrid`      |
| `chk_events_visibility`       | Visibility level | `public`, `members_only`, `volunteer_only`, `private`, `campus_specific` |
| `chk_events_location_type`    | Location type    | `online`, `offline`, `hybrid`                                            |
| `chk_events_virtual_platform` | Virtual platform | `youtube`, `zoom`, `meet`, `custom`, `NULL`                              |
| `chk_events_status`           | Event status     | `draft`, `published`, `cancelled`, `completed`                           |

### Business Logic Constraints

| Constraint                           | Purpose                   | Logic                                           |
| ------------------------------------ | ------------------------- | ----------------------------------------------- |
| `chk_events_dates`                   | Date validation           | `end_date > start_date`                         |
| `chk_events_recurrence_end_date`     | Recurrence logic          | If recurring, end date must be after start date |
| `chk_events_template_not_recurring`  | Template rules            | Templates cannot be recurring                   |
| `chk_events_template_self_reference` | Self-reference prevention | Event cannot reference itself as template       |

---

## 📈 Performance Considerations

### Partial Indexes

Most indexes include `WHERE deleted_at IS NULL` to:

- **Reduce index size** (exclude deleted records)
- **Improve query performance** (smaller index to scan)
- **Match query patterns** (most queries filter out deleted records)

### Index Selectivity

Indexes are ordered by selectivity (most selective first):

- ✅ **Good:** `(creator_community_id, status, start_date)` - High selectivity
- ✅ **Good:** `(visibility, status, start_date)` - Medium-high selectivity
- ⚠️ **Acceptable:** `(status, start_date)` - Medium selectivity (few status values)

### GIN Indexes

Used for:

- **Array columns** - Fast `ANY()` and `@>` operations
- **JSONB columns** - Fast key/value lookups
- **Full-text search** - Fast text search without external tools

### Index Maintenance

- **Auto-vacuum:** PostgreSQL handles automatically
- **Reindex:** Consider quarterly for heavily updated tables
- **Monitor:** Use `pg_stat_user_indexes` to track index usage

---

## 📝 Documentation Strategy

### Table Comment

Comprehensive description of the table's purpose and supported use cases.

### Column Comments

Each column has detailed documentation including:

- **Purpose** - What the column stores
- **Format** - Expected data format/structure
- **Examples** - Sample values
- **Constraints** - Validation rules
- **Relationships** - Foreign key references

### Benefits

1. **Self-documenting database** - New developers understand schema quickly
2. **IDE integration** - Comments appear in database tools
3. **API documentation** - Can be extracted for API docs
4. **Maintenance** - Easier to understand legacy code

---

## 🚀 Query Optimization Tips

### Use Covered Indexes

When possible, select only columns in the index:

```sql
-- ✅ Covered by idx_events_status_start_date
SELECT id, code, title, status, start_date
FROM events
WHERE status = 'published' AND deleted_at IS NULL
ORDER BY start_date DESC;
```

### Avoid Index Bloat

Don't create indexes for:

- ❌ Columns rarely used in WHERE/ORDER BY
- ❌ Very low selectivity columns (e.g., boolean with 99% same value)
- ❌ Columns that change frequently (high write overhead)

### Monitor Index Usage

```sql
-- Check index usage statistics
SELECT
    schemaname, tablename, indexname,
    idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
WHERE tablename = 'events'
ORDER BY idx_scan DESC;
```

---

## 📊 Expected Query Performance

| Query Type            | Without Index                         | With Index        | Improvement     |
| --------------------- | ------------------------------------- | ----------------- | --------------- |
| List published events | Full table scan (~500ms for 10K rows) | Index scan (~5ms) | **100x faster** |
| Filter by visibility  | Full table scan (~500ms)              | Index scan (~3ms) | **166x faster** |
| Search by creator     | Full table scan (~500ms)              | Index scan (~2ms) | **250x faster** |
| Full-text search      | Sequential scan (~1000ms)             | GIN index (~10ms) | **100x faster** |
| Array containment     | Sequential scan (~800ms)              | GIN index (~8ms)  | **100x faster** |

_Performance estimates based on 10,000 events. Actual performance varies by hardware and data distribution._

---

## 🔄 Migration Rollback

If you need to rollback this migration, create `000015_events_setup.down.sql`:

```sql
-- Drop all constraints
ALTER TABLE events DROP CONSTRAINT IF EXISTS chk_events_category;
ALTER TABLE events DROP CONSTRAINT IF EXISTS chk_events_visibility;
ALTER TABLE events DROP CONSTRAINT IF EXISTS chk_events_location_type;
ALTER TABLE events DROP CONSTRAINT IF EXISTS chk_events_virtual_platform;
ALTER TABLE events DROP CONSTRAINT IF EXISTS chk_events_status;
ALTER TABLE events DROP CONSTRAINT IF EXISTS chk_events_dates;
ALTER TABLE events DROP CONSTRAINT IF EXISTS chk_events_recurrence_end_date;
ALTER TABLE events DROP CONSTRAINT IF EXISTS chk_events_template_not_recurring;
ALTER TABLE events DROP CONSTRAINT IF EXISTS chk_events_template_self_reference;

-- Drop all indexes (except those created by PRIMARY KEY and UNIQUE constraints)
DROP INDEX IF EXISTS idx_events_creator;
DROP INDEX IF EXISTS idx_events_series;
DROP INDEX IF EXISTS idx_events_template;
DROP INDEX IF EXISTS idx_events_status_start_date;
DROP INDEX IF EXISTS idx_events_visibility_status_start;
DROP INDEX IF EXISTS idx_events_category_status_start;
DROP INDEX IF EXISTS idx_events_creator_status_start;
DROP INDEX IF EXISTS idx_events_date_range;
DROP INDEX IF EXISTS idx_events_recurring;
DROP INDEX IF EXISTS idx_events_templates;
DROP INDEX IF EXISTS idx_events_deleted_at;
DROP INDEX IF EXISTS idx_events_allowed_user_types;
DROP INDEX IF EXISTS idx_events_allowed_roles;
DROP INDEX IF EXISTS idx_events_allowed_campuses;
DROP INDEX IF EXISTS idx_events_allowed_community_ids;
DROP INDEX IF EXISTS idx_events_organizer_community_ids;
DROP INDEX IF EXISTS idx_events_recurrence_pattern;
DROP INDEX IF EXISTS idx_events_reminder_config;
DROP INDEX IF EXISTS idx_events_search;

-- Drop table
DROP TABLE IF EXISTS events;
```

---

## 📚 References

- [PostgreSQL Index Types](https://www.postgresql.org/docs/current/indexes-types.html)
- [GIN Indexes](https://www.postgresql.org/docs/current/gin.html)
- [Partial Indexes](https://www.postgresql.org/docs/current/indexes-partial.html)
- [Full-Text Search](https://www.postgresql.org/docs/current/textsearch.html)
- Event Management System Specification v2.1

---

**Last Updated:** February 7, 2026  
**Maintained By:** Development Team
