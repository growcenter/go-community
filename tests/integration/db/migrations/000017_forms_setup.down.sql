-- Drop triggers first
DROP TRIGGER IF EXISTS trg_form_associations_updated_at ON form_associations;

DROP TRIGGER IF EXISTS trg_form_answers_updated_at ON form_answers;

DROP TRIGGER IF EXISTS trg_form_questions_updated_at ON form_questions;

DROP TRIGGER IF EXISTS trg_forms_updated_at ON forms;

-- Drop helper function
DROP FUNCTION IF EXISTS update_updated_at_column ();

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS form_associations;

DROP TABLE IF EXISTS form_answers;

DROP TABLE IF EXISTS form_questions;

DROP TABLE IF EXISTS forms;