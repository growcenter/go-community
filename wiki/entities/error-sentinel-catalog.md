---
tags: [concept]
sources: [raw/2026-07-06-bootstrap-go-backend-models.md]
updated: 2026-07-06
ingested_at: 91a596d
---

# Error Sentinel Catalog

Every domain error in the service is a package-level `errors.New(...)` sentinel declared in `internal/models/error_model.go`, grouped by comment banners (Event, Auth, Rate Limiter, Pagination, etc.) rather than split across per-domain files. [internal/models/error_model.go:12-93](internal/models/error_model.go#L12-L93)

## Mapping to HTTP responses

`ErrorMapping(err error)` is a single large switch that maps each sentinel to an HTTP status code, a string `Status` code (e.g. `"DATA_NOT_FOUND"`, `"FORBIDDEN_REGISTRATION"`), and a message. It's called centrally by the HTTP layer to translate errors returned from the [Usecase Layer](usecase-layer.md) into responses. [internal/models/error_model.go:95-407](internal/models/error_model.go#L95-L407)

## Adding a new error

Adding a new domain error requires touching two places: declaring the sentinel and adding a case to `ErrorMapping`. There's no error-code/status metadata attached to the sentinel itself (e.g. via a typed error struct), so the switch grows linearly and by hand with every new error. A typed error carrying its own HTTP status/code could collapse this into one declaration site, if the churn becomes a problem.
