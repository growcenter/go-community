SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "locations" (
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "code" varchar(5) UNIQUE NOT NULL,
    "campus_code" varchar(3) NOT NULL,
    "name" varchar(255) UNIQUE NOT NULL,
    "region" varchar(10) NOT NULL,
    "status" varchar(8) NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now()
);
