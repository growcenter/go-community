SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "event_generals" (
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "code" varchar(30) UNIQUE NOT NULL,
    "name" varchar(255) NOT NULL,
    "campus_code" varchar(3) NOT NULL,
    "description" varchar(255),
    "open_registration" TIMESTAMPTZ NOT NULL,
    "closed_registration" TIMESTAMPTZ NOT NULL,
    "status" varchar(8) NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMP
);