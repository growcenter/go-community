---
tags: [architecture]
sources: [raw/2026-07-06-bootstrap-go-backend-server-wiring.md]
updated: 2026-07-06
ingested_at: 91a596d
---

# Composition Root

`internal/contract/contract.go` is the single place the app wires all its dependencies together. `cmd/api/main.go` owns process lifecycle only — it builds config, hands off to the contract, and waits for `SIGINT`/`SIGTERM` to trigger a graceful shutdown.

## Startup order

`Contract.New` constructs dependencies in a fixed sequence, each layer depending on the one before it:

1. Connect to Postgres via GORM, then ping the underlying `*sql.DB` — fatal on any failure. [internal/contract/contract.go:24-45](internal/contract/contract.go#L24-L45)
2. Build the Google OAuth client and the JWT `Authorization` helper. [internal/contract/contract.go:47-57](internal/contract/contract.go#L47-L57)
3. Build the repository aggregate via `pgsql.New(psql)`. [internal/contract/contract.go:59-60](internal/contract/contract.go#L59-L60)
4. Build the usecase aggregate via `usecases.New(usecases.Dependencies{...})`, passing the repository aggregate, Google client, auth, and config. [internal/contract/contract.go:62-68](internal/contract/contract.go#L62-L68)
5. Register all HTTP routes via `handler.New(e, usecase, config, auth)` onto a shared `echo.Echo` instance. [internal/contract/contract.go:70-71](internal/contract/contract.go#L70-L71)

## Process lifecycle

`main` starts serving in a goroutine and blocks on OS signals; on shutdown it calls `contract.Stop` with a 10-second timeout context. [cmd/api/main.go:17-47](cmd/api/main.go#L17-L47)

`Contract.Start`/`Stop` delegate directly to `echo.Echo.Start`/`Shutdown` — the app runs Echo's built-in HTTP server, not a hand-rolled `net/http.Server`. [internal/contract/contract.go:78-84](internal/contract/contract.go#L78-L84) See [Unused HTTP Server Wrapper](unused-http-server-wrapper.md) for a related dead-code note.

## Related

- [Configuration Loading](configuration-loading.md) — how `config.Configuration` is loaded and shaped before being passed into the composition root.
- [HTTP Handler Conventions](http-handler-conventions.md) — what `handler.New` registers in step 5.
- [Usecase Layer](usecase-layer.md) — what `usecases.New` builds in step 4.
