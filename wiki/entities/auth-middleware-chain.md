---
tags: [architecture, concept]
sources: [raw/2026-07-06-bootstrap-go-backend-handlers-api.md]
updated: 2026-07-06
ingested_at: 91a596d
---

# Auth Middleware Chain

Auth is enforced through Echo middleware applied to route (sub-)groups in [HTTP Handler Conventions](http-handler-conventions.md), not through per-handler checks. The core middleware is `UserMiddleware`, defined in `internal/deliveries/http/middleware/authorization.go`.

## UserMiddleware

`UserMiddleware(config, usecase, allowedRoles)` parses the JWT from the `Authorization` header. The token's `kid` header selects which secret to verify against from [Configuration Loading](configuration-loading.md)'s `Auth.BearerSecret` map (keyed by base64url-decoded `kid`). [internal/deliveries/http/middleware/authorization.go:26-109](internal/deliveries/http/middleware/authorization.go#L26-L109)

After signature verification it checks:

- `exp`/`iat` validity and that the token type claim is `"access"` (not a refresh token).
- If `allowedRoles` is non-nil: a `superadmin` user type bypasses the role check entirely; otherwise the caller's `roles` claim must intersect `allowedRoles`.

On success, `id`, `userTypes`, and `roles` are stashed into the Echo context via `ctx.Set(...)` for downstream handlers to read.

## Usage pattern

Handlers pass `nil` for `allowedRoles` to require only a valid token (any authenticated user), or a specific role list to restrict a sub-group further — e.g. `middleware.UserMiddleware(c, u, []string{"event-internal-view", "event-internal-edit"})` on the internal-user endpoints. [internal/deliveries/http/v2/user_v2_handler.go:48-52](internal/deliveries/http/v2/user_v2_handler.go#L48-L52)

## Related

- [Feature-Flag-Gated Auth](feature-flag-gated-auth.md) — `GeneralMiddleware`/`RefreshMiddleware`, applied above `UserMiddleware` at the group level, whose entire strategy varies by feature flag.
