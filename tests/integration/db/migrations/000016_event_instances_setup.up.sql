SET
    TIME ZONE 'Asia/Jakarta';

CREATE TABLE "event_instances" (
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "code" varchar(15) UNIQUE NOT NULL,
    "event_code" varchar(7) NOT NULL,
    "title" varchar(255) NOT NULL,
    "description" TEXT,
    -- Instance Type (NEW)
    "instance_type" VARCHAR(30) NOT NULL DEFAULT 'registration',
    "instance_start_at" TIMESTAMPTZ,
    "instance_end_at" TIMESTAMPTZ,
    "register_start_at" TIMESTAMPTZ,
    "register_end_at" TIMESTAMPTZ,
    "allow_verify_at" TIMESTAMPTZ,
    "disallow_verify_at" TIMESTAMPTZ,
    "location_type" varchar(6) not null,
    "location_name" varchar(255) NOT NULL,
    "max_per_transaction" INT NOT NULL,
    "is_one_per_account" BOOLEAN DEFAULT FALSE,
    "is_one_per_ticket" BOOLEAN DEFAULT FALSE,
    "register_flow" VARCHAR(8),
    "check_type" VARCHAR(9),
    "total_seats" INT NOT NULL,
    "booked_seats" INT NOT NULL,
    "scanned_seats" INT NOT NULL,
    -- Age-Based Registration (NEW)
    "min_age" INT,
    "max_age" INT,
    "require_parent_info" BOOLEAN NOT NULL DEFAULT FALSE,
    -- Family Registration (NEW)
    "is_family_registration" BOOLEAN NOT NULL DEFAULT FALSE,
    "max_family_members" INT,
    "status" varchar(8) NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT now (),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now (),
    "deleted_at" TIMESTAMP,
    -- Constraints (NEW)
    CONSTRAINT check_instance_type CHECK (
        instance_type IN (
            'registration',
            'announcement',
            'volunteer-attendance'
        )
    ),
    CONSTRAINT check_age_range CHECK (
        (
            min_age IS NULL
            AND max_age IS NULL
        )
        OR (
            min_age IS NOT NULL
            AND max_age IS NOT NULL
            AND min_age >= 0
            AND max_age >= 0
            AND min_age <= max_age
        )
    ),
    CONSTRAINT check_family_members CHECK (
        (
            is_family_registration = FALSE
            AND max_family_members IS NULL
        )
        OR (
            is_family_registration = TRUE
            AND max_family_members IS NOT NULL
            AND max_family_members > 0
            AND max_family_members <= 20
        )
    )
);

-- Indexes (NEW)
CREATE INDEX idx_event_instances_type ON event_instances (instance_type)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_event_instances_age ON event_instances (min_age, max_age)
WHERE
    min_age IS NOT NULL
    AND deleted_at IS NULL;

CREATE INDEX idx_event_instances_family ON event_instances (is_family_registration)
WHERE
    is_family_registration = TRUE
    AND deleted_at IS NULL;

CREATE INDEX idx_event_instances_type_status ON event_instances (instance_type, status)
WHERE
    deleted_at IS NULL;

-- Comments (NEW)
COMMENT ON COLUMN event_instances.instance_type IS 'Type of instance: registration (requires registration), announcement (info only), volunteer-attendance (QR-based attendance tracking)';

COMMENT ON COLUMN event_instances.min_age IS 'Minimum age requirement for registration (optional, must be set with max_age)';

COMMENT ON COLUMN event_instances.max_age IS 'Maximum age requirement for registration (optional, must be set with min_age)';

COMMENT ON COLUMN event_instances.require_parent_info IS 'Whether parent/guardian information is required (auto-set based on max_age and config threshold)';

COMMENT ON COLUMN event_instances.is_family_registration IS 'Whether this instance supports family registration (multiple family members per registration)';

COMMENT ON COLUMN event_instances.max_family_members IS 'Maximum number of family members allowed per registration (required if is_family_registration is true)';