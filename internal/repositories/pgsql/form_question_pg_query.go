package pgsql

import (
	"go-community/internal/models"
	"strings"
)

const (
	initialQueryGetFormQuestionsByEntities = `
		SELECT fq.*
		FROM form_questions fq
		JOIN form_associations fa ON fq.form_code = fa.form_code::uuid
	`
)

// BuildGetFormQuestionsByEntitiesQuery constructs a SQL query to fetch form questions
// associated with a flexible number of entities (like events and event instances).
func BuildGetFormQuestionsByEntitiesQuery(entities []models.FormQuestionEntityFilter) (string, []interface{}) {
	// If there are no entities to filter by, return an empty query.
	if len(entities) == 0 {
		return "", nil
	}

	var conditions []string
	var args []interface{}

	// Dynamically build the WHERE clause conditions.
	// For each entity, it creates a condition like: (fa.entity_code = ? AND fa.entity_type = ?)
	for _, entity := range entities {
		conditions = append(conditions, "(fa.entity_code = ? AND fa.entity_type = ?)")
		args = append(args, entity.Code, entity.Type)
	}

	// Start with the base query.
	query := initialQueryGetFormQuestionsByEntities

	// Join the individual conditions with "OR" and append to the main query.
	query += " WHERE " + strings.Join(conditions, " OR ")

	// Add ordering to sort questions by their display order, with a secondary sort on creation time to ensure a stable order for duplicate display_order values.
	query += " ORDER BY fq.display_order ASC, fq.created_at ASC"

	return query, args
}
