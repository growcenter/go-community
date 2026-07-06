# Ingest Log

Append-only chronological record of ingests. See [SCHEMA.md](SCHEMA.md#log-entry-format) for format.

## [2026-07-06] ingest | Bootstrap: Go Backend — Server Wiring | author: Jeremy | commit: 91a596d | at: 91a596d

Documented the app's composition root (`internal/contract/contract.go` + `cmd/api/main.go`), how `config.Configuration` is loaded from env-specific YAML, and flagged `internal/server/server.go` as likely-dead code since the composition root starts Echo's server directly instead.

Touched:
- entities/composition-root.md
- entities/configuration-loading.md
- entities/unused-http-server-wrapper.md

## [2026-07-06] ingest | Bootstrap: Go Backend — Handlers / API | author: Jeremy | commit: 91a596d | at: 91a596d

Documented the versioned Echo HTTP layer's route registration and handler conventions, the JWT-based `UserMiddleware` auth chain, and flagged the feature-flag-gated `GeneralMiddleware`/`RefreshMiddleware` as a tech-debt concern since two structurally different auth flows can be live simultaneously.

Touched:
- entities/http-handler-conventions.md
- entities/auth-middleware-chain.md
- entities/feature-flag-gated-auth.md
- entities/composition-root.md (cross-link added)

## [2026-07-06] ingest | Bootstrap: Go Backend — Usecases | author: Jeremy | commit: 91a596d | at: 91a596d

Documented the usecase layer's role and shape, its inconsistent repository-dependency style across domains, a specific conflated create/upgrade flow in `userUsecase.Create`, and flagged `authorization_usecase.go` as a possibly-dead file despite its `// NOT USED` marker.

Touched:
- entities/usecase-layer.md
- entities/usecase-dependency-style.md
- entities/user-create-upgrade-conflation.md
- entities/unused-authorization-usecase.md
- entities/composition-root.md (cross-link added)

## [2026-07-06] ingest | Bootstrap: Go Backend — Models | author: Jeremy | commit: 91a596d | at: 91a596d

Documented the models layer's file-per-domain bundling of GORM entities, DB-output structs, and per-endpoint request/response DTOs, the centralized error-sentinel-to-HTTP-response mapping, and flagged the duplicated availability-status type-switch as tech-debt.

Touched:
- entities/models-layer.md
- entities/error-sentinel-catalog.md
- entities/availability-status-duplication.md
- entities/usecase-layer.md (cross-links added)

## [2026-07-06] ingest | Bootstrap: Go Backend — Repositories | author: Jeremy | commit: 91a596d | at: 91a596d

Documented the repository layer's GORM-plus-raw-SQL query strategy, cursor-based pagination, query-time RBAC role derivation, and flagged the two competing transaction abstractions along with a likely transaction-scoping bug in `userRepository`'s `Create`/`Update` methods.

Touched:
- entities/repository-layer.md
- entities/cursor-pagination.md
- entities/rbac-role-derivation.md
- entities/competing-transaction-abstractions.md
- entities/usecase-layer.md (cross-link added)
