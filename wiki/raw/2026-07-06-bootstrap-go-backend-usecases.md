# Bootstrap: Go Backend — Usecases

_Surveyed 2026-07-06 at commit a0da240._

## Overview

`internal/usecases/` is the business-logic layer sitting between HTTP handlers and repositories. Each domain has an interface + struct pair (e.g. `UserUsecase`/`userUsecase`) constructed with `New*Usecase(...)`, taking multiple repository interfaces, `config.Configuration`, and sometimes other usecases (composition) as dependencies. Usecases own request validation against runtime config (campus/department maps loaded from YAML), password hashing, code/token generation, and orchestration across repositories — there is no separate "service" or "domain" layer beneath this.

## Key structures / flows

- [internal/usecases/user_usecase.go:36-64](internal/usecases/user_usecase.go#L36-L64) — `userUsecase` depends directly on 7 repository interfaces plus `config.Configuration` and `authorization.Auth`, injected via constructor; no repository facade/aggregate is used here (contrast with `eventUsecase` below).
- [internal/usecases/user_usecase.go:66-113](internal/usecases/user_usecase.go#L66-L113) — `Create` validates department/cool/campus existence against in-memory config maps (`uu.cfg.Department`, `uu.cfg.Campus`) before ever touching the DB, then checks `userTypes` via `uu.utr.GetByArray`.
- [internal/usecases/user_usecase.go:114-186](internal/usecases/user_usecase.go#L114-L186) — `Create` branches on whether a user already exists by email/phone: existing internal/cool-category users get "upgraded" in place (`Update`), others are rejected with `ErrorAlreadyExist` — user creation and "claim an existing shell record" are the same endpoint.
- [internal/usecases/event_usecase.go:28-42](internal/usecases/event_usecase.go#L28-L42) — `eventUsecase` instead depends on a single `pgsql.PostgreRepositories` aggregate (`eu.r.Event`, `eu.r.Role`, ...) plus a `FeatureFlagUsecase` — usecases are inconsistent about whether they take individual repo interfaces or the aggregate.
- [internal/usecases/event_usecase.go:44-117](internal/usecases/event_usecase.go#L44-L117) — `Create` (event) generates a unique `eventCode` via hash of nanosecond timestamps, then validates `AllowedFor` public/private rules: private events must specify non-nil roles/users/campuses, each checked for existence against repos/config before the event row is built.
- Cross-cutting: every usecase method wraps its body in `defer func() { LogService(ctx, err) }()`, mirroring the repository layer's `LogRepository` pattern — errors are logged once at the usecase boundary in addition to any handler-level logging.

## Open questions

- [internal/usecases/authorization_usecase.go:1-45](internal/usecases/authorization_usecase.go#L1-L45) — file is headed `// NOT USED` but still defines a live `Auth`/`Authorization` JWT-generation type; worth confirming whether this is dead code safe to delete or a fallback still referenced somewhere.
- Repository dependency style is inconsistent across usecases (individual interfaces vs. the `PostgreRepositories` aggregate) — no apparent convention for when to use which.
