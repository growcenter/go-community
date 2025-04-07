SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE event_questions (
    id UUID PRIMARY KEY,
    event_code VARCHAR(255) NOT NULL,
    instance_code VARCHAR(255), -- Nullable
    question TEXT NOT NULL,
    description TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,
    options TEXT[], -- Array of strings for options
    is_required BOOLEAN,
    is_registrant_required BOOLEAN,
    display_order INTEGER, -- Nullable
    is_visible_to_registrant BOOLEAN,
    rules JSONB, -- Optional JSONB for rule validation
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP -- Nullable (for soft delete)
);

-- Index on event_code to quickly find questions for an event
CREATE INDEX idx_event_questions_event_code ON event_questions(event_code);

-- Index on instance_code to support filtering by session/instance
CREATE INDEX idx_event_questions_instance_code ON event_questions(instance_code);

-- Index on id for fast lookup (redundant with PRIMARY KEY but sometimes useful for explicit queries)
CREATE INDEX idx_event_questions_id ON event_questions(id);