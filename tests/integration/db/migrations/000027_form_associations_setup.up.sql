CREATE TABLE form_associations (
    id SERIAL PRIMARY KEY,
    form_code UUID NOT NULL,
    entity_type VARCHAR(255) NOT NULL,
    entity_code VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    UNIQUE(form_code, entity_code)
);