# AGENTS.md

This file provides guidance to AI coding agents (Claude Code, and others that read `AGENTS.md`) when working with code in this repository.

## Project

Go backend ("go-community") for GROW IT Team's Church Community Dashboard. Echo HTTP framework, GORM over PostgreSQL, Viper config, `swaggo/swag` for API docs.

## Commands

```bash
cp config/config.local.template.yaml config/config.local.yaml   # first-time setup

make docker-start        # start Postgres + app via docker-compose
make docker-stop          # stop docker-compose stack
make database_up          # create the community_db database
make database_down        # drop the community_db database

make run                  # generate swagger docs, tidy modules, then run the API
make run_api              # tidy modules and run ./cmd/api/main.go directly (export ENV=DEV first)
make tidy                 # go mod tidy && go mod download
make generate-docs        # regenerate docs/ via swag init -g cmd/api/main.go

make migration name=<n>       # create a new up/down migration pair under tests/integration/db/migrations/
make migration_up             # apply migrations (edit the DB connection string args first)
make migration_down           # roll back migrations
```

There is no configured `test` or `lint` Makefile target yet, despite README mentioning them — run `go test ./...` directly for tests (there are currently no `*_test.go` files in the repo) and `golangci-lint run` if you need to lint (no `.golangci.yml` is checked in, so it runs on defaults).

The app reads `ENV` (case-insensitive) to select `config/config.<env>.yaml` (e.g. `ENV=DEV` → `config.dev.yaml`) via Viper. `config.local.yaml` (gitignored, copied from the template) is used for local dev.

## Architecture

Layered, dependency-injected, single composition root — see `wiki/entities/composition-root.md` for the full narrative and `wiki/README.md` for the catalog of everything else documented there. The wiki (`wiki/`) is the curated knowledge base for this repo: check it before re-deriving architecture from scratch, and read `wiki/SCHEMA.md` if you're asked to add to it.

**Startup wiring** (`cmd/api/main.go` → `internal/contract/contract.go`): config is loaded once, then `Contract.New` builds dependencies in a fixed order — Postgres connect/ping → Google OAuth + JWT auth → repository aggregate (`pgsql.New`) → usecase aggregate (`usecases.New`) → HTTP routes (`handler.New`). The app runs Echo's own server directly (`echo.Echo.Start`/`Shutdown`); `internal/server/server.go`'s `net/http.Server` wrapper is unused dead code.

**Layers**, each with one interface+struct pair per domain and its own `New*` aggregate constructor:

- `internal/models/` — GORM entities, DB-output query structs, and per-endpoint request/response DTOs all bundled in one file per domain (e.g. `user_model.go`). There's no shared request/response shape per entity — each endpoint declares its own, so validation rules differ across endpoints for the same underlying entity. Domain errors are centralized sentinel `errors.New(...)` values in `error_model.go`, mapped to HTTP status/code/message by a single `ErrorMapping` switch.
- `internal/repositories/pgsql/` — GORM query builder for simple CRUD; raw SQL (package-level string vars in sibling `*_pg_query.go` files, assembled by `Build*` functions) for joins/aggregates/cursor pagination. `pgsql.New(db)` returns the `PostgreRepositories` aggregate injected into usecases. Two transaction helpers coexist (`Transaction` vs. the newer `Atomic`, which correctly threads a tx-scoped repository set into the callback) — see `wiki/entities/competing-transaction-abstractions.md` before adding transactional writes.
- `internal/usecases/` — business logic: validates against runtime config (department/campus maps), orchestrates repositories, generates codes/tokens/hashes. No separate service/domain layer beneath it. Dependency style is inconsistent across usecases (some take individual repository interfaces, others take the whole `PostgreRepositories` aggregate) — check `wiki/entities/usecase-dependency-style.md` before picking a pattern for a new one. `usecases.New(Dependencies{...})` returns the `Usecases` aggregate.
- `internal/deliveries/http/` — Echo handlers, versioned into `v1/` and `v2/` packages mounted under `/api`. Each domain's `New*Handler(group, usecases, config, ...)` registers its own route group and splits sub-groups by auth requirement (public / token-holder via `middleware.UserMiddleware` / role-restricted). Handler shape is fixed: `ctx.Bind` → `validator.Validate` → call usecase → `response.Success`/`response.Error`. Swagger annotations (`@Summary`, `@Router`, ...) above each handler drive `docs/` generation via `make generate-docs`.

**Auth**: JWT bearer tokens verified per-`kid` against `config.Auth.BearerSecret`/`RefreshSecret` maps (supports key rotation). `UserMiddleware` enforces role checks with a `superadmin` bypass. Two middlewares (`GeneralMiddleware`, `RefreshMiddleware`) branch their entire auth strategy on live feature flags (`event_be_enablegeneralheader`, `event_be_customrefreshheader`) checked per-request — both code paths must be assumed live in production simultaneously; see `wiki/entities/feature-flag-gated-auth.md`.

**RBAC**: a user's effective roles are derived at query time — `user_types` unnested and joined against a `user_types`→`roles` mapping table — not stored as a flattened list on the user row beyond `user.roles`/`user.user_types`.

Config (`internal/config/config.go`) nests `Application`, `PostgreSQL`, `Google`, `Auth` (per-kid JWT secrets, static API key, allowed client-ID map), and flat `Department`/`Campus` code→name maps used by usecases for validation without a DB round-trip.
