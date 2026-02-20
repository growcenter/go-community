SET
    TIME ZONE 'Asia/Jakarta';

-- ============================================================================
-- TABLE: forms
-- Description: Universal form definitions (domain-agnostic)
-- Supports: Event registrations, volunteer applications, surveys, feedback
-- ============================================================================
CREATE TABLE forms (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(30) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    form_type VARCHAR(30),
    is_template BOOLEAN DEFAULT FALSE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_by_community_id VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW (),
    updated_at TIMESTAMPTZ DEFAULT NOW (),
    deleted_at TIMESTAMPTZ
);

-- ============================================================================
-- TABLE: form_fields
-- Description: Individual questions/fields within a form
-- ============================================================================
CREATE TABLE form_fields (
    id BIGSERIAL PRIMARY KEY,
    form_id BIGINT NOT NULL REFERENCES forms (id) ON DELETE CASCADE,
    field_key VARCHAR(50) NOT NULL,
    field_type VARCHAR(30) NOT NULL,
    label VARCHAR(255) NOT NULL,
    placeholder TEXT,
    help_text TEXT,
    options JSONB,
    validation_rules JSONB,
    display_order INTEGER NOT NULL,
    is_required BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW (),
    updated_at TIMESTAMPTZ DEFAULT NOW (),
    UNIQUE (form_id, field_key)
);

-- ============================================================================
-- TABLE: form_submissions
-- Description: Tracks form submissions (polymorphic)
-- ============================================================================
CREATE TABLE form_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    form_id BIGINT NOT NULL REFERENCES forms (id),
    submission_type VARCHAR(30) NOT NULL,
    reference_id UUID,
    submitted_by_community_id VARCHAR(50),
    submitted_at TIMESTAMPTZ DEFAULT NOW (),
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW (),
    updated_at TIMESTAMPTZ DEFAULT NOW ()
);

-- ============================================================================
-- TABLE: form_answers
-- Description: Individual answers to form fields (optimized for querying)
-- ============================================================================
CREATE TABLE form_answers (
    id BIGSERIAL PRIMARY KEY,
    submission_id UUID NOT NULL REFERENCES form_submissions (id) ON DELETE CASCADE,
    field_id BIGINT NOT NULL REFERENCES form_fields (id),
    answer_value TEXT,
    answer_values TEXT[],
    answer_number DECIMAL(15, 2),
    answer_date DATE,
    answer_boolean BOOLEAN,
    created_at TIMESTAMPTZ DEFAULT NOW (),
    updated_at TIMESTAMPTZ DEFAULT NOW (),
    UNIQUE (submission_id, field_id)
);

-- ============================================================================
-- INDEXES: Performance Optimization
-- ============================================================================
-- Forms indexes
CREATE INDEX idx_forms_type ON forms (form_type)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_forms_creator ON forms (created_by_community_id)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_forms_status ON forms (status)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_forms_template ON forms (is_template)
WHERE
    is_template = TRUE
    AND deleted_at IS NULL;

-- Form fields indexes
CREATE INDEX idx_form_fields_form ON form_fields (form_id, display_order);

CREATE INDEX idx_form_fields_type ON form_fields (field_type);

-- Form submissions indexes
CREATE INDEX idx_form_submissions_form ON form_submissions (form_id);

CREATE INDEX idx_form_submissions_type_ref ON form_submissions (submission_type, reference_id);

CREATE INDEX idx_form_submissions_submitted_by ON form_submissions (submitted_by_community_id);

-- Form answers indexes (CRITICAL for querying)
CREATE INDEX idx_form_answers_field_value ON form_answers (field_id, answer_value);

CREATE INDEX idx_form_answers_field_values ON form_answers USING GIN (answer_values);

CREATE INDEX idx_form_answers_field_number ON form_answers (field_id, answer_number)
WHERE
    answer_number IS NOT NULL;

CREATE INDEX idx_form_answers_field_date ON form_answers (field_id, answer_date)
WHERE
    answer_date IS NOT NULL;

CREATE INDEX idx_form_answers_submission ON form_answers (submission_id);

-- ============================================================================
-- CONSTRAINTS: Data Integrity
-- ============================================================================
-- Forms constraints
ALTER TABLE forms ADD CONSTRAINT chk_forms_type CHECK (
    form_type IS NULL
    OR form_type IN (
        'event_registration',
        'volunteer_application',
        'survey',
        'membership',
        'feedback'
    )
);

ALTER TABLE forms ADD CONSTRAINT chk_forms_status CHECK (status IN ('active', 'archived', 'draft'));

-- Form fields constraints
ALTER TABLE form_fields ADD CONSTRAINT chk_form_fields_type CHECK (
    field_type IN (
        'text',
        'textarea',
        'select',
        'multiselect',
        'radio',
        'checkbox',
        'date',
        'number',
        'email',
        'phone'
    )
);

ALTER TABLE form_fields ADD CONSTRAINT chk_form_fields_display_order CHECK (display_order > 0);

-- Form submissions constraints
ALTER TABLE form_submissions ADD CONSTRAINT chk_form_submissions_type CHECK (
    submission_type IN (
        'event_registration',
        'volunteer_application',
        'survey_response'
    )
);

-- ============================================================================
-- COMMENTS: Documentation
-- ============================================================================
COMMENT ON TABLE forms IS 'Universal form definitions. Domain-agnostic and reusable across events, volunteer applications, surveys, and any future use case requiring custom data collection.';

COMMENT ON COLUMN forms.form_type IS 'Categorizes forms by domain: event_registration, volunteer_application, survey, membership, feedback';

COMMENT ON COLUMN forms.is_template IS 'Whether this form is a template for creating similar forms';

COMMENT ON TABLE form_fields IS 'Individual questions/fields within a form. Each field has a stable field_key for querying.';

COMMENT ON COLUMN form_fields.field_key IS 'Stable identifier for querying (e.g., dietary_preference). Does not change if label changes.';

COMMENT ON COLUMN form_fields.field_type IS 'Input type: text, textarea, select, multiselect, radio, checkbox, date, number, email, phone';

COMMENT ON COLUMN form_fields.options IS 'JSONB array of options for select/multiselect/radio/checkbox fields. Example: [{"value":"vegetarian","label":"Vegetarian"}]';

COMMENT ON COLUMN form_fields.validation_rules IS 'JSONB validation configuration. Example: {"required":true,"min_length":5,"max_length":100,"pattern":"regex"}';

COMMENT ON TABLE form_submissions IS 'Tracks form submissions. Uses polymorphic pattern (submission_type + reference_id) to link to any domain.';

COMMENT ON COLUMN form_submissions.submission_type IS 'What this submission is for: event_registration, volunteer_application, survey_response';

COMMENT ON COLUMN form_submissions.reference_id IS 'UUID reference to the related record (registration_id, application_id, etc.)';

COMMENT ON TABLE form_answers IS 'Individual answers to form fields. Optimized for querying with multiple answer columns for different data types.';

COMMENT ON COLUMN form_answers.answer_value IS 'Text answer for text, textarea, select, radio fields';

COMMENT ON COLUMN form_answers.answer_values IS 'Array answer for multiselect, checkbox fields';

COMMENT ON COLUMN form_answers.answer_number IS 'Numeric answer for number fields';

COMMENT ON COLUMN form_answers.answer_date IS 'Date answer for date fields';

COMMENT ON COLUMN form_answers.answer_boolean IS 'Boolean answer for single checkbox fields';