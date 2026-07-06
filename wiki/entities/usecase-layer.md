---
tags: [architecture]
sources: [raw/2026-07-06-bootstrap-go-backend-usecases.md]
updated: 2026-07-06
ingested_at: 91a596d
---

# Usecase Layer

`internal/usecases/` is the business-logic layer sitting between [HTTP Handler Conventions](http-handler-conventions.md) and the repository layer. Each domain gets an interface + struct pair (e.g. `UserUsecase`/`userUsecase`), constructed via `New*Usecase(...)` and assembled into a single `usecases.Usecases` aggregate that handlers depend on. There is no separate "service" or "domain" layer beneath usecases — they own request validation against runtime config, password hashing, code/token generation, and orchestration across repositories directly.

## Responsibilities

- Validate input against in-memory config maps loaded at startup (department/campus codes) before touching the DB. [internal/usecases/user_usecase.go:66-113](internal/usecases/user_usecase.go#L66-L113)
- Generate derived values — e.g. `eventUsecase.Create` builds a unique event code by hashing nanosecond timestamps. [internal/usecases/event_usecase.go:44-117](internal/usecases/event_usecase.go#L44-L117)
- Orchestrate multiple repositories/usecases per operation (e.g. checking role/user-type/campus existence before creating a private event).

## Cross-cutting logging

Every usecase method wraps its body in `defer func() { LogService(ctx, err) }()`, mirroring the repository layer's `LogRepository` pattern — errors get logged once at the usecase boundary in addition to any handler-level logging.

## Related

- [Usecase Dependency Style Inconsistency](usecase-dependency-style.md) — individual repo interfaces vs. the repository aggregate.
- [Conflated User Create/Upgrade Flow](user-create-upgrade-conflation.md) — a specific usecase worth calling out.
- [Composition Root](composition-root.md) — where the usecase aggregate is built and handed to handlers.
- [Models Layer](models-layer.md) — the request/response DTOs and entities usecases validate, transform, and return.
- [Error Sentinel Catalog](error-sentinel-catalog.md) — the sentinel errors usecases return on validation/orchestration failure.
- [Repository Layer](repository-layer.md) — the data-access layer usecases call into.
