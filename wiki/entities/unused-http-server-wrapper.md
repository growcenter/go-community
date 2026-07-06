---
tags: [tech-debt]
sources: [raw/2026-07-06-bootstrap-go-backend-server-wiring.md]
updated: 2026-07-06
ingested_at: 91a596d
---

# Unused HTTP Server Wrapper

`internal/server/server.go` defines a `Server` type wrapping `net/http.Server` with its own `Run`/`Stop` methods. [internal/server/server.go:10-29](internal/server/server.go#L10-L29)

The actual [Composition Root](composition-root.md) (`internal/contract/contract.go`) does not use this type — it starts and stops Echo's built-in server directly via `echo.Echo.Start`/`Shutdown`. [internal/contract/contract.go:78-84](internal/contract/contract.go#L78-L84)

This package appears to be dead wiring left over from before the switch to Echo's own server lifecycle, or from an earlier refactor. Not confirmed dead — worth checking for references in tests or other entry points before removing.
