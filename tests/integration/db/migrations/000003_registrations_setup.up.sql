CREATE TYPE registration_status AS ENUM ('BOOKED', 'CANCELLED', 'ATTENDED');
CREATE TYPE attendee_role AS ENUM ('LEADER', 'MEMBER', 'GUEST', 'EXTERNAL');

CREATE TABLE registrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_instance_id UUID NOT NULL REFERENCES event_instances(id) ON DELETE CASCADE,
    user_id UUID NOT NULL, -- The user who made the booking
    status registration_status NOT NULL DEFAULT 'BOOKED',
    quantity INT NOT NULL, -- The total number of attendees in this booking
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Individual people attending the event, part of a single registration
CREATE TABLE attendees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    registration_id UUID NOT NULL REFERENCES registrations(id) ON DELETE CASCADE,
    role attendee_role NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    -- Each attendee gets their own unique QR code for individual check-in
    attendee_qr_code_data TEXT UNIQUE,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
