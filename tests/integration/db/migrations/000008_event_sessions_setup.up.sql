SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "event_sessions" (
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "code" varchar(30) UNIQUE NOT NULL,
    "name" varchar(255) NOT NULL,
    "event_code" varchar(30) NOT NULL,
    "description" varchar(255),
    "time" TIMESTAMPTZ NOT NULL,
    "max_seating" INT NOT NULL,
    "available_seats" INT NOT NULL,
    "registered_seats" INT NOT NULL,
    "scanned_seats" INT NOT NULL,
    "status" varchar(8) NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMP
);