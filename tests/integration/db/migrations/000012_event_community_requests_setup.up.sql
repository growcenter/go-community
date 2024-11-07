SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "event_community_requests" (
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "full_name" VARCHAR(100) NOT NULL,
    "request_type" VARCHAR(10) NOT NULL CHECK ("request_type" IN ('Prayer', 'Grateful')), -- Type of request ('Prayer' or 'Grateful')
    "email" VARCHAR(100) NOT NULL, -- Email address (unique)
    "phone_number" VARCHAR(20) NOT NULL, -- Optional phone number
    "request_information" TEXT NOT NULL, -- Detailed request information
    "want_contacted" BOOLEAN NOT NULL DEFAULT FALSE, -- Flag for whether the user wants to be contacted
    "account_number" VARCHAR(50) NOT NULL, -- Account number (no longer unique or a foreign key)
    "created_at" TIMESTAMP NOT NULL DEFAULT now(), -- Timestamp for when the request is created
    "updated_at" TIMESTAMP NOT NULL DEFAULT now(), -- Timestamp for when the request is updated
    "deleted_at" TIMESTAMP -- Soft delete timestamp
);

