SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "cools" (
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "name" varchar(255) UNIQUE NOT NULL,
    "status" varchar(8) NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now()
);
