# Wiki

This folder is a collaborative, LLM-curated wiki for this repository.

- Add raw source docs to `raw/` (meeting notes, decision memos, transcripts, architectural writeups).
- Run `/codewiki:ingest` in a Claude Code session to fold a raw doc into the wiki.
- Run `/codewiki:lint` periodically to surface drift and health issues.

See [SCHEMA.md](SCHEMA.md) for conventions.

## Recently updated

- [Auth Middleware Chain](entities/auth-middleware-chain.md) — 2026-07-06 — How `UserMiddleware` verifies JWTs and enforces role checks on route sub-groups.
- [Availability Status Duplication](entities/availability-status-duplication.md) — 2026-07-06 — `DefineAvailabilityStatus` repeats field-extraction logic across 4+ near-identical DB-output shapes.
- [Competing Transaction Abstractions](entities/competing-transaction-abstractions.md) — 2026-07-06 — `Transaction` vs `Atomic`, and a likely transaction-scoping bug in `userRepository` writes.
- [Composition Root](entities/composition-root.md) — 2026-07-06 — Startup wiring: `cmd/api/main.go` and `internal/contract/contract.go` compose the app in a fixed dependency order.
- [Conflated User Create/Upgrade Flow](entities/user-create-upgrade-conflation.md) — 2026-07-06 — `userUsecase.Create` handles both new-user creation and upgrading a pre-existing shell record through the same endpoint.
- [Configuration Loading](entities/configuration-loading.md) — 2026-07-06 — How `config.Configuration` is loaded from env-specific YAML and shaped.
- [Cursor Pagination](entities/cursor-pagination.md) — 2026-07-06 — `GetAllWithCursor`'s over-fetch-and-encrypt keyset pagination approach.
- [Error Sentinel Catalog](entities/error-sentinel-catalog.md) — 2026-07-06 — Centralized `errors.New` sentinels mapped to HTTP responses via one large switch.
- [Feature-Flag-Gated Auth](entities/feature-flag-gated-auth.md) — 2026-07-06 — `GeneralMiddleware`/`RefreshMiddleware` branch their entire auth strategy on live feature flags.
- [HTTP Handler Conventions](entities/http-handler-conventions.md) — 2026-07-06 — Route registration, handler shape, and Swagger annotation conventions in the Echo HTTP layer.
- [Models Layer](entities/models-layer.md) — 2026-07-06 — GORM entities, DB-output structs, and per-endpoint request/response DTOs bundled per domain file.
- [RBAC Role Derivation](entities/rbac-role-derivation.md) — 2026-07-06 — Combined roles are computed at query time from `user_types`, not stored directly on the user.
- [Repository Layer](entities/repository-layer.md) — 2026-07-06 — GORM + raw-SQL data access conventions between usecases and Postgres.
- [Unused authorization_usecase.go](entities/unused-authorization-usecase.md) — 2026-07-06 — A `// NOT USED`-marked file still defining a live JWT-generation type.
- [Unused HTTP Server Wrapper](entities/unused-http-server-wrapper.md) — 2026-07-06 — `internal/server/server.go` appears to be dead code, unused by the actual composition root.
- [Usecase Dependency Style Inconsistency](entities/usecase-dependency-style.md) — 2026-07-06 — Usecases inconsistently take individual repository interfaces vs. the repository aggregate.
- [Usecase Layer](entities/usecase-layer.md) — 2026-07-06 — The business-logic layer between HTTP handlers and repositories: shape, responsibilities, and cross-cutting logging.

## Browse by category

### architecture

- [Auth Middleware Chain](entities/auth-middleware-chain.md)
- [Composition Root](entities/composition-root.md)
- [Configuration Loading](entities/configuration-loading.md)
- [HTTP Handler Conventions](entities/http-handler-conventions.md)
- [Models Layer](entities/models-layer.md)
- [Repository Layer](entities/repository-layer.md)
- [Usecase Layer](entities/usecase-layer.md)

### concept

- [Auth Middleware Chain](entities/auth-middleware-chain.md)
- [Configuration Loading](entities/configuration-loading.md)
- [Conflated User Create/Upgrade Flow](entities/user-create-upgrade-conflation.md)
- [Cursor Pagination](entities/cursor-pagination.md)
- [Error Sentinel Catalog](entities/error-sentinel-catalog.md)
- [Feature-Flag-Gated Auth](entities/feature-flag-gated-auth.md)
- [RBAC Role Derivation](entities/rbac-role-derivation.md)

### tech-debt

- [Availability Status Duplication](entities/availability-status-duplication.md)
- [Competing Transaction Abstractions](entities/competing-transaction-abstractions.md)
- [Feature-Flag-Gated Auth](entities/feature-flag-gated-auth.md)
- [Unused authorization_usecase.go](entities/unused-authorization-usecase.md)
- [Unused HTTP Server Wrapper](entities/unused-http-server-wrapper.md)
- [Usecase Dependency Style Inconsistency](entities/usecase-dependency-style.md)

## Raw sources

- [2026-07-06-bootstrap-go-backend-handlers-api.md](raw/2026-07-06-bootstrap-go-backend-handlers-api.md) — Jeremy — Bootstrap survey of the Echo HTTP handler/middleware layer (v1/v2 routes, auth middleware). **Ingested.**
- [2026-07-06-bootstrap-go-backend-models.md](raw/2026-07-06-bootstrap-go-backend-models.md) — Jeremy — Bootstrap survey of `internal/models/` (GORM entities, request/response DTOs, error catalog). **Ingested.**
- [2026-07-06-bootstrap-go-backend-repositories.md](raw/2026-07-06-bootstrap-go-backend-repositories.md) — Jeremy — Bootstrap survey of `internal/repositories/pgsql/` (GORM + raw SQL data access, transactions). **Ingested.**
- [2026-07-06-bootstrap-go-backend-server-wiring.md](raw/2026-07-06-bootstrap-go-backend-server-wiring.md) — Jeremy — Bootstrap survey of startup wiring: `cmd/api/main.go`, `internal/contract/`, `internal/config/`. **Ingested.**
- [2026-07-06-bootstrap-go-backend-usecases.md](raw/2026-07-06-bootstrap-go-backend-usecases.md) — Jeremy — Bootstrap survey of `internal/usecases/` (business logic layer). **Ingested.**
