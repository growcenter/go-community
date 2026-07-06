# Bootstrap: Go Backend — Server Wiring

_Surveyed 2026-07-06 at commit a0da240._

## Overview

Startup wiring is split across `cmd/api/main.go` (process lifecycle) and `internal/contract/contract.go` (dependency construction). `main.go` loads config, builds a `Contract`, starts it in a goroutine, and blocks on `SIGINT`/`SIGTERM` for graceful shutdown. `Contract.New` is the composition root: it connects Postgres, builds Google OAuth + JWT auth, wires the repository layer, then the usecase layer, then the HTTP handler layer, in that fixed order. `internal/config` loads environment-specific YAML (`config.<env>.yaml`) via Viper, driven by an `ENV` env var.

## Key structures / flows

- [cmd/api/main.go:17-47](cmd/api/main.go#L17-L47) — `main` builds config, then `contract.New(config)`, starts serving in a goroutine, and on receiving a signal calls `contract.Stop` with a 10-second shutdown timeout.
- [internal/contract/contract.go:24-76](internal/contract/contract.go#L24-L76) — `Contract.New` is the single composition root: Postgres connect + ping (fatal on failure) → Google OAuth client → JWT `Authorization` → `pgsql.New(psql)` repository aggregate → `usecases.New(usecases.Dependencies{...})` → `handler.New(e, usecase, config, auth)` registers all HTTP routes onto the `echo.Echo` instance.
- [internal/contract/contract.go:78-84](internal/contract/contract.go#L78-L84) — `Start`/`Stop` delegate directly to `echo.Echo.Start`/`Shutdown` — the app runs its own HTTP server via Echo, not the `net/http.Server` wrapper.
- [internal/server/server.go:10-29](internal/server/server.go#L10-L29) — a separate `server.Server` type wraps `net/http.Server` with its own `Run`/`Stop`, but `Contract` does not use it (it starts Echo directly) — this package appears to be unused/dead wiring.
- [internal/config/config.go:60-80](internal/config/config.go#L60-L80) — `config.New` reads `ENV` via `viper.AutomaticEnv()`, resolves `config.<env>.yaml` from `./config`, and unmarshals into `Configuration`, which nests `Application`, `PostgreSQL`, `Google`, `Auth` (including per-`kid` bearer/refresh secret maps), and flat `Department`/`Campus` code-to-name maps used across usecases for validation.

## Open questions

- `internal/server/server.go` looks unreferenced by `internal/contract/contract.go` (which starts Echo directly) — worth confirming it's genuinely dead code, or used in a path not surveyed here (e.g. tests).
- `Auth.ClientId` conflicts in name with `mapstructure:"client_id"` also used for `BearerSecret`'s kid lookup context in the HTTP middleware — worth double-checking config key collisions don't cause confusion between "OAuth client ID" and "allowed client ID map" concepts.
