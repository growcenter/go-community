---
tags: [tech-debt]
sources: [raw/2026-07-06-bootstrap-go-backend-usecases.md]
updated: 2026-07-06
ingested_at: 91a596d
---

# Unused authorization_usecase.go

`internal/usecases/authorization_usecase.go` is headed with a `// NOT USED` comment but still defines a live `Auth`/`Authorization` type that generates JWTs via `jwt.NewWithClaims`. [internal/usecases/authorization_usecase.go:1-45](internal/usecases/authorization_usecase.go#L1-L45)

This is separate from the JWT verification logic in [Auth Middleware Chain](auth-middleware-chain.md) (`internal/deliveries/http/middleware/authorization.go`), which reads tokens rather than generates them, and from `internal/pkg/authorization`, which the [Composition Root](composition-root.md) actually wires in as the app's `Authorization` dependency.

Not confirmed dead — the `// NOT USED` comment suggests intent, but worth grepping for any remaining references to `usecases.Auth`/`usecases.NewAuthorization` before removing it, since a stale comment is equally plausible.
