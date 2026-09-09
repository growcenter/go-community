# Spec: `policy-kernel`

**Module id:** `policy-kernel` · **Depends on:** `foundation` · **Build order:** 2nd (parallel with `qr`)
**Parent:** [SPEC.md](SPEC.md)

---

## Objective

The four decisions that must be identical everywhere they are asked, expressed as pure functions: *may this person register*, *are they close enough*, *is this answer valid*, *when does this recurring event next happen*. Every module that needs one of these calls the same function — that sameness is the whole point, because ADR 0007 exists precisely because eligibility was being decided in one place and skipped in three others.

## Scope

**In:** `internal/pkg/eligibility`, `internal/pkg/geo`, `internal/pkg/form`, `internal/pkg/recurrence`. Table-driven unit tests, no database, no HTTP, no clock reads except an injected `now`.

**Out:** Deciding *when* to call them. That is each consuming module's job and is specified there.

---

## Contracts

```go
// eligibility
type User struct {
	CommunityID string
	Roles       []string
	UserTypes   []string
	CampusCode  string
}
// Check reports whether u may register under cfg. A zero User (anonymous) is
// never eligible unless cfg.Audience == "everyone".
func Check(cfg v3.Eligibility, u User) (ok bool, reason string)

// geo
type Coords struct{ Lat, Lng, AccuracyM float64 }
type Result struct {
	GeoResult string // ok|out_of_range|overridden|not_provided|not_required
	Err       error  // v3.ErrOutOfRange when blocked
}
func Validate(cfg v3.GeoConfig, mode string, c *Coords, override bool) Result

// form
func Visible(field v3.FormField, answers map[string]any) bool
func Validate(fields []v3.FormField, audience string, answers map[string]any) error
func ValidateSchema(fields []v3.FormField) error

// recurrence
func Expand(r v3.Recurrence, from time.Time, aheadPeriods int) []Occurrence
type Occurrence struct {
	Key      string    // stable; feeds event_sessions.occurrence_key
	StartAt  time.Time
	EndAt    time.Time
	Defaults v3.SessionDefaults
}
```

---

## Acceptance Criteria

### `eligibility.Check`

1. `audience: everyone` and `audience: members` both admit any user with a non-empty `CommunityID`. They are aliases; `members` exists for admin-UI clarity.
2. `audience: rules` admits a user matching **any** non-empty list — OR across `roles`/`user_types`/`campuses`/`community_ids`, and OR within each list.
3. An empty `rules` object under `audience: rules` admits nobody, and `ValidateSchema` rejects it at publish rather than shipping an event nobody can join.
4. **An anonymous user (empty `CommunityID`) is eligible only under `everyone`.** This is the function-level half of ADR 0007: restricted events reject anonymous registrations because there is no account for the rules to evaluate.
5. `reason` names which constraint failed, in a form the caller can render — refusals must say why, never fail generically (ADR 0007).
6. The full audience × rules matrix is table-driven, including empty lists, unknown roles, and a user matching one list but not another.

### `geo.Validate`

7. Distance is Haversine, in Go, no PostGIS.
8. Per-mode knob: `off` → `not_required`, no error. `warn` → checks, records `ok` or `out_of_range`, **never errors**. `require` → errors `OUT_OF_RANGE` when outside the radius **or when coordinates are missing**.
9. `staff_override: true` plus `override: true` on a `require` failure proceeds and returns `overridden` — with the actual coordinates preserved, never discarded.
10. `staff_override: false` ignores `override: true` entirely; a self-scanner cannot override their own geo check.
11. `desk_registration` is a valid mode key and **defaults to `off`** (ADR 0006).
12. Missing coordinates under `warn` return `not_provided`, not an error.
13. Coordinates and reported accuracy are returned for storage as evidence; they are never treated as trustworthy input.

### `form.Visible` / `Validate` / `ValidateSchema`

14. Leaf operators: `eq`, `neq`, `in`, `not_in`, `answered`, `not_answered`, `gt`, `gte`, `lt`, `lte`, `contains`. Combinators `all` and `any`, nestable to any depth.
15. **A hidden field's `required` is ignored, and any answer submitted for it is dropped server-side.** Visibility is re-evaluated on the server and never trusted from the client.
16. `answered_by` audiences: `primary` (default) — registrant only; `everyone` — registrant and each companion; `companions` — companion seats only. `required` applies within that audience only.
17. Built-in types validate for free: `email` format, Indonesian phone (`^\+?62\d{8,13}$` or `^0\d{8,12}$`), 16-digit `nik`. Plus `name`, `text`, `textarea`, `number`, `select`, `multiselect`, `checkbox`, `date`.
18. `ValidateSchema` rejects condition cycles (`a` depends on `b` depends on `a`) and references to unknown field keys — at publish, not at registration.
19. Validation failures render as field-keyed entries suitable for herr's `FieldError`, naming the field and the problem.

### `recurrence.Expand`

20. `weekly` with `by_day`, `monthly` with `by_month_day`, and `custom_days` with `custom_dates` all expand correctly, honouring `interval` and `until`.
21. Expansion is **deterministic** — the same rule and the same `from` produce the same occurrences, so generation can be re-run safely.
22. `Occurrence.Key` is stable across runs; it is what makes `UNIQUE(event_id, occurrence_key)` idempotent.
23. **`Defaults` carries `attendance_modes`, `enable_checkout` and `enable_late_marking`** alongside `start_time`/`duration_min`/`total_seats`. A test asserts a rule specifying `attendance_modes: [personal_qr]` produces occurrences carrying it — the failure this guards is 52 generated weeks that silently record nobody (ADR 0005).
24. Business "today" and all date arithmetic use `Asia/Jakarta` via `common.GetLocation()`.

---

## Verification

```bash
go test ./internal/pkg/eligibility/ ./internal/pkg/geo/ ./internal/pkg/form/ ./internal/pkg/recurrence/ -v
go test ./internal/pkg/... -cover     # these are the cheapest tests in the codebase; expect high coverage
```

Every test in this module is table-driven, runs in milliseconds, and needs no database. If a test here needs a database, the logic is in the wrong module.

---

## Boundaries (module-specific)

- **Always** keep these functions pure — values in, values out. An injected `now` is the only concession.
- **Never** call a repository, read config from disk, or touch `time.Now()` directly from these packages.
- **Ask first** before changing a signature once a consuming module has shipped against it.

---

## Notes for consumers

This module changes *what the rules are*, not *who asks*. The 2026-09-09 spec is explicit that its eligibility work changes **who calls** `Check`, not the rule logic — so existing coverage stands, with additions only where a new caller introduces a genuinely new case (notably: the anonymous user under a restricted audience, criterion 4).
