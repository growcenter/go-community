SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "events" (
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "code" varchar(7) UNIQUE NOT NULL,
    "title" varchar(255) NOT NULL,
    "category" varchar(255) NOT NULL,
    "redirect_url" varchar(255),
    "topics" TEXT[],
    "description" TEXT,
    "terms_and_conditions" TEXT,
    "image_links" TEXT[],
    "slug" varchar(255),
    "created_by" varchar(255) NOT NULL,
    "location_type" varchar(10) NOT NULL,
    "location_offline_venue" varchar(255),
    "location_online_link" varchar(255),
    "location_detail" TEXT,
    "visibility" varchar(10) NOT NULL,
    "location_visibility" varchar(255),
    "allowed_community_ids" TEXT[],
    "allowed_user_types" TEXT[],
    "allowed_roles" TEXT[],
    "allowed_campuses" TEXT[],
    "organizer_community_ids" TEXT[],
    "contact_community_ids" TEXT[],
    "is_recurring" BOOL DEFAULT FALSE,
    "start_at" TIMESTAMPTZ NOT NULL,
    "end_at" TIMESTAMPTZ NOT NULL,
    "timezone" TEXT NOT NULL,
    "post_details" JSONB DEFAULT '{}',
    "status" varchar(8) NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMP
);

-- Index for looking up events by their URL-friendly slug
CREATE INDEX idx_events_slug ON "events" ("slug");
CREATE INDEX idx_events_code ON "events" ("code");

-- Composite index to optimize filtering and sorting for event lists
CREATE INDEX idx_events_status_visibility_start_at ON "events" ("status", "visibility", "start_at" DESC);

-- GIN indexes for efficient searching within array columns (for access control)
CREATE INDEX idx_events_gin_allowed_roles ON "events" USING GIN ("allowed_roles");
CREATE INDEX idx_events_gin_allowed_user_types ON "events" USING GIN ("allowed_user_types");
CREATE INDEX idx_events_gin_allowed_community_ids ON "events" USING GIN ("allowed_community_ids");
CREATE INDEX idx_events_gin_allowed_campuses ON "events" USING GIN ("allowed_campuses");

-- Index to efficiently filter out soft-deleted records
CREATE INDEX idx_events_deleted_at ON "events" ("deleted_at");