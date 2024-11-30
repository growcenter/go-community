SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "event_instances" (
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "code" varchar(30) UNIQUE NOT NULL,
    "title" varchar(255) NOT NULL,
    "location" varchar(255) NOT NULL,
    "event_code" varchar(30) NOT NULL,
    "instance_start_at" TIMESTAMPTZ,
    "instance_end_at" TIMESTAMPTZ,
    "register_start_at" TIMESTAMPTZ,
    "register_end_at" TIMESTAMPTZ,
    "description" varchar(255),
    "max_register" INT NOT NULL,
    "total_seats" INT NOT NULL,
    "booked_seats" INT NOT NULL,
    "scanned_seats" INT NOT NULL,
    "is_required" BOOLEAN DEFAULT TRUE,
    "status" varchar(8) NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMP
);