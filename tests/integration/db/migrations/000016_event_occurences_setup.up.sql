SET
    TIME ZONE 'Asia/Jakarta';

-- ============================================================================
-- TABLE: event_occurrences
-- Description: Individual occurrences of recurring events
-- Supports: Sunday Services, Weekly meetings, Multi-day conferences
-- ============================================================================
CREATE TABLE event_occurrences (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(30) UNIQUE NOT NULL,
    event_id BIGINT NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    session_id BIGINT REFERENCES event_sessions (id),
    -- Occurrence Date
    occurrence_date DATE NOT NULL,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    -- Override (if this specific occurrence is modified)
    is_modified BOOLEAN DEFAULT FALSE,
    is_cancelled BOOLEAN DEFAULT FALSE,
    cancellation_reason TEXT,
    -- Check-in Window
    check_in_start_at TIMESTAMPTZ,
    check_in_end_at TIMESTAMPTZ,
    -- Stats (Denormalized for performance)
    registered_count INTEGER DEFAULT 0,
    checked_in_count INTEGER DEFAULT 0,
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled',
    created_at TIMESTAMPTZ DEFAULT NOW (),
    updated_at TIMESTAMPTZ DEFAULT NOW ()
);

-- ============================================================================
-- INDEXES: Performance Optimization
-- ============================================================================
-- Primary lookup indexes
CREATE INDEX idx_occurrences_event ON event_occurrences (event_id);

CREATE INDEX idx_occurrences_session ON event_occurrences (session_id)
WHERE
    session_id IS NOT NULL;

-- Composite indexes for common query patterns
-- Pattern: List occurrences for an event by date
CREATE INDEX idx_occurrences_event_date ON event_occurrences (event_id, occurrence_date ASC);

-- Pattern: Filter by event and status
CREATE INDEX idx_occurrences_event_status ON event_occurrences (event_id, status, occurrence_date ASC);

-- Pattern: Session-specific occurrences
CREATE INDEX idx_occurrences_session_date ON event_occurrences (session_id, occurrence_date ASC)
WHERE
    session_id IS NOT NULL;

-- Pattern: Date range queries (upcoming/past events)
CREATE INDEX idx_occurrences_date_range ON event_occurrences (occurrence_date, start_at);

-- Partial indexes for specific scenarios
-- Pattern: Upcoming occurrences only
CREATE INDEX idx_occurrences_upcoming ON event_occurrences (event_id, occurrence_date ASC)
WHERE
    occurrence_date >= CURRENT_DATE
    AND is_cancelled = FALSE;

-- Pattern: Modified occurrences (for admin review)
CREATE INDEX idx_occurrences_modified ON event_occurrences (event_id, occurrence_date)
WHERE
    is_modified = TRUE;

-- Pattern: Cancelled occurrences
CREATE INDEX idx_occurrences_cancelled ON event_occurrences (event_id, occurrence_date)
WHERE
    is_cancelled = TRUE;

-- Pattern: Check-in window queries
CREATE INDEX idx_occurrences_checkin_window ON event_occurrences (check_in_start_at, check_in_end_at)
WHERE
    check_in_start_at IS NOT NULL;

-- Unique constraint: One occurrence per event/session/date combination
CREATE UNIQUE INDEX idx_occurrences_unique ON event_occurrences (
    event_id,
    COALESCE(session_id, 0),
    occurrence_date
);

-- ============================================================================
-- CONSTRAINTS: Data Integrity
-- ============================================================================
-- Status validation
ALTER TABLE event_occurrences ADD CONSTRAINT chk_occurrences_status CHECK (
    status IN ('scheduled', 'ongoing', 'completed', 'cancelled')
);

-- Date validation
ALTER TABLE event_occurrences ADD CONSTRAINT chk_occurrences_dates CHECK (end_at > start_at);

-- Check-in window validation
ALTER TABLE event_occurrences ADD CONSTRAINT chk_occurrences_checkin_window CHECK (
    (
        check_in_start_at IS NULL
        AND check_in_end_at IS NULL
    )
    OR (
        check_in_start_at IS NOT NULL
        AND check_in_end_at IS NOT NULL
        AND check_in_end_at > check_in_start_at
    )
);

-- Count validation
ALTER TABLE event_occurrences ADD CONSTRAINT chk_occurrences_counts CHECK (
    registered_count >= 0
    AND checked_in_count >= 0
    AND checked_in_count <= registered_count
);

-- ============================================================================
-- COMMENTS: Documentation
-- ============================================================================
-- Table comment
COMMENT ON TABLE event_occurrences IS 'Individual occurrences of recurring events. Auto-generated for recurring events (e.g., Sunday Services every week) or manually created for multi-day conferences. Each occurrence can be independently modified or cancelled.';

-- Column comments: Basic Info
COMMENT ON COLUMN event_occurrences.id IS 'Primary key, auto-incrementing integer';

COMMENT ON COLUMN event_occurrences.code IS 'Unique occurrence code for QR generation (e.g., OCC-ABC123)';

-- Column comments: Event Reference
COMMENT ON COLUMN event_occurrences.event_id IS 'Reference to parent event (required). Cascade delete when event is deleted';

COMMENT ON COLUMN event_occurrences.session_id IS 'Reference to specific session if this occurrence is for a session (optional)';

-- Column comments: Occurrence Date
COMMENT ON COLUMN event_occurrences.occurrence_date IS 'Date of this occurrence (used for grouping and filtering)';

COMMENT ON COLUMN event_occurrences.start_at IS 'Start date and time of this occurrence (timezone-aware)';

COMMENT ON COLUMN event_occurrences.end_at IS 'End date and time of this occurrence (timezone-aware). Must be after start_at';

-- Column comments: Override
COMMENT ON COLUMN event_occurrences.is_modified IS 'Whether this occurrence has been modified from the default event/session settings (e.g., different time, location)';

COMMENT ON COLUMN event_occurrences.is_cancelled IS 'Whether this specific occurrence is cancelled (event continues for other dates)';

COMMENT ON COLUMN event_occurrences.cancellation_reason IS 'Reason for cancellation (e.g., weather, speaker unavailable)';

-- Column comments: Check-in Window
COMMENT ON COLUMN event_occurrences.check_in_start_at IS 'When check-in opens for this occurrence (optional, overrides event/session settings)';

COMMENT ON COLUMN event_occurrences.check_in_end_at IS 'When check-in closes for this occurrence (optional, overrides event/session settings)';

-- Column comments: Stats
COMMENT ON COLUMN event_occurrences.registered_count IS 'Denormalized count of registrations for this occurrence. Updated via triggers or application logic';

COMMENT ON COLUMN event_occurrences.checked_in_count IS 'Denormalized count of check-ins for this occurrence. Updated via triggers or application logic';

-- Column comments: Status
COMMENT ON COLUMN event_occurrences.status IS 'Occurrence status: scheduled (future), ongoing (currently happening), completed (finished), cancelled (this occurrence cancelled)';

-- Column comments: Metadata
COMMENT ON COLUMN event_occurrences.created_at IS 'Timestamp when record was created';

COMMENT ON COLUMN event_occurrences.updated_at IS 'Timestamp when record was last updated';