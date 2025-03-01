SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "events" (
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "code" varchar(7) UNIQUE NOT NULL,
    "title" varchar(255) NOT NULL,
    "topics" TEXT[],
    "description" TEXT,
    "terms_and_conditions" TEXT,
    "allowed_for" varchar(7) NOT NULL,
    "allowed_users" TEXT[],
    "allowed_roles" TEXT[],
    "allowed_campuses" TEXT[],
    "is_recurring" BOOL DEFAULT FALSE,
    "recurrence" VARCHAR(20),
    "event_start_at" TIMESTAMPTZ,
    "event_end_at" TIMESTAMPTZ,
    "register_start_at" TIMESTAMPTZ,
    "register_end_at" TIMESTAMPTZ,
    "location_type" varchar(6) not null,
    "location_name" varchar(255) NOT NULL,
    "image_links" TEXT[],
    "status" varchar(8) NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMP
);