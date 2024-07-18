CREATE TABLE "event_users" (
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "account_number" varchar(15) UNIQUE NOT NULL,
    "name" varchar(100) NOT NULL,
    "phone_number" varchar(15) NULL,
    "email" varchar(50) NULL,
    "password" varchar(255),
    "state" varchar(20) NOT NULL,
    "status" varchar(20) NOT NULL,
    "role" varchar(20) NOT NULL,
    "token" varchar(255),
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMP
);

CREATE INDEX idx_event_users_account_number ON "event_users" ("account_number");
CREATE INDEX idx_event_users_email ON "event_users" ("email");