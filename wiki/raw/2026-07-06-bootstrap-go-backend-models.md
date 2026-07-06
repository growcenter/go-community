# Bootstrap: Go Backend — Models

_Surveyed 2026-07-06 at commit a0da240._

## Overview

`internal/models/` holds all domain structs for the service: GORM-mapped DB entities, per-endpoint request/response DTOs, and a centralized error catalog. There's no separate DTO/entity split by file — each file (e.g. `user_model.go`, `event_model.go`) bundles the DB struct, its DB-output query structs, and every request/response pair for that domain's handlers, tagged with `json` and `validate` struct tags read by the validator middleware.

## Key structures / flows

- [internal/models/user_model.go:12-44](internal/models/user_model.go#L12-L44) — `User` is the canonical GORM entity; array fields (`UserTypes`, `Roles`) are `pq.StringArray` mapped to Postgres `text[]`, and it has GORM `foreignKey` associations to `Campus` and `CoolCategory`.
- [internal/models/user_model.go:117-161](internal/models/user_model.go#L117-L161) — request/response pairs are declared per use case (`CreateUserRequest`/`CreateUserResponse`) rather than a single shared user shape, so validation rules differ per endpoint (e.g. `CreateVolunteerRequest` requires `IsKOM100`/`IsBaptized`, but `CreateUserRequest` does not).
- [internal/models/user_model.go:91-115](internal/models/user_model.go#L91-L115) — `To*()` methods on response types (e.g. `ToCreateUser()`) convert an already-populated response struct into itself, used as an explicit "shape this for the wire" step called from usecases.
- [internal/models/event_model.go:15-39](internal/models/event_model.go#L15-L39) — `Event` entity mirrors the request/response pattern; `AllowedFor` (`public`/`private`) gates whether `AllowedUsers`/`AllowedRoles`/`AllowedCampuses` are enforced.
- [internal/models/event_model.go:491-582](internal/models/event_model.go#L491-L582) — `DefineAvailabilityStatus` computes one of `available/unavailable/full/soon/walkin` from seat counts and registration window via a type-switch over multiple DB-output shapes (`GetAllEventsDBOutput`, `GetEventByCodeDBOutput`, etc.) — the same status logic is duplicated per shape rather than through a shared interface.
- [internal/models/error_model.go:12-93](internal/models/error_model.go#L12-L93) — every domain error is a single package-level `errors.New(...)` sentinel, grouped by comment banners (Event, Auth, Rate Limiter, Pagination, etc.) rather than per-file.
- [internal/models/error_model.go:95-407](internal/models/error_model.go#L95-L407) — `ErrorMapping(err error)` is a giant switch mapping each sentinel to an HTTP status + string `Status` code + message, called centrally by the HTTP layer to translate usecase errors into responses.

## Open questions

- Sentinel-error switch in `ErrorMapping` grows linearly with every new error — no error-code/status metadata attached to the sentinel itself (e.g. via a typed error struct), so adding an error requires touching two places (declaration + switch case).
- `DefineAvailabilityStatus` takes `interface{}` and type-switches on 4+ concrete DB-output structs with near-identical field extraction — a shared interface with accessor methods could collapse this.
