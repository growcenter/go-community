CREATE TABLE form_questions (
    id SERIAL PRIMARY KEY,
    code UUID NOT NULL UNIQUE,
    form_code UUID NOT NULL,
    text TEXT NOT NULL,
    type VARCHAR(255) NOT NULL,
    mandatory_for TEXT[],
    apply_for TEXT[],
    options JSONB,
    rules JSONB,
    correct_answer TEXT,
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
