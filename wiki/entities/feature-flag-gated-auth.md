---
tags: [tech-debt, concept]
sources: [raw/2026-07-06-bootstrap-go-backend-handlers-api.md]
updated: 2026-07-06
ingested_at: 91a596d
---

# Feature-Flag-Gated Auth

`GeneralMiddleware` and `RefreshMiddleware` (`internal/deliveries/http/middleware/authorization.go`) each branch their *entire* auth strategy on a feature flag checked live, per-request, against the usecase layer's `FeatureFlag` service — meaning two structurally different auth flows can be live in production simultaneously depending on flag state.

## GeneralMiddleware

Gated on `event_be_enablegeneralheader`:

- **Flag off**: legacy behavior — requires a static `X-API-Key` header matching `config.Auth.APIKey`.
- **Flag on**: newer scheme — extracts/generates `X-Request-Id` and `X-Timestamp`, validates an optional base64-encoded `X-Client-Id` against `config.Auth.ClientId`, and still checks `X-API-Key` if present.

[internal/deliveries/http/middleware/authorization.go:187-251](internal/deliveries/http/middleware/authorization.go#L187-L251)

## RefreshMiddleware

Gated on `event_be_customrefreshheader`:

- **Flag on**: reads the refresh token from a custom `X-Refresh-Token` header.
- **Flag off**: reads it from the `refresh_token` cookie (the same cookie [HTTP Handler Conventions](http-handler-conventions.md) notes is set inline in the `Login` handler).

[internal/deliveries/http/middleware/authorization.go:111-185](internal/deliveries/http/middleware/authorization.go#L111-L185)

## Why this matters

Because the flag is evaluated per-request rather than at deploy time, both code paths must be assumed live and correct simultaneously — there is no single "current" auth behavior to reason about without also checking flag rollout status. Before changing either middleware, confirm which path (or both) is actually enabled via the feature flag service.
