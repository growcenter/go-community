# RFC: Refactor Authorization with Relationship-Based Access Control (ReBAC)

**Status:** Draft
**Author:** Gemini
**Date:** 2025-11-07

## 1. Summary

This RFC proposes a fundamental shift in our authorization model, moving from our current system to a more powerful and flexible Relationship-Based Access Control (ReBAC) system. This will be accomplished by introducing a dedicated, centralized authorization service that models permissions as a graph of relationships between our core entities (users, communities, events, etc.).

This approach will replace hardcoded permission logic and denormalized array columns (e.g., `AllowedCommunityIds`) with a scalable, data-driven system that can handle complex authorization scenarios gracefully. This provides a future-proof foundation for access control that aligns with the complexity of our domain.

## 2. Motivation

The current authorization system, which relies on arrays of IDs stored directly in model columns, is difficult to maintain, query, and scale. It violates database normalization principles and cannot express complex, real-world access rules, such as:

*   "A user can edit an event *if* they are an admin of the community that organized it."
*   "A user can view the attendee list *if* they are a designated contact for the event."

A simple role-based system (RBAC) is also insufficient, as it doesn't natively capture the rich context of our application. The core of our platform is the *relationship* between users, communities, and events. Our authorization model should reflect that reality.

Adopting ReBAC is necessary to:
*   **Model Complex Permissions:** Natively handle rules based on ownership, membership, roles within a specific context, and hierarchies.
*   **Centralize and Decouple Logic:** Move all authorization logic out of the main application and into a specialized, single-source-of-truth service.
*   **Improve Database Design:** Eliminate denormalized array columns, leading to a cleaner and more maintainable primary database schema.
*   **Enhance Performance:** Utilize a dedicated engine that is highly optimized for answering complex access control questions at scale.

## 3. Detailed Design

We will adopt a model inspired by Google's Zanzibar, implemented using a mature open-source engine such as **SpiceDB** or **OpenFGA**. The design consists of three main components: the Authorization Schema, the database schema for storing relationships, and the API for interacting with the new authorization service.

### 3.1. Authorization Schema (Zanzibar-style SDL)

The heart of the new system is the authorization schema, which defines our resources, the relationships between them, and the permissions that derive from those relationships. This schema is a plain text file loaded into the ReBAC engine.

```
definition user {}

definition community {
    relation admin: user
    relation member: user

    permission view = member + admin
    permission edit = admin
    permission create_event = admin
}

definition event {
    relation organizer: community
    relation contact: user
    relation attendee: user

    // An event can be viewed by its attendees, its contacts, or any member of the organizing community.
    permission view = attendee + contact + organizer->view

    // An event can be edited by its contacts or by an admin of the organizing community.
    permission edit = contact + organizer->edit

    // A user can register for an event if they are a member of the organizing community.
    permission register = organizer->view
}
```

This schema becomes the single, auditable source of truth for all permissions. The `->` (arrow) operator is key, as it allows the system to traverse the relationship graph. For example, `organizer->edit` delegates the `edit` permission check on an `event` to the `edit` permission on the `community` that is its `organizer`.

### 3.2. Database Schema

The ReBAC engine manages its own data store, which is conceptually a single, highly-indexed table of relationship tuples. This completely replaces the need for `event_allowed_roles`, `event_allowed_communities`, etc., tables or columns in our main application database.

The schema for storing these relationships is as follows:

```sql
-- This table will be managed internally by the dedicated authorization service.
CREATE TABLE relation_tuples (
    id UUID PRIMARY KEY,
    object_type VARCHAR(50) NOT NULL, -- e.g., 'event'
    object_id VARCHAR(50) NOT NULL,   -- e.g., 'gocon-2025'
    relation VARCHAR(50) NOT NULL,    -- e.g., 'organizer'
    subject_type VARCHAR(50) NOT NULL, -- e.g., 'community'
    subject_id VARCHAR(50) NOT NULL,   -- e.g., 'go-developers'
    -- with a unique constraint on all columns except id
);
```

Our application's database becomes much cleaner, and the responsibility for authorization data is correctly delegated to the authorization service.

### 3.3. Authorization Service API

The application will communicate with the ReBAC engine via a simple, well-defined API.

#### 1. Check Permission
This is the primary endpoint, used to ask "Can this user do this action on this resource?"

*   **Endpoint:** `POST /v1/permissions/check`
*   **Request Body:**
    ```json
    {
      "resource": { "type": "event", "id": "gocon-2025" },
      "permission": "edit",
      "subject": { "type": "user", "id": "user-jeremy" }
    }
    ```
*   **Success Response:** `{"allowed": true}`

#### 2. Write Relationships
Creates or deletes relationships. This is called when the state of the application changes (e.g., a user joins a community, an event is created, an admin is assigned).

*   **Endpoint:** `POST /v1/relationships/write`
*   **Request Body:**
    ```json
    {
      "updates": [
        {
          "op": "CREATE", // or "DELETE"
          "relationship": {
            "resource": { "type": "community", "id": "go-developers" },
            "relation": "member",
            "subject": { "type": "user", "id": "new-user-anna" }
          }
        }
      ]
    }
    ```
*   **Success Response:** `200 OK`

#### 3. Expand Permissions
Returns the full set of subjects that have a given permission on a resource. This is useful for UI elements and admin dashboards.

*   **Endpoint:** `POST /v1/permissions/expand`
*   **Request Body:**
    ```json
    {
      "resource": { "type": "event", "id": "gocon-2025" },
      "permission": "view"
    }
    ```
*   **Success Response:** A tree structure representing all subjects with the `view` permission.

### 3.4. Application Integration

A new middleware will be created to protect our API endpoints.

```go
// internal/deliveries/http/middleware/rebac_middleware.go

// Can checks if the user has permission to perform an action on a resource.
func Can(authClient *rebac.Client, permission string, resourceType string, resourceIDParam string) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            claims, ok := c.Get("claims").(models.TokenClaims)
            if !ok {
                return response.Error(c, models.ErrorUnauthorized)
            }

            resourceID := c.Param(resourceIDParam) // Get resource ID from URL param

            allowed, err := authClient.Check(
                c.Request().Context(),
                resourceType,
                resourceID,
                permission,
                "user",
                claims.CommunityId, // Assumes CommunityId in claims is the user's unique ID
            )

            if err != nil || !allowed {
                return response.Error(c, models.ErrorForbidden)
            }
            return next(c)
        }
    }
}

// Example of a refactored route definition:
// The middleware extracts the event "code" from the URL and checks the "view" permission.
server.GET("/events/:code", h.GetEvent, middleware.Can(h.AuthClient, "view", "event", "code"))
```

When a resource is created or a relationship changes, the relevant use case will call the `authClient.Write()` method to keep the authorization service in sync.

## 4. Adoption Strategy

This refactor can be rolled out incrementally and safely.

1.  **Setup:** Deploy an instance of the chosen ReBAC engine (e.g., SpiceDB) and load the initial schema.
2.  **Implement:** Build the Go client for the authorization service and the new `Can` middleware.
3.  **Dual-Write:** Modify the application's use cases (e.g., `user_usecase`, `event_usecase`) to write relationship data to the new authorization service *in addition* to modifying the old database columns. This keeps both systems in sync.
4.  **Shadow Mode:** In the middleware, check permissions against *both* the old system and the new ReBAC system. Log any discrepancies for analysis but continue to use the old system's decision to grant or deny access. This allows for safe, production-based validation.
5.  **Cut-over:** Once discrepancies are resolved and confidence in the new system is high, switch the middleware to rely solely on the ReBAC system for all authorization decisions.
6.  **Cleanup:** After a stabilization period, remove the old authorization logic, the now-unused array columns from the database (e.g., `AllowedCommunityIds`), and the dual-write code.

## 5. Drawbacks

*   **Operational Complexity:** This introduces a new, stateful service to our stack that must be deployed, monitored, and maintained.
*   **Learning Curve:** The development team will need to become familiar with ReBAC concepts and the specific Schema Definition Language (SDL) of the chosen engine.

## 6. Alternatives Considered

*   **Status Quo:** Continuing with the current array-based approach is not scalable and will lead to an unmaintainable, error-prone system.
*   **Build In-House ReBAC:** Building a correct, performant, and scalable graph authorization engine is a massive undertaking. Leveraging a mature, open-source solution is far more practical and allows us to focus on our core product.
