SET
    TIME ZONE 'Asia/Jakarta';

-- ============================================================================
-- TABLE: registrations
-- Description: User registrations for events and sessions
-- Supports: Personal QR, Event QR, Session QR, Manual, Web registrations
-- ============================================================================
CREATE TABLE registrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    code VARCHAR(30) UNIQUE NOT NULL, -- Ticket code for QR
    -- Event Reference
    event_id BIGINT NOT NULL REFERENCES events (id),
    session_id BIGINT REFERENCES event_sessions (id), -- Can be parent or child session
    occurrence_id BIGINT REFERENCES event_occurrences (id),
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
    group_registration_id UUID REFERENCES registrations (id),
    -- Custom Form Submission
    form_submission_id UUID REFERENCES form_submissions (id),
    -- Status Flow
    status VARCHAR(30) NOT NULL DEFAULT 'pending', -- pending, confirmed, waitlisted, cancelled, expired
    waitlist_position INTEGER,
    -- Timestamps
    registered_at TIMESTAMPTZ DEFAULT NOW (),
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
    registration_location_accuracy DECIMAL(10, 2),
    registration_distance_from_venue DECIMAL(10, 2),
    geolocation_validation_passed BOOLEAN,
    geolocation_override_by VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW (),
    updated_at TIMESTAMPTZ DEFAULT NOW (),
    deleted_at TIMESTAMPTZ
);

-- ============================================================================
-- INDEXES: Performance Optimization
-- ============================================================================
-- Primary lookup indexes (unique constraint already creates index for id, code)
CREATE INDEX idx_registrations_event ON registrations (event_id)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_registrations_session ON registrations (session_id)
WHERE
    session_id IS NOT NULL
    AND deleted_at IS NULL;

CREATE INDEX idx_registrations_occurrence ON registrations (occurrence_id)
WHERE
    occurrence_id IS NOT NULL
    AND deleted_at IS NULL;

CREATE INDEX idx_registrations_registrant ON registrations (registrant_community_id)
WHERE
    registrant_community_id IS NOT NULL
    AND deleted_at IS NULL;

CREATE INDEX idx_registrations_registered_by ON registrations (registered_by_community_id)
WHERE
    registered_by_community_id IS NOT NULL
    AND deleted_at IS NULL;

CREATE INDEX idx_registrations_group ON registrations (group_registration_id)
WHERE
    group_registration_id IS NOT NULL
    AND deleted_at IS NULL;

-- Composite indexes for common query patterns
-- Pattern: List registrations for an event by status
CREATE INDEX idx_registrations_event_status ON registrations (event_id, status, registered_at DESC)
WHERE
    deleted_at IS NULL;

-- Pattern: List registrations for a session by status
CREATE INDEX idx_registrations_session_status ON registrations (session_id, status, registered_at DESC)
WHERE
    session_id IS NOT NULL
    AND deleted_at IS NULL;

-- Pattern: User's registration history
CREATE INDEX idx_registrations_registrant_status ON registrations (
    registrant_community_id,
    status,
    registered_at DESC
)
WHERE
    registrant_community_id IS NOT NULL
    AND deleted_at IS NULL;

-- Pattern: Waitlist management (ordered by position)
CREATE INDEX idx_registrations_waitlist ON registrations (event_id, session_id, waitlist_position ASC)
WHERE
    status = 'waitlisted'
    AND deleted_at IS NULL;

-- Pattern: Group registration queries (find all members of a group)
CREATE INDEX idx_registrations_group_members ON registrations (group_registration_id, is_primary_registrant)
WHERE
    group_registration_id IS NOT NULL
    AND deleted_at IS NULL;

-- Pattern: Registration method analytics
CREATE INDEX idx_registrations_method ON registrations (registration_method, registered_at DESC)
WHERE
    deleted_at IS NULL;

-- Time-based indexes
CREATE INDEX idx_registrations_registered_at ON registrations (registered_at DESC)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_registrations_confirmed_at ON registrations (confirmed_at DESC)
WHERE
    confirmed_at IS NOT NULL
    AND deleted_at IS NULL;

-- Partial indexes for specific statuses
CREATE INDEX idx_registrations_pending ON registrations (event_id, registered_at DESC)
WHERE
    status = 'pending'
    AND deleted_at IS NULL;

CREATE INDEX idx_registrations_confirmed ON registrations (event_id, registered_at DESC)
WHERE
    status = 'confirmed'
    AND deleted_at IS NULL;

-- Soft delete support
CREATE INDEX idx_registrations_deleted_at ON registrations (deleted_at)
WHERE
    deleted_at IS NOT NULL;

-- JSONB index for form answers
CREATE INDEX idx_registrations_form_answers ON registrations USING GIN (form_answers)
WHERE
    form_answers IS NOT NULL
    AND deleted_at IS NULL;

-- Geolocation indexes
CREATE INDEX idx_registrations_geolocation_failed ON registrations (event_id, registration_method)
WHERE
    geolocation_validation_passed = FALSE
    AND deleted_at IS NULL;

-- ============================================================================
-- CONSTRAINTS: Data Integrity
-- ============================================================================
-- Status validation
ALTER TABLE registrations ADD CONSTRAINT chk_registrations_status CHECK (
    status IN (
        'pending',
        'confirmed',
        'waitlisted',
        'cancelled',
        'expired'
    )
);

-- Registration method validation
ALTER TABLE registrations ADD CONSTRAINT chk_registrations_method CHECK (
    registration_method IN (
        'personal-qr',
        'event-qr',
        'session-qr',
        'manual',
        'web'
    )
);

-- Notification channel validation
ALTER TABLE registrations ADD CONSTRAINT chk_registrations_channel CHECK (
    confirmation_channel IS NULL
    OR confirmation_channel IN ('email', 'sms', 'whatsapp')
);

-- Waitlist position validation
ALTER TABLE registrations ADD CONSTRAINT chk_registrations_waitlist_position CHECK (
    (
        status != 'waitlisted'
        AND waitlist_position IS NULL
    )
    OR (
        status = 'waitlisted'
        AND waitlist_position IS NOT NULL
        AND waitlist_position > 0
    )
);

-- Timestamp logic validation
ALTER TABLE registrations ADD CONSTRAINT chk_registrations_confirmed_at CHECK (
    (
        status = 'confirmed'
        AND confirmed_at IS NOT NULL
    )
    OR (status != 'confirmed')
);

-- Geolocation accuracy validation
ALTER TABLE registrations ADD CONSTRAINT chk_registrations_geolocation_accuracy CHECK (
    registration_location_accuracy IS NULL
    OR registration_location_accuracy >= 0
);

-- Geolocation distance validation
ALTER TABLE registrations ADD CONSTRAINT chk_registrations_geolocation_distance CHECK (
    registration_distance_from_venue IS NULL
    OR registration_distance_from_venue >= 0
);

-- ============================================================================
-- COMMENTS: Documentation
-- ============================================================================
-- Table comment
COMMENT ON TABLE registrations IS 'User registrations for events and sessions. Supports multiple registration methods (Personal QR, Event QR, Session QR, Manual entry, Web registration), family/group registrations, waitlist management, and geolocation validation for physical attendance verification.';

-- Column comments: Basic Info
COMMENT ON COLUMN registrations.id IS 'Primary key, UUID for distributed system compatibility';

COMMENT ON COLUMN registrations.code IS 'Unique registration code for QR ticket generation (e.g., REG-ABC123)';

-- Column comments: Event Reference
COMMENT ON COLUMN registrations.event_id IS 'Reference to parent event (required)';

COMMENT ON COLUMN registrations.session_id IS 'Reference to specific session within event (optional, can be parent or child session)';

COMMENT ON COLUMN registrations.occurrence_id IS 'Reference to specific occurrence for recurring events (optional)';

-- Column comments: Registrant Info
COMMENT ON COLUMN registrations.registrant_community_id IS 'Community ID of registrant if they are a registered user. NULL for guest registrations';

COMMENT ON COLUMN registrations.registrant_name IS 'Full name of registrant (required for all registrations)';

COMMENT ON COLUMN registrations.registrant_email IS 'Email address of registrant (optional, depends on session identifier_config)';

COMMENT ON COLUMN registrations.registrant_phone IS 'Phone number of registrant (optional, depends on session identifier_config)';

COMMENT ON COLUMN registrations.registrant_legal_id IS 'Legal ID (KTP/Passport) of registrant (optional, for high-security events)';

-- Column comments: Registration Source
COMMENT ON COLUMN registrations.registration_method IS 'How the registration was created: personal-qr (staff scanned user QR), event-qr (user scanned event QR), session-qr (user scanned session QR), manual (staff manual entry), web (online registration)';

COMMENT ON COLUMN registrations.registered_by_community_id IS 'Community ID of the person who created this registration. For family/group registrations, this is the primary registrant who registered others';

-- Column comments: Family/Group Registration
COMMENT ON COLUMN registrations.is_primary_registrant IS 'Whether this is the primary registrant in a group registration. TRUE for single registrations and group leaders';

COMMENT ON COLUMN registrations.group_registration_id IS 'Reference to the primary registrant registration ID for group members. NULL for primary registrants and single registrations';

-- Column comments: Custom Form Answers
COMMENT ON COLUMN registrations.form_answers IS 'JSONB storage for custom registration form answers (e.g., dietary preferences, t-shirt size, special requirements)';

-- Column comments: Status Flow
COMMENT ON COLUMN registrations.status IS 'Registration status: pending (awaiting confirmation/approval), confirmed (registration confirmed), waitlisted (on waitlist due to capacity), cancelled (registration cancelled), expired (registration expired)';

COMMENT ON COLUMN registrations.waitlist_position IS 'Position in waitlist queue (1 = first in line). Required if status = waitlisted, NULL otherwise';

-- Column comments: Timestamps
COMMENT ON COLUMN registrations.registered_at IS 'When the registration was created';

COMMENT ON COLUMN registrations.confirmed_at IS 'When the registration was confirmed. Required if status = confirmed';

COMMENT ON COLUMN registrations.cancelled_at IS 'When the registration was cancelled';

COMMENT ON COLUMN registrations.cancellation_reason IS 'Reason for cancellation (user-provided or system-generated)';

-- Column comments: Notification Tracking
COMMENT ON COLUMN registrations.confirmation_sent_at IS 'When confirmation notification was sent to registrant';

COMMENT ON COLUMN registrations.confirmation_channel IS 'Channel used for confirmation notification: email, sms, whatsapp';

COMMENT ON COLUMN registrations.reminder_sent_at IS 'When reminder notification was last sent';

-- Column comments: Geolocation Tracking
COMMENT ON COLUMN registrations.registration_latitude IS 'Latitude coordinate captured at registration time (for geolocation validation)';

COMMENT ON COLUMN registrations.registration_longitude IS 'Longitude coordinate captured at registration time (for geolocation validation)';

COMMENT ON COLUMN registrations.registration_location_accuracy IS 'GPS accuracy in meters at registration time';

COMMENT ON COLUMN registrations.registration_distance_from_venue IS 'Calculated distance from event venue in meters';

COMMENT ON COLUMN registrations.geolocation_validation_passed IS 'Whether the registration location was within acceptable radius of venue. NULL if validation not required';

COMMENT ON COLUMN registrations.geolocation_override_by IS 'Community ID of staff who manually overrode failed geolocation validation';

-- Column comments: Metadata
COMMENT ON COLUMN registrations.created_at IS 'Timestamp when record was created';

COMMENT ON COLUMN registrations.updated_at IS 'Timestamp when record was last updated';

COMMENT ON COLUMN registrations.deleted_at IS 'Soft delete timestamp. NULL = active, NOT NULL = deleted';