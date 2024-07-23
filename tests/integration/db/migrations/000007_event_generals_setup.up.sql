CREATE TABLE "event_generals" (
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "code" varchar(3) UNIQUE NOT NULL,
    "name" varchar(255) UNIQUE NOT NULL,
    "campus_code" varchar(3) UNIQUE NOT NULL,
    "description" varchar(255),
    "open_registration" TIMESTAMP NOT NULL,
    "closed_registration" TIMESTAMP NOT NULL,
    "status" varchar(8) NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMP
);