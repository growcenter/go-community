SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE IF NOT EXISTS "events" (
    -- Primary Key
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    
    -- Core Event Information
    "code" VARCHAR(50) UNIQUE NOT NULL,
    "title" VARCHAR(255) NOT NULL,
    "slug" VARCHAR(255) UNIQUE NOT NULL,
    
    -- Content & Media
    "topics" TEXT[],
    "category" VARCHAR(50) NOT NULL,
    "description" TEXT,
    "terms_and_conditions" TEXT,
    "image_links" TEXT[],
    "post_registration_details" JSONB,
    
    -- Creator & Organizers
    "creator_community_id" VARCHAR(50) NOT NULL,
    "organizer_community_ids" TEXT[],
    "contact_community_ids" TEXT[],
    
    -- Location Information
    "location_type" VARCHAR(20) NOT NULL,
    "physical_address" TEXT,
    "virtual_link" TEXT,
    "meeting_cta_text" VARCHAR(100),
    "location_details" TEXT,
    "location_visibility" VARCHAR(30),
    
    -- Access Control
    "access_level" VARCHAR(20) NOT NULL,
    "allowed_user_types" TEXT[],
    "allowed_roles" TEXT[],
    "allowed_campuses" TEXT[],
    "allowed_community_ids" TEXT[],
    
    -- Scheduling
    "recurrence" VARCHAR(255),
    "start_at" TIMESTAMPTZ NOT NULL,
    "end_at" TIMESTAMPTZ NOT NULL,
    "timezone" VARCHAR(50) NOT NULL,
    
    -- Confirmation
    "confirmation_method" VARCHAR(20),
    
    -- Status & Metadata
    "status" VARCHAR(20) NOT NULL DEFAULT 'draft',
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "deleted_at" TIMESTAMPTZ
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS "idx_events_code" ON "events" ("code");
CREATE INDEX IF NOT EXISTS "idx_events_slug" ON "events" ("slug");
CREATE INDEX IF NOT EXISTS "idx_events_title" ON "events" ("title");
CREATE INDEX IF NOT EXISTS "idx_events_creator" ON "events" ("creator_community_id");
CREATE INDEX IF NOT EXISTS "idx_events_access_level" ON "events" ("access_level");
CREATE INDEX IF NOT EXISTS "idx_events_status" ON "events" ("status");
CREATE INDEX IF NOT EXISTS "idx_events_start_at" ON "events" ("start_at");
CREATE INDEX IF NOT EXISTS "idx_events_end_at" ON "events" ("end_at");
CREATE INDEX IF NOT EXISTS "idx_events_deleted_at" ON "events" ("deleted_at");
CREATE INDEX IF NOT EXISTS "idx_events_category" ON "events" ("category");

-- Composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS "idx_events_status_start" ON "events" ("status", "start_at");
CREATE INDEX IF NOT EXISTS "idx_events_access_status" ON "events" ("access_level", "status");

-- Check constraints for data integrity
ALTER TABLE "events" ADD CONSTRAINT "chk_events_category" 
    CHECK ("category" IN ('announcement', 'registerable'));

ALTER TABLE "events" ADD CONSTRAINT "chk_events_location_type" 
    CHECK ("location_type" IN ('online', 'offline', 'hybrid'));

ALTER TABLE "events" ADD CONSTRAINT "chk_events_access_level" 
    CHECK ("access_level" IN ('public', 'private'));

ALTER TABLE "events" ADD CONSTRAINT "chk_events_status" 
    CHECK ("status" IN ('draft', 'published', 'cancelled', 'completed'));

ALTER TABLE "events" ADD CONSTRAINT "chk_events_dates" 
    CHECK ("end_at" > "start_at");

ALTER TABLE "events" ADD CONSTRAINT "chk_events_location_visibility"
    CHECK ("location_visibility" IS NULL OR "location_visibility" IN ('all', 'pre-registration', 'post-registration'));

ALTER TABLE "events" ADD CONSTRAINT "chk_events_confirmation_method"
    CHECK ("confirmation_method" IS NULL OR "confirmation_method" IN ('whatsapp', 'email', 'both'));

ALTER TABLE "events" ADD CONSTRAINT "chk_events_recurrence"
    CHECK ("recurrence" IS NULL OR "recurrence" IN ('', 'daily', 'weekly', 'monthly', 'yearly') OR LENGTH("recurrence") > 10);

-- Comment on table
COMMENT ON TABLE "events" IS 'Stores church events with comprehensive scheduling and access control';

-- Column comments for documentation
COMMENT ON COLUMN "events"."code" IS 'Unique event code (e.g., EVT12345)';
COMMENT ON COLUMN "events"."slug" IS 'URL-friendly slug for the event';
COMMENT ON COLUMN "events"."category" IS 'Event category: announcement or registerable';
COMMENT ON COLUMN "events"."post_registration_details" IS 'JSONB details shown after registration';
COMMENT ON COLUMN "events"."location_type" IS 'Location type: online, offline, or hybrid';
COMMENT ON COLUMN "events"."location_visibility" IS 'When to show location: all, pre-registration, or post-registration';
COMMENT ON COLUMN "events"."access_level" IS 'Access level: public or private';
COMMENT ON COLUMN "events"."recurrence" IS 'Recurrence pattern: daily, weekly, monthly, yearly, or cron expression';
COMMENT ON COLUMN "events"."timezone" IS 'IANA timezone (e.g., Asia/Jakarta)';
COMMENT ON COLUMN "events"."status" IS 'Event status: draft, published, cancelled, or completed';