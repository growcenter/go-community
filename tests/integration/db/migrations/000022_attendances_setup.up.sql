SET
    TIME ZONE 'Asia/Jakarta';

-- ============================================================================
-- TABLE: attendances
-- Description: Attendance tracking for events and sessions
-- Supports: Check-in/check-out, Walk-ins, QR-based attendance, Manual entry
-- ============================================================================
CREATE TABLE attendances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    -- Registration Reference
    registration_id UUID REFERENCES registrations (id),
    -- Or Direct Attendance (for walk-ins or attendance-only events)
    event_id BIGINT NOT NULL REFERENCES events (id),
    session_id BIGINT REFERENCES event_sessions (id), -- Can be parent or child session
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
    check_in_location_accuracy DECIMAL(10, 2),
    check_in_distance_from_venue DECIMAL(10, 2),
    check_in_geolocation_passed BOOLEAN,
    check_in_geolocation_override_by VARCHAR(50),
    check_out_latitude DECIMAL(10, 8),
    check_out_longitude DECIMAL(11, 8),
    check_out_location_accuracy DECIMAL(10, 2),
    check_out_distance_from_venue DECIMAL(10, 2),
    -- Metadata
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW (),
    updated_at TIMESTAMPTZ DEFAULT NOW ()
);

-- ============================================================================
-- INDEXES: Performance Optimization
-- ============================================================================
-- Primary lookup indexes
CREATE INDEX idx_attendances_registration ON attendances (registration_id)
WHERE
    registration_id IS NOT NULL;

CREATE INDEX idx_attendances_event ON attendances (event_id);

CREATE INDEX idx_attendances_session ON attendances (session_id)
WHERE
    session_id IS NOT NULL;

CREATE INDEX idx_attendances_attendee ON attendances (attendee_community_id)
WHERE
    attendee_community_id IS NOT NULL;

-- Composite indexes for common query patterns
-- Pattern: List attendances for an event by status
CREATE INDEX idx_attendances_event_status ON attendances (event_id, status, check_in_at DESC);

-- Pattern: List attendances for a session by status
CREATE INDEX idx_attendances_session_status ON attendances (session_id, status, check_in_at DESC)
WHERE
    session_id IS NOT NULL;

-- Pattern: User's attendance history
CREATE INDEX idx_attendances_attendee_event ON attendances (attendee_community_id, event_id, check_in_at DESC)
WHERE
    attendee_community_id IS NOT NULL;

-- Pattern: Staff check-in activity tracking
CREATE INDEX idx_attendances_check_in_by ON attendances (check_in_by_community_id, check_in_at DESC)
WHERE
    check_in_by_community_id IS NOT NULL;

-- Pattern: Check-out tracking
CREATE INDEX idx_attendances_check_out_by ON attendances (check_out_by_community_id, check_out_at DESC)
WHERE
    check_out_by_community_id IS NOT NULL;

-- Time-based indexes
-- Pattern: Daily attendance reports
CREATE INDEX idx_attendances_check_in_date ON attendances (check_in_at, event_id)
WHERE
    check_in_at IS NOT NULL;

-- Pattern: Check-in to check-out duration analysis
CREATE INDEX idx_attendances_duration ON attendances (check_in_at, check_out_at)
WHERE
    check_in_at IS NOT NULL;

-- Partial indexes for specific statuses
CREATE INDEX idx_attendances_present ON attendances (event_id, check_in_at DESC)
WHERE
    status = 'present';

CREATE INDEX idx_attendances_late ON attendances (event_id, late_minutes DESC)
WHERE
    status = 'late';

CREATE INDEX idx_attendances_absent ON attendances (event_id)
WHERE
    status = 'absent';

-- Pattern: Walk-ins (no registration)
CREATE INDEX idx_attendances_walk_ins ON attendances (event_id, check_in_at DESC)
WHERE
    registration_id IS NULL;

-- JSONB index for device info
CREATE INDEX idx_attendances_device_info ON attendances USING GIN (check_in_device_info)
WHERE
    check_in_device_info IS NOT NULL;

-- Geolocation indexes
CREATE INDEX idx_attendances_geolocation_failed ON attendances (event_id, check_in_method)
WHERE
    check_in_geolocation_passed = FALSE;

-- ============================================================================
-- CONSTRAINTS: Data Integrity
-- ============================================================================
-- Status validation
ALTER TABLE attendances ADD CONSTRAINT chk_attendances_status CHECK (
    status IN ('present', 'late', 'absent', 'excused', 'permit')
);

-- Check-in method validation
ALTER TABLE attendances ADD CONSTRAINT chk_attendances_check_in_method CHECK (
    check_in_method IS NULL
    OR check_in_method IN (
        'personal-qr',
        'ticket-qr',
        'manual',
        'facial-recognition'
    )
);

-- Check-out method validation
ALTER TABLE attendances ADD CONSTRAINT chk_attendances_check_out_method CHECK (
    check_out_method IS NULL
    OR check_out_method IN (
        'personal-qr',
        'ticket-qr',
        'manual',
        'facial-recognition'
    )
);

-- Check-out must be after check-in
ALTER TABLE attendances ADD CONSTRAINT chk_attendances_check_out_after_check_in CHECK (
    check_out_at IS NULL
    OR check_in_at IS NULL
    OR check_out_at > check_in_at
);

-- Late minutes validation
ALTER TABLE attendances ADD CONSTRAINT chk_attendances_late_minutes CHECK (
    (
        status != 'late'
        AND late_minutes IS NULL
    )
    OR (
        status = 'late'
        AND late_minutes IS NOT NULL
        AND late_minutes > 0
    )
);

-- Geolocation accuracy validation
ALTER TABLE attendances ADD CONSTRAINT chk_attendances_check_in_geolocation_accuracy CHECK (
    check_in_location_accuracy IS NULL
    OR check_in_location_accuracy >= 0
);

ALTER TABLE attendances ADD CONSTRAINT chk_attendances_check_out_geolocation_accuracy CHECK (
    check_out_location_accuracy IS NULL
    OR check_out_location_accuracy >= 0
);

-- Geolocation distance validation
ALTER TABLE attendances ADD CONSTRAINT chk_attendances_geolocation_distance CHECK (
    (
        check_in_distance_from_venue IS NULL
        OR check_in_distance_from_venue >= 0
    )
    AND (
        check_out_distance_from_venue IS NULL
        OR check_out_distance_from_venue >= 0
    )
);

-- ============================================================================
-- COMMENTS: Documentation
-- ============================================================================
-- Table comment
COMMENT ON TABLE attendances IS 'Attendance tracking for events and sessions. Supports check-in/check-out workflows, walk-in attendees (no prior registration), QR-based attendance, manual entry by staff, and geolocation validation for physical presence verification.';

-- Column comments: Basic Info
COMMENT ON COLUMN attendances.id IS 'Primary key, UUID for distributed system compatibility';

-- Column comments: Registration Reference
COMMENT ON COLUMN attendances.registration_id IS 'Reference to registration record if attendee pre-registered. NULL for walk-ins or attendance-only events';

-- Column comments: Event Reference
COMMENT ON COLUMN attendances.event_id IS 'Reference to parent event (required)';

COMMENT ON COLUMN attendances.session_id IS 'Reference to specific session within event (optional, can be parent or child session)';

-- Column comments: Attendee Info
COMMENT ON COLUMN attendances.attendee_community_id IS 'Community ID of attendee if they are a registered user. NULL for guest attendees';

COMMENT ON COLUMN attendances.attendee_name IS 'Full name of attendee (required, copied from registration or entered manually for walk-ins)';

-- Column comments: Check-in Details
COMMENT ON COLUMN attendances.check_in_at IS 'Timestamp when attendee checked in';

COMMENT ON COLUMN attendances.check_in_method IS 'How check-in was performed: personal-qr (staff scanned user QR), ticket-qr (staff scanned ticket QR), manual (staff manual entry), facial-recognition (AI-based check-in)';

COMMENT ON COLUMN attendances.check_in_by_community_id IS 'Community ID of staff member who performed check-in. NULL for self-service check-in';

COMMENT ON COLUMN attendances.check_in_device_info IS 'JSONB storage for device information (device type, OS, browser, app version, location services status)';

-- Column comments: Check-out Details
COMMENT ON COLUMN attendances.check_out_at IS 'Timestamp when attendee checked out (optional, for events that track exit time)';

COMMENT ON COLUMN attendances.check_out_method IS 'How check-out was performed (same options as check_in_method)';

COMMENT ON COLUMN attendances.check_out_by_community_id IS 'Community ID of staff member who performed check-out';

-- Column comments: Attendance Status
COMMENT ON COLUMN attendances.status IS 'Attendance status: present (on time), late (arrived after grace period), absent (registered but did not attend), excused (pre-approved absence), permit (special permission to be absent)';

COMMENT ON COLUMN attendances.late_minutes IS 'Number of minutes late. Required if status = late, NULL otherwise. Calculated as difference between check_in_at and session start time';

COMMENT ON COLUMN attendances.excuse_reason IS 'Reason for excused absence or permit (user-provided or staff-entered)';

-- Column comments: Geolocation Tracking - Check-in
COMMENT ON COLUMN attendances.check_in_latitude IS 'Latitude coordinate captured at check-in time (for geolocation validation)';

COMMENT ON COLUMN attendances.check_in_longitude IS 'Longitude coordinate captured at check-in time (for geolocation validation)';

COMMENT ON COLUMN attendances.check_in_location_accuracy IS 'GPS accuracy in meters at check-in time';

COMMENT ON COLUMN attendances.check_in_distance_from_venue IS 'Calculated distance from event venue in meters at check-in';

COMMENT ON COLUMN attendances.check_in_geolocation_passed IS 'Whether the check-in location was within acceptable radius of venue. NULL if validation not required';

COMMENT ON COLUMN attendances.check_in_geolocation_override_by IS 'Community ID of staff who manually overrode failed geolocation validation at check-in';

-- Column comments: Geolocation Tracking - Check-out
COMMENT ON COLUMN attendances.check_out_latitude IS 'Latitude coordinate captured at check-out time';

COMMENT ON COLUMN attendances.check_out_longitude IS 'Longitude coordinate captured at check-out time';

COMMENT ON COLUMN attendances.check_out_location_accuracy IS 'GPS accuracy in meters at check-out time';

COMMENT ON COLUMN attendances.check_out_distance_from_venue IS 'Calculated distance from event venue in meters at check-out';

-- Column comments: Metadata
COMMENT ON COLUMN attendances.notes IS 'Additional notes about the attendance (e.g., special circumstances, staff observations)';

COMMENT ON COLUMN attendances.created_at IS 'Timestamp when record was created';

COMMENT ON COLUMN attendances.updated_at IS 'Timestamp when record was last updated';