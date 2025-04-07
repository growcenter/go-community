SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "event_instances" (
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "code" varchar(15) UNIQUE NOT NULL,
    "event_code" varchar(7) NOT NULL,
    "title" varchar(255) NOT NULL,
    "description" TEXT,
    "instance_start_at" TIMESTAMPTZ,
    "instance_end_at" TIMESTAMPTZ,
    "register_start_at" TIMESTAMPTZ,
    "register_end_at" TIMESTAMPTZ,
    "allow_verify_at" TIMESTAMPTZ,
    "disallow_verify_at" TIMESTAMPTZ,
    "location_type" varchar(6) not null,
    "location_name" varchar(255) NOT NULL,
    "max_per_transaction" INT NOT NULL,
    "is_one_per_account" BOOLEAN DEFAULT FALSE,
    "is_one_per_ticket" BOOLEAN DEFAULT FALSE,
    "register_flow" VARCHAR(8),
    "check_type" VARCHAR(9),
    "total_seats" INT NOT NULL,
    "booked_seats" INT NOT NULL,
    "scanned_seats" INT NOT NULL,
    "status" varchar(8) NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMP
);