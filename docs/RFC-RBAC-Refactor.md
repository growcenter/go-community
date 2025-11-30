# RFC: Centralized, Data-Driven RBAC

**Status:** Draft
**Author:** Gemini
**Date:** 2025-11-07

## 1. Summary

This RFC proposes a refactoring of the application's Role-Based Access Control (RBAC) system. The goal is to move from a pattern of hardcoded role lists scattered throughout the HTTP handlers to a centralized, data-driven authorization service.

This change will dramatically improve the maintainability, scalability, and security of our permission system by making authorization logic explicit, declarative, and manageable via the database rather than through code changes.

## 2. Motivation

Currently, authorization is handled by passing a hardcoded list of role strings to the `UserMiddleware` at the route-definition level. A typical example from `event_v2_handler.go` is:

```go
endpointUserInternal.Use(middleware.UserMiddleware(c, u, []string{"event-internal-view", "event-internal-edit"}))
```

This approach has several critical drawbacks:

*   **Scattered Logic:** The definition of "who can do what" is spread across dozens of lines in multiple handler files. There is no single source of truth for permissions.
*   **High Maintenance Overhead:** To change permissions (e.g., grant a new role access to an endpoint), a developer must find and modify every relevant hardcoded list, then redeploy the application. This is slow, error-prone, and inefficient.
*   **Lack of Flexibility:** The business cannot easily create or modify roles and their capabilities without a full development cycle.
*   **Poor Readability:** Route definitions are cluttered with implementation details rather than declaring the intended permission (e.g., `"events:create"`).

This RFC aims to solve these problems by introducing a centralized authorization service that leverages our existing data model.

## 3. Detailed Design

The proposal consists of three main components: creating a central authorization use case, defining a new declarative middleware, and refactoring the HTTP handlers to use it.

### Component 1: The Centralized `authorization_usecase`

We will create a new use case, `authorization_usecase.go`, which will be the single source of truth for all permission checks. It will expose one primary function: `Can()`.

The core logic will live in a helper function that correctly resolves a user's complete set of permissions by leveraging the existing `UserType` model.

```go
// internal/usecases/authorization_usecase.go

// getEffectiveRoles resolves all permissions a user has from their direct roles and UserTypes.
func (au *AuthUsecase) getEffectiveRoles(ctx context.Context, claims models.TokenClaims) ([]string, error) {
    // Use a map to handle duplicates automatically
    effectiveRoles := make(map[string]bool)

    // 1. Add roles assigned directly to the user
    for _, role := range claims.Roles {
        effectiveRoles[role] = true
    }

    // 2. Fetch UserType objects from the database (or a cache)
    userTypeObjects, err := au.userTypeUsecase.GetByTypes(ctx, claims.UserTypes)
    if err != nil {
        return nil, err
    }

    // 3. Add all roles from each UserType
    for _, ut := range userTypeObjects {
        for _, role := range ut.Roles {
            effectiveRoles[role] = true
        }
    }

    // 4. Convert map keys to a slice
    finalRoles := make([]string, 0, len(effectiveRoles))
    for role := range effectiveRoles {
        finalRoles = append(finalRoles, role)
    }

    return finalRoles, nil
}

// Can checks if a user has permission to perform an action.
func (au *AuthUsecase) Can(
    ctx context.Context,
    claims models.TokenClaims,
    requiredRole string, // e.g., "events:create"
    resource interface{}, // Optional: for resource-specific checks like ownership
) (bool, error) {
    effectiveRoles, err := au.getEffectiveRoles(ctx, claims)
    if err != nil {
        return false, err
    }

    // The primary check
    hasPermission := common.CheckOneDataInList(effectiveRoles, []string{requiredRole})

    // Optional: Add resource-specific logic here if needed
    // e.g., check if user is the owner of the 'resource' object

    return hasPermission, nil
}
```
**Performance Note:** The `UserType -> Roles` mapping is ideal for an in-memory cache with a short TTL (e.g., 1-5 minutes) to ensure near-zero latency on lookups.

### Component 2: The Declarative `Can` Middleware

We will create a new, lean middleware that connects the HTTP layer to the authorization use case.

```go
// internal/deliveries/http/middleware/access_control.go

func Can(authUsecase usecases.AuthorizationUsecase, action string) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(ctx echo.Context) error {
            claims, ok := ctx.Get("claims").(models.TokenClaims)
            if !ok {
                return response.Error(ctx, models.ErrorUnauthorized)
            }

            hasPermission, err := authUsecase.Can(ctx, claims, action, nil)
            if err != nil || !hasPermission {
                return response.Error(ctx, models.ErrorForbidden)
            }

            return next(ctx)
        }
    }
}
```

### Component 3: Refactoring HTTP Handlers

With the new components, our route definitions become clean and declarative.

**Before:**
```go
// Handler defines *who* can access this endpoint
endpointUserInternal.Use(middleware.UserMiddleware(c, u, []string{"event-internal-view", "event-internal-edit", "superadmin"}))
endpointUserInternal.POST("", h.CreateEvent)
```

**After:**
```go
// Handler declares *what action* is being performed
// The UserMiddleware now only handles authentication, not authorization.
endpointUserInternal.POST("", h.CreateEvent, middleware.Can(h.AuthUsecase, "events:create"))
```

## 4. Adoption Strategy

This refactor can be rolled out incrementally without a "big bang" rewrite.

1.  **Implement:** Create the `authorization_usecase` and the new `Can` middleware.
2.  **Define & Map:** Define a clear list of action-based roles (e.g., `events:create`, `events:edit`). Update the `Roles` for each `UserType` in the database.
3.  **Migrate:** Refactor one handler at a time, starting with the new `v2` handlers. Replace the old `UserMiddleware` call with the new `Can` middleware for each route. The old and new systems can coexist during this period.
4.  **Deprecate:** Once all handlers are migrated, the authorization logic can be removed from the old `UserMiddleware`, leaving it to handle only authentication.

## 5. Drawbacks

*   **Initial Effort:** There is an upfront development cost to build the new service and refactor existing handlers.
*   **Caching:** A caching strategy for the UserType-to-Role map is required for optimal performance, which adds a minor piece of complexity.

## 6. Alternatives Considered

*   **Status Quo:** Continuing with the current approach is not viable as it will lead to an unmaintainable and error-prone system as the application scales.
*   **Third-Party Libraries (e.g., Casbin):** While powerful, introducing a large third-party dependency for access control adds significant complexity. The proposed in-house solution is tailored to our existing architecture and provides the necessary flexibility without the overhead of a new library.
