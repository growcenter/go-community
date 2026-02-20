SET TIME ZONE 'Asia/Jakarta';

-- ============================================================================
-- TABLE: event_sessions
-- Description: Unified sessions/classes/tracks within an event
-- Supports: Services, Classes, Tracks, Breakouts, Workshops (hierarchical)
-- Aligns with: models.EventSession
-- ============================================================================
CREATE TABLE event_sessions (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    event_code VARCHAR(50) NOT NULL REFERENCES events(code) ON DELETE CASCADE,
    parent_session_code VARCHAR(50) REFERENCES event_sessions(code) ON DELETE CASCADE,
    -- NULL = top-level session; set for child sessions (e.g., track → breakout)

    -- Basic Info
    title TEXT NOT NULL,
    description TEXT,
    session_type VARCHAR(50) NOT NULL,
    -- service, class, track, breakout, workshop, general, kids, youth, teen, adult

    -- Location (inherits from event if not set — see inheritance chain in §4.2)
    location_type VARCHAR(20) NOT NULL DEFAULT 'offline', -- online, offline, hybrid
    location_visibility VARCHAR(30) NOT NULL DEFAULT 'all', -- all, pre-registration, post-registration
    physical_place_name TEXT,                -- venue / room name
    physical_address TEXT,
    virtual_link TEXT,
    virtual_platform TEXT,                   -- youtube, zoom, meet, etc.
    location_details TEXT,
    cta_text VARCHAR(100),                   -- call-to-action button label
    cta_link TEXT,                           -- CTA URL or "NORMAL_FLOW"

    -- Geolocation Validation
    geolocation JSONB,
    -- Structure: { enabled, venueLatitude, venueLongitude, radiusMeters,
    --              validationRules: { personal_qr: {...}, event_qr: {...} },
    --              errorAction: "reject"|"warn" }

    -- Schedule
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    timezone VARCHAR(50) NOT NULL DEFAULT 'Asia/Jakarta',

    -- Registration Window
    registration_start_at TIMESTAMPTZ,
    registration_end_at TIMESTAMPTZ,

    -- Capacity
    capacity INTEGER DEFAULT 0,              -- 0 = unlimited
    waitlist_enabled BOOLEAN DEFAULT FALSE,
    waitlist_capacity INTEGER DEFAULT 0,

    -- Registration Rules
    require_approval BOOLEAN DEFAULT FALSE,
    registration_methods TEXT[],             -- personal_qr, event_qr, session_qr, registration_qr
    registration_mode VARCHAR(30) NOT NULL DEFAULT 'self_and_others',
    -- self_only, self_and_registered, self_and_others
    max_registrations_per_user INTEGER NOT NULL DEFAULT 1,
    one_session_per_event BOOLEAN NOT NULL DEFAULT FALSE,

    -- Check-in Configuration
    check_in_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    check_in_required BOOLEAN NOT NULL DEFAULT FALSE,
    check_in_start_at TIMESTAMPTZ,           -- custom window start (NULL = open from session start)
    check_in_end_at TIMESTAMPTZ,             -- custom window end   (NULL = open indefinitely)
    check_in_allow_late BOOLEAN NOT NULL DEFAULT TRUE,
    check_in_late_threshold INTEGER NOT NULL DEFAULT 10, -- minutes after start_at to mark as "late"

    -- Check-out Configuration
    check_out_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    check_out_required BOOLEAN NOT NULL DEFAULT FALSE,
    check_out_start_at TIMESTAMPTZ,
    check_out_end_at TIMESTAMPTZ,
    check_out_allow_late BOOLEAN NOT NULL DEFAULT TRUE,
    check_out_late_threshold INTEGER NOT NULL DEFAULT 30,

    -- Age / Eligibility
    min_age INTEGER,
    max_age INTEGER,
    prerequisites TEXT,                      -- free-text or structured description

    -- Identifier Configuration (JSONB — controls which fields primary/additional registrants fill)
    -- Maps to: SessionIdentifierConfig { Primary: PrimaryIdentifierConfig, Additional: AdditionalIdentifierConfig }
    -- Each FieldConfig: { visible: bool, required: bool }
    identifier_config JSONB DEFAULT '{
        "primary": {
            "name":        {"visible": true,  "required": true},
            "email":       {"visible": true,  "required": true},
            "phone":       {"visible": true,  "required": false},
            "communityId": {"visible": false, "required": false}
        },
        "additional": {
            "name":        {"visible": true,  "required": true},
            "email":       {"visible": false, "required": false},
            "phone":       {"visible": false, "required": false},
            "communityId": {"visible": false, "required": false}
        }
    }',

    -- Status & Metadata
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    -- draft, active, inactive (matches model validate tag)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- ============================================================================
-- INDEXES: Performance Optimization
-- ============================================================================

-- Primary lookup
CREATE INDEX idx_sessions_event ON event_sessions(event_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_sessions_parent ON event_sessions(parent_session_code)
    WHERE parent_session_code IS NOT NULL AND deleted_at IS NULL;

-- Common query patterns
CREATE INDEX idx_sessions_event_start ON event_sessions(event_code, start_at ASC)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_sessions_event_status_start ON event_sessions(event_code, status, start_at ASC)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_sessions_type_status ON event_sessions(session_type, status)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_sessions_parent_start ON event_sessions(parent_session_code, start_at ASC)
    WHERE parent_session_code IS NOT NULL AND deleted_at IS NULL;

-- Capacity management
CREATE INDEX idx_sessions_waitlist ON event_sessions(event_code, waitlist_enabled)
    WHERE waitlist_enabled = TRUE AND deleted_at IS NULL;

-- Time window queries
CREATE INDEX idx_sessions_registration_window ON event_sessions(registration_start_at, registration_end_at)
    WHERE registration_start_at IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_sessions_checkin_window ON event_sessions(check_in_start_at, check_in_end_at)
    WHERE check_in_start_at IS NOT NULL AND deleted_at IS NULL;

-- Age-restricted
CREATE INDEX idx_sessions_age ON event_sessions(session_type, min_age, max_age)
    WHERE (min_age IS NOT NULL OR max_age IS NOT NULL) AND deleted_at IS NULL;

-- JSONB
CREATE INDEX idx_sessions_identifier_config ON event_sessions USING GIN(identifier_config)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_sessions_geolocation ON event_sessions USING GIN(geolocation)
    WHERE geolocation IS NOT NULL AND deleted_at IS NULL;

-- Registration methods (GIN for array contains queries)
CREATE INDEX idx_sessions_registration_methods ON event_sessions USING GIN(registration_methods)
    WHERE deleted_at IS NULL;

-- ============================================================================
-- CONSTRAINTS: Data Integrity
-- ============================================================================

ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_type
    CHECK (session_type IN ('service','class','track','breakout','workshop','general','kids','youth','teen','adult'));

ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_location_type
    CHECK (location_type IN ('online','offline','hybrid'));

ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_location_visibility
    CHECK (location_visibility IN ('all','pre-registration','post-registration'));

ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_status
    CHECK (status IN ('draft','active','inactive'));

ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_registration_mode
    CHECK (registration_mode IN ('self_only','self_and_registered','self_and_others'));

ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_dates
    CHECK (end_at > start_at);

ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_registration_window
    CHECK (
        (registration_start_at IS NULL AND registration_end_at IS NULL) OR
        (registration_start_at IS NOT NULL AND registration_end_at IS NOT NULL
            AND registration_end_at > registration_start_at)
    );

ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_checkin_window
    CHECK (
        (check_in_start_at IS NULL AND check_in_end_at IS NULL) OR
        (check_in_start_at IS NOT NULL AND check_in_end_at IS NOT NULL
            AND check_in_end_at > check_in_start_at)
    );

ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_checkout_window
    CHECK (
        (check_out_start_at IS NULL AND check_out_end_at IS NULL) OR
        (check_out_start_at IS NOT NULL AND check_out_end_at IS NOT NULL
            AND check_out_end_at > check_out_start_at)
    );

ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_capacity
    CHECK (capacity >= 0);

ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_waitlist_capacity
    CHECK (waitlist_capacity >= 0);

ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_max_registrations
    CHECK (max_registrations_per_user > 0);

ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_age_range
    CHECK (
        (min_age IS NULL AND max_age IS NULL) OR
        (min_age IS NOT NULL AND max_age IS NULL AND min_age >= 0) OR
        (min_age IS NULL AND max_age IS NOT NULL AND max_age > 0) OR
        (min_age IS NOT NULL AND max_age IS NOT NULL AND max_age > min_age AND min_age >= 0)
    );

ALTER TABLE event_sessions ADD CONSTRAINT chk_sessions_late_thresholds
    CHECK (check_in_late_threshold >= 0 AND check_out_late_threshold >= 0);

-- ============================================================================
-- AUTO-UPDATE TRIGGER
-- ============================================================================
CREATE OR REPLACE FUNCTION update_sessions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_event_sessions_updated_at
    BEFORE UPDATE ON event_sessions
    FOR EACH ROW EXECUTE FUNCTION update_sessions_updated_at();

-- ============================================================================
-- COMMENTS: Documentation
-- ============================================================================
COMMENT ON TABLE event_sessions IS 'Unified sessions/classes/tracks within an event. Handles all sub-event types. Supports hierarchical structure via parent_session_code.';

COMMENT ON COLUMN event_sessions.code IS 'Unique session code (e.g., SES-001). Used for QR codes and API lookup.';
COMMENT ON COLUMN event_sessions.event_code IS 'Code of the parent event. CASCADE delete.';
COMMENT ON COLUMN event_sessions.parent_session_code IS 'Code of parent session for hierarchical events. NULL = top-level session.';
COMMENT ON COLUMN event_sessions.session_type IS 'service, class, track, breakout, workshop, general, kids, youth, teen, adult';
COMMENT ON COLUMN event_sessions.geolocation IS 'JSONB geolocation config for physical attendance verification.';
COMMENT ON COLUMN event_sessions.registration_methods IS 'Array of allowed methods: personal_qr, event_qr, session_qr, registration_qr';
COMMENT ON COLUMN event_sessions.capacity IS '0 = unlimited. Positive = max registrants.';
COMMENT ON COLUMN event_sessions.identifier_config IS 'JSONB controlling which identifier fields (name, email, phone, communityId) are visible/required for primary and additional registrants.';
COMMENT ON COLUMN event_sessions.check_in_late_threshold IS 'Minutes after start_at to classify a check-in as late.';
COMMENT ON COLUMN event_sessions.check_out_late_threshold IS 'Minutes after end_at to classify a check-out as late.';
COMMENT ON COLUMN event_sessions.status IS 'draft, active, inactive';