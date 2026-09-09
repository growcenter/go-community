# Spec: `foundation`

**Module id:** `foundation` · **Depends on:** — · **Build order:** 1st
**Parent:** [SPEC.md](SPEC.md)

---

## Objective

Everything the other eight modules stand on: the toolchain bump, the error catalog, the schema, the typed config structs, the entities, and the repository aggregate. Nothing here is user-visible; everything here is depended on. Getting a shape wrong at this layer is expensive to change once four modules have consumed it.

## Scope

**In:** Go 1.26 bump; herr integration + bilingual error catalog; migration `000025`; `internal/models/v3/{errors,config,entities}.go`; repository aggregate wiring; `V3` config block; `make test`/`make lint` targets; integration test harness.

**Out:** Any usecase, any handler, any validator logic. Those belong to their modules.

---

## Acceptance Criteria

### Toolchain

1. ✅ **Done 2026-09-10.** `go.mod` declares `go 1.26.1` and `Dockerfile` builds on `golang:1.26.1-alpine`; `go build ./...` succeeds. Remaining: add `herr` as a direct dependency.
2. ✅ **Verified.** Go 1.26.1 is installed and first on PATH. No CI workflows exist in this repository, so the Dockerfile is the only other pinned toolchain — pin any future CI to 1.26.1 as well. A stale Go 1.25.5 at `/usr/local/go/bin/go` is shadowed but still reachable by explicit path.

### Error catalog

3. Every domain error is a `herr.Define` class in package `v3`. Matching uses `Class.Is(err)`.
4. A completeness test asserts that **every** defined class has an Indonesian translation under `errors.<lowercase_code>.message`. A class without one fails the build.
5. `Accept-Language: en` returns the English message; anything else, including absent, returns Indonesian.
6. Internal detail attached via `Internal`/`With`/`Wrap` never appears in a response body. A test asserts this for at least one class carrying internal detail.

The catalog must define at minimum:

| Code | Kind → HTTP | Raised by |
|---|---|---|
| `NOT_FOUND` | 404 | all |
| `FORBIDDEN` | 403 | all |
| `INVALID_INPUT` | 400 | all — always with field-level `errors[]` |
| `REGISTRATION_CLOSED` | 409 | `registration` |
| `CHANNEL_CLOSED` | 409 | `registration` |
| `QUOTA_FULL` | 409 | `registration`, `attendance` |
| `MARKED_SOLD_OUT` | 409 | `registration` |
| `ALREADY_REGISTERED` | 409 | `registration` |
| `MAX_SESSIONS_EXCEEDED` | 409 | `registration` |
| `NOT_ELIGIBLE` | 403 | `registration`, `attendance` |
| `ANONYMOUS_NOT_ALLOWED` | 403 | `registration` |
| `INVALID_ACCESS_CODE` | 403 | `registration` |
| `EDIT_WINDOW_CLOSED` | 409 | `registration` |
| `ALREADY_CHECKED_IN` | 409 | `attendance` |
| `NOT_CHECKED_IN` | 409 | `attendance` |
| `CATEGORY_REQUIRED` | 400 | `attendance` |
| `ELIGIBILITY_REVOKED` | 403 | `attendance` |
| `NOTHING_TO_UNDO` | 409 | `attendance` |
| `OUT_OF_RANGE` | 403 | `attendance`, `registration` |
| `CHECKIN_WINDOW_CLOSED` | 409 | `attendance` |
| `INVALID_TOKEN` | 400 | `qr` |
| `TOKEN_EXPIRED` | 400 | `qr` |

Adding a class later is routine; adding one without its `id` translation is a review-rejectable defect.

### Migration `000025_v3_event_management`

7. Runs clean up and down against a database already carrying `000022_configs_setup` and the COOL migrations `000023`/`000024`.
8. Creates the v3 tables. **v2 tables are untouched** — a test asserts `events`, `event_instances`, `event_registration_records`, `event_questions` are unmodified.

Schema, with every ADR amendment folded in (§9 of the parent spec lists what changed and why):

```sql
-- events: all four JSONB configs; no event_organizers table (ADR 0001)
events_v3(
  id, code UK, slug UK, title, description, topics text[],
  terms_and_conditions, image_links text[], campus_codes text[],
  status,                         -- draft|published|completed|cancelled|archived
  publish_at, unpublish_at,       -- nullable scheduled visibility flips
  content_sections jsonb, contacts jsonb, venue_address jsonb,
  eligibility jsonb,              -- + recheck_at_checkin: off|note|block  (ADR 0008)
  registration_config jsonb,      -- + channels[], max_sessions_per_person (ADR 0006, 0004)
  geo_config jsonb,               -- + modes.desk_registration             (ADR 0006)
  recurrence jsonb,               -- session_defaults carries attendance settings (ADR 0005)
  created_by, created_at, updated_at, deleted_at
)

event_sessions(
  id, code UK, event_id FK, title, description,
  start_at, end_at,                        -- session period
  register_start_at, register_end_at,      -- registration window
  checkin_open_at, checkin_close_at,       -- check-in window
  location_type, location_name, online_url,
  total_seats int,                         -- 0 = unlimited; ignored when categories exist
  booked_seats int,                        -- row-locked counter
  marked_sold_out bool,
  attendance_modes text[],                 -- EMPTY IS VALID: registration-only
  enable_checkout bool DEFAULT false,
  enable_late_marking bool DEFAULT false,  -- ADR 0005
  generated_from_recurrence bool, occurrence_key,
  status, created_at, updated_at, deleted_at,
  UNIQUE(event_id, occurrence_key)         -- idempotent recurrence generation
)

session_ticket_categories(
  id, session_id FK, code, name, description,
  total_seats int, booked_seats int, marked_sold_out bool, sort_order int
)

registrations(
  id uuid PK, session_id FK, category_id FK NULL,
  registered_by NULL,                      -- NULLABLE — anonymous desk reg (ADR 0002)
  party_size int, status,                  -- confirmed|cancelled
  source,                                  -- web|admin|walk_in
  registered_at, cancelled_at
)

registration_attendees(
  id uuid PK, registration_id FK, community_id NULL, name,
  answers jsonb, answer_revisions jsonb,
  status,                                  -- registered|attended|checked_out|no_show|cancelled
  attended_at, checked_out_at
)

attendance_logs(                           -- APPEND-ONLY. Never updated, never deleted.
  id uuid PK, session_id FK, attendee_id FK NULL, community_id NULL,
  action,                                  -- checkin|checkout
  mode,                                    -- personal_qr|session_qr|registration_qr|manual
                                           --   FOUR values. Headcount is a `manual` sub-case (ADR 0005)
  is_headcount bool DEFAULT false,         --   ...distinguished by this flag, not a fifth mode
  checked_by NULL,                         -- staff actor; NULL for session_qr (self-scanned)
  lat, lng, accuracy_m numeric NULL,
  geo_result,                              -- ok|out_of_range|overridden|not_provided|not_required
  eligibility_result NULL,                 -- ok|not_eligible|not_checked  — STORED (ADR 0008)
  outcome,                                 -- checked_in|already_checked_in|checked_out|
                                           --   already_checked_out|rejected_geo|rejected_not_eligible|
                                           --   rejected_no_category|rejected_other|headcount_undone
  created_at
)

notification_outbox(
  id uuid PK, channel, recipient, template, payload jsonb,
  status,                                  -- pending|sent|failed
  attempts int, scheduled_at, sent_at
)

event_templates(id, name, description, snapshot jsonb, is_system bool, created_by)
```

9. Indexes: `attendance_logs(session_id, created_at)`, `attendance_logs(attendee_id)`, `registrations(session_id, registered_by) WHERE status='confirmed'`, `notification_outbox(status, scheduled_at)`, plus the unique `(event_id, occurrence_key)`.
10. **No `event_organizers` table is created** (ADR 0001). A test asserts it does not exist.

### Config structs

11. `internal/models/v3/config.go` defines typed Go structs for `Eligibility`, `RegistrationConfig`, `GeoConfig`, `Recurrence`, `Contacts`, `ContentSections`, `VenueAddress`, each with `Parse` and `ValidateStrict`.
12. `RegistrationConfig.Channels` defaults to `["web", "desk"]` when absent, and **an empty list fails `ValidateStrict`** — `mode: none` already expresses "nobody registers," and two ways to say it invites drift (ADR 0006).
13. `RegistrationConfig.MaxSessionsPerPerson` defaults to `0` (unlimited).
14. `Eligibility.RecheckAtCheckin` defaults to `off`.
15. `GeoConfig.Modes` carries six rows: `web_registration`, `desk_registration` (**default `off`**), `personal_qr`, `session_qr`, `registration_qr`, `manual`.
16. `Recurrence.SessionDefaults` carries `start_time`, `duration_min`, `total_seats`, **and** `attendance_modes`, `enable_checkout`, `enable_late_marking` (ADR 0005). A struct-completeness test asserts the last three are present — their absence is the silent failure that would generate 52 weeks of untracked sessions.
17. `attendance_modes` accepts an empty list without error at both loose and strict validation.

### Repository aggregate & harness

18. `pgsql.New(db)` returns a `PostgreRepositories` carrying a `V3` sub-aggregate. Existing v1/v2 repositories are unchanged.
19. Each v3 repository exposes small methods; the seat-critical ones — `GetByCodeForUpdate`, `AddBooked`, `ReleaseBooked` — take a context whose transaction is supplied by `Atomic`.
20. `tests/integration/v3/harness_test.go` applies migrations, truncates v3 tables between tests, and `t.Skip`s when `TEST_DATABASE_URL` is unset. `go test ./...` is green on a machine with no database.

### Repo hygiene

21. `make test` and `make lint` targets exist and do what their names say (the README already promises them).
22. `config/config.local.template.yaml` gains the `v3:` block — `app_domain`, `qr_secret`, `email.{host,port,username,password,from}`. Real values go only in the gitignored `config.local.yaml`.

---

## Verification

```bash
go build ./...
go test ./internal/models/v3/...                       # catalog completeness, config defaults
TEST_DATABASE_URL=... go test ./tests/integration/v3/  # migration up/down, harness
golangci-lint run
```

Manual: `make migration_up` then `make migration_down` on a database carrying `000022`–`000024`, twice, leaves no residue.

---

## Boundaries (module-specific)

- **Ask first** before changing any JSONB shape here once a downstream module has consumed it — these structs are the frontend's contract.
- **Never** add a fifth value to `attendance_modes`, a fifth `mode` to `attendance_logs`, or an `event_organizers` table.
- **Never** put logic in this module. If it decides something, it belongs in `policy-kernel` or a usecase.

---

## Open Questions (blocking this module)

1. **Go 1.26 availability** — verify before writing a line. See parent Open Question 5.
2. **Eligibility re-check placement** — this spec puts `recheck_at_checkin` inside the `eligibility` JSONB. Parent Open Question 1.
3. **Headcount undo linkage** — this schema nets by counting outcomes and carries no `undoes_log_id`. If the stricter self-reference is wanted, it must be added here, not later. Parent Open Question 2.
