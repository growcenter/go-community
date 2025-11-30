package authz

// This file defines the application's authorization model using the schema structures.
// It is the Go representation of the Zanzibar-style schema defined in the RFC.

// NewAuthzModel initializes and returns the application's authorization schema.
// This schema is the single source of truth for all authorization logic.
func NewAuthzModel() *Schema {
	return &Schema{
		Definitions: map[string]*Definition{
			// A user is a basic entity and has no specific relations or permissions defined on it directly.
			"user": {
				Name: "user",
			},

			// A community has members and admins.
			"community": {
				Name: "community",
				Relations: map[string]*Relation{
					"admin":  {Name: "admin", With: []string{"user"}},
					"member": {Name: "member", With: []string{"user"}},
				},
				Permissions: map[string]*Permission{
					// Viewing is allowed for members or admins.
					"view": {Name: "view", Rule: "member | admin"},
					// Editing is restricted to admins.
					"edit": {Name: "edit", Rule: "admin"},
					// Creating events is restricted to admins.
					"create_event": {Name: "create_event", Rule: "admin"},
				},
			},

			// An event is organized by a community and has contacts and attendees.
			"event": {
				Name: "event",
				Relations: map[string]*Relation{
					"organizer": {Name: "organizer", With: []string{"community"}},
					"contact":   {Name: "contact", With: []string{"user"}},
					"attendee":  {Name: "attendee", With: []string{"user"}},
				},
				Permissions: map[string]*Permission{
					// View permission is granted to attendees, contacts, or anyone who can view the organizing community.
					// The "organizer->view" syntax means we traverse the graph to check the 'view' permission on the related 'organizer' community.
					"view": {Name: "view", Rule: "attendee | contact | organizer->view"},

					// Edit permission is granted to contacts or admins of the organizing community.
					// The "organizer->edit" syntax traverses the graph to check the 'edit' permission on the related 'organizer' community.
					"edit": {Name: "edit", Rule: "contact | organizer->edit"},

					// Register permission is granted to anyone who can view the organizing community.
					"register": {Name: "register", Rule: "organizer->view"},
				},
			},
		},
	}
}
