---
tags: [tech-debt]
sources: [raw/2026-07-06-bootstrap-go-backend-usecases.md]
updated: 2026-07-06
ingested_at: 91a596d
---

# Usecase Dependency Style Inconsistency

Within the [Usecase Layer](usecase-layer.md), there's no consistent convention for how a usecase gets its repository dependencies:

- `userUsecase` takes 7 individual repository interfaces directly as constructor params (`UserRepository`, `UserRelationRepository`, `CampusRepository`, ...) plus `config.Configuration` and `authorization.Auth` — no repository facade is used. [internal/usecases/user_usecase.go:36-64](internal/usecases/user_usecase.go#L36-L64)
- `eventUsecase` instead takes a single `pgsql.PostgreRepositories` aggregate (accessed as `eu.r.Event`, `eu.r.Role`, ...) plus a `FeatureFlagUsecase` for composition with another usecase. [internal/usecases/event_usecase.go:28-42](internal/usecases/event_usecase.go#L28-L42)

Both styles coexist across the codebase with no apparent rule for when to use which. Worth establishing a convention (or documenting the existing rationale, if one exists) before adding new usecases, since it affects how easy each usecase is to unit-test and how much surface area its constructor exposes.
