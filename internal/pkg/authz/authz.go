package authz

import (
	"strings"
)

// Check performs an authorization check to see if a subject has a specific permission on a resource.
// It does not call the database. It evaluates the request against the provided schema and a set of relationships.
//
// Parameters:
//   - schema: The authorization model of the application.
//   - relationships: A list of all relevant relationships for this check.
//   - subjectType: The type of the subject (e.g., "user").
//   - subjectID: The ID of the subject.
//   - permission: The permission being checked (e.g., "edit").
//   - objectType: The type of the resource (e.g., "event").
//   - objectID: The ID of the resource.
//
// Returns:
//   - bool: True if the subject has the permission, false otherwise.
func Check(schema *Schema, relationships []Relationship, subjectType, subjectID, permission, objectType, objectID string) bool {
	// Find the definition for the resource type in the schema.
	def, ok := schema.Definitions[objectType]
	if !ok {
		return false // If the resource type doesn't exist, access is denied.
	}

	// Look for the specific permission in the resource's definition.
	perm, ok := def.Permissions[permission]
	if !ok {
		return false // If the permission isn't defined for this resource, access is denied.
	}

	// Evaluate the permission rule.
	return checkRule(schema, relationships, subjectType, subjectID, perm.Rule, objectType, objectID)
}

// checkRule evaluates a permission rule expression (e.g., "admin | member" or "organizer->edit").
// It supports OR conditions ( | ) and graph traversal ( -> ).
func checkRule(schema *Schema, relationships []Relationship, subjectType, subjectID, rule, objectType, objectID string) bool {
	// The rule is split by "|" to handle OR conditions.
	// For example, in "admin | member", if the subject satisfies either part, the rule is met.
	for _, part := range strings.Split(rule, " | ") {
		if checkRulePart(schema, relationships, subjectType, subjectID, strings.TrimSpace(part), objectType, objectID) {
			return true // If any part of the OR condition is true, the entire rule is true.
		}
	}
	return false // If no parts of the OR condition were met.
}

// checkRulePart evaluates a single segment of a permission rule.
// This can be a direct relationship check (e.g., "admin") or a delegated check (e.g., "organizer->edit").
func checkRulePart(schema *Schema, relationships []Relationship, subjectType, subjectID, rulePart, objectType, objectID string) bool {
	// Check for delegated permissions (e.g., "organizer->edit").
	if strings.Contains(rulePart, "->") {
		parts := strings.Split(rulePart, "->")
		relationName := parts[0]
		nextPermission := parts[1]

		// Find all resources related to the current object via the specified relation.
		// For example, find the "community" that is the "organizer" of the "event".
		for _, rel := range relationships {
			if rel.ObjectType == objectType && rel.ObjectID == objectID && rel.Relation == relationName {
				// Recursively call Check on the related resource.
				// This traverses the graph. For example, we now check if the user has "edit" permission on the "community".
				if Check(schema, relationships, subjectType, subjectID, nextPermission, rel.SubjectType, rel.SubjectID) {
					return true
				}
			}
		}
		return false
	}

	// This is a direct relationship check.
	// We check if a direct relationship exists between the subject and the object.
	// For example, is the user a "contact" for the "event"?
	relationName := rulePart
	for _, rel := range relationships {
		if rel.ObjectType == objectType &&
			rel.ObjectID == objectID &&
			rel.Relation == relationName &&
			rel.SubjectType == subjectType &&
			rel.SubjectID == subjectID {
			return true // A direct relationship was found.
		}
	}

	return false // No direct or delegated permission found.
}

// Relationship is a simplified struct for passing relationship data into the Check function.
// This avoids a direct dependency on the GORM model within the pure logic package.
type Relationship struct {
	ObjectType  string
	ObjectID    string
	Relation    string
	SubjectType string
	SubjectID   string
}
