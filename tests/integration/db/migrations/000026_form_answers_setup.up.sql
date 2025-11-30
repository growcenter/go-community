CREATE TABLE form_answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    form_code UUID,
    question_code UUID NOT NULL,
    identifier_type VARCHAR(255) NOT NULL,
    identifier VARCHAR(255) NOT NULL,
    answer TEXT NOT NULL,
    is_correct BOOL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Foreign key constraints can be added separately for clarity
ALTER TABLE form_answers ADD CONSTRAINT fk_form_answers_form_code FOREIGN KEY (form_code) REFERENCES forms(code);
ALTER TABLE form_answers ADD CONSTRAINT fk_form_answers_question_code FOREIGN KEY (question_code) REFERENCES form_questions(code);

-- Index for faster lookups on identifier
CREATE INDEX idx_form_answers_identifier ON form_answers(identifier_type, identifier);
