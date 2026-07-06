---
tags: [architecture]
sources: [raw/2026-07-06-bootstrap-go-backend-handlers-api.md]
updated: 2026-07-06
ingested_at: 91a596d
---

# HTTP Handler Conventions

`internal/deliveries/http/` is the Echo-based HTTP layer that `handler.New` (called from the [Composition Root](composition-root.md)) registers onto the shared `echo.Echo` instance. Routes are versioned into `v1/` and `v2/` packages, each mounted under a shared root handler.

## Route registration

Each domain gets its own `New*Handler(group, usecases, config, ...)` function that registers a route group and its sub-groups. [internal/deliveries/http/v2/v2_handler.go:12-24](internal/deliveries/http/v2/v2_handler.go#L12-L24) shows `NewV2Handler` mounting `/v2`, applying [General Middleware](feature-flag-gated-auth.md) to the whole group, then calling each domain's `New*Handler` in sequence — this one function enumerates the entire v2 API surface.

Within a domain, routes are further split into sub-groups by auth requirement — see `NewUserHandler`: public routes on the base group, a token-holder-only sub-group wrapped in [User Middleware](auth-middleware-chain.md), and a role-restricted sub-group. [internal/deliveries/http/v2/user_v2_handler.go:24-53](internal/deliveries/http/v2/user_v2_handler.go#L24-L53)

## Handler shape

Every handler follows the same shape: bind the request into a `models.*Request` struct, validate it, call into the usecase layer, then translate the result via `response.Success`/`response.Error`. [internal/deliveries/http/v2/user_v2_handler.go:67-83](internal/deliveries/http/v2/user_v2_handler.go#L67-L83)

Swagger annotations (`@Summary`, `@Router`, etc.) are hand-written directly above each handler function and drive `docs/swagger.yaml` generation — API docs are kept in sync by convention, not enforced by tooling.

## Notable deviation: inline cookie policy

`Login` sets the refresh token as an `HttpOnly`/`Secure`/`SameSite=Strict` cookie directly in the handler body, rather than through a shared response helper. [internal/deliveries/http/v2/user_v2_handler.go:97-124](internal/deliveries/http/v2/user_v2_handler.go#L97-L124) If more endpoints need to set this cookie, this logic should likely move into `response` or a middleware.

## Related

- [Auth Middleware Chain](auth-middleware-chain.md) — how sub-groups enforce token/role requirements.
- [Feature-Flag-Gated Auth](feature-flag-gated-auth.md) — runtime-branching auth middlewares applied at the group level.
