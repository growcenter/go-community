package authz

// This file is inspired by the Zanzibar schema and defines the structure of our authorization model.
// It allows us to define resources, the relationships they can have, and how permissions are derived from those relationships.

// Schema is the top-level container for our authorization model.
type Schema struct {
	// Definitions is a map of resource type names (e.g., "community", "event") to their definitions.
	Definitions map[string]*Definition
}

// Definition represents a resource type in our system (e.g., a user, a community, an event).
type Definition struct {
	// Name is the type of the resource (e.g., "community").
	Name string
	// Relations is a map of relation names (e.g., "admin", "member") to their Relation definitions.
	Relations map[string]*Relation
	// Permissions is a map of permission names (e.g., "view", "edit") to their Permission definitions.
	Permissions map[string]*Permission
}

// Relation defines a relationship between two resources.
// For example, a "community" can have an "admin" relation with a "user".
type Relation struct {
	// Name is the name of the relation (e.g., "admin").
	Name string
	// With is a list of subject types that can have this relation.
	// For example, a "community" might have an "organizer" relation with an "event".
	With []string
}

// Permission defines an action that can be performed on a resource.
// Permissions are computed from relationships.
type Permission struct {
	// Name is the name of the permission (e.g., "edit").
	Name string
	// Rule is the expression that defines how this permission is granted.
	Rule string
}
