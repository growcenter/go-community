SET TIME ZONE 'Asia/Jakarta';

-- ============================================================================
-- TABLE: events
-- Description: Main event container for church event management system
-- Supports: Sunday Services, Conferences, Volunteer Meetings, Announcements
-- ============================================================================

CREATE TABLE events (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(20) UNIQUE NOT NULL,

    -- Basic Info
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    pre_description TEXT,
    post_description JSONB,
    terms_and_conditions TEXT,
    category VARCHAR(30) NOT NULL, -- registration, attendance, announcement, volunteer, hybrid

    -- Media
    image_urls TEXT[],
    banner_url TEXT,

    -- Organization
    creator_community_id VARCHAR(50) NOT NULL REFERENCES users(community_id),
    organizer_community_ids TEXT[],
    contact_community_ids TEXT[],

    -- Visibility & Access
    access_level VARCHAR(30) NOT NULL DEFAULT 'public',
    allowed_user_types TEXT[],
    allowed_roles TEXT[],
    allowed_campuses TEXT[],
    allowed_community_ids TEXT[],

    -- Location (Default for all sessions)
    location_type VARCHAR(20) NOT NULL, -- online, offline, hybrid
    location_visibility VARCHAR(20) NOT NULL DEFAULT 'all', -- pre-registration, post-registration, all
    physical_place_name VARCHAR(100),
    physical_address TEXT,
    virtual_link TEXT,
    virtual_platform VARCHAR(50), -- youtube, zoom, meet, custom
    location_details TEXT,
    cta_text VARCHAR(50),
    cta_link TEXT,

    -- Schedule
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    timezone VARCHAR(50) NOT NULL DEFAULT 'Asia/Jakarta',
    
    -- Recurrence (For Sunday Service type)
    is_recurring BOOLEAN DEFAULT FALSE,
    recurrence_pattern JSONB, -- {type: "weekly", days: ["sunday"], interval: 1}
    recurrence_end_date TIMESTAMPTZ,

    -- Template/Series
    is_template BOOLEAN DEFAULT FALSE,
    template_id BIGINT REFERENCES events(id),
    series_id BIGINT REFERENCES event_series(id),

    -- Notification Config
    notification_channels TEXT[], -- email, sms, whatsapp
    reminder_config JSONB, -- {enabled: true, intervals: ["24h", "1h"]}

    -- Metadata
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- ============================================================================
-- INDEXES: Performance Optimization
-- ============================================================================

-- Primary lookup indexes (unique constraints already create indexes for code, slug)
CREATE INDEX idx_events_creator ON events(creator_community_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_events_series ON events(series_id) WHERE series_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_events_template ON events(template_id) WHERE template_id IS NOT NULL AND deleted_at IS NULL;

-- Composite indexes for common query patterns
-- Pattern: List published events ordered by date
CREATE INDEX idx_events_status_start_date ON events(status, start_date DESC) WHERE deleted_at IS NULL;

-- Pattern: Filter events by visibility and status (public events, member-only events)
CREATE INDEX idx_events_access_level_status_start ON events(access_level, status, start_date DESC) WHERE deleted_at IS NULL;

-- Pattern: Filter events by category and status (attendance events, registration events)
CREATE INDEX idx_events_category_status_start ON events(category, status, start_date DESC) WHERE deleted_at IS NULL;

-- Pattern: Find events by creator and status (organizer's event management)
CREATE INDEX idx_events_creator_status_start ON events(creator_community_id, status, start_date DESC) WHERE deleted_at IS NULL;

-- Pattern: Date range queries (upcoming events, past events)
CREATE INDEX idx_events_date_range ON events(start_date, end_date) WHERE deleted_at IS NULL AND status != 'draft';

-- Pattern: Recurring events management
CREATE INDEX idx_events_recurring ON events(is_recurring, start_date) WHERE is_recurring = TRUE AND deleted_at IS NULL;

-- Pattern: Template/series queries
CREATE INDEX idx_events_templates ON events(is_template) WHERE is_template = TRUE AND deleted_at IS NULL;

-- Soft delete support (queries excluding deleted records)
CREATE INDEX idx_events_deleted_at ON events(deleted_at) WHERE deleted_at IS NOT NULL;

-- Array column indexes for access control filtering
-- Pattern: Find events allowed for specific user types, roles, campuses
CREATE INDEX idx_events_allowed_user_types ON events USING GIN(allowed_user_types) WHERE deleted_at IS NULL;
CREATE INDEX idx_events_allowed_roles ON events USING GIN(allowed_roles) WHERE deleted_at IS NULL;
CREATE INDEX idx_events_allowed_campuses ON events USING GIN(allowed_campuses) WHERE deleted_at IS NULL;
CREATE INDEX idx_events_allowed_community_ids ON events USING GIN(allowed_community_ids) WHERE deleted_at IS NULL;
CREATE INDEX idx_events_organizer_community_ids ON events USING GIN(organizer_community_ids) WHERE deleted_at IS NULL;
CREATE INDEX idx_events_contact_community_ids ON events USING GIN(contact_community_ids) WHERE deleted_at IS NULL;

-- JSONB indexes for notification and recurrence pattern queries
CREATE INDEX idx_events_recurrence_pattern ON events USING GIN(recurrence_pattern) WHERE recurrence_pattern IS NOT NULL;
CREATE INDEX idx_events_reminder_config ON events USING GIN(reminder_config) WHERE reminder_config IS NOT NULL;

-- Full-text search on title and description (for event search functionality)
CREATE INDEX idx_events_search ON events USING GIN(to_tsvector('indonesian', COALESCE(title, '') || ' ' || COALESCE(description, ''))) WHERE deleted_at IS NULL;

-- ============================================================================
-- CONSTRAINTS: Data Integrity
-- ============================================================================

-- Category validation
ALTER TABLE events ADD CONSTRAINT chk_events_category 
    CHECK (category IN ('registration', 'attendance', 'announcement', 'volunteer', 'hybrid'));

-- Visibility validation
ALTER TABLE events ADD CONSTRAINT chk_events_visibility 
    CHECK (visibility IN ('public', 'members_only', 'volunteer_only', 'private', 'campus_specific'));

-- Location type validation
ALTER TABLE events ADD CONSTRAINT chk_events_location_type 
    CHECK (location_type IN ('online', 'offline', 'hybrid'));

-- Virtual platform validation
ALTER TABLE events ADD CONSTRAINT chk_events_virtual_platform 
    CHECK (virtual_platform IS NULL OR virtual_platform IN ('youtube', 'zoom', 'meet', 'custom'));

-- Status validation
ALTER TABLE events ADD CONSTRAINT chk_events_status 
    CHECK (status IN ('draft', 'published', 'cancelled', 'completed'));

-- Date logic validation
ALTER TABLE events ADD CONSTRAINT chk_events_dates 
    CHECK (end_date > start_date);

-- Recurrence validation
ALTER TABLE events ADD CONSTRAINT chk_events_recurrence_end_date 
    CHECK (
        (is_recurring = FALSE AND recurrence_end_date IS NULL) OR
        (is_recurring = TRUE AND recurrence_end_date IS NULL) OR
        (is_recurring = TRUE AND recurrence_end_date > start_date)
    );

-- Template validation (templates cannot be recurring)
ALTER TABLE events ADD CONSTRAINT chk_events_template_not_recurring 
    CHECK (NOT (is_template = TRUE AND is_recurring = TRUE));

-- Self-reference validation (event cannot be its own template)
ALTER TABLE events ADD CONSTRAINT chk_events_template_self_reference 
    CHECK (template_id IS NULL OR template_id != id);

-- ============================================================================
-- COMMENTS: Documentation
-- ============================================================================

-- Table comment
COMMENT ON TABLE events IS 'Main event container for church event management system. Supports diverse event types including Sunday Services (recurring attendance), Christmas Celebrations (multi-session registration), Conferences (hierarchical sessions), Volunteer Meetings, and Announcements.';

-- Column comments: Basic Info
COMMENT ON COLUMN events.id IS 'Primary key, auto-incrementing event ID';
COMMENT ON COLUMN events.code IS 'Unique event code for QR generation and public reference (e.g., EVT-XMS24, EVT-SUN01)';
COMMENT ON COLUMN events.title IS 'Event title displayed to users (max 255 chars)';
COMMENT ON COLUMN events.slug IS 'URL-friendly slug for event pages (e.g., christmas-celebration-2026)';
COMMENT ON COLUMN events.pre_description IS 'Detailed event description supporting markdown/rich text';
COMMENT ON COLUMN events.post_description IS 'Detailed event description supporting markdown/rich text';
COMMENT ON COLUMN events.terms_and_conditions IS 'Detailed event description supporting markdown/rich text';
COMMENT ON COLUMN events.category IS 'Event category: registration (standard events), attendance (recurring tracking like Sunday Service), announcement (info-only), volunteer (internal events), hybrid (registration + attendance)';

-- Column comments: Media
COMMENT ON COLUMN events.image_url IS 'Event thumbnail/card image URL (recommended: 16:9 aspect ratio)';
COMMENT ON COLUMN events.banner_url IS 'Event banner/hero image URL (recommended: 21:9 aspect ratio)';

-- Column comments: Organization
COMMENT ON COLUMN events.creator_community_id IS 'Community ID of the user who created this event (references users.community_id)';
COMMENT ON COLUMN events.organizer_community_ids IS 'Array of community IDs for additional organizers who can manage this event';
COMMENT ON COLUMN events.contact_community_ids IS 'Array of community IDs for contact person for this event';

-- Column comments: Visibility & Access
COMMENT ON COLUMN events.access_level IS 'Event visibility level: public (everyone including guests), members_only (authenticated users), volunteer_only (church volunteers + staff), private (invited users only), campus_specific (specific campus users)';
COMMENT ON COLUMN events.allowed_user_types IS 'Array of user types allowed to view/register (e.g., [''member'', ''volunteer'']). NULL = all types allowed';
COMMENT ON COLUMN events.allowed_roles IS 'Array of roles allowed to view/register (e.g., [''youth_leader'', ''worship_team'']). NULL = all roles allowed';
COMMENT ON COLUMN events.allowed_campuses IS 'Array of campus codes allowed to view/register (e.g., [''JKT'', ''BDG'']). NULL = all campuses allowed';
COMMENT ON COLUMN events.allowed_community_ids IS 'Array of specific community IDs allowed (for private/invite-only events). NULL = no specific restriction';

-- Column comments: Location
COMMENT ON COLUMN events.location_type IS 'Location type: online (virtual only), offline (physical only), hybrid (both virtual and physical)';
COMMENT ON COLUMN events.physical_address IS 'Full physical address for offline/hybrid events';
COMMENT ON COLUMN events.virtual_link IS 'Virtual meeting/streaming link (YouTube, Zoom, Google Meet, etc.)';
COMMENT ON COLUMN events.virtual_platform IS 'Virtual platform identifier: youtube, zoom, meet, custom';
COMMENT ON COLUMN events.location_details IS 'Additional location information (parking, room number, landmarks, etc.)';

-- Column comments: Schedule
COMMENT ON COLUMN events.timezone IS 'IANA timezone identifier (e.g., Asia/Jakarta, America/New_York) for accurate time display';
COMMENT ON COLUMN events.start_date IS 'Event start date and time (timezone-aware)';
COMMENT ON COLUMN events.end_date IS 'Event end date and time (timezone-aware). Must be after start_date';

-- Column comments: Recurrence
COMMENT ON COLUMN events.is_recurring IS 'Whether this event recurs (e.g., Sunday Service). If true, occurrences are auto-generated';
COMMENT ON COLUMN events.recurrence_pattern IS 'JSONB recurrence configuration. Example: {"type": "weekly", "days": ["sunday"], "interval": 1} for weekly Sunday Service';
COMMENT ON COLUMN events.recurrence_end_date IS 'Optional end date for recurring events. NULL = recurs indefinitely';

-- Column comments: Template/Series
COMMENT ON COLUMN events.is_template IS 'Whether this event is a template for creating similar events (cannot be recurring)';
COMMENT ON COLUMN events.template_id IS 'Reference to template event if this was created from a template';
COMMENT ON COLUMN events.series_id IS 'Reference to event_series for grouping related events (e.g., "Leadership Training Series")';

-- Column comments: Notification Config
COMMENT ON COLUMN events.notification_channels IS 'Array of enabled notification channels: email, sms, whatsapp. NULL = no notifications';
COMMENT ON COLUMN events.reminder_config IS 'JSONB reminder configuration. Example: {"enabled": true, "intervals": ["24h", "1h"]} sends reminders 24 hours and 1 hour before event';

-- Column comments: Metadata
COMMENT ON COLUMN events.status IS 'Event lifecycle status: draft (not visible to public), published (active and visible), cancelled (event cancelled), completed (event finished)';
COMMENT ON COLUMN events.created_at IS 'Timestamp when event was created';
COMMENT ON COLUMN events.updated_at IS 'Timestamp when event was last updated';
COMMENT ON COLUMN events.deleted_at IS 'Soft delete timestamp. NULL = active, NOT NULL = deleted';