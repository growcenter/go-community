SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "event_registrations" (
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "name" varchar(100) NOT NULL,
    "identifier" varchar(100) NULL,
    "address" varchar(255) NOT NULL,
    "account_number" varchar(15) NULL,
    "code" varchar(255) UNIQUE NOT NULL,
    "event_code" varchar(30) NOT NULL,
    "session_code" varchar(30) NOT NULL,
    "registered_by" varchar(100) NOT NULL,
    "account_number_origin" varchar(15) NULL,
    "updated_by" varchar(100) NULL,
    "status" varchar(20) NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMP,

    FOREIGN KEY ("event_code") REFERENCES "event_generals"("code"),
    FOREIGN KEY ("session_code") REFERENCES "event_sessions"("code")
);