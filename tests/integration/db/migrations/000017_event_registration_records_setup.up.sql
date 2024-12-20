SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "event_registration_records" (
    "id" UUID NOT NULL PRIMARY KEY,
    "name" varchar(100) NOT NULL,
    "identifier" varchar(100) NULL,
    "community_id" varchar(15) NULL,
    "event_code" varchar(30) NOT NULL,
    "instance_code" varchar(30) NOT NULL,
    "identifier_origin" varchar(100) NOT NULL,
    "community_id_origin" varchar(15) NULL,
    "updated_by" varchar(100) NULL,
    "status" varchar(20) NOT NULL,
    "registered_at" timestamptz NOT NULL,
    "verified_at" timestamptz NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMP
);