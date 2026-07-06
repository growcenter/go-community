---
tags: [architecture]
sources: [raw/2026-07-06-bootstrap-go-backend-models.md]
updated: 2026-07-06
ingested_at: 91a596d
---

# Models Layer

`internal/models/` holds every domain struct for the service: GORM-mapped DB entities, DB-output query structs, and per-endpoint request/response DTOs, all bundled into one file per domain (e.g. `user_model.go`, `event_model.go`) rather than split by concern. Structs are tagged with `json` (wire shape) and `validate` (rules read by the validator middleware used in [HTTP Handler Conventions](http-handler-conventions.md)).

## Entity shape

`User` is the canonical GORM entity: array fields (`UserTypes`, `Roles`) are `pq.StringArray` mapped to Postgres `text[]`, with GORM `foreignKey` associations to `Campus` and `CoolCategory`. [internal/models/user_model.go:12-44](internal/models/user_model.go#L12-L44)

`Event` mirrors the same pattern, with `AllowedFor` (`public`/`private`) gating whether `AllowedUsers`/`AllowedRoles`/`AllowedCampuses` are enforced. [internal/models/event_model.go:15-39](internal/models/event_model.go#L15-L39)

## Per-endpoint request/response DTOs

Rather than one shared shape per entity, each endpoint gets its own request/response pair with its own validation rules — e.g. `CreateVolunteerRequest` requires `IsKOM100`/`IsBaptized` while `CreateUserRequest` does not, even though both create a `User`. [internal/models/user_model.go:117-161](internal/models/user_model.go#L117-L161)

Response types carry `To*()` methods (e.g. `ToCreateUser()`) as an explicit "shape this for the wire" step, called from the [Usecase Layer](usecase-layer.md) before returning to handlers. [internal/models/user_model.go:91-115](internal/models/user_model.go#L91-L115)

## Related

- [Error Sentinel Catalog](error-sentinel-catalog.md) — the centralized error values and HTTP-status mapping also living in this package.
- [Duplicated Availability Status Logic](availability-status-duplication.md) — a specific piece of model logic worth calling out separately.
