---
tags: [concept]
sources: [raw/2026-07-06-bootstrap-go-backend-usecases.md]
updated: 2026-07-06
ingested_at: 91a596d
---

# Conflated User Create/Upgrade Flow

`userUsecase.Create` (in the [Usecase Layer](usecase-layer.md)) handles two conceptually different flows through the same endpoint, branching on whether a user record already exists by email/phone:

- **No existing record** — creates a brand-new `User` row with a generated `CommunityID`, hashed password, and the submitted profile fields. [internal/usecases/user_usecase.go:186-233](internal/usecases/user_usecase.go#L186-L233)
- **Existing record found** — if the matched user's type is `internal` or `cool` category, the existing "shell" record is "upgraded" in place via `Update` (setting password, campus, cool, department, etc.); otherwise the request is rejected with `ErrorAlreadyExist`. [internal/usecases/user_usecase.go:114-186](internal/usecases/user_usecase.go#L114-L186)

This implies the system pre-populates shell user records (likely internal/cool-category imports) that get "claimed" on first real signup, distinct from ordinary duplicate-account prevention. Anyone modifying `Create` needs to keep both flows in mind — a change aimed at "new user validation" can silently affect the shell-upgrade path too.
