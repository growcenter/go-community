SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "event_registrations" (
    "code" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "event_code" VARCHAR(7) NOT NULL,
    "instance_code" VARCHAR(15) NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "email" VARCHAR(255),
    "phone_number" VARCHAR(20),
    "community_id" VARCHAR(15) NOT NULL,
    "method" VARCHAR(50) NOT NULL,
    "quantity" INT NOT NULL,
    "status" VARCHAR(20) NOT NULL,
    "register_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "deleted_at" TIMESTAMP
);

CREATE INDEX idx_event_registrations_community_id ON event_registrations(community_id);
CREATE INDEX idx_event_registrations_instance_code ON event_registrations(instance_code);