SET TIME ZONE 'Asia/Jakarta';

-- ============================================================================
-- TABLE: forms
-- Description: Universal form definitions (domain-agnostic)
-- Supports: Event registrations, surveys, quizzes
-- Aligns with: models.Form
-- ============================================================================
CREATE TABLE forms (
    id BIGSERIAL PRIMARY KEY,
    code UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    form_type VARCHAR(50),           -- 'registration', 'survey', 'quiz'
    is_template BOOLEAN DEFAULT FALSE,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    creator_community_id VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT forms_name_not_empty CHECK (LENGTH(TRIM(name)) > 0),
    CONSTRAINT chk_forms_type CHECK (
        form_type IS NULL
        OR form_type IN ('registration', 'survey', 'quiz')
    ),
    CONSTRAINT chk_forms_status CHECK (
        status IN ('active', 'archived', 'draft')
    )
);

-- ============================================================================
-- TABLE: form_questions
-- Description: Individual questions within a form
-- Aligns with: models.FormQuestion
-- ============================================================================
CREATE TABLE form_questions (
    id BIGSERIAL PRIMARY KEY,
    code UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    form_code UUID NOT NULL REFERENCES forms (code) ON DELETE CASCADE,

    -- Question content
    text TEXT NOT NULL,
    category VARCHAR(50) NOT NULL,       -- question_type: short_text, long_text, number, email, phone, single_choice, multiple_choice, date, time

    -- Context-aware filtering
    -- required_for: contexts where this question is required. Values: 'parent', 'child'
    required_for TEXT[] DEFAULT ARRAY[]::TEXT[],
    -- visible_for: contexts where this question is shown. Values: 'parent', 'child'
    visible_for TEXT[] DEFAULT ARRAY[]::TEXT[],

    -- Options for single_choice / multiple_choice
    -- Structure: {"choices": ["Option A", "Option B", ...]}
    options JSONB,

    -- Validation rules
    -- Structure: {"minLength": 5, "maxLength": 100, "pattern": "^[A-Z]", "minValue": 0, ...}
    rules JSONB,

    -- For quiz/assessment questions
    correct_answer TEXT,

    display_order INTEGER NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT form_questions_text_not_empty CHECK (LENGTH(TRIM(text)) > 0),
    CONSTRAINT form_questions_display_order_non_negative CHECK (display_order >= 0),
    CONSTRAINT chk_form_questions_category CHECK (
        category IN (
            'short_text', 'long_text', 'number',
            'email', 'phone',
            'single_choice', 'multiple_choice',
            'date', 'time'
        )
    )
);

-- ============================================================================
-- TABLE: form_answers
-- Description: Individual answers submitted to form questions
-- Aligns with: models.FormAnswer
-- ============================================================================
CREATE TABLE form_answers (
    id BIGSERIAL PRIMARY KEY,
    code UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    form_code UUID REFERENCES forms (code) ON DELETE SET NULL,
    question_code UUID NOT NULL REFERENCES form_questions (code) ON DELETE CASCADE,

    -- Flexible identifier (supports authenticated users and walk-ins)
    identifier_type VARCHAR(50) NOT NULL,   -- 'community_id', 'eventAttendance'
    identifier_code VARCHAR(255) NOT NULL,

    answer TEXT NOT NULL,
    is_correct BOOLEAN,
    status VARCHAR(50) NOT NULL DEFAULT 'active',

    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT form_answers_answer_not_empty CHECK (LENGTH(TRIM(answer)) > 0),
    CONSTRAINT chk_form_answers_identifier_type CHECK (
        identifier_type IN ('community_id', 'eventAttendance')
    ),
    -- One answer per question per identifier
    CONSTRAINT unique_answer_per_question_per_identifier
        UNIQUE (question_code, identifier_type, identifier_code, deleted_at)
);

-- ============================================================================
-- TABLE: form_associations
-- Description: Polymorphic many-to-many link between forms and entities
-- Aligns with: models.FormAssociation
-- ============================================================================
CREATE TABLE form_associations (
    id BIGSERIAL PRIMARY KEY,
    code UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    form_code UUID NOT NULL REFERENCES forms (code) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL,   -- 'event', 'event_session'
    entity_code VARCHAR(50) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT form_associations_entity_code_not_empty
        CHECK (LENGTH(TRIM(entity_code)) > 0),
    CONSTRAINT chk_form_associations_entity_type CHECK (
        entity_type IN ('event', 'event_session')
    ),
    -- Prevent duplicate associations
    CONSTRAINT unique_form_entity_association
        UNIQUE (form_code, entity_type, entity_code, deleted_at)
);

-- ============================================================================
-- INDEXES: Performance Optimization
-- ============================================================================
-- forms
CREATE INDEX idx_forms_status ON forms (status) WHERE deleted_at IS NULL;
CREATE INDEX idx_forms_type ON forms (form_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_forms_template ON forms (is_template) WHERE is_template = TRUE AND deleted_at IS NULL;
CREATE INDEX idx_forms_creator ON forms (creator_community_id) WHERE deleted_at IS NULL;

-- form_questions
CREATE INDEX idx_form_questions_form ON form_questions (form_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_form_questions_type ON form_questions (type) WHERE deleted_at IS NULL;
CREATE INDEX idx_form_questions_display_order ON form_questions (form_code, display_order) WHERE deleted_at IS NULL;
CREATE INDEX idx_form_questions_mandatory_for ON form_questions USING GIN (mandatory_for) WHERE deleted_at IS NULL;
CREATE INDEX idx_form_questions_apply_for ON form_questions USING GIN (apply_for) WHERE deleted_at IS NULL;

-- form_answers
CREATE INDEX idx_form_answers_question ON form_answers (question_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_form_answers_form ON form_answers (form_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_form_answers_identifier ON form_answers (identifier_type, identifier_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_form_answers_submitted_at ON form_answers (submitted_at DESC) WHERE deleted_at IS NULL;

-- form_associations
CREATE INDEX idx_form_associations_form ON form_associations (form_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_form_associations_entity ON form_associations (entity_type, entity_code) WHERE deleted_at IS NULL;

-- ============================================================================
-- AUTO-UPDATE TRIGGERS: Keep updated_at current
-- ============================================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_forms_updated_at
    BEFORE UPDATE ON forms
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_form_questions_updated_at
    BEFORE UPDATE ON form_questions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_form_answers_updated_at
    BEFORE UPDATE ON form_answers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_form_associations_updated_at
    BEFORE UPDATE ON form_associations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- COMMENTS: Documentation
-- ============================================================================
COMMENT ON TABLE forms IS 'Universal form definitions. Reusable across events, surveys, and quizzes.';
COMMENT ON COLUMN forms.form_type IS 'Classifies form purpose: registration, survey, quiz';
COMMENT ON COLUMN forms.is_template IS 'If true, this form serves as a reusable template';

COMMENT ON TABLE form_questions IS 'Questions within a form. Context-aware via mandatory_for and apply_for arrays.';
COMMENT ON COLUMN form_questions.type IS 'Question input type: short_text, long_text, number, email, phone, single_choice, multiple_choice, date, time';
COMMENT ON COLUMN form_questions.mandatory_for IS 'Registrant contexts that must answer this question. Values: parent, child';
COMMENT ON COLUMN form_questions.apply_for IS 'Registrant contexts that see this question. Values: parent, child';
COMMENT ON COLUMN form_questions.options IS 'Choices for single_choice/multiple_choice. Structure: {"choices": ["A", "B"]}';
COMMENT ON COLUMN form_questions.rules IS 'Validation rules. Supports: minLength, maxLength, minValue, maxValue, notBefore, notAfter, minSelection, maxSelection, pattern';

COMMENT ON TABLE form_answers IS 'Individual answers to form questions. Supports both authenticated users (community_id) and walk-ins (eventAttendance).';
COMMENT ON COLUMN form_answers.identifier_type IS 'How the submitter is identified: community_id or eventAttendance';
COMMENT ON COLUMN form_answers.identifier_code IS 'The actual identifier value (community_id value or attendance reference)';
COMMENT ON COLUMN form_answers.is_correct IS 'Populated for quiz/assessment questions with a correct_answer defined';

COMMENT ON TABLE form_associations IS 'Polymorphic many-to-many link between forms and entities (event, event_session).';
COMMENT ON COLUMN form_associations.entity_type IS 'Type of linked entity: event, event_session';
COMMENT ON COLUMN form_associations.entity_code IS 'Code of the linked entity (event code or session code)';