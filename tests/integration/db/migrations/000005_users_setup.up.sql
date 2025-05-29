SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "users" (
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "community_id" varchar(15) UNIQUE NOT NULL,
    "name" varchar(100) NOT NULL,
    "phone_number" varchar(15) NULL,
    "email" varchar(50) NULL,
    "password" varchar(70) NOT NULL,
    "user_types" TEXT[] NOT NULL,
    "roles" TEXT[],
    "status" varchar(20) NOT NULL,
    "token" varchar(255),
    "gender" varchar(6) NOT NULL,
    "address" text,
    "campus_code" VARCHAR(3) NOT NULL,
    "cool_category_code" VARCHAR(3),
    "cool_code" varchar(30),
    "department" VARCHAR(50),
    "date_of_birth" TIMESTAMP,
    "place_of_birth" VARCHAR(100),
    "marital_status" VARCHAR(20) NOT NULL,
    "date_of_marriage" TIMESTAMP,
    "employment_status" VARCHAR(50),
    "education_level" VARCHAR(50),
    "kkj_number" VARCHAR(50),
    "jemaat_id" VARCHAR(50),
    "is_baptized" BOOLEAN DEFAULT FALSE,
    "is_kom100" BOOLEAN DEFAULT FALSE,
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMP
);

CREATE INDEX idx_users_community_id ON "users" ("community_id");
CREATE INDEX idx_users_email ON "users" ("email");
