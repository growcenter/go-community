SET TIME ZONE 'Asia/Jakarta';

-- Create the main table for storing relationships
CREATE TABLE access_relationships (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    object_type VARCHAR(50) NOT NULL,
    object_id VARCHAR(50) NOT NULL,
    relation VARCHAR(50) NOT NULL,
    subject_type VARCHAR(50) NOT NULL,
    subject_id VARCHAR(50) NOT NULL,
    created_at TIMESTIMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTIMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTIMPTZ
);

-- Index 1: For "Check" queries (e.g., "Can user:A edit doc:B?")
-- This unique index enforces data integrity and makes "check" lookups fast.
-- It is partial (WHERE deleted_at IS NULL) to only index active tuples.
CREATE UNIQUE INDEX uq_access_relationships_partial ON access_relationships (
    object_type,
    object_id,
    relation,
    subject_type,
    subject_id
)
WHERE deleted_at IS NULL;


-- Index 2: For "List" queries (e.g., "What can user:A see?")
-- This index makes "list" lookups fast by starting with the subject.
-- It is also partial to improve performance.
CREATE INDEX idx_relation_tuples_subject_partial ON access_relationships (
    subject_type,
    subject_id
)
WHERE deleted_at IS NULL;

