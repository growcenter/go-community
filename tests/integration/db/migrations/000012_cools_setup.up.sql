SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "cools" (
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "code" varchar(30) UNIQUE NOT NULL,
    "name" varchar(255) UNIQUE NOT NULL,
    "description" TEXT,
    "campus_code" varchar(3) NOT NULL,
    "facilitator_community_ids" TEXT[] NOT NULL,
    "leader_community_ids" TEXT[] NOT NULL,
    "core_community_ids" TEXT[],
    "category" varchar(40) NOT NULL,
    "gender" varchar(6),
    "recurrence" VARCHAR(50),
    "location_type" varchar(10),
    "location_area" varchar(255),
    "location_district" varchar(255),
    "status" varchar(8) NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMP
);