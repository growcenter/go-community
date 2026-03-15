SET
    TIME ZONE 'Asia/Jakarta';

-- ============================================================================
-- TABLE: event_series
-- Description: Event series/collections for grouping related events
-- Supports: Conference series, Training programs, Seasonal events
-- ============================================================================
CREATE TABLE event_series (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(20) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    image_url TEXT,
    created_by_community_id VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW (),
    updated_at TIMESTAMPTZ DEFAULT NOW ()
);

-- ============================================================================
-- INDEXES: Performance Optimization
-- ============================================================================
-- Primary lookup indexes
CREATE INDEX idx_series_creator ON event_series (created_by_community_id);

CREATE INDEX idx_series_status ON event_series (status)
WHERE
    status = 'active';

-- Composite indexes for common query patterns
-- Pattern: Creator's active series
CREATE INDEX idx_series_creator_status ON event_series (created_by_community_id, status, created_at DESC);

-- ============================================================================
-- CONSTRAINTS: Data Integrity
-- ============================================================================
-- Status validation
ALTER TABLE event_series ADD CONSTRAINT chk_series_status CHECK (status IN ('active', 'archived', 'draft'));

-- ============================================================================
-- COMMENTS: Documentation
-- ============================================================================
-- Table comment
COMMENT ON TABLE event_series IS 'Event series or collections for grouping related events. Examples: "Leadership Conference 2024-2026" (multi-year series), "Discipleship Training Program" (course series), "Christmas Services" (seasonal series). Events reference series via series_id.';

-- Column comments: Basic Info
COMMENT ON COLUMN event_series.id IS 'Primary key, auto-incrementing integer';

COMMENT ON COLUMN event_series.code IS 'Unique series code (e.g., SER-ABC123)';

-- Column comments: Series Info
COMMENT ON COLUMN event_series.title IS 'Series title (e.g., "Leadership Conference Series", "Discipleship Training Program")';

COMMENT ON COLUMN event_series.description IS 'Detailed description of the series, its purpose, and what events it includes';

COMMENT ON COLUMN event_series.image_url IS 'Series banner/logo image URL';

-- Column comments: Creator
COMMENT ON COLUMN event_series.created_by_community_id IS 'Community ID of the person who created this series';

-- Column comments: Status
COMMENT ON COLUMN event_series.status IS 'Series status: active (currently active), archived (past series, kept for reference), draft (being planned)';

-- Column comments: Metadata
COMMENT ON COLUMN event_series.created_at IS 'Timestamp when record was created';

COMMENT ON COLUMN event_series.updated_at IS 'Timestamp when record was last updated';