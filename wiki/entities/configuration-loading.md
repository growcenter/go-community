---
tags: [architecture, concept]
sources: [raw/2026-07-06-bootstrap-go-backend-server-wiring.md]
updated: 2026-07-06
ingested_at: 91a596d
---

# Configuration Loading

`config.New` loads environment-specific YAML into a single `Configuration` struct, read once at startup and passed by reference through the whole [Composition Root](composition-root.md).

## Resolution

The environment is read from the `ENV` env var (via `viper.AutomaticEnv()`), which selects `config.<env>.yaml` from the `./config` directory (e.g. `config.dev.yaml`, `config.prod.yaml`). [internal/config/config.go:60-80](internal/config/config.go#L60-L80)

## Shape

`Configuration` nests:

- `Application` — name, version, port, environment, host, timeouts, log options.
- `PostgreSQL` — DB connection fields.
- `Google` — OAuth client ID/secret/redirect/state.
- `Auth` — per-`kid` bearer and refresh JWT secret maps, bearer/refresh durations, a static API key, and a `ClientId` map of allowed client IDs.
- `Department` / `Campus` — flat maps of code → name, used across usecases (e.g. user/event creation) to validate department and campus codes without a DB round-trip.

[internal/config/config.go:12-58](internal/config/config.go#L12-L58)

## Open question: naming collision

`Auth.ClientId` (`map[string]bool` of allowed client IDs, checked in `GeneralMiddleware`) and Google's `ClientID` (OAuth app client ID) are distinct concepts that share a very similar name across two different config sections. Worth double-checking call sites don't confuse the two. [internal/config/config.go:44-57](internal/config/config.go#L44-L57)
