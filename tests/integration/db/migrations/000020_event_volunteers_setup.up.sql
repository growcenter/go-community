SET
    TIME ZONE 'Asia/Jakarta';

-- ============================================================================
-- TABLE: event_volunteers
-- Description: Volunteer assignments for events and sessions
-- Supports: Scanners, Ushers, Coordinators, Custom roles
-- ============================================================================
CREATE TABLE event_volunteers (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    session_id BIGINT REFERENCES event_sessions (id),
    community_id VARCHAR(50) NOT NULL,
    role VARCHAR(50) NOT NULL, -- scanner, usher, coordinator
    assigned_at TIMESTAMPTZ DEFAULT NOW (),
    assigned_by_community_id VARCHAR(50) NOT NULL,
    UNIQUE (event_id, session_id, community_id)
);

-- ============================================================================
-- INDEXES: Performance Optimization
-- ============================================================================
-- Primary lookup indexes
CREATE INDEX idx_volunteers_event ON event_volunteers (event_id);

CREATE INDEX idx_volunteers_session ON event_volunteers (session_id)
WHERE
    session_id IS NOT NULL;

CREATE INDEX idx_volunteers_community ON event_volunteers (community_id);

-- Composite indexes for common query patterns
-- Pattern: Filter volunteers by event and role
CREATE INDEX idx_volunteers_event_role ON event_volunteers (event_id, role);

-- Pattern: User's volunteer assignments
CREATE INDEX idx_volunteers_community_role ON event_volunteers (community_id, role, assigned_at DESC);

-- Pattern: Who assigned volunteers (audit trail)
CREATE INDEX idx_volunteers_assigned_by ON event_volunteers (assigned_by_community_id, assigned_at DESC);

-- ============================================================================
-- CONSTRAINTS: Data Integrity
-- ============================================================================
-- Role validation
ALTER TABLE event_volunteers ADD CONSTRAINT chk_volunteers_role CHECK (
    role IN (
        'scanner',
        'usher',
        'coordinator',
        'registration-desk',
        'parking',
        'security',
        'tech-support',
        'hospitality',
        'custom'
    )
);

-- ============================================================================
-- COMMENTS: Documentation
-- ============================================================================
-- Table comment
COMMENT ON TABLE event_volunteers IS 'Volunteer assignments for events and sessions. Tracks who is assigned to help with event operations (scanning QR codes, ushering, coordinating, etc.). Supports both event-level and session-specific assignments.';

-- Column comments: Basic Info
COMMENT ON COLUMN event_volunteers.id IS 'Primary key, auto-incrementing integer';

-- Column comments: Event Reference
COMMENT ON COLUMN event_volunteers.event_id IS 'Reference to parent event (required). Cascade delete when event is deleted';

COMMENT ON COLUMN event_volunteers.session_id IS 'Reference to specific session if volunteer is assigned to a session (optional)';

-- Column comments: Volunteer Info
COMMENT ON COLUMN event_volunteers.community_id IS 'Community ID of the volunteer';

COMMENT ON COLUMN event_volunteers.role IS 'Volunteer role: scanner (QR code scanning), usher (seating/directing), coordinator (overall management), registration-desk (check-in desk), parking (parking management), security (security), tech-support (technical support), hospitality (welcoming/refreshments), custom (custom role)';

-- Column comments: Assignment Info
COMMENT ON COLUMN event_volunteers.assigned_at IS 'When the volunteer was assigned to this role';

COMMENT ON COLUMN event_volunteers.assigned_by_community_id IS 'Community ID of the person who assigned this volunteer (for audit trail)';

-- Column comments: Metadata
COMMENT ON COLUMN event_volunteers.created_at IS 'Timestamp when record was created';

COMMENT ON COLUMN event_volunteers.updated_at IS 'Timestamp when record was last updated';