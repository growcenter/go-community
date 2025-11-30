SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "event_attendances" (
    "code" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "instance_code" VARCHAR(15) NOT NULL,
    "registration_code" UUID NOT NULL,
    "role" VARCHAR(50) NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "email" VARCHAR(255),
    "phone_number" VARCHAR(20),
    "legal_id" VARCHAR(16),
    "reference_code" VARCHAR(255),
    "remarks" TEXT,
    "status" VARCHAR(20) NOT NULL,
    "register_at" TIMESTAMPTZ NOT NULL,
    "verified_by" VARCHAR(255),
    "verified_at" TIMESTAMPTZ,
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMP
);

-- Indexes for faster lookups
CREATE INDEX idx_event_attendances_instance_code ON "event_attendances" ("instance_code");
CREATE INDEX idx_event_attendances_registration_code ON "event_attendances" ("registration_code");
CREATE INDEX idx_event_attendances_email ON "event_attendances" ("email");
CREATE INDEX idx_event_attendances_phone_number ON "event_attendances" ("phone_number");
CREATE INDEX idx_event_attendances_deleted_at ON "event_attendances" ("deleted_at");
