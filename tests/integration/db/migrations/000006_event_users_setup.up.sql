SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "event_users" (
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "account_number" varchar(15) UNIQUE NOT NULL,
    "name" varchar(100) NOT NULL,
    "phone_number" varchar(15) NULL,
    "email" varchar(50) NULL,
    "password" varchar(255),
    "address" varchar(255) NOT NULL,
    "state" varchar(20) NOT NULL,
    "status" varchar(20) NOT NULL,
    "role" varchar(20) NOT NULL,
    "token" varchar(255),
    "gender" varchar(20) NULL,           -- Added gender field
    "marriage_status" varchar(20) NULL,  -- Added marriage status field
    "department" varchar(100) NULL,      -- Added department field
    "kkj" varchar(50) NULL,              -- Added KKJ field (boolean)
    "cool" varchar(100) NULL,            -- Added Cool field (boolean)
    "kom100" BOOLEAN DEFAULT FALSE,      -- Added KOM100 field (boolean)
    "baptis" BOOLEAN DEFAULT FALSE,      -- Added Baptis field (boolean)
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMP
);

CREATE INDEX idx_event_users_account_number ON "event_users" ("account_number");
CREATE INDEX idx_event_users_email ON "event_users" ("email");
