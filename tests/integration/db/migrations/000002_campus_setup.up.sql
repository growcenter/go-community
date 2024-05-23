CREATE TABLE "campus" (
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "code" varchar(3) UNIQUE NOT NULL,
    "region" varchar(10) UNIQUE NOT NULL,
    "name" varchar(255) UNIQUE NOT NULL,
    "location" varchar(255) UNIQUE NOT NULL,
    "address" varchar(255) NOT NULL,
    "status" varchar(8) NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now()
);
