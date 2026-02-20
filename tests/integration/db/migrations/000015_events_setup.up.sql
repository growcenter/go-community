SET TIME ZONE 'Asia/Jakarta';

-- ============================================================================
-- TABLE: events
-- Description: Main event container for church event management system
-- Aligns with: models.Event
-- ============================================================================

CREATE TABLE events (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,   -- VARCHAR(50) matches gorm tag

    -- Core Info
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,  -- URL-safe, auto-generated from code when omitted
    pre_description TEXT,               -- shown before registration
    post_description JSONB,             -- shown after registration (rich content / JSONB)
    terms_and_conditions TEXT,
    category VARCHAR(30) NOT NULL,      -- announcement, registerable

    -- Media
    image_urls TEXT[],                  -- array of image URLs (pq.StringArray)
    banner_url TEXT,

    -- Organization
    creator_community_id VARCHAR(50) NOT NULL,
    organizer_community_ids TEXT[],
    contact_community_ids TEXT[],

    -- Visibility & Access
    access_level VARCHAR(20) NOT NULL DEFAULT 'public',
    -- public, private, members_only, volunteer_only, campus_specific
    allowed_user_types TEXT[],
    allowed_roles TEXT[],
    allowed_campuses TEXT[],
    allowed_community_ids TEXT[],

    -- Location (default, inherited by sessions)
    location_type VARCHAR(20) NOT NULL,             -- online, offline, hybrid
    location_visibility VARCHAR(30) NOT NULL DEFAULT 'all',
    -- all, pre-registration, post-registration
    physical_place_name TEXT,
    physical_address TEXT,
    virtual_link TEXT,
    virtual_platform TEXT,
    location_details TEXT,
    cta_text VARCHAR(100),
    cta_link TEXT,

    -- Schedule
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    timezone VARCHAR(50) NOT NULL DEFAULT 'Asia/Jakarta',

    -- Recurrence (occurrences calculated on-demand — no pre-generated rows)
    is_recurring BOOLEAN NOT NULL DEFAULT FALSE,
    recurrence_pattern JSONB,
    -- { frequency, interval, weekDays, count, endDate,
    --   monthlyPattern, excludeDates, additionalDates }

    -- Template / Series (soft references — no FK)
    is_template BOOLEAN NOT NULL DEFAULT FALSE,
    template_id VARCHAR(255),
    series_id VARCHAR(255),

    -- Registration
    session_per_user INTEGER NOT NULL DEFAULT 0,    -- 0 = unlimited

    -- Notifications
    notification_channels TEXT[],
    reminder_config JSONB,

    -- Status & Metadata
    status VARCHAR(20) NOT NULL DEFAULT 'draft',    -- draft, active, inactive
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- ============================================================================
-- INDEXES: Performance Optimization
-- ============================================================================

-- Composite index for common listing queries
CREATE INDEX idx_events_status_start ON events(status, start_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_events_access_level_status ON events(access_level, status, start_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_events_category_status ON events(category, status, start_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_events_creator_status ON events(creator_community_id, status, start_at DESC) WHERE deleted_at IS NULL;

-- Lookup helpers
CREATE INDEX idx_events_creator ON events(creator_community_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_events_date_range ON events(start_at, end_at) WHERE deleted_at IS NULL AND status != 'draft';
CREATE INDEX idx_events_recurring ON events(is_recurring, start_at) WHERE is_recurring = TRUE AND deleted_at IS NULL;
CREATE INDEX idx_events_templates ON events(is_template) WHERE is_template = TRUE AND deleted_at IS NULL;

-- Soft-delete
CREATE INDEX idx_events_deleted_at ON events(deleted_at) WHERE deleted_at IS NOT NULL;

-- Access control array columns (GIN for @> / && queries)
CREATE INDEX idx_events_allowed_user_types ON events USING GIN(allowed_user_types) WHERE deleted_at IS NULL;
CREATE INDEX idx_events_allowed_roles ON events USING GIN(allowed_roles) WHERE deleted_at IS NULL;
CREATE INDEX idx_events_allowed_campuses ON events USING GIN(allowed_campuses) WHERE deleted_at IS NULL;
CREATE INDEX idx_events_allowed_community_ids ON events USING GIN(allowed_community_ids) WHERE deleted_at IS NULL;
CREATE INDEX idx_events_organizer_community_ids ON events USING GIN(organizer_community_ids) WHERE deleted_at IS NULL;
CREATE INDEX idx_events_contact_community_ids ON events USING GIN(contact_community_ids) WHERE deleted_at IS NULL;

-- JSONB
CREATE INDEX idx_events_recurrence_pattern ON events USING GIN(recurrence_pattern) WHERE recurrence_pattern IS NOT NULL;
CREATE INDEX idx_events_reminder_config ON events USING GIN(reminder_config) WHERE reminder_config IS NOT NULL;

-- Full-text search (title + pre_description)
CREATE INDEX idx_events_search ON events USING GIN(
    to_tsvector('indonesian', COALESCE(title, '') || ' ' || COALESCE(pre_description, ''))
) WHERE deleted_at IS NULL;

-- ============================================================================
-- CONSTRAINTS: Data Integrity
-- ============================================================================

ALTER TABLE events ADD CONSTRAINT chk_events_category
    CHECK (category IN ('announcement', 'registerable'));

ALTER TABLE events ADD CONSTRAINT chk_events_access_level
    CHECK (access_level IN ('public', 'private', 'members_only', 'volunteer_only', 'campus_specific'));

ALTER TABLE events ADD CONSTRAINT chk_events_location_type
    CHECK (location_type IN ('online', 'offline', 'hybrid'));

ALTER TABLE events ADD CONSTRAINT chk_events_location_visibility
    CHECK (location_visibility IN ('all', 'pre-registration', 'post-registration'));

ALTER TABLE events ADD CONSTRAINT chk_events_status
    CHECK (status IN ('draft', 'active', 'inactive'));

ALTER TABLE events ADD CONSTRAINT chk_events_dates
    CHECK (end_at > start_at);

ALTER TABLE events ADD CONSTRAINT chk_events_recurrence
    CHECK (is_recurring = FALSE OR recurrence_pattern IS NOT NULL);

ALTER TABLE events ADD CONSTRAINT chk_events_template_not_recurring
    CHECK (NOT (is_template = TRUE AND is_recurring = TRUE));

-- ============================================================================
-- AUTO-UPDATE TRIGGER
-- ============================================================================

CREATE OR REPLACE FUNCTION update_events_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_events_updated_at
    BEFORE UPDATE ON events
    FOR EACH ROW EXECUTE FUNCTION update_events_updated_at();

-- ============================================================================
-- COMMENTS: Documentation
-- ============================================================================

COMMENT ON TABLE events IS 'Main event container. Supports Sunday Services, Christmas events, Conferences, Volunteer Meetings, and Announcements.';

COMMENT ON COLUMN events.code IS 'Unique event code. Used for QR codes and API lookup. VARCHAR(50).';
COMMENT ON COLUMN events.slug IS 'URL-safe slug. Auto-generated from event code when not provided.';
COMMENT ON COLUMN events.pre_description IS 'Description shown before registration.';
COMMENT ON COLUMN events.post_description IS 'JSONB content shown after registration.';
COMMENT ON COLUMN events.category IS 'announcement | registerable';
COMMENT ON COLUMN events.access_level IS 'public | private | members_only | volunteer_only | campus_specific';
COMMENT ON COLUMN events.location_visibility IS 'all | pre-registration | post-registration';
COMMENT ON COLUMN events.is_recurring IS 'True = occurrences calculated on-demand from recurrence_pattern. No separate occurrence rows stored.';
COMMENT ON COLUMN events.recurrence_pattern IS 'JSONB: { frequency, interval, weekDays, count, endDate, monthlyPattern, excludeDates, additionalDates }';
COMMENT ON COLUMN events.template_id IS 'Soft reference to template event (VARCHAR, no FK).';
COMMENT ON COLUMN events.series_id IS 'Soft reference to event series (VARCHAR, no FK).';
COMMENT ON COLUMN events.session_per_user IS '0 = unlimited. Max sessions a user can register for within this event.';
COMMENT ON COLUMN events.status IS 'draft | active | inactive';