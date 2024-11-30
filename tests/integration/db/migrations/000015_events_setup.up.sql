SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "events" (
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "code" varchar(30) UNIQUE NOT NULL,
    "title" varchar(255) NOT NULL,
    "location" varchar(255) NOT NULL,
    "description" varchar(255),
    "campus_code" TEXT[],
    "allowed_users" TEXT[],
    "allowed_roles" TEXT[],
    "is_recurring" BOOL DEFAULT FALSE,
    "recurrence" VARCHAR(20),
    "event_start_at" TIMESTAMPTZ,
    "event_end_at" TIMESTAMPTZ,
    "register_start_at" TIMESTAMPTZ,
    "register_end_at" TIMESTAMPTZ,
    "status" varchar(8) NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMP
);