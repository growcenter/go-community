CREATE TABLE "user_types" (
    "id" BIGSERIAL PRIMARY KEY,
    "type" varchar(50) UNIQUE NOT NULL,
    "name" varchar(100) NOT NULL,
    "roles" TEXT[] NOT NULL,
    "description" text,
    "category" varchar(20) NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMP
);
