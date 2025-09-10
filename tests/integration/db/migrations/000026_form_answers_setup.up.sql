CREATE TABLE form_answers (
    id UUID PRIMARY KEY,
    form_code UUID NOT NULL REFERENCES forms(code),
    user_id VARCHAR(255) NOT NULL,
    question_code UUID NOT NULL REFERENCES form_questions(code),
    answer TEXT NOT NULL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
