# Bootstrap: Go Backend — Handlers / API

_Surveyed 2026-07-06 at commit a0da240._

## Overview

`internal/deliveries/http/` is the Echo-based HTTP layer, versioned into `v1/` and `v2/` packages that both mount under a shared root handler. Each domain handler (`UserHandler`, `EventHandler`, ...) registers its own route group, binds+validates the request into a `models.*Request`, calls into `usecases.Usecases` (an aggregate of all usecases), and converts the result through a `response.Success`/`response.Error` helper. Auth is enforced via Echo middleware chains rather than per-handler checks.

## Key structures / flows

- [internal/deliveries/http/v2/v2_handler.go:12-24](internal/deliveries/http/v2/v2_handler.go#L12-L24) — `NewV2Handler` mounts `/v2`, applies `GeneralMiddleware` (API-key/client-id gate) to the whole group, then registers each domain's `New*Handler` — one place enumerates the entire v2 surface.
- [internal/deliveries/http/v2/user_v2_handler.go:24-53](internal/deliveries/http/v2/user_v2_handler.go#L24-L53) — `NewUserHandler` shows the per-route-group auth pattern: public routes on the base group, a `endpointUserAuth` sub-group wrapped in `middleware.UserMiddleware(c, u, nil)` for token-holder-only routes, and an `userInternalEndpoint` sub-group wrapped with role-restricted `UserMiddleware(c, u, []string{"event-internal-view", "event-internal-edit"})`.
- [internal/deliveries/http/v2/user_v2_handler.go:67-83](internal/deliveries/http/v2/user_v2_handler.go#L67-L83) — Standard handler shape: `ctx.Bind` → `validator.Validate` → call usecase → `response.Error`/`response.Success`; Swagger annotations (`@Summary`, `@Router`, ...) are hand-written above each handler and drive `docs/swagger.yaml` generation.
- [internal/deliveries/http/v2/user_v2_handler.go:97-124](internal/deliveries/http/v2/user_v2_handler.go#L97-L124) — `Login` sets the refresh token as an `HttpOnly`/`Secure`/`SameSite=Strict` cookie directly in the handler, coupling cookie policy to this one endpoint rather than a shared response helper.
- [internal/deliveries/http/middleware/authorization.go:26-109](internal/deliveries/http/middleware/authorization.go#L26-L109) — `UserMiddleware` parses a JWT (`kid` header selects the verification secret from `config.Auth.BearerSecret`), checks `exp`/`iat`/token type `"access"`, then enforces `allowedRoles` with a `superadmin` bypass; on success it stashes `id`/`userTypes`/`roles` into Echo context via `ctx.Set`.
- [internal/deliveries/http/middleware/authorization.go:111-185](internal/deliveries/http/middleware/authorization.go#L111-L185) — `RefreshMiddleware` reads the refresh token from either a custom header or the `refresh_token` cookie, gated by a feature flag (`event_be_customrefreshheader`) checked live against `usecase.FeatureFlag` on every request.
- [internal/deliveries/http/middleware/authorization.go:187-251](internal/deliveries/http/middleware/authorization.go#L187-L251) — `GeneralMiddleware` similarly branches its entire auth strategy (legacy static API key vs. newer request-id/timestamp/client-id scheme) on a feature flag (`event_be_enablegeneralheader`), so the API's auth behavior differs at runtime based on flag state, not deploy.

## Open questions

- Feature-flag-gated auth branching in `GeneralMiddleware`/`RefreshMiddleware` means two structurally different auth flows are live in production simultaneously — worth checking flag rollout status before assuming either path is "the" current behavior.
- Cookie-setting logic lives inline in the `Login` handler; if other endpoints ever set the refresh cookie, this should probably move into `response` or a middleware helper.
