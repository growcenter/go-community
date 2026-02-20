# Event Management System - Database Schema

**Version:** 3.0  
**Last Updated:** February 2026  
**Related Documents:** [Event Management System Specification](./event_management_system_spec.md)

---

## Overview

This document contains the complete database schema for the Event Management System, including all tables, indexes, constraints, triggers, and helper functions.

### Key Features

- **Production-Ready**: Foreign keys, comprehensive indexes, ENUM types
- **Flexible Forms**: Normalized form system with context-aware filtering
- **Polymorphic Associations**: Many-to-many relationships via `form_associations`
- **Soft Deletes**: `deleted_at` columns with appropriate indexes
- **Auto-Update Triggers**: Automatic `updated_at` timestamp management

### Schema Organization

1. **Core Event Tables** - Events and sessions
2. **Form System** - Production-ready normalized forms
3. **Registration & Attendance** - User registrations and check-in/out
4. **Supporting Tables** - Series, volunteers, etc.
5. **Migration Notes** - Form schema migration strategy

---

## 1. Core Event Tables

#### `events` - Main Event Container

```sql
CREATE TABLE events (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,       -- matches varchar(50) in model

    -- Core Info
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,      -- slugified, URL-safe
    pre_description TEXT,                   -- shown before registration
    post_description JSONB,                 -- shown after registration (rich content)
    terms_and_conditions TEXT,
    category VARCHAR(30) NOT NULL,          -- announcement, registerable

    -- Media
    image_urls TEXT[],                      -- array of image URLs
    banner_url TEXT,

    -- Organization
    creator_community_id VARCHAR(50) NOT NULL,
    organizer_community_ids TEXT[],
    contact_community_ids TEXT[],

    -- Visibility & Access
    access_level VARCHAR(20) NOT NULL DEFAULT 'public',  -- public, private, members_only, etc.
    allowed_user_types TEXT[],
    allowed_roles TEXT[],
    allowed_campuses TEXT[],
    allowed_community_ids TEXT[],

    -- Location (defaults inherited by sessions)
    location_type VARCHAR(20) NOT NULL,     -- online, offline, hybrid
    physical_place_name TEXT,               -- venue name
    physical_address TEXT,
    virtual_link TEXT,
    virtual_platform TEXT,                  -- youtube, zoom, meet, etc.
    location_details TEXT,
    location_visibility VARCHAR(30) NOT NULL DEFAULT 'all', -- all, pre-registration, post-registration
    cta_text VARCHAR(100),                  -- call-to-action button label
    cta_link TEXT,                          -- CTA target URL or "NORMAL_FLOW"

    -- Schedule
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    timezone VARCHAR(50) NOT NULL DEFAULT 'Asia/Jakarta',

    -- Recurrence (occurrences calculated on-demand, never pre-generated)
    is_recurring BOOLEAN DEFAULT FALSE,
    recurrence_pattern JSONB,
    -- Structure: { frequency, interval, weekDays, count, endDate,
    --              monthlyPattern, excludeDates, additionalDates }

    -- Template / Series
    is_template BOOLEAN DEFAULT FALSE,
    template_id VARCHAR(255),               -- no FK — external or soft reference
    series_id   VARCHAR(255),               -- no FK — external or soft reference

    -- Registration
    session_per_user INTEGER DEFAULT 0,     -- 0 = unlimited

    -- Notifications
    notification_channels TEXT[],           -- email, sms, whatsapp, push
    reminder_config JSONB,                  -- { enabled, intervals: ["24h","1h"] }

    -- Metadata
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Composite index used by event listing queries
CREATE INDEX idx_events_status_start ON events(status, start_at);
CREATE INDEX idx_events_access_level ON events(access_level);
CREATE INDEX idx_events_category ON events(category);
CREATE INDEX idx_events_creator ON events(creator_community_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_events_is_recurring ON events(is_recurring) WHERE is_recurring = TRUE AND deleted_at IS NULL;
```



### Recurrence Pattern Calculation

For recurring events, occurrence dates are calculated on-demand using the `recurrence_pattern` JSONB field in the `events` table. This field contains:

```json
{
  "frequency": "weekly|daily|monthly|yearly",
  "interval": 1,
  "weekDays": ["sunday", "monday", ...],
  "monthlyPattern": "day_of_month|nth_weekday",
  "dayOfMonth": 15,
  "nthWeekday": "first|second|third|fourth|last",
  "weekday": "monday|tuesday|...",
  "count": 12,
  "endDate": "2026-12-31T00:00:00Z",
  "excludeDates": ["2026-03-15", "2026-04-20"],
  "additionalDates": ["2026-05-01"]
}
```

**Calculation Logic:**

- When querying events for a date range (e.g., "events this week"), the backend calculates which occurrences fall within that range
- Calculation is done in the application layer using the `RecurrencePattern.CalculateOccurrences()` method
- Results can be cached per query for performance

---

## 2. Universal Form System (Production-Ready)

The form system is **domain-agnostic** and can be used across events, volunteer applications, surveys, and any future use case requiring custom data collection.

> [!IMPORTANT]
> This schema uses **normalized tables** with proper foreign keys, indexes, and ENUM types for production readiness. It supports **polymorphic associations** via `form_associations` and **context-aware question filtering** via `mandatory_for`/`apply_for` arrays.

#### ENUM Types (Type Safety)

```sql
CREATE TYPE form_status_enum AS ENUM ('draft', 'active', 'archived', 'template');

CREATE TYPE question_type_enum AS ENUM (
    'short_text',
    'long_text',
    'number',
    'email',
    'phone',
    'single_choice',
    'multiple_choice',
    'date',
    'time'
);

CREATE TYPE entity_type_enum AS ENUM (
    'event',
    'event_session'
    -- Future: 'user', 'organization', 'volunteer_application', 'registration'
);

CREATE TYPE identifier_type_enum AS ENUM (
    'community_id',      -- Authenticated community member
    'eventAttendance'    -- Walk-in / attendance-only check-in
    -- Future: 'registration_code', 'email', 'phone', 'session_code', 'event_code'
);
```

#### `forms` - Form definitions

```sql
CREATE TABLE forms (
    id BIGSERIAL PRIMARY KEY,
    code UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    form_type VARCHAR(50), -- registration, survey, quiz
    status form_status_enum NOT NULL DEFAULT 'draft',
    is_template BOOLEAN DEFAULT FALSE,
    
    -- Audit fields
    created_by_community_id VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- Constraints
    CONSTRAINT forms_name_not_empty CHECK (LENGTH(TRIM(name)) > 0)
);

-- Indexes
CREATE INDEX idx_forms_status ON forms(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_forms_type ON forms(form_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_forms_template ON forms(is_template) WHERE is_template = TRUE AND deleted_at IS NULL;
CREATE INDEX idx_forms_created_at ON forms(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_forms_creator ON forms(created_by_community_id) WHERE deleted_at IS NULL;
```

#### `form_questions` - Individual questions/fields

> [!TIP]
> Use `mandatory_for` and `apply_for` arrays to create context-aware questions that adapt based on who's filling out the form (primary registrant, additional registrant, kids session, etc.)

```sql
CREATE TABLE form_questions (
    id BIGSERIAL PRIMARY KEY,
    code UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    form_code UUID NOT NULL,
    
    -- Question content
    text TEXT NOT NULL,
    type question_type_enum NOT NULL,
    
    -- Context-aware filtering
    -- mandatory_for: contexts in which answering this question is required.
    --   Valid values: 'parent' (primary registrant), 'child' (additional registrant)
    mandatory_for TEXT[] DEFAULT ARRAY[]::TEXT[],
    -- apply_for: contexts in which this question is shown.
    --   Valid values: 'parent', 'child'
    apply_for TEXT[] DEFAULT ARRAY[]::TEXT[],
    
    -- Options for single_choice / multiple_choice questions
    -- Structure: {"choices": ["Option A", "Option B", ...]}
    options JSONB,
    
    -- Validation rules
    rules JSONB, -- {"min_length": 5, "max_length": 100, "pattern": "^[A-Z]"}
    
    -- For quiz/assessment questions
    correct_answer TEXT,
    
    -- Display
    display_order INT NOT NULL,
    
    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- Foreign keys
    CONSTRAINT fk_form_questions_form 
        FOREIGN KEY (form_code) 
        REFERENCES forms(code) 
        ON DELETE CASCADE,
    
    -- Constraints
    CONSTRAINT form_questions_text_not_empty CHECK (LENGTH(TRIM(text)) > 0),
    CONSTRAINT form_questions_display_order_positive CHECK (display_order >= 0),
    CONSTRAINT form_questions_options_required_for_select 
        CHECK (
            type NOT IN ('select', 'multiselect', 'radio', 'checkbox') 
            OR options IS NOT NULL
        )
);

-- Indexes
CREATE INDEX idx_form_questions_form ON form_questions(form_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_form_questions_type ON form_questions(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_form_questions_display_order ON form_questions(form_code, display_order) WHERE deleted_at IS NULL;
CREATE INDEX idx_form_questions_apply_for ON form_questions USING GIN(apply_for) WHERE deleted_at IS NULL;
CREATE INDEX idx_form_questions_mandatory_for ON form_questions USING GIN(mandatory_for) WHERE deleted_at IS NULL;
```

#### `form_answers` - Individual answers (flexible identifier support)

> [!NOTE]
> This table uses a flexible `identifier_type` + `identifier` pattern to support both authenticated users (via `community_id`) and guests (via `registration_code`, `email`, etc.)

```sql
CREATE TABLE form_answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    form_code UUID,
    question_code UUID NOT NULL,
    
    -- Flexible identifier (supports both authenticated users and guests)
    identifier_type identifier_type_enum NOT NULL,
    identifier VARCHAR(255) NOT NULL,
    
    -- Answer
    answer TEXT NOT NULL,
    
    -- For quiz/assessment
    is_correct BOOLEAN,
    
    -- Audit fields
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- Foreign keys
    CONSTRAINT fk_form_answers_form
        FOREIGN KEY (form_code)
        REFERENCES forms(code)
        ON DELETE SET NULL, -- Keep answers even if form is deleted
    
    CONSTRAINT fk_form_answers_question
        FOREIGN KEY (question_code)
        REFERENCES form_questions(code)
        ON DELETE CASCADE,
    
    -- Constraints
    CONSTRAINT form_answers_answer_not_empty CHECK (LENGTH(TRIM(answer)) > 0),
    
    -- Prevent duplicate answers (one answer per question per identifier)
    CONSTRAINT unique_answer_per_question_per_identifier
        UNIQUE (question_code, identifier_type, identifier, deleted_at)
);

-- Indexes
CREATE INDEX idx_form_answers_question ON form_answers(question_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_form_answers_form ON form_answers(form_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_form_answers_identifier ON form_answers(identifier_type, identifier) WHERE deleted_at IS NULL;
CREATE INDEX idx_form_answers_submitted_at ON form_answers(submitted_at DESC) WHERE deleted_at IS NULL;
```

#### `form_associations` - Polymorphic associations (NEW!)

> [!TIP]
> This table enables **many-to-many relationships** between forms and any entity type (events, sessions, users, etc.). One form can be associated with multiple entities, and one entity can have multiple forms.

```sql
CREATE TABLE form_associations (
    id BIGSERIAL PRIMARY KEY,
    form_code UUID NOT NULL,
    entity_type entity_type_enum NOT NULL,
    entity_code VARCHAR(255) NOT NULL,
    
    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- Foreign keys
    CONSTRAINT fk_form_associations_form
        FOREIGN KEY (form_code)
        REFERENCES forms(code)
        ON DELETE CASCADE,
    
    -- Constraints
    CONSTRAINT form_associations_entity_code_not_empty 
        CHECK (LENGTH(TRIM(entity_code)) > 0),
    
    -- Prevent duplicate associations
    CONSTRAINT unique_form_entity_association
        UNIQUE (form_code, entity_type, entity_code, deleted_at)
);

-- Indexes
CREATE INDEX idx_form_associations_form ON form_associations(form_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_form_associations_entity ON form_associations(entity_type, entity_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_form_associations_entity_type ON form_associations(entity_type) WHERE deleted_at IS NULL;
```

#### Auto-Update Triggers

```sql
-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply to all form tables
CREATE TRIGGER trg_forms_updated_at
    BEFORE UPDATE ON forms
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_form_questions_updated_at
    BEFORE UPDATE ON form_questions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_form_answers_updated_at
    BEFORE UPDATE ON form_answers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_form_associations_updated_at
    BEFORE UPDATE ON form_associations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

#### Helper Functions

**Clone form from template:**

```sql
CREATE OR REPLACE FUNCTION clone_form_from_template(
    p_template_code UUID,
    p_new_name VARCHAR,
    p_created_by VARCHAR DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    v_new_form_code UUID;
BEGIN
    -- Create new form from template
    INSERT INTO forms (code, name, description, form_type, status, created_by_community_id)
    SELECT 
        gen_random_uuid(),
        p_new_name,
        description,
        form_type,
        'draft',
        p_created_by
    FROM forms
    WHERE code = p_template_code
      AND is_template = TRUE
      AND deleted_at IS NULL
    RETURNING code INTO v_new_form_code;
    
    IF v_new_form_code IS NULL THEN
        RAISE EXCEPTION 'Template form not found: %', p_template_code;
    END IF;
    
    -- Clone all questions
    INSERT INTO form_questions (
        code, form_code, text, type, mandatory_for, apply_for,
        options, rules, correct_answer, display_order
    )
    SELECT 
        gen_random_uuid(),
        v_new_form_code,
        text,
        type,
        mandatory_for,
        apply_for,
        options,
        rules,
        correct_answer,
        display_order
    FROM form_questions
    WHERE form_code = p_template_code
      AND deleted_at IS NULL;
    
    RETURN v_new_form_code;
END;
$$ LANGUAGE plpgsql;
```

**Get questions for context:**

```sql
CREATE OR REPLACE FUNCTION get_form_questions_for_context(
    p_form_code UUID,
    p_context TEXT DEFAULT 'all'
) RETURNS TABLE (
    question_code UUID,
    text TEXT,
    type question_type_enum,
    is_mandatory BOOLEAN,
    options JSONB,
    rules JSONB,
    display_order INT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        fq.code,
        fq.text,
        fq.type,
        (p_context = ANY(fq.mandatory_for)) AS is_mandatory,
        fq.options,
        fq.rules,
        fq.display_order
    FROM form_questions fq
    WHERE fq.form_code = p_form_code
      AND fq.deleted_at IS NULL
      AND (p_context = ANY(fq.apply_for) OR 'all' = ANY(fq.apply_for))
    ORDER BY fq.display_order;
END;
$$ LANGUAGE plpgsql;
```

#### Form System Usage Examples

**1. Creating a form with context-aware questions:**

```sql
-- Create form
INSERT INTO forms (code, name, form_type, status, created_by_community_id)
VALUES (
    gen_random_uuid(),
    'Christmas Event Registration',
    'event_registration',
    'active',
    'admin123'
) RETURNING code; -- Returns form_code

-- Context values: 'parent' = primary registrant, 'child' = additional registrant
INSERT INTO form_questions (code, form_code, text, type, mandatory_for, apply_for, display_order)
VALUES
  -- Name: required for parent only
  (gen_random_uuid(), form_code, 'Full Name', 'short_text',
   ARRAY['parent'], ARRAY['parent'], 1),
  
  -- Email: required for parent, visible for both
  (gen_random_uuid(), form_code, 'Email Address', 'email',
   ARRAY['parent'], ARRAY['parent', 'child'], 2),
  
  -- Dietary: optional for parent
  (gen_random_uuid(), form_code, 'Dietary Restrictions', 'multiple_choice',
   ARRAY[]::TEXT[], ARRAY['parent'], 3);

-- Associate with event
INSERT INTO form_associations (form_code, entity_type, entity_code)
VALUES (form_code, 'event', 'EVT-CHRISTMAS-2024');
```

**2. Using polymorphic associations:**

```sql
-- One form, multiple events
INSERT INTO form_associations (form_code, entity_type, entity_code)
VALUES 
  ('form-uuid-123', 'event', 'EVT-CHRISTMAS-2024'),
  ('form-uuid-123', 'event', 'EVT-EASTER-2024');

-- Multiple forms, one event
INSERT INTO form_associations (form_code, entity_type, entity_code)
VALUES
  ('form-registration-uuid', 'event', 'EVT-CONF-2024'),
  ('form-dietary-uuid', 'event', 'EVT-CONF-2024'),
  ('form-workshop-uuid', 'event', 'EVT-CONF-2024');
```

**3. Querying answers with flexible identifiers:**

```sql
-- Get answers from registered user
SELECT fq.text, fa.answer
FROM form_answers fa
JOIN form_questions fq ON fq.code = fa.question_code
WHERE fa.identifier_type = 'community_id'
  AND fa.identifier = 'CID123456'
  AND fa.deleted_at IS NULL;

-- Get answers from guest (by registration code)
SELECT fq.text, fa.answer
FROM form_answers fa
JOIN form_questions fq ON fq.code = fa.question_code
WHERE fa.identifier_type = 'registration_code'
  AND fa.identifier = 'REG-ABC123'
  AND fa.deleted_at IS NULL;
```

**4. Clone template and customize:**

```sql
-- Clone template
SELECT clone_form_from_template(
    'template-form-uuid',
    'My Custom Event Form',
    'admin123'
) AS new_form_code;

-- Get questions for specific context
SELECT * FROM get_form_questions_for_context(
    'form-uuid',
    'primary' -- Only show questions for primary registrant
);
```

## 3. Registration & Attendance Tables

#### `registrations` - User registrations

```sql
CREATE TABLE registrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(30) UNIQUE NOT NULL, -- Ticket code for QR

    -- Event Reference
    event_id BIGINT NOT NULL REFERENCES events(id),
    session_id BIGINT REFERENCES event_sessions(id), -- Can be parent or child session
    occurrence_id BIGINT REFERENCES event_occurrences(id),

    -- Registrant Info
    registrant_community_id VARCHAR(50), -- NULL for guests
    registrant_name VARCHAR(255) NOT NULL,
    registrant_email VARCHAR(255),
    registrant_phone VARCHAR(50),
    registrant_legal_id VARCHAR(100),

    -- Registration Source
    registration_method VARCHAR(30) NOT NULL, -- personal-qr, event-qr, session-qr, manual, web
    registered_by_community_id VARCHAR(50), -- Who registered (for family/group)

    -- For Family/Group Registration
    is_primary_registrant BOOLEAN DEFAULT TRUE,
    group_registration_id UUID REFERENCES registrations(id),

    -- Custom Form Submission
    form_submission_id UUID REFERENCES form_submissions(id),

    -- Status Flow
    status VARCHAR(30) NOT NULL DEFAULT 'pending', -- pending, confirmed, waitlisted, cancelled, expired
    waitlist_position INTEGER,

    -- Timestamps
    registered_at TIMESTAMPTZ DEFAULT NOW(),
    confirmed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    cancellation_reason TEXT,

    -- Notification Tracking
    confirmation_sent_at TIMESTAMPTZ,
    confirmation_channel VARCHAR(20), -- email, sms, whatsapp
    reminder_sent_at TIMESTAMPTZ,

    -- Geolocation Tracking (captured at registration time)
    registration_latitude DECIMAL(10, 8),
    registration_longitude DECIMAL(11, 8),
    registration_location_accuracy DECIMAL(10, 2), -- in meters
    registration_distance_from_venue DECIMAL(10, 2), -- calculated distance in meters
    geolocation_validation_passed BOOLEAN,
    geolocation_override_by VARCHAR(50), -- community_id of staff who overrode validation

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_registrations_event ON registrations(event_id);
CREATE INDEX idx_registrations_session ON registrations(session_id);
CREATE INDEX idx_registrations_registrant ON registrations(registrant_community_id);
CREATE INDEX idx_registrations_status ON registrations(status);
CREATE INDEX idx_registrations_code ON registrations(code);
```

#### `attendance_records` - Check-in/Check-out tracking

```sql
CREATE TABLE attendance_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Registration Reference
    registration_id UUID REFERENCES registrations(id),

    -- Or Direct Attendance (for walk-ins or attendance-only events)
    event_id BIGINT NOT NULL REFERENCES events(id),
    session_id BIGINT REFERENCES event_sessions(id), -- Can be parent or child session
    occurrence_id BIGINT REFERENCES event_occurrences(id),

    -- Attendee Info
    attendee_community_id VARCHAR(50),
    attendee_name VARCHAR(255) NOT NULL,

    -- Check-in Details
    check_in_at TIMESTAMPTZ,
    check_in_method VARCHAR(30), -- personal-qr, ticket-qr, manual
    check_in_by_community_id VARCHAR(50), -- Staff who scanned
    check_in_device_info JSONB,

    -- Check-out Details
    check_out_at TIMESTAMPTZ,
    check_out_method VARCHAR(30),
    check_out_by_community_id VARCHAR(50),

    -- Attendance Status
    status VARCHAR(30) NOT NULL, -- present, late, absent, excused, permit
    late_minutes INTEGER,
    excuse_reason TEXT,

    -- Geolocation Tracking (captured at check-in/check-out)
    check_in_latitude DECIMAL(10, 8),
    check_in_longitude DECIMAL(11, 8),
    check_in_location_accuracy DECIMAL(10, 2), -- in meters
    check_in_distance_from_venue DECIMAL(10, 2), -- calculated distance in meters
    check_in_geolocation_passed BOOLEAN,
    check_in_geolocation_override_by VARCHAR(50), -- community_id of staff who overrode

    check_out_latitude DECIMAL(10, 8),
    check_out_longitude DECIMAL(11, 8),
    check_out_location_accuracy DECIMAL(10, 2),
    check_out_distance_from_venue DECIMAL(10, 2),

    -- Metadata
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_attendance_registration ON attendance_records(registration_id);
CREATE INDEX idx_attendance_event ON attendance_records(event_id);
CREATE INDEX idx_attendance_occurrence ON attendance_records(occurrence_id);
CREATE INDEX idx_attendance_attendee ON attendance_records(attendee_community_id);
CREATE INDEX idx_attendance_check_in_at ON attendance_records(check_in_at);
```

## 3. Registration & Attendance Tables

## 4. Supporting Tables

#### `event_series` - For grouping related events

```sql
CREATE TABLE event_series (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(20) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    image_url TEXT,
    created_by_community_id VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### `event_volunteers` - Volunteer assignments per event

```sql
CREATE TABLE event_volunteers (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    session_id BIGINT REFERENCES event_sessions(id),
    community_id VARCHAR(50) NOT NULL,
    role VARCHAR(50) NOT NULL, -- scanner, usher, coordinator
    assigned_at TIMESTAMPTZ DEFAULT NOW(),
    assigned_by_community_id VARCHAR(50) NOT NULL,

    UNIQUE(event_id, session_id, community_id)
);

CREATE INDEX idx_event_volunteers_event ON event_volunteers(event_id);
CREATE INDEX idx_event_volunteers_community ON event_volunteers(community_id);
```

#### QR Code Strategy (Frontend-Generated)

QR codes are generated by the frontend using the data provided by the backend:

| QR Type     | Data Format                             | API Endpoint                   |
| ----------- | --------------------------------------- | ------------------------------ |
| Personal QR | `{"type":"USER","id":"CID123456"}`      | User profile response          |
| Ticket QR   | `{"type":"TICKET","code":"REG-ABC123"}` | Registration response          |
| Event QR    | `{"type":"EVENT","code":"EVT-XMS24"}`   | Event details response         |
| Session QR  | `{"type":"SESSION","code":"SES-001"}`   | Session/class details response |

> [!NOTE]
> No `user_qr_codes` table needed. Frontend generates QR images from the structured data above.

````

---

## 5. Form Schema Migration Notes

> [!WARNING]
> **Breaking Change**: The form system has been migrated from a JSONB-based `registration_forms` schema to a normalized `forms`/`form_questions`/`form_answers`/`form_associations` structure.

#### What Changed

| Old Schema | New Schema | Change Type |
|------------|------------|-------------|
| `registration_forms` table | `forms` table | **Replaced** |
| `fields` JSONB column | `form_questions` table | **Normalized** |
| N/A | `form_associations` table | **New** (polymorphic) |
| `form_submissions` table | Removed | **Deleted** |
| `form_fields` table | `form_questions` table | **Renamed + Enhanced** |
| `form_answers.field_id` | `form_answers.question_code` | **Changed** |
| `form_answers.submission_id` | `form_answers.identifier` + `identifier_type` | **Replaced** |

#### Key Improvements

1. **Type Safety**: ENUM types for `form_status`, `question_type`, `entity_type`, `identifier_type`
2. **Data Integrity**: Foreign key constraints with CASCADE deletes
3. **Performance**: Comprehensive indexes on all foreign keys and query patterns
4. **Flexibility**: Polymorphic associations via `form_associations`
5. **Context-Aware**: `mandatory_for` and `apply_for` arrays for adaptive questions
6. **Guest Support**: Flexible `identifier_type` + `identifier` pattern

#### Migration Strategy

**Phase 1: Create New Tables** (Zero Downtime)

```sql
-- 1. Create ENUMs
CREATE TYPE form_status_enum AS ENUM ('draft', 'active', 'archived', 'template');
CREATE TYPE question_type_enum AS ENUM (...);
CREATE TYPE entity_type_enum AS ENUM (...);
CREATE TYPE identifier_type_enum AS ENUM (...);

-- 2. Create new tables (forms, form_questions, form_answers, form_associations)
-- See section 4.2 for full schema

-- 3. Create triggers and helper functions
```

**Phase 2: Migrate Data**

```sql
-- Migrate forms
INSERT INTO forms (code, name, description, status, created_by_community_id, is_template, created_at)
SELECT 
    code::uuid,
    title,
    description,
    status::form_status_enum,
    created_by_community_id,
    is_template,
    created_at
FROM registration_forms
WHERE deleted_at IS NULL;

-- Migrate questions from JSONB to normalized structure
-- (Requires custom migration script to parse JSONB fields array)
INSERT INTO form_questions (code, form_code, text, type, display_order, ...)
SELECT 
    gen_random_uuid(),
    f.code,
    field->>'label',
    (field->>'type')::question_type_enum,
    (field->>'display_order')::int,
    ...
FROM registration_forms rf
CROSS JOIN LATERAL jsonb_array_elements(rf.fields) AS field
JOIN forms f ON f.code = rf.code::uuid;

-- Migrate form associations from event_sessions
INSERT INTO form_associations (form_code, entity_type, entity_code)
SELECT DISTINCT
    f.code,
    'event_session'::entity_type_enum,
    es.code
FROM event_sessions es
JOIN registration_forms rf ON rf.id = es.registration_form_id
JOIN forms f ON f.code = rf.code::uuid
WHERE es.deleted_at IS NULL;
```

**Phase 3: Update Application Code**

- Update all references from `registration_forms` to `forms`
- Update queries to use `form_questions` instead of JSONB field parsing
- Update form association logic to use `form_associations` table
- Update answer submission to use flexible `identifier_type` + `identifier`

**Phase 4: Deprecate Old Tables**

```sql
-- After grace period (e.g., 30 days)
DROP TABLE IF EXISTS form_submissions CASCADE;
DROP TABLE IF EXISTS form_fields CASCADE;
DROP TABLE IF EXISTS registration_forms CASCADE;
```

#### Backward Compatibility

**Option A: Dual-Write Pattern** (Recommended)

- Write to both old and new schemas during transition
- Read from new schema only
- Allows rollback if issues arise

**Option B: View-Based Compatibility Layer**

```sql
-- Create view that mimics old registration_forms structure
CREATE VIEW registration_forms AS
SELECT 
    id,
    code::varchar(30),
    name as title,
    description,
    -- Aggregate form_questions back into JSONB for compatibility
    (
        SELECT jsonb_agg(
            jsonb_build_object(
                'id', fq.code::text,
                'type', fq.type::text,
                'label', fq.text,
                'required', CASE WHEN 'primary' = ANY(fq.mandatory_for) THEN true ELSE false END,
                'options', fq.options,
                'display_order', fq.display_order
            ) ORDER BY fq.display_order
        )
        FROM form_questions fq
        WHERE fq.form_code = f.code
          AND fq.deleted_at IS NULL
    ) as fields,
    created_by_community_id,
    is_template,
    status::varchar(20),
    created_at,
    updated_at,
    deleted_at
FROM forms f;
```

#### Testing Checklist

- [ ] Verify all forms migrated correctly
- [ ] Verify all questions migrated with correct types
- [ ] Verify form associations created for all sessions
- [ ] Test form creation with new schema
- [ ] Test form cloning from templates
- [ ] Test context-aware question filtering
- [ ] Test polymorphic associations
- [ ] Test answer submission with different identifier types
- [ ] Verify foreign key constraints work correctly
- [ ] Verify indexes improve query performance

---

