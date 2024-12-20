CREATE TABLE "roles" (
    "id" BIGSERIAL PRIMARY KEY,
    "role" varchar(50) UNIQUE NOT NULL,
    "description" text,
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMP
);
