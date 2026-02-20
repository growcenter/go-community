SET TIME ZONE 'Asia/Jakarta';

-- ============================================================================
-- TABLE: event_sessions
-- Description: Unified sessions/classes/tracks within an event
-- Supports: Services, Classes, Tracks, Breakouts, Workshops (hierarchical)
-- ============================================================================

CREATE TABLE event_sessions (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(30) UNIQUE NOT NULL,
    event_code BIGINT NOT NULL REFERENCES events(code) ON DELETE CASCADE,
    parent_session_id BIGINT REFERENCES event_sessions(code) ON DELETE CASCADE, -- For hierarchy (e.g., Day 1 → Track A)

    -- Basic Info
    title VARCHAR(255) NOT NULL,
    description TEXT,
    session_type VARCHAR(30) NOT NULL, -- service, class, track, breakout, workshop, general, kids, youth, teen, adult

    -- Instructor/Speaker (for workshops, tracks, classes)
    instructor_name TEXT[],
    instructor_community_id TEXT[],

    -- Override Location (if different from event/parent)
    location_type VARCHAR(20) NOT NULL, -- online, offline, hybrid
    location_visibility VARCHAR(20) NOT NULL DEFAULT 'all', -- pre-registration, post-registration, all
    physical_place_name VARCHAR(100),
    physical_address TEXT,
    virtual_link TEXT,
    virtual_platform VARCHAR(50), -- youtube, zoom, meet, custom
    location_details TEXT,
    cta_text VARCHAR(50),
    cta_link TEXT,

    -- Geolocation Validation (for physical attendance verification)
    geolocation JSONB, -- {enabled, venue_latitude, venue_longitude, radius_meters, validation_rules, error_action}

    -- Schedule
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    timezone VARCHAR(50),

    -- Registration Window
    registration_start_at TIMESTAMPTZ,
    registration_end_at TIMESTAMPTZ,

    -- Verification Window (Check-in)
    check_in_start_at TIMESTAMPTZ,
    check_in_end_at TIMESTAMPTZ,

    -- Capacity
    capacity INTEGER,
    current_count INTEGER DEFAULT 0,
    waitlist_enabled BOOLEAN DEFAULT FALSE,
    waitlist_capacity INTEGER,

    -- Registration Rules
    require_approval BOOLEAN DEFAULT FALSE,

    -- Group/Family Registration Config
    registration_mode VARCHAR(30) DEFAULT 'self_and_others',
    -- Values: 'self_only' (user can only register themselves)
    --         'self_and_registered' (user can register themselves + other registered users)
    --         'self_and_others' (user can register themselves + guests without accounts)

    max_registrations_per_user INTEGER DEFAULT 1, -- How many people a user can register (including self)
    one_session_per_event BOOLEAN DEFAULT FALSE, -- If true, user can only register for ONE session in this event

    -- Form Requirements for Additional Registrants
    additional_registrant_form_mode VARCHAR(30) DEFAULT 'name_only',
    -- Values: 'same_as_primary' (same form as primary registrant)
    --         'name_only' (only name required for additional registrants)
    --         'custom' (use additional_registrant_form_id)
    additional_registrant_form_id BIGINT REFERENCES registration_forms(id),

    -- Check-in Config
    check_in_required BOOLEAN DEFAULT TRUE,
    check_out_required BOOLEAN DEFAULT FALSE,

    -- Age/Eligibility (for kids, youth, etc.)
    min_age INTEGER,
    max_age INTEGER,
    prerequisites TEXT[],

    -- Custom Questions (for primary registrant)
    registration_form_id BIGINT REFERENCES registration_forms(id),

    -- Identifier Configuration (JSONB for flexibility)
    -- Separate configs for primary registrant and additional registrants
    identifier_config JSONB DEFAULT '{
        "primary": {
            "name": {"visible": true, "required": true},
            "email": {"visible": true, "required": true},
            "phone": {"visible": true, "required": false},
            "legal_id": {"visible": false, "required": false}
        },
        "additional": {
            "name": {"visible": true, "required": true},
            "email": {"visible": false, "required": false},
            "phone": {"visible": false, "required": false},
            "legal_id": {"visible": false, "required": false}
        }
    }',

    -- Form Field Overrides (session-specific field visibility/requirements)
    form_field_overrides JSONB, -- {"field_key": {"visible": true, "required": true}}

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- ============================================================================
-- INDEXES: Performance Optimization
-- ============================================================================

-- Primary lookup indexes (unique constraint already creates index for code)
CREATE INDEX idx_sessions_event ON event_sessions(event_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_sessions_parent ON event_sessions(parent_session_id) WHERE parent_session_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_sessions_instructor ON event_sessions(instructor_community_id) WHERE instructor_community_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_sessions_registration_form ON event_sessions(registration_form_id) WHERE registration_form_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_sessions_additional_form ON event_sessions(additional_registrant_form_id) WHERE additional_registrant_form_id IS NOT NULL AND deleted_at IS NULL;

-- Composite indexes for common query patterns
-- Pattern: List sessions for an event, ordered by start time
CREATE INDEX idx_sessions_event_start ON event_sessions(event_id, start_at ASC) WHERE deleted_at IS NULL;

-- Pattern: List sessions by event and status (published sessions for an event)
CREATE INDEX idx_sessions_event_status_start ON event_sessions(event_id, status, start_at ASC) WHERE deleted_at IS NULL;

-- Pattern: Filter sessions by type and status (all published kids sessions)
CREATE INDEX idx_sessions_type_status_start ON event_sessions(session_type, status, start_at ASC) WHERE deleted_at IS NULL;

-- Pattern: Get child sessions of a parent (conference tracks under a day)
CREATE INDEX idx_sessions_parent_start ON event_sessions(parent_session_id, start_at ASC) WHERE parent_session_id IS NOT NULL AND deleted_at IS NULL;

-- Pattern: Find sessions by instructor
CREATE INDEX idx_sessions_instructor_start ON event_sessions(instructor_community_id, start_at DESC) WHERE instructor_community_id IS NOT NULL AND deleted_at IS NULL;

-- Pattern: Capacity management queries (sessions with available capacity)
CREATE INDEX idx_sessions_capacity_available ON event_sessions(event_id, capacity, current_count) WHERE capacity IS NOT NULL AND current_count < capacity AND deleted_at IS NULL;

-- Pattern: Waitlist management (sessions with waitlist enabled)
CREATE INDEX idx_sessions_waitlist ON event_sessions(event_id, waitlist_enabled) WHERE waitlist_enabled = TRUE AND deleted_at IS NULL;

-- Pattern: Date range queries (upcoming sessions, sessions in progress)
CREATE INDEX idx_sessions_date_range ON event_sessions(start_at, end_at) WHERE deleted_at IS NULL AND status = 'published';

-- Pattern: Registration window queries (sessions currently accepting registrations)
CREATE INDEX idx_sessions_registration_window ON event_sessions(registration_start_at, registration_end_at) WHERE registration_start_at IS NOT NULL AND deleted_at IS NULL;

-- Pattern: Check-in window queries (sessions currently accepting check-ins)
CREATE INDEX idx_sessions_checkin_window ON event_sessions(check_in_start_at, check_in_end_at) WHERE check_in_start_at IS NOT NULL AND deleted_at IS NULL;

-- Pattern: Age-restricted sessions (kids, youth, teen sessions)
CREATE INDEX idx_sessions_age_restricted ON event_sessions(session_type, min_age, max_age) WHERE min_age IS NOT NULL OR max_age IS NOT NULL;

-- Soft delete support
CREATE INDEX idx_sessions_deleted_at ON event_sessions(deleted_at) WHERE deleted_at IS NOT NULL;

-- Array column indexes for prerequisites
CREATE INDEX idx_sessions_prerequisites ON event_sessions USING GIN(prerequisites) WHERE prerequisites IS NOT NULL AND deleted_at IS NULL;

-- JSONB index for identifier configuration queries
CREATE INDEX idx_sessions_identifier_config ON event_sessions USING GIN(identifier_config) WHERE deleted_at IS NULL;

-- JSONB index for form field overrides
CREATE INDEX idx_sessions_form_field_overrides ON event_sessions USING GIN(form_field_overrides) WHERE form_field_overrides IS NOT NULL AND deleted_at IS NULL;

-- Full-text search on title and description
CREATE INDEX idx_sessions_search ON event_sessions USING GIN(to_tsvector('indonesian', COALESCE(title, '') || ' ' || COALESCE(description, ''))) WHERE deleted_at IS NULL;

-- ============================================================================
-- CONSTRAINTS: Data Integrity
-- ============================================================================

-- Session type validation
ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_type 
    CHECK (session_type IN ('service', 'class', 'track', 'breakout', 'workshop', 'general', 'kids', 'youth', 'teen', 'adult'));

-- Location type validation (if overriding event location)
ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_location_type 
    CHECK (location_type IS NULL OR location_type IN ('online', 'offline', 'hybrid'));

-- Virtual platform validation
ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_virtual_platform 
    CHECK (virtual_platform IS NULL OR virtual_platform IN ('youtube', 'zoom', 'meet', 'custom'));

-- Status validation
ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_status 
    CHECK (status IN ('draft', 'published', 'cancelled', 'completed', 'full'));

-- Registration mode validation
ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_registration_mode 
    CHECK (registration_mode IN ('self_only', 'self_and_registered', 'self_and_others'));

-- Additional registrant form mode validation
ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_additional_form_mode 
    CHECK (additional_registrant_form_mode IN ('same_as_primary', 'name_only', 'custom'));

-- Date logic validation
ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_dates 
    CHECK (end_at > start_at);

-- Registration window validation
ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_registration_window 
    CHECK (
        (registration_start_at IS NULL AND registration_end_at IS NULL) OR
        (registration_start_at IS NOT NULL AND registration_end_at IS NOT NULL AND registration_end_at > registration_start_at)
    );

-- Check-in window validation
ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_checkin_window 
    CHECK (
        (check_in_start_at IS NULL AND check_in_end_at IS NULL) OR
        (check_in_start_at IS NOT NULL AND check_in_end_at IS NOT NULL AND check_in_end_at > check_in_start_at)
    );

-- Capacity validation
ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_capacity 
    CHECK (capacity IS NULL OR capacity > 0);

-- Current count validation
ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_current_count 
    CHECK (current_count >= 0 AND (capacity IS NULL OR current_count <= capacity + COALESCE(waitlist_capacity, 0)));

-- Waitlist capacity validation
ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_waitlist_capacity 
    CHECK (waitlist_capacity IS NULL OR waitlist_capacity >= 0);

-- Max registrations validation
ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_max_registrations 
    CHECK (max_registrations_per_user > 0);

-- Age range validation
ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_age_range 
    CHECK (
        (min_age IS NULL AND max_age IS NULL) OR
        (min_age IS NOT NULL AND max_age IS NULL AND min_age >= 0) OR
        (min_age IS NULL AND max_age IS NOT NULL AND max_age > 0) OR
        (min_age IS NOT NULL AND max_age IS NOT NULL AND max_age > min_age AND min_age >= 0)
    );

-- Custom form reference validation (if mode is 'custom', form_id must be provided)
ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_custom_form_reference 
    CHECK (
        additional_registrant_form_mode != 'custom' OR 
        (additional_registrant_form_mode = 'custom' AND additional_registrant_form_id IS NOT NULL)
    );

-- Self-reference validation (session cannot be its own parent)
ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_self_reference 
    CHECK (parent_session_id IS NULL OR parent_session_id != id);

-- ============================================================================
-- COMMENTS: Documentation
-- ============================================================================

-- Table comment
COMMENT ON TABLE event_sessions IS 'Unified sessions/classes/tracks within an event. Handles all sub-event types including Sunday Service sessions, Christmas service times, Conference tracks and breakouts, Workshop sessions, and Age-specific sessions (kids, youth, teen, adult). Supports hierarchical structure via parent_session_id for complex events like multi-day conferences.';

-- Column comments: Basic Info
COMMENT ON COLUMN event_sessions.id IS 'Primary key, auto-incrementing session ID';
COMMENT ON COLUMN event_sessions.code IS 'Unique session code for QR generation and public reference (e.g., SES-001, SES-TRACK-A)';
COMMENT ON COLUMN event_sessions.event_id IS 'Reference to parent event (CASCADE delete - sessions deleted when event is deleted)';
COMMENT ON COLUMN event_sessions.parent_session_id IS 'Reference to parent session for hierarchical events (e.g., Conference Day 1 → Track A). NULL for top-level sessions. CASCADE delete - child sessions deleted when parent is deleted';
COMMENT ON COLUMN event_sessions.title IS 'Session title displayed to users (e.g., "Service 1", "Track A: Leadership", "Kids Session")';
COMMENT ON COLUMN event_sessions.description IS 'Detailed session description supporting markdown/rich text';
COMMENT ON COLUMN event_sessions.session_type IS 'Session type: service (Sunday Service), class (educational), track (conference track), breakout (small group), workshop (hands-on), general (default), kids/youth/teen/adult (age-specific)';

-- Column comments: Instructor/Speaker
COMMENT ON COLUMN event_sessions.instructor_name IS 'Name of instructor/speaker/facilitator for this session';
COMMENT ON COLUMN event_sessions.instructor_community_id IS 'Community ID of instructor if they are a registered user';

-- Column comments: Location Override
COMMENT ON COLUMN event_sessions.location_type IS 'Location type override (if different from event): online, offline, hybrid. NULL = inherit from event';
COMMENT ON COLUMN event_sessions.physical_address IS 'Physical address override for this session (if different from event)';
COMMENT ON COLUMN event_sessions.room_name IS 'Specific room/venue name (e.g., "Room 201", "Main Auditorium", "Kids Hall")';
COMMENT ON COLUMN event_sessions.virtual_link IS 'Virtual meeting/streaming link override for this session';
COMMENT ON COLUMN event_sessions.virtual_platform IS 'Virtual platform override: youtube, zoom, meet, custom. NULL = inherit from event';
COMMENT ON COLUMN event_sessions.location_details IS 'Additional location information specific to this session';

-- Column comments: Schedule
COMMENT ON COLUMN event_sessions.start_at IS 'Session start date and time (timezone-aware)';
COMMENT ON COLUMN event_sessions.end_at IS 'Session end date and time (timezone-aware). Must be after start_at';
COMMENT ON COLUMN event_sessions.timezone IS 'IANA timezone identifier override. NULL = inherit from event';

-- Column comments: Registration Window
COMMENT ON COLUMN event_sessions.registration_start_at IS 'When registration opens for this session. NULL = open immediately';
COMMENT ON COLUMN event_sessions.registration_end_at IS 'When registration closes for this session. NULL = open until session starts';

-- Column comments: Check-in Window
COMMENT ON COLUMN event_sessions.check_in_start_at IS 'When check-in opens (e.g., 30 minutes before session). NULL = open anytime';
COMMENT ON COLUMN event_sessions.check_in_end_at IS 'When check-in closes (e.g., at session start). NULL = open indefinitely';

-- Column comments: Capacity
COMMENT ON COLUMN event_sessions.capacity IS 'Maximum number of registrants allowed. NULL = unlimited capacity';
COMMENT ON COLUMN event_sessions.current_count IS 'Current number of confirmed registrants (denormalized for performance)';
COMMENT ON COLUMN event_sessions.waitlist_enabled IS 'Whether waitlist is enabled when capacity is full';
COMMENT ON COLUMN event_sessions.waitlist_capacity IS 'Maximum waitlist size. NULL = unlimited waitlist';

-- Column comments: Registration Rules
COMMENT ON COLUMN event_sessions.require_approval IS 'Whether registrations require organizer approval before confirmation';

-- Column comments: Group/Family Registration
COMMENT ON COLUMN event_sessions.registration_mode IS 'Registration mode: self_only (user can only register themselves), self_and_registered (user + other registered users), self_and_others (user + guests without accounts)';
COMMENT ON COLUMN event_sessions.max_registrations_per_user IS 'Maximum number of people one user can register including themselves (e.g., 5 for family registration)';
COMMENT ON COLUMN event_sessions.one_session_per_event IS 'If true, user can only register for ONE session within the parent event (prevents double-booking for parallel sessions)';

-- Column comments: Additional Registrant Form
COMMENT ON COLUMN event_sessions.additional_registrant_form_mode IS 'Form requirements for additional registrants: same_as_primary (same custom form), name_only (only name required), custom (use additional_registrant_form_id)';
COMMENT ON COLUMN event_sessions.additional_registrant_form_id IS 'Reference to custom form for additional registrants (required if mode = custom)';

-- Column comments: Check-in Config
COMMENT ON COLUMN event_sessions.check_in_required IS 'Whether check-in is required for this session';
COMMENT ON COLUMN event_sessions.check_out_required IS 'Whether check-out is required for this session (for tracking attendance duration)';

-- Column comments: Age/Eligibility
COMMENT ON COLUMN event_sessions.min_age IS 'Minimum age requirement (e.g., 3 for kids session). NULL = no minimum';
COMMENT ON COLUMN event_sessions.max_age IS 'Maximum age requirement (e.g., 12 for kids session). NULL = no maximum';
COMMENT ON COLUMN event_sessions.prerequisites IS 'Array of prerequisite requirements (e.g., ["Completed Level 1", "Member for 6 months"])';

-- Column comments: Custom Forms
COMMENT ON COLUMN event_sessions.registration_form_id IS 'Reference to custom registration form for primary registrant (e.g., dietary preferences, t-shirt size)';

-- Column comments: Identifier Configuration
COMMENT ON COLUMN event_sessions.identifier_config IS 'JSONB configuration for which identifier fields (name, email, phone, legal_id) are visible and required for primary and additional registrants. Allows flexible data collection per session.';

-- Column comments: Form Field Overrides
COMMENT ON COLUMN event_sessions.form_field_overrides IS 'JSONB session-specific overrides for form field visibility and requirements. Allows same form to have different field requirements per session. Example: {"question_b": {"visible": true, "required": true}, "question_c": {"visible": false}}';

-- Column comments: Metadata
COMMENT ON COLUMN event_sessions.status IS 'Session lifecycle status: draft (not visible), published (active and visible), cancelled (session cancelled), completed (session finished), full (capacity reached)';
COMMENT ON COLUMN event_sessions.created_at IS 'Timestamp when session was created';
COMMENT ON COLUMN event_sessions.updated_at IS 'Timestamp when session was last updated';
COMMENT ON COLUMN event_sessions.deleted_at IS 'Soft delete timestamp. NULL = active, NOT NULL = deleted';