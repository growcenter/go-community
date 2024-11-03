SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "users" (
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "community_id" varchar(15) UNIQUE NOT NULL,
    "name" varchar(100) UNIQUE NOT NULL,
    "phone_number" varchar(15) UNIQUE NOT NULL,
    "email" varchar(50) UNIQUE NOT NULL,
    "password" varchar(20) NOT NULL,
    "user_type" varchar(20) NOT NULL,
    "status" varchar(20) NOT NULL,
    "roles" varchar(20) NOT NULL,
    "token" varchar(255),
    "gender" varchar(6) NOT NULL,
    "address" text,
    "campus_code" VARCHAR(3) NOT NULL,
    "cool_category_code" VARCHAR(3),
    "cool_id" INT,
    "department" VARCHAR(50),
    "date_of_birth" DATE,
    "place_of_birth" VARCHAR(100),
    "marital_status" VARCHAR(20) NOT NULL,
    "date_of_marriage" DATE,
    "employment_status" VARCHAR(50),
    "education_level" VARCHAR(50),
    "kkj_number" VARCHAR(50),
    "jemaat_id" VARCHAR(50),
    "is_baptized" BOOLEAN DEFAULT FALSE,
    "is_kom100" BOOLEAN DEFAULT FALSE,
    "age" INT NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMP,

    FOREIGN KEY ("campus_code") REFERENCES "campus"("code"),
    FOREIGN KEY ("cool_category_code") REFERENCES "cool_categories"("code")
);

CREATE INDEX idx_users_account_number ON "users" ("account_number");
CREATE INDEX idx_users_email ON "users" ("email");
