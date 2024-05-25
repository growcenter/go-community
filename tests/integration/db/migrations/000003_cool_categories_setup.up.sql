CREATE TABLE "cool_categories" (
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "code" varchar(3) UNIQUE NOT NULL,
    "name" varchar(255) UNIQUE NOT NULL,
    "age_start" INT UNIQUE NOT NULL,
    "age_end" INT UNIQUE NOT NULL,
    "status" varchar(8) NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now()
);
