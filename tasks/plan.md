# Implementation Plan: v3 Platform

Source of truth: `docs/intent/v3-platform.md`, `CAPABILITY-MAP.md` and the seven `SPEC-*.md` files. Task list: `tasks/todo.md`. This plan holds every task in full, grouped by module, plus the cross-module order, the export map and the risks.

## Overview

Build the seven modules of the capability map on `/api/v3`, in dependency order, as 82 small tasks (S or M). Each task is one vertical slice with its own acceptance criteria and verification commands. v1 and v2 stay compiled and return 410 once the deprecation middleware lands in Phase 1.

## Architecture decisions

- Build order follows the capability map: platform, identity, permissions, qr, then events and cool in parallel, then calendar.
- Identity is split around permissions. Sign-in tasks (I1 to I9, I12) come first, then the permissions core (M1 to M4), then the identity admin endpoints (I10, I11, I13). No temporary guard is ever built for admin routes.
- Migration numbers are fixed here so parallel work does not collide: identity 000023, qr 000024, events 000025, cool 000026. Permissions grants and backfill use 000027 and 000028 if they land after cool, otherwise the next free numbers in that order.
- Two new dependencies need approval before their task starts: `github.com/skip2/go-qrcode` (Q7) and a PDF library, `gofpdf` or `maroto` (E11). `herr` itself was approved in the spec.
- The project logs with zap, not slog. Where a spec snippet says `slog`, use the existing zap logger.
- Every write uses `Atomic`. Every non-public v3 route is guarded by the identity JWT middleware plus `permission.Require`.

## Cross-module order and checkpoints

### Phase 1: platform foundation

- P1: Go 1.26 bump and herr dependency compile
- P2: herrx registry with COMMON_ classes and completeness test
- P3: response.OK/Fail envelope with tests
- P5: configdb typed getters and boot-time defaults
- P4: deprecation middleware returns 410 on v1 and v2
- P6: outbox table, repository and Notifier interface
- P7: retention purge and scheduler endpoint
- P8: dispatch endpoint and email Notifier

Checkpoint:
- [ ] `go build ./... && go test ./internal/pkg/... ./internal/platform/...` pass
- [ ] Every v1 and v2 route returns 410 and `/health` returns 200
- [ ] Outbox concurrency test sends each row once
- [ ] Review with the owner before the next phase

### Phase 2: identity core (sign-in)

- I1: Migration 000023 with account status backfill
- I2: `USER_` error catalog and locale
- I3: Models, repositories and audit writer
- I4: OTP package with limits and cooldown
- I12: v3 JWT issuer and middleware
- I5: User check, OTP send and verification, create user with claim
- I6: Login, new device verification, device limit
- I7: Refresh and logout with device revoke
- I8: PIN reset
- I9: Me endpoints and the email-required gate

Checkpoint:
- [ ] Register, login, new device, refresh and logout pass integration tests
- [ ] `community_id` unchanged for every backfilled row
- [ ] Three sends in two hours trigger the cooldown
- [ ] Review with the owner before the next phase

### Phase 3: permissions core

- M1: Grants table migration and `PERM_` catalog
- M2: Action registry, title table and matrix test
- M3: Evaluator `Can` and `Scopes` with parents
- M4: `Require` middleware, resolvers and `Actor`

Checkpoint:
- [ ] Matrix test covers every title times every action
- [ ] `campus_admin` of BSD passes `Can` on an event with parent BSD and fails on PIK
- [ ] `permission.Require` returns 403 `PERM_FORBIDDEN` on a miss
- [ ] Review with the owner before the next phase

### Phase 4: identity admin and permissions endpoints

- I10: Admin create-pending, status change, admin reset
- I11: Admin list, detail, patch, delete
- I13: Dashboard and audit-log endpoints
- I14: OTP on/off switch
- M5: Grant and revoke endpoints with audit
- M6: List, me, check and titles endpoints
- M7: Backfill migration from frozen tables
- M8: `EnsureSeeds` and boot wiring in `contract.go`
- I15: Wire v3 identity into the composition root and configdb defaults

Checkpoint:
- [ ] Admin endpoints are guarded by `permission.Require`, no allow-list exists
- [ ] Grant, revoke and backfill integration tests pass
- [ ] `grep -rn 'UserMiddleware(' internal/deliveries/http/v3` returns nothing
- [ ] Review with the owner before the next phase

### Phase 5: qr

- Q1: Token codec with kid map and config validation
- Q2: Action registry with Allowed and Dispatch
- Q3: QR_ error catalog and locale
- Q4: Scan log table, repository and purge hook
- Q5: Resolve endpoint
- Q6: Act endpoint
- Q7: Personal QR PNG endpoint (needs dependency approval)
- Q8: Wire into the composition root

Checkpoint:
- [ ] Personal token is byte-identical across calls and verifies after key rotation
- [ ] `act` on a not-allowed action returns 403 and never runs the handler
- [ ] Every URL is under 160 characters in the length test
- [ ] Review with the owner before the next phase

### Phase 6a: events (can run in parallel with 6b)

- E1: Migration 000025 and EVENT_ catalog
- E2: JSONB configs with category, conditions and self_scan_protection
- E3: Eligibility package with AND conditions and reader interfaces
- E4: Geo, form and recurrence pure packages
- E5: Entities, repositories and audit writer
- E6: Event and session CRUD with permission guards and audit
- E7: Publish gate and templates without Sunday Service
- E8: Recurrence generation per occurrence
- E9: Registration with seat lock, eligibility and idempotent re-registration
- E10: Registration channels and desk registration
- E11: Attendee tokens and PDF ticket for windowed events
- E12: Check-in core with personal_qr and session_qr
- E13: registration_qr, manual lookup and eligibility recheck
- E14: Checkout, lateness, headcount and undo
- E21: Scan labels on check-in
- E22: Roster sync endpoint
- E23: Offline batch check-in sync
- E15: QR action registration for events
- E16: Attendee list, live counts and computed session state
- E17: Reports CSV, XLSX and JSON
- E18: Attendance matrix report
- E24: Label counts in reports
- E19: Staff dashboard, audit endpoint and permission scopes
- E20: Pool caps, wiring, seeds and v2 410

Checkpoint:
- [ ] 200 concurrent registrations for one seat leave `booked_seats = 1`
- [ ] Poster scan from home with `window_geo` returns `EVENT_GEO_OUT_OF_RANGE`
- [ ] Matrix report has one column per occurrence in CSV, XLSX and JSON
- [ ] Review with the owner before the next phase

### Phase 6b: cool (can run in parallel with 6a)

- C1: Migration and COOL_ catalog
- C2: COOL entities and DTOs
- C3: COOL create, edit, list and detail
- C4: Members add and remove
- C5: Titles and facilitators endpoints
- C6: Meeting create, edit and cancel
- C7: Meeting QR image and scan handler
- C8: Mark attendance and close
- C9: Edit attendance after close with revisions
- C10: Guests add and link
- C11: Member "mine" endpoints
- C12: Stats package and reporting endpoints
- C13: Wiring and registration

Checkpoint:
- [ ] Close with one `pending` row returns `COOL_MEETING_PENDING`
- [ ] A scan never changes `status`, only `scanned_at`
- [ ] Removing a member revokes their COOL titles in the same transaction
- [ ] Review with the owner before the next phase

### Phase 7: calendar

- L1: Pure calendar package: DTOs, query validation, status mapping, `CAL_` catalog
- L2: `.ics` serializer with category line
- L3: Sessions query on v3 tables with organizer refs and parents
- L4: Meetings query on cool tables
- L5: Usecase tiering, eligibility, category filter, merge
- L6: JSON and `.ics` endpoints, mount, docs

Checkpoint:
- [ ] A plain member sees only eligible published sessions plus their own COOL meetings
- [ ] A cancelled meeting is absent from JSON and `.ics`
- [ ] `grep -rn 'event_organizers\|cool_facilitators\|ErrorMapping' internal/**/calendar` returns nothing
- [ ] Review with the owner before the next phase

## Export map

Task dependency lines name upstream work by export. This table maps each name to the task that delivers it.

| Export name used in task dependencies | Task |
|---|---|
| `platform: herrx.Def`, `locale.Register`, `AssertCatalogComplete` | P2 |
| `platform: response.OK/Fail` | P3 |
| `platform: deprecate.Middleware` | P4 |
| `platform: configdb.*`, `EnsureDefaults` | P5 |
| `platform: outbox producer`, `Notifier` | P6 |
| `platform: retention.Purge`, `RegisterTable` | P7 |
| `identity: migration 000023` | I1 |
| `identity: user_audit_logs writer` | I3 |
| `identity: Reader.Profile` | I9 |
| `identity: pending account creator (POST /v3/admin/users)` | I10 |
| `identity: v3 JWT middleware`, `Actor(c)`, `DeviceID(c)` | I12 |
| `permissions: permission.Register`, `Titles()` | M2 |
| `permissions: permission.Can`, `Scopes` | M3 |
| `permissions: permission.Require`, `FromParam`, `FromQuery`, `Global`, `WithParents`, `Actor` | M4 |
| `permissions: grants usecase Grant/Revoke` | M5 |
| `qr: qr.Codec`, `Kind*`, `Payload` | Q1 |
| `qr: qr.Registry`, `Action`, `ActionCtx` | Q2 |
| `qr: migration 000024` | Q4 |
| `qr: qr.RenderPNG` | Q7 |
| `events: migration 000025`, `v3_*` tables | E1 |
| `events: category constant set` | E2 |
| `events: eligibility.Check` | E3 |
| `cool: cool_meetings and cool_members migration` | C1 |
| `cool: MeetingRepository.ListInRange`, `MemberRepository.ActiveCoolID` | C2, C6 |

## Risks

| Risk | Impact | Mitigation |
|---|---|---|
| herr helper names (`From`, `HTTPStatus`, `Body`) differ from the events plan | Medium | P1 runs `go doc` on herr first and P2/P3 adapt the two call sites. `KindGone` is added to herr before P4 |
| Production `user_types` holds values the backfill does not know | High | M7 runs `SELECT DISTINCT unnest(user_types)` on production before the migration is written and maps every value |
| Supabase free-tier pool exhausted on event day | High | E20 sets pool caps, E9 keeps the seat lock short, and the 200-concurrency tests in E9 and E12 run in CI |
| Two parallel tracks (events, cool) both edit `contract.go` | Medium | E20 and C13 are the only tasks that touch `contract.go`, and they run one after the other |
| Email provider quota on spike days | Medium | Outbox retries with `max_attempts`, and P8 picks a provider with at least 3,000 sends a day |

## Open questions

- Approve `skip2/go-qrcode` before Q7 and a PDF library before E11.
- Choose the free email provider before P8.
- Decide whether Cloud Scheduler gets its own key or reuses `X-API-Key` before P7.
- Confirm the web app owns the `/q/<token>` route and the value of `application.public_url` before Q5.

## Module tasks


---

# Module: platform

# Plan: platform module

Source: `SPEC-platform.md`. Build order: P1 → P8. Every task leaves `go build ./...` green.

## Task P1: Go 1.26 bump and herr dependency compile

**Description:** Add `github.com/jeremygprawira/herr`, with `KindGone` released, to `go.mod`. Bump the toolchain from `go 1.23.0` to `go 1.26` and make sure that the existing tree still builds. Record the real names of `herr.From`, `herr.HTTPStatus`, `Body` and `Attrs` from `go doc`, so P3 does not guess.

**Acceptance criteria:**
- [ ] `go.mod` says `go 1.26` and lists `github.com/jeremygprawira/herr`.
- [ ] `go doc github.com/jeremygprawira/herr` shows `KindGone` and the helper names P3 will call, noted in the P3 task before P3 starts.
- [ ] Existing v1/v2 code compiles unchanged.

**Verification:**
- [ ] `go get github.com/jeremygprawira/herr@latest && go mod tidy`
- [ ] `go build ./...`
- [ ] Manual: `go doc github.com/jeremygprawira/herr | grep -i 'gone\|func From\|HTTPStatus'`

**Dependencies:** None
**Files likely touched:** `go.mod`, `go.sum`
**Estimated scope:** S

## Task P2: herrx registry with COMMON_ classes and completeness test

**Description:** Create `internal/pkg/herrx` with `Def`, `AllClasses`, `EnglishMessages` and the six `COMMON_` classes. Create `internal/pkg/locale` with `Register(lang, msgs)`, `Install()` and `Pick(acceptLanguage)`. Ship the shared completeness test that every module reuses. Each class in `AllClasses` must have a non-empty `id` message that differs from English, and a duplicate code panics.

**Acceptance criteria:**
- [ ] `herrx.Def("X", kind, en)` twice with the same code panics at init.
- [ ] Removing one `id` entry from `common_locale.go` makes `go test ./internal/pkg/herrx/` fail.
- [ ] `locale.Pick("")` and `locale.Pick("fr")` return `id`. `locale.Pick("en-US")` returns `en`.

**Verification:**
- [ ] `go test ./internal/pkg/herrx/ ./internal/pkg/locale/`
- [ ] `go build ./...`

**Dependencies:** P1
**Files likely touched:** `internal/pkg/herrx/herrx.go`, `internal/pkg/herrx/common.go`, `internal/pkg/herrx/common_locale.go`, `internal/pkg/herrx/herrx_test.go`, `internal/pkg/locale/locale.go`
**Estimated scope:** M

## Task P3: response.OK/Fail envelope with tests

**Description:** Add `internal/pkg/response/v3.go` with `Meta`, `OK`, `Created`, `Fail` and the private `withTopLevelMessage`. `Fail` converts any error to a herr value and picks the language from `Accept-Language`. It logs internal detail through the existing zap logger in `internal/pkg/logger`, then writes the body. The old `response.go` is not edited.

**Acceptance criteria:**
- [ ] `OK` body has `data`, optional `meta`, never `errors`. `Fail` body always has `code` and `message`. With field errors it also has `errors[]`.
- [ ] A plain `errors.New("boom")` through `Fail` yields 500 `COMMON_INTERNAL` and the word `boom` is absent from the body.
- [ ] No `Accept-Language` header yields the Indonesian message. `Accept-Language: en` yields English.

**Verification:**
- [ ] `go test ./internal/pkg/response/`
- [ ] `go build ./...`
- [ ] Manual: `grep -n 'slog' internal/pkg/response/v3.go` returns nothing (zap is the project logger).

**Dependencies:** P2
**Files likely touched:** `internal/pkg/response/v3.go`, `internal/pkg/response/v3_test.go`
**Estimated scope:** S

## Checkpoint: error contract
- [ ] `go build ./... && go test ./internal/pkg/...` green.
- [ ] One throwaway handler returning `herrx.ErrNotFound.New()` produces `{"code":"COMMON_NOT_FOUND","message":"..."}` with 404 in an `httptest` run.
- [ ] Human review of the catalog and envelope before any module registers codes.

## Task P4: deprecation middleware returns 410 on v1 and v2

**Description:** Add `internal/platform/deprecate` with `Middleware()`. Apply it in `internal/deliveries/http/main_handler.go` to the `api` group before the v1 and v2 handlers register. It matches on the path prefixes `/api/v1` and `/api/v2`. No file under `v1/` or `v2/` changes. `/api/health` and `/` stay live.

**Acceptance criteria:**
- [ ] `GET /api/v1/campus` returns 410 with `code = COMMON_DEPRECATED_ENDPOINT` and an Indonesian message.
- [ ] `GET /api/health` still returns 200.
- [ ] `git diff --stat internal/deliveries/http/v1 internal/deliveries/http/v2` is empty.

**Verification:**
- [ ] `go test ./internal/platform/deprecate/`
- [ ] `go build ./...`
- [ ] Manual: `export ENV=DEV && go run ./cmd/api/main.go` then `curl -i localhost:<port>/api/v1/campus` shows 410.

**Dependencies:** P3
**Files likely touched:** `internal/platform/deprecate/middleware.go`, `internal/platform/deprecate/middleware_test.go`, `internal/deliveries/http/main_handler.go`
**Estimated scope:** S

## Task P5: configdb typed getters and boot-time defaults

**Description:** Add `internal/pkg/configdb` over the existing `pgsql.ConfigRepository.GetByKey`. It offers `Get`, `GetInt`, `GetBool` and `GetDuration`, each falling back to a default on a missing or bad value. `EnsureDefaults(ctx, map[string]string)` upserts missing keys. Register the five `platform.*` keys from the spec table in `internal/contract/contract.go` after `pgsql.New`.

**Acceptance criteria:**
- [ ] If the row is missing or `Value = "abc"`, `GetInt(ctx, "platform.outbox.batch_size", 50)` returns 50.
- [ ] After boot, `SELECT key FROM config WHERE key LIKE 'platform.%'` returns the five keys. A second boot inserts nothing new.
- [ ] `ConfigRepository` gains an `Upsert` used only by `EnsureDefaults`.

**Verification:**
- [ ] `go test ./internal/pkg/configdb/`
- [ ] `TEST_DATABASE_URL="postgres://postgres:$DB_PASSWORD@localhost:5888/community_db?sslmode=disable" go test ./tests/integration/platform/ -run Configdb -v`
- [ ] `go build ./...`

**Dependencies:** P1
**Files likely touched:** `internal/pkg/configdb/configdb.go`, `internal/pkg/configdb/configdb_test.go`, `internal/repositories/pgsql/config_pg_repository.go`, `internal/contract/contract.go`, `tests/integration/platform/configdb_test.go`
**Estimated scope:** M

## Task P6: outbox table, repository and Notifier interface

**Description:** Add migration `000023_notification_outbox`, the next free number after `000022_configs_setup`. Add the `notification_outbox` model and a repository with `Insert(tx, row)`, `ClaimPending(ctx, batch)` using `FOR UPDATE SKIP LOCKED`, `MarkSent` and `MarkFailed`. Define the `Notifier` interface. No sender yet. Producers insert inside their own `Atomic` callback.

**Acceptance criteria:**
- [ ] Migration up and down both apply cleanly on an empty database.
- [ ] Two goroutines calling `ClaimPending(50)` against 100 pending rows claim 100 distinct rows.
- [ ] `MarkFailed` past `platform.outbox.max_attempts` sets `status = failed`. Below the limit it leaves `pending` with `attempts + 1`.

**Verification:**
- [ ] `make migration name=notification_outbox` then `make migration_up` (after editing the placeholder connection string in `Makefile`)
- [ ] `TEST_DATABASE_URL="postgres://postgres:$DB_PASSWORD@localhost:5888/community_db?sslmode=disable" go test ./tests/integration/platform/ -run Outbox -v`
- [ ] `go build ./...`

**Dependencies:** P5
**Files likely touched:** `tests/integration/db/migrations/000023_notification_outbox.up.sql`, `tests/integration/db/migrations/000023_notification_outbox.down.sql`, `internal/platform/outbox/model.go`, `internal/platform/outbox/repository.go`, `tests/integration/platform/outbox_test.go`
**Estimated scope:** M

## Checkpoint: deprecation and tunables
- [ ] App boots with `ENV=DEV`, seeds the five keys, and v1/v2 answer 410.
- [ ] `go test ./...` green with `TEST_DATABASE_URL` set.
- [ ] Human review of the purge list and defaults.

## Task P7: retention purge and scheduler endpoint

**Description:** Add `internal/platform/retention` with `Purge(ctx)`. It runs one short `DELETE` per listed table using the `platform.retention_days.*` keys and returns counts. Mount `POST /api/v3/internal/retention/purge` in `main_handler.go` under a new `/api/v3/internal` group guarded by the existing `X-API-Key` check. The purge list starts with `auth_challenges`, `notification_outbox` and `qr_scan_logs`. A table that does not exist yet is skipped, so the endpoint works before identity and qr land.

**Acceptance criteria:**
- [ ] With `platform.retention_days.outbox = 30`, a `sent` row 31 days old is deleted and a `sent` row 29 days old stays.
- [ ] A `pending` outbox row of any age is never deleted.
- [ ] Calling the endpoint without `X-API-Key` returns 401.

**Verification:**
- [ ] `go test ./internal/platform/retention/`
- [ ] `TEST_DATABASE_URL="postgres://postgres:$DB_PASSWORD@localhost:5888/community_db?sslmode=disable" go test ./tests/integration/platform/ -run Purge -v`
- [ ] `make generate-docs`

**Dependencies:** P5, P6
**Files likely touched:** `internal/platform/retention/purge.go`, `internal/platform/retention/handler.go`, `internal/platform/retention/purge_test.go`, `internal/deliveries/http/main_handler.go`, `tests/integration/platform/purge_test.go`
**Estimated scope:** M

## Task P8: dispatch endpoint and email Notifier

**Description:** Add the dispatcher. It claims a batch, routes each row to the `Notifier` for its `channel`, and marks `sent` or bumps attempts. An unknown channel is marked `failed` with `last_error = "no sender"`. Expose `POST /api/v3/internal/notifications/dispatch` beside the purge route, and an in-process `Dispatch(ctx)` for the fast path. Add one SMTP email `Notifier`. Its provider fields go under a new `Email` section in `internal/config/config.go` and the YAML template.

**Acceptance criteria:**
- [ ] With a fake `Notifier` that counts calls, two concurrent dispatch calls over 100 pending rows send exactly 100.
- [ ] A row with `channel = whatsapp` becomes `failed` with `last_error = "no sender"` on first dispatch.
- [ ] A `Notifier` error leaves the row `pending` with `attempts = 1` and `last_error` set.

**Verification:**
- [ ] `go test ./internal/platform/outbox/`
- [ ] `TEST_DATABASE_URL="postgres://postgres:$DB_PASSWORD@localhost:5888/community_db?sslmode=disable" go test ./tests/integration/platform/ -run Dispatch -v`
- [ ] `make generate-docs`
- [ ] Manual: point `Email` at a Mailpit or Brevo sandbox, insert one row, call dispatch, see the email.

**Dependencies:** P6, P7
**Files likely touched:** `internal/platform/outbox/dispatch.go`, `internal/platform/outbox/email.go`, `internal/platform/outbox/handler.go`, `internal/platform/outbox/dispatch_test.go`, `internal/config/config.go`
**Estimated scope:** M

## Checkpoint: platform complete
- [ ] All nine success criteria in `SPEC-platform.md` pass, each with a named test.
- [ ] `go test ./...` green locally with and without `TEST_DATABASE_URL`.
- [ ] Human review before identity starts registering `USER_` codes.

## Exports

Other modules import only these:

- `go-community/internal/pkg/herrx`: `Def`, `AllClasses`, `EnglishMessages`, `ErrInvalidInput`, `ErrNotFound`, `ErrForbidden`, `ErrUnauthorized`, `ErrInternal`, `ErrDeprecatedEndpoint`, test helper `AssertCatalogComplete(t, lang)`
- `go-community/internal/pkg/locale`: `Register(lang string, msgs map[string]string)`, `Install()`, `Pick(acceptLanguage string) string`
- `go-community/internal/pkg/response`: `OK`, `Created`, `Fail`, `Meta`
- `go-community/internal/pkg/configdb`: `Get`, `GetInt`, `GetBool`, `GetDuration`, `EnsureDefaults`
- `go-community/internal/platform/outbox`: `Notifier` (interface), `Row`, `Repository.Insert`, `Dispatch`
- `go-community/internal/platform/retention`: `RegisterTable(name, configKey)` so identity and qr add their tables

## Risks

| Risk | Impact | Mitigation |
|---|---|---|
| herr helper names in the spec (`From`, `HTTPStatus`, `Body`, `Attrs`) differ from the real API | Med | P1 records the real names from `go doc` before P3 starts |
| Spec snippet uses `slog`; the project logs with zap | Low | P3 uses `logger.Logger` (zap); P3 manual check greps for `slog` |
| `make migration_up` has placeholder credentials and existing migrations may not match production state | Med | P7 runs the migration on a fresh Docker database first; production apply is a human step |

---

# Module: identity

# Plan: identity

Source: `SPEC-identity.md`. Upstream: `platform` (herrx, response, configdb, locale, outbox, Atomic). Every task writes through `TransactionRepository.Atomic`, never `Transaction`. Person key is `community_id` (existing Luhn number from `internal/pkg/generator/account_number.go`).

Unit test command: `go test ./internal/pkg/otp/ ./internal/usecases/identity/...`
Integration command: `TEST_DATABASE_URL="postgres://postgres:$DB_PASSWORD@localhost:5888/community_db?sslmode=disable" go test ./tests/integration/identity/ -v`
Build: `go build ./...`

## Task I1: Migration 000023 with account status backfill
**Description:** Add the new columns to `users`. Create `user_departments`, `kom_certifications`, `user_devices`, `auth_challenges` and `user_audit_logs`. Backfill `users.status`: `active` when a password hash exists, `pending` when there is no password and a `user_types` category of `internal` or `cool`. `community_id` and `id` are untouched. Down migration drops only what up added.
**Acceptance criteria:**
- [ ] `make migration_up` on a copy of the v2 schema creates the six changes and every pre-existing `community_id` is unchanged.
- [ ] A row with a password hash is `active`. A shell row (no password, internal/cool type) is `pending`. Nothing else changes status.
- [ ] `make migration_down` returns the schema to 000022.
**Verification:**
- [ ] `go build ./...`
- [ ] Integration test `TestBackfill` in `tests/integration/identity/migration_test.go` seeds three legacy rows and asserts the three statuses.
- [ ] Manual: `SELECT status, count(*) FROM users GROUP BY 1` on a production copy, compare against the shell-upgrade count (Open Question 1).
**Dependencies:** None.
**Files likely touched:**
- `tests/integration/db/migrations/000023_identity_v3.up.sql`
- `tests/integration/db/migrations/000023_identity_v3.down.sql`
- `tests/integration/identity/migration_test.go`
**Estimated scope:** S

## Task I2: `USER_` error catalog and locale
**Description:** Define every `USER_` class from the spec table with `herrx.Def`. Register the Indonesian messages with `locale.Register`. Add `identity.FieldErrors(err)`, which turns validator errors into `herrx.ErrInvalidInput` field entries.
**Acceptance criteria:**
- [ ] Every code in the spec table exists with the listed Kind and EN message.
- [ ] The shared platform completeness test passes for the `USER_` prefix, and removing one `id` message fails it.
- [ ] `FieldErrors` produces `code`, top-level `message` and `errors[]` through `response.Fail`.
**Verification:**
- [ ] `go test ./internal/models/identity/`
- [ ] `go build ./...`
**Dependencies:** `platform: herrx.Def`, `platform: locale.Register`, `platform: response.Fail`.
**Files likely touched:**
- `internal/models/identity/errors.go`
- `internal/models/identity/locale.go`
- `internal/models/identity/errors_test.go`
**Estimated scope:** S

## Task I3: Models, repositories and audit writer
**Description:** GORM entities for the six tables. One repository per table under `internal/repositories/pgsql/identity`. An `audit.Write(ctx, r, actor, target, action, before, after)` helper that stores a field-level diff and strips `pin_hash`, `code_hash` and `password`. Add the aggregate to `PostgreRepositories` so `Atomic` threads it.
**Acceptance criteria:**
- [ ] Each repository has Create, Get and the list queries the endpoints in I5 to I11 need, and nothing more.
- [ ] `audit.Diff` returns only changed fields and never a secret field.
- [ ] Repositories are reachable from the tx-scoped `*PostgreRepositories` inside `Atomic`.
**Verification:**
- [ ] `go test ./internal/usecases/identity/...` (diff unit test)
- [ ] `go build ./...`
**Dependencies:** I1.
**Files likely touched:**
- `internal/models/identity/entities.go`
- `internal/repositories/pgsql/identity/*.go`
- `internal/repositories/pgsql/pgsql.go` (aggregate)
- `internal/usecases/identity/audit.go`, `audit_test.go`
**Estimated scope:** M

## Checkpoint: I1 to I3
- [ ] `go build ./...` and both test commands pass.
- [ ] Migration applied and rolled back once on a local database.
- [ ] Human review of the backfill rule before it touches production data.

## Task I4: OTP package with limits and cooldown
**Description:** Pure package `internal/pkg/otp` with an injected clock. It creates a 6-digit code, hashes it and verifies in constant time. It enforces expiry, `max_attempts`, `resend_seconds`, `max_sends` per `window_hours` per identifier and per IP, and `cooldown_hours`. Limits come in as a struct so configdb stays outside.
**Acceptance criteria:**
- [ ] Third wrong code starts a cooldown. Fourth send inside the window starts a cooldown. A send 2 minutes after the last is refused.
- [ ] A used code cannot be reused and an expired code fails with the expired reason.
- [ ] `Verify` uses `subtle.ConstantTimeCompare` on the hash.
**Verification:**
- [ ] `go test ./internal/pkg/otp/`
**Dependencies:** None.
**Files likely touched:**
- `internal/pkg/otp/otp.go`
- `internal/pkg/otp/otp_test.go`
**Estimated scope:** S

## Task I5: User check, OTP send and verification, create user with claim
**Description:** The four public endpoints and the usecases behind them. `POST /v3/users/check` returns the next step (`register` or `login`) and the login method. It sends nothing and writes nothing. It sits behind Echo's in-memory rate limiter with `identity.check_max_per_minute` per client IP. `POST /v3/auth/otp` sends one OTP through the outbox for purpose `register` or `pin_reset` and rejects purpose `device`. It returns the real error for an `active` account, a cooled-down identifier or a missing email. `POST /v3/auth/otp/verification` returns a short-lived verification token that holds the identifier and the purpose. `POST /v3/users` takes a `register` token, sets the PIN, campus and name and trusts the device. For a `pending` row it sets `active` and `claimed_at`. Then it issues tokens.
**Acceptance criteria:**
- [ ] A new identifier ends with one `users` row, one `auth_challenges` row, one outbox row, one `user_devices` row and a token pair.
- [ ] A `pending` identifier ends `active` with the same `community_id` and one audit row with action `claimed`.
- [ ] `auth/otp` with purpose `register` on an `active` identifier returns 409 `USER_ALREADY_EXISTS` and inserts no outbox row. A cooled-down identifier returns 403 `USER_OTP_COOLDOWN` and writes one `otp_cooldown` audit row.
- [ ] `users/check` returns `register` for a new and for a `pending` identifier and `login` for an `active` one, and inserts no row. Call 61 inside one minute from one IP returns 403 `USER_TOO_MANY_REQUESTS`.
- [ ] `POST /v3/users` with a `pin_reset` token returns 401 `USER_VERIFICATION_TOKEN_INVALID`.
**Verification:**
- [ ] Integration `TestRegister`, `TestClaim`, `TestUsersCheck`, `TestOTPCooldown`.
- [ ] `make generate-docs`
**Dependencies:** I2, I3, I4, I12, `platform: outbox producer`, `platform: configdb.GetInt`.
**Files likely touched:**
- `internal/usecases/identity/register.go`, `otp.go`
- `internal/models/identity/dto.go`
- `internal/deliveries/http/v3/identity/users_handler.go`, `auth_handler.go`
- `tests/integration/identity/register_test.go`
**Estimated scope:** M

## Task I6: Login, new device verification, device limit
**Description:** `login` with identifier plus PIN on a trusted device returns tokens. Unknown device returns `USER_NEW_DEVICE_VERIFICATION_REQUIRED` and sends an OTP with purpose `device`. A second `login` call with a `device` verification token trusts the device. At `max_active_devices` it requires `replace_oldest` before it revokes the oldest row. PIN lockout counts wrong PINs and locks for `pin_lockout_minutes`. Legacy password login works only for rows with a password hash while `password_login_enabled` is true.
**Acceptance criteria:**
- [ ] Wrong PIN five times returns `USER_PIN_LOCKED` for the configured minutes.
- [ ] Second device without `replace_oldest` returns 409. With it the first device's row is revoked.
- [ ] Two concurrent `login` calls with a `device` token for the last slot produce exactly one trusted device (user row locked in `Atomic`).
- [ ] `login` with a `register` token returns 401 `USER_VERIFICATION_TOKEN_INVALID`.
**Verification:**
- [ ] Integration `TestLogin`, `TestNewDevice`, `TestDeviceLimitRace`, `TestPinLockout`, `TestLegacyPassword`.
- [ ] Unit test for lockout math in `internal/usecases/identity/pin_test.go`.
**Dependencies:** I5, I12.
**Files likely touched:**
- `internal/usecases/identity/login.go`, `pin.go`, `pin_test.go`
- `internal/deliveries/http/v3/identity/auth_handler.go`
- `tests/integration/identity/login_test.go`
**Estimated scope:** M

## Checkpoint: I4 to I6
- [ ] A new member can register and log in end to end against a local database.
- [ ] Both test commands pass.
- [ ] Human review of the OTP limits against the spec numbers (3 sends / 2 h, 3 min gap, 3 attempts, 2 h cooldown).

## Task I7: Refresh and logout with device revoke
**Description:** `refresh` reads `X-Refresh-Token` and checks `typ = refresh`. It rejects a revoked or missing `did` row with `USER_DEVICE_REVOKED` and rejects a deactivated account. Then it bumps `last_seen_at` and returns a new pair. `logout` revokes the calling device.
**Acceptance criteria:**
- [ ] Refresh after logout returns 401 `USER_DEVICE_REVOKED`.
- [ ] Refresh for a deactivated account returns 403 `USER_ACCOUNT_DEACTIVATED`.
- [ ] Logout writes one audit row with action `device_revoked`.
**Verification:**
- [ ] Integration `TestRefreshAfterRevoke`, `TestRefreshDeactivated`.
**Dependencies:** I6, I12.
**Files likely touched:**
- `internal/usecases/identity/session.go`
- `internal/deliveries/http/v3/identity/auth_handler.go`
- `tests/integration/identity/session_test.go`
**Estimated scope:** S

## Task I8: PIN reset
**Description:** The code comes from `auth/otp` with purpose `pin_reset` (I5). `POST /v3/auth/pin/reset` takes a `pin_reset` verification token and sets a new PIN hash. It clears the lockout counter and writes one audit row with action `pin_reset`. With OTP off, `auth/otp` with purpose `pin_reset` returns `USER_OTP_DISABLED`.
**Acceptance criteria:**
- [ ] After reset the old PIN fails and the new PIN logs in.
- [ ] `auth/otp` with purpose `pin_reset` on an unknown identifier returns 404 `USER_NOT_FOUND` and sends nothing.
- [ ] One audit row per successful reset, with no `pin_hash` in `changes`.
**Verification:**
- [ ] Integration `TestPinReset`.
**Dependencies:** I5.
**Files likely touched:**
- `internal/usecases/identity/pin.go`
- `internal/deliveries/http/v3/identity/auth_handler.go`
- `tests/integration/identity/pin_test.go`
**Estimated scope:** S

## Task I9: Me endpoints and the email-required gate
**Description:** `GET /v3/me` returns profile, departments, max KOM level and the `email_required` flag. `PATCH /v3/me` edits own fields, including email. `GET` and `DELETE /v3/me/devices` manage devices. A middleware on every member route except `GET /v3/me` and `PATCH /v3/me` returns `USER_EMAIL_REQUIRED` when the account has no email.
**Acceptance criteria:**
- [ ] A legacy account with no email gets 422 `USER_EMAIL_REQUIRED` on `GET /v3/me/devices`, sets an email through `PATCH /v3/me`, and the next call succeeds.
- [ ] `DELETE /v3/me/devices/{id}` revokes only the caller's own device and writes one audit row.
- [ ] `PATCH /v3/me` cannot change `community_id`, `status` or `pin_hash`.
**Verification:**
- [ ] Integration `TestEmailRequiredGate`, `TestMeDevices`.
- [ ] `make generate-docs`
**Dependencies:** I7, I12.
**Files likely touched:**
- `internal/usecases/identity/me.go`
- `internal/deliveries/http/v3/identity/me_handler.go`
- `internal/deliveries/http/v3/identity/email_gate.go`
- `tests/integration/identity/me_test.go`
**Estimated scope:** M

## Checkpoint: I7 to I9
- [ ] Full member lifecycle (register, login, refresh, forgot PIN, edit profile, logout) passes in integration.
- [ ] `grep -rn 'userTypes\|Roles' internal/deliveries/http/v3/identity` returns nothing.
- [ ] Human review of the email-required behaviour on a legacy account.

## Task I10: Admin create-pending, status change, admin reset
**Description:** `POST /v3/admin/users` creates a `pending` row with a generated `community_id`, a name and one identifier. It sends no OTP. `PATCH .../status` sets account status (`active`, `deactivated`) or `volunteer_status`. Deactivation revokes every device. `POST .../reset` issues a one-time code shown once in the response. The code is stored hashed as an `auth_challenges` row with purpose `admin_reset` and works in `auth/otp/verification` for the purposes `device` and `pin_reset`. Route guards name the actions `user.create_pending` and `user.manage`. The evaluator arrives with `permissions`, so until then the guard is the identity middleware plus a named TODO.
**Acceptance criteria:**
- [ ] Create returns `community_id` and status `pending`, one audit row, no outbox row.
- [ ] Deactivate revokes all devices, next login and refresh return 403 `USER_ACCOUNT_DEACTIVATED`.
- [ ] Admin reset code works once and writes one audit row with action `admin_reset`.
**Verification:**
- [ ] Integration `TestAdminCreatePending`, `TestDeactivate`, `TestAdminReset`.
**Dependencies:** I8, I9.
**Files likely touched:**
- `internal/usecases/identity/admin.go`
- `internal/deliveries/http/v3/identity/admin_users_handler.go`
- `tests/integration/identity/admin_test.go`
**Estimated scope:** M

## Task I11: Admin list, detail, patch, delete
**Description:** Cursor-paginated list that reuses `internal/pkg/cursor`. Detail with departments and KOM. Patch for profile, `departments[]` with `is_primary` and `max_departments_per_user`, KOM rows and file URLs. Soft delete. Every write is one `Atomic` transaction with one audit row.
**Acceptance criteria:**
- [ ] List pages forward and backward with the same cursor format as v2.
- [ ] Patch with two departments and one primary stores both rows and one audit row with the field diff.
- [ ] Delete sets `deleted_at`, revokes devices, and the row is absent from list and detail.
**Verification:**
- [ ] Integration `TestAdminList`, `TestAdminPatchDepartments`, `TestAdminDelete`.
- [ ] `make generate-docs`
**Dependencies:** I10.
**Files likely touched:**
- `internal/usecases/identity/admin.go`
- `internal/repositories/pgsql/identity/users.go`
- `internal/deliveries/http/v3/identity/admin_users_handler.go`
- `tests/integration/identity/admin_list_test.go`
**Estimated scope:** M

## Task I12: v3 JWT issuer and middleware
**Description:** Token issuer with claims `sub`, `did`, `jti`, `iat`, `exp` and `typ` only. `kid` comes from `config.Auth`. TTLs come from `identity.access_ttl_minutes` and `identity.refresh_ttl_days`. Echo middleware `identity.RequireAuth()` verifies the access token and loads the user. It rejects `deactivated`, `pending` and a revoked `did`. It stores `community_id` and `device_id` on the context for `identity.Actor(c)`. `middleware.UserMiddleware` is not touched.
**Acceptance criteria:**
- [ ] The access token decodes to exactly the six claims.
- [ ] A revoked device or a deactivated account gets 401 or 403 with the spec code.
- [ ] `identity.Actor(c)` returns the `community_id` on every guarded route.
**Verification:**
- [ ] Unit `internal/usecases/identity/token_test.go` (claim set, wrong `typ`, wrong `kid`).
- [ ] `go build ./...`
**Dependencies:** I2, I3, `platform: configdb.GetInt`.
**Files likely touched:**
- `internal/usecases/identity/token.go`, `token_test.go`
- `internal/deliveries/http/v3/identity/middleware.go`
**Estimated scope:** S

## Task I13: Dashboard and audit-log endpoints
**Description:** `GET /v3/admin/users/dashboard` returns the demographic aggregates from the user-management design. Groups are campus, gender, age band, marital status, department and KOM level, one query each. `GET /v3/admin/audit-logs` lists `user_audit_logs` with cursor pagination and filters by actor, target, action and date range.
**Acceptance criteria:**
- [ ] Dashboard counts match seeded rows and exclude soft-deleted users.
- [ ] Audit list returns one row per admin mutation from I10 and I11 and never includes `pin_hash`.
- [ ] Both routes name `user.dashboard.view` and `user.audit.view`.
**Verification:**
- [ ] Integration `TestDashboard`, `TestAuditList`.
- [ ] `make generate-docs`
**Dependencies:** I11.
**Files likely touched:**
- `internal/usecases/identity/dashboard.go`, `auditlog.go`
- `internal/repositories/pgsql/identity/dashboard.go`
- `internal/deliveries/http/v3/identity/admin_users_handler.go`
- `tests/integration/identity/dashboard_test.go`
**Estimated scope:** M

## Task I14: OTP on/off switch
**Description:** `identity.otp_enabled` read per request through configdb. When false: `auth/otp` with purpose `register` returns the verification token at once, `login` on an unknown device with the right PIN trusts it, and `auth/otp` with purpose `pin_reset` returns `USER_OTP_DISABLED`. `auth/otp/verification` returns `USER_OTP_DISABLED` for every code except an admin-assisted code. Flipping the key writes one audit row with action `otp_switch`. The audit write hooks the configdb setter for this key.
**Acceptance criteria:**
- [ ] With the switch off, register completes with zero `auth_challenges` and zero outbox rows.
- [ ] With the switch off, a new device is trusted on the first correct PIN.
- [ ] Flipping the switch writes exactly one `otp_switch` audit row.
**Verification:**
- [ ] Integration `TestOTPDisabled`, `TestOTPSwitchAudit`.
**Dependencies:** I6, I8.
**Files likely touched:**
- `internal/usecases/identity/register.go`, `login.go`, `pin.go`
- `internal/usecases/identity/otp_switch.go`
- `tests/integration/identity/otp_switch_test.go`
**Estimated scope:** S

## Task I15: Wire v3 identity into the composition root and configdb defaults
**Description:** Mount `/api/v3/users`, `/api/v3/auth`, `/api/v3/me` and `/api/v3/admin` in `main_handler.go`. Construct the identity usecases in `contract.go`. Call `configdb.EnsureDefaults` with every `identity.*` key and default from the spec table. v2 user and token routes are left for the platform 410 middleware.
**Acceptance criteria:**
- [ ] App boots with an empty `config` table and every `identity.*` key exists afterwards with its default.
- [ ] `GET /api/v3/me` without a token returns 401 `COMMON_UNAUTHORIZED`.
- [ ] Swagger lists every v3 identity route.
**Verification:**
- [ ] `go build ./... && make generate-docs`
- [ ] Manual: `make run`, call `POST /api/v3/auth/otp` with a test email and purpose `register`, see one outbox row.
**Dependencies:** I5 to I14, `platform: configdb.EnsureDefaults`, `platform: deprecate middleware` (for the v2 410).
**Files likely touched:**
- `internal/deliveries/http/main_handler.go`
- `internal/deliveries/http/v3/identity/handler.go`
- `internal/contract/contract.go`
- `internal/usecases/usecases.go`
**Estimated scope:** S

## Checkpoint: I10 to I15
- [ ] Every success criterion in `SPEC-identity.md` has a passing test.
- [ ] `go build ./...`, both test commands, `golangci-lint run ./internal/...` clean.
- [ ] Human review of the admin surface and the OTP switch before `permissions` starts.

## Exports

Other modules import these:

- `internal/deliveries/http/v3/identity.RequireAuth() echo.MiddlewareFunc` (I12), the JWT guard every v3 route uses.
- `internal/deliveries/http/v3/identity.Actor(c echo.Context) string` (I12), the caller's `community_id`.
- `internal/deliveries/http/v3/identity.DeviceID(c echo.Context) string` (I12).
- `internal/usecases/identity.AuditWriter` with `Write(ctx, r *pgsql.PostgreRepositories, actor, target, action string, before, after any) error` (I3), used by `permissions` and `cool` for grant and membership rows.
- `internal/usecases/identity.AdminUsecase.CreatePending(ctx, actor string, req CreatePendingRequest) (CreatePendingResponse, error)` (I10), the guest account path for `cool`.
- `internal/usecases/identity.Reader` with `Profile(ctx, communityID) (campus string, departments []string, status string, err error)` (I9), the reader the events eligibility checker uses.
- `internal/models/identity` DTOs and `USER_` classes (I2).
- `internal/repositories/pgsql/identity` aggregate on `PostgreRepositories` (I3).

## Risks

| Risk | Impact | Mitigation |
|---|---|---|
| Backfill marks the wrong legacy rows `pending`, locking real members out | High | I1 integration test on a production copy, count reviewed by a human before migration (Open Question 1) |
| Admin routes in I10, I11 and I13 need a permission guard | High | Resolved by ordering: M1 to M4 are built before I10, so these routes use `permission.Require` from day one. No temporary allow-list. |
| `users/check` lets anyone test an identifier for membership | Medium | Accepted on 2026-10-01 for the identifier-first UI. Per-IP limit `identity.check_max_per_minute`, in memory per app instance. `users/check` sends nothing and writes nothing |

---

# Module: permissions

# Plan: permissions

Source: `SPEC-permissions.md`. Upstream exports assumed: `platform: herrx.Def`, `platform: response.OK/Fail`, `platform: locale.Register`, `identity: v3 JWT middleware` (sets `community_id` on `echo.Context`), `identity: user_audit_logs writer` (a repository method callable inside `Atomic`).

Existing code this plan reuses: `pgsql.PostgreRepositories` aggregate and `Atomic` in `internal/repositories/pgsql/transaction_pg_repository.go`, route mounting in `internal/deliveries/http/main_handler.go`, the `v1`/`v2` handler pattern for a `NewV3Handler`. Existing code this plan retires for v3: `middleware.UserMiddleware` in `internal/deliveries/http/middleware/authorization.go` (reads `claims.UserTypes`/`claims.Roles`, hardcodes `superadmin`).

## Task M1: Grants table migration and `PERM_` catalog

**Description:** Create the `permission_grants` table with the partial unique index on active rows and the check that exactly one of `title` or `action` is set. Define the eight `PERM_` error classes with `herrx.Def` and their Indonesian messages, so every later task can return them.

**Acceptance criteria:**
- [ ] `migrate up` creates `permission_grants` with the columns in the spec, a `CHECK ((title IS NULL) <> (action IS NULL))`, and a unique index on `(community_id, title, action, scope_kind, scope_ref) WHERE revoked_at IS NULL`.
- [ ] Inserting two identical active rows fails at the database. Inserting the same row after the first is revoked succeeds.
- [ ] Every `PERM_` class has an `id` message, and the shared platform completeness test passes for the `permissions` catalog.

**Verification:**
- [ ] `make migration name=permission_grants_v3` then `go build ./...`
- [ ] `go test ./internal/pkg/herrx/ ./internal/models/permissions/`
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/permissions/ -run TestGrantsTable -v`

**Dependencies:** `platform: herrx.Def`, `platform: locale.Register`

**Files likely touched:**
- `tests/integration/db/migrations/0000NN_permission_grants_v3.up.sql` and `.down.sql` (NN after identity's 000023)
- `internal/models/permissions/grant.go`
- `internal/models/permissions/errors.go`
- `internal/models/permissions/locale.go`
- `tests/integration/permissions/table_test.go`

**Estimated scope:** M

## Task M2: Action registry, title table and matrix test

**Description:** Build the pure core. `permission.Register(action, module)` panics on a duplicate or a wrong prefix. `titles.go` holds the seed title table, one fixed scope kind per title, with COOL strength inclusion. A generated matrix test walks every title times every registered action. This task has no database and no HTTP.

**Acceptance criteria:**
- [ ] `Register` panics on a duplicate name and on a name whose prefix is not the registering module.
- [ ] If any title in `titles.go` names an unregistered action, a test fails. If `cool_leader` lacks an action of `cool_core_team`, or `cool_facilitator` lacks one of `cool_leader`, the same test fails.
- [ ] The matrix test is generated from `titles.go` and the registry with no hand-written list.

**Verification:**
- [ ] `go test ./internal/pkg/permission/...`
- [ ] Manual: add a fake action to a title in a scratch branch and see the completeness test fail.

**Dependencies:** None

**Files likely touched:**
- `internal/pkg/permission/actions.go`
- `internal/pkg/permission/titles.go`
- `internal/pkg/permission/titles_test.go`
- `internal/pkg/permission/actions_test.go`

**Estimated scope:** S

## Task M3: Evaluator `Can` and `Scopes` with parents

**Description:** Implement the pure evaluation over an already-loaded grant slice. It covers superadmin bypass, title expansion, and scope match on the resource or one of its parents. `Scopes` returns direct refs, parent refs by kind, and `all`. Add the grants repository method that loads active rows for one actor. Add the per-request loader that caches them on the Echo context only.

**Acceptance criteria:**
- [ ] `Can` returns true for a `campus` grant on BSD against an event resource with `Parents{"campus": "BSD"}` and false against PIK. A `cool` title never matches an `event` resource.
- [ ] `Scopes(actor, "event.view_internal", "event")` for a `department_leader` of MUSIC returns no refs, `parents["department"] == ["MUSIC"]`, `all == false`. For `event_admin` it returns `all == true`.
- [ ] A revoked row is ignored, and a `superadmin` row returns true with no other rows loaded.

**Verification:**
- [ ] `go test ./internal/pkg/permission/...`
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/permissions/ -run TestEvaluate -v`

**Dependencies:** M1, M2

**Files likely touched:**
- `internal/pkg/permission/evaluator.go`
- `internal/pkg/permission/evaluator_test.go`
- `internal/repositories/pgsql/permissions/grants_pg_repository.go`
- `internal/repositories/pgsql/main_pg_repository.go` (add `Grants` to the aggregate)
- `tests/integration/permissions/evaluate_test.go`

**Estimated scope:** M

## Checkpoint: Core

- [ ] `go build ./... && go test ./internal/pkg/permission/... ./internal/models/permissions/` pass.
- [ ] The matrix test lists every seed title from the spec table and nothing else.
- [ ] Review the title table with the human before any route depends on it.

## Task M4: `Require` middleware, resolvers and `Actor`

**Description:** Add the Echo middleware `permission.Require(action, resolver)`. It runs after the identity JWT middleware, loads the actor's grants once per request, and evaluates. On a miss it returns `PERM_FORBIDDEN` through `response.Fail`. Ship the resolvers `FromParam(kind, param)`, `FromQuery(kind, param)`, `Global()`, and `WithParents(resolver, loader)`. A consumer uses the last one to attach campus and department parents. `Actor(c)` reads the `community_id` the identity middleware set.

**Acceptance criteria:**
- [ ] A route guarded with `Require("cool.confirm_attendance", FromParam("cool", "coolId"))` returns 403 `PERM_FORBIDDEN` for an actor without the grant and 200 with it.
- [ ] When two `Require` calls or `Can` calls run in the same request, grants are loaded once.
- [ ] `Actor(c)` returns the `community_id` from the identity middleware. When the request is unauthenticated, it returns an empty string.

**Verification:**
- [ ] `go test ./internal/pkg/permission/...` (uses `httptest` with a fake grant loader)
- [ ] `go build ./...`

**Dependencies:** M3, `identity: v3 JWT middleware`, `platform: response.OK/Fail`

**Files likely touched:**
- `internal/pkg/permission/middleware.go`
- `internal/pkg/permission/resolvers.go`
- `internal/pkg/permission/middleware_test.go`

**Estimated scope:** S

## Task M5: Grant and revoke endpoints with audit

**Description:** Implement `POST /v3/permissions/grants` and `DELETE /v3/permissions/grants/{id}` behind `Require("permission.manage", Global())`. The usecase makes sure that the title exists, the scope kind matches, the scope ref is present, the title is not superadmin, no active duplicate exists, and a raw action is registered. Then it writes the grant and one `user_audit_logs` row inside one `Atomic` call. If identity has not created a `NewV3Handler` yet, add one and mount the permissions group under `/api/v3`.

**Acceptance criteria:**
- [ ] A `POST` with `title = cool_leader, scope_kind = campus` returns 400 `PERM_SCOPE_MISMATCH`. A duplicate active grant returns 409 `PERM_ALREADY_GRANTED`. `title = superadmin` returns 403 `PERM_SUPERADMIN_LOCKED`.
- [ ] A successful `POST` leaves exactly one `user_audit_logs` row with `permission.grant_added`. A `DELETE` leaves one with `permission.grant_removed`. A raw action grant carries `raw: true` in `changes`.
- [ ] After `DELETE`, the very next `Can` for that actor is false.

**Verification:**
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/permissions/ -run TestGrantRevoke -v`
- [ ] `make generate-docs && go build ./...`

**Dependencies:** M4, `identity: user_audit_logs writer`

**Files likely touched:**
- `internal/usecases/permissions/grants_usecase.go`
- `internal/deliveries/http/v3/permissions/grants_handler.go`
- `internal/deliveries/http/v3/v3_handler.go` (new group, or extend identity's)
- `internal/models/permissions/dto.go`
- `tests/integration/permissions/grant_revoke_test.go`

**Estimated scope:** M

## Task M6: List, me, check and titles endpoints

**Description:** Add the four read endpoints. One returns the seed title table. One returns the caller's own active grants. One is a cursor-paginated grant list filtered by community id, title, scope kind and scope ref, behind `permission.view` or self. `GET /v3/permissions/check` answers `Can` for the caller only. Reuse the existing cursor helper in `internal/pkg/cursor`.

**Acceptance criteria:**
- [ ] `GET /v3/permissions/me` for a person with three grants lists all three with title, scope kind and scope ref.
- [ ] `GET /v3/permissions/grants?community_id=X` returns 403 for a caller who is neither `X` nor a `permission.view` holder.
- [ ] `GET /v3/permissions/check?action=cool.edit&kind=cool&ref=C-12` returns `{ "data": { "allowed": true } }` for a `cool_leader` of C-12 and `false` for C-13.

**Verification:**
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/permissions/ -run TestReadEndpoints -v`
- [ ] `make generate-docs && go build ./...`

**Dependencies:** M5

**Files likely touched:**
- `internal/usecases/permissions/grants_usecase.go`
- `internal/deliveries/http/v3/permissions/grants_handler.go`
- `internal/deliveries/http/v3/permissions/titles_handler.go`
- `internal/repositories/pgsql/permissions/grants_pg_repository.go`
- `tests/integration/permissions/read_test.go`

**Estimated scope:** M

## Checkpoint: API

- [ ] Success criteria 1 to 6 of the spec pass as integration tests.
- [ ] `grep -rn 'UserMiddleware(' internal/deliveries/http/v3` returns nothing.
- [ ] Review the endpoint shapes with the human before the web app builds against them.

## Task M7: Backfill migration from frozen tables

**Description:** Write the one-time migration that inserts grants with `granted_by = 'migration'`. From `users.user_types`, `superadmin` becomes `superadmin` global and `admin` becomes `event_admin` global. From `users.roles`, the four `event-*` roles become `event_admin` global. If the table `user_departments` exists with a `position` column, `leader` and `core_team` become department titles. Old rows are untouched. The migration is idempotent (safe to run twice) through `ON CONFLICT DO NOTHING` against the partial unique index.

**Acceptance criteria:**
- [ ] On a fixture copy of `users`, every `superadmin` user gets an active `superadmin` grant. Every `admin` user gets `event_admin` global. `volunteer` and `event-view-volunteer` produce no rows.
- [ ] Running the migration twice leaves the same row count.
- [ ] `users.user_types` and `users.roles` are byte-identical before and after.

**Verification:**
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/permissions/ -run TestBackfill -v`
- [ ] Manual, before production: `SELECT DISTINCT unnest(user_types) FROM users;` and add any unknown value to the spec table first.

**Dependencies:** M1

**Files likely touched:**
- `tests/integration/db/migrations/0000NN_permission_backfill.up.sql` and `.down.sql`
- `tests/integration/permissions/backfill_test.go`
- `tests/integration/permissions/fixtures/users_legacy.sql`

**Estimated scope:** S

## Task M8: `EnsureSeeds` and boot wiring in `contract.go`

**Description:** Wire the module into the composition root. Register the two `permission.*` actions. Run `EnsureSeeds`, which makes sure that the title table and the registry agree. If no active `superadmin` row exists, `EnsureSeeds` logs a loud warning. Construct the grants usecase with the repository aggregate and mount the v3 permissions handler. Mark `middleware.UserMiddleware` as v1/v2 only with a one-line comment.

**Acceptance criteria:**
- [ ] The app starts and `GET /api/v3/permissions/titles` returns the seed table. When no `superadmin` grant exists, startup logs one warning.
- [ ] `contract.go` builds permissions after identity and before any module that calls `permission.Require`.
- [ ] `go vet ./...` and `golangci-lint run ./internal/pkg/permission/...` are clean.

**Verification:**
- [ ] `go build ./... && ENV=DEV go run ./cmd/api/main.go` then `curl localhost:8080/api/v3/permissions/titles`
- [ ] `go test ./...`

**Dependencies:** M6, M7

**Files likely touched:**
- `internal/contract/contract.go`
- `internal/pkg/permission/seeds.go`
- `internal/deliveries/http/main_handler.go`
- `internal/deliveries/http/middleware/authorization.go` (comment only)

**Estimated scope:** S

## Checkpoint: Module complete

- [ ] All nine spec success criteria pass.
- [ ] `go test ./...` and `go build ./...` pass with `TEST_DATABASE_URL` set.
- [ ] Human review of the backfill row counts against a production snapshot before the migration runs there.

## Exports

Other modules import these and nothing else from this module:

- `go-community/internal/pkg/permission` functions: `Register(action, module string)`, `Require(action string, r Resolver) echo.MiddlewareFunc`, `FromParam(kind, param string) Resolver`, `FromQuery(kind, param string) Resolver`, `Global() Resolver`, `WithParents(r Resolver, load func(echo.Context) map[string]string) Resolver`, `Actor(c echo.Context) string`.
- `go-community/internal/pkg/permission` evaluation: `Can(ctx, actor, action string, res Resource) (bool, error)` and `Scopes(ctx, actor, action, kind string) (refs []string, parents map[string][]string, all bool, err error)`.
- `go-community/internal/pkg/permission` types: `Resource{Kind, Ref string; Parents map[string]string}` and the `Evaluator` interface that wraps `Can` and `Scopes` for consumers that inject it.
- `go-community/internal/usecases/permissions`: `GrantsUsecase.GrantTitle(ctx, actor, target, title, scopeKind, scopeRef string) error` and `RevokeByTitle(ctx, actor, target, title, scopeKind, scopeRef string) error`. The `cool` module uses them for `POST /v3/cools/{id}/titles` and for member removal.
- `go-community/internal/pkg/permission`: `Titles() []Title` for the titles endpoint and for `cool` to validate `cool_leader`/`cool_core_team`.

## Risks

| Risk | Impact | Mitigation |
|---|---|---|
| Production `user_types` holds values the spec never saw, so the backfill skips people | High | M7 runs `SELECT DISTINCT unnest(user_types)` on a snapshot first. Startup warns when no `superadmin` grant exists |
| `user_departments.position` does not exist in the current schema, because users have one `department_code` column. Department titles have no backfill source | Medium | M7 guards that step with `IF EXISTS`. Department titles are granted by hand after the move |
| Every v3 request runs one grants query. A burst of 3,000 check-ins multiplies it | Medium | One indexed query per request on `(community_id) WHERE revoked_at IS NULL`, cached on the Echo context. A cross-request cache stays "ask first" per the spec |

---

# Module: qr

# Plan: qr

Source: `SPEC-qr.md`. Depends on `platform` and `identity`. Consumers: `events`, `cool`.

## Task Q1: Token codec with kid map and config validation
**Description:** Add `internal/pkg/qr/codec.go` with `Payload{Kind, Ref, IssuedAt, ExpiresAt}`, `Codec` (`Encode`, `Decode`, `URL`), the four kind constants, and HMAC-SHA256 signing keyed by `kid`. Add `QR{ActiveKid, Secret map[string]string}` and `Application.PublicURL` to `internal/config/config.go`. `qr.NewCodec(cfg)` refuses an empty secret map or an `active_kid` that is not in the map. Stdlib only.
**Acceptance criteria:**
- [ ] `Encode` then `Decode` round-trips every kind, and a `personal` token for one `community_id` is byte-identical across calls with no `exp`.
- [ ] Tampered payload, wrong key, unknown kid, missing part and unknown kind all fail with `QR_INVALID_TOKEN`. A good signature with past `exp` fails with `QR_EXPIRED_TOKEN`.
- [ ] Every kind's URL is under 160 characters with a 30 character public URL and a 16 digit ref.
**Verification:**
- [ ] `go test ./internal/pkg/qr/`
- [ ] `go build ./...`
- [ ] Manual: start the app with `qr.active_kid` missing from `qr.secret` and see it refuse to boot.
**Dependencies:** `platform: herrx.Def` (for the two error classes, or stub them until Q3).
**Files likely touched:**
- `internal/pkg/qr/codec.go`
- `internal/pkg/qr/codec_test.go`
- `internal/config/config.go`
- `config/config.local.template.yaml`
**Estimated scope:** M

## Task Q2: Action registry with Allowed and Dispatch
**Description:** Add `internal/pkg/qr/registry.go` with `ActionCtx`, `Action{Name, Allowed, Handle}` and `Registry` (`Register`, `AllowedActions`, `Dispatch`). `Dispatch` runs `Allowed` before `Handle` and returns `QR_ACTION_NOT_ALLOWED` on false, `QR_ACTION_NOT_FOUND` on an unknown name. Duplicate `(kind, Name)` panics.
**Acceptance criteria:**
- [ ] `AllowedActions` returns only actions whose `Allowed` is true for the given context.
- [ ] `Dispatch` on a refused action never calls `Handle`.
- [ ] Registering the same kind and name twice panics.
**Verification:**
- [ ] `go test ./internal/pkg/qr/`
**Dependencies:** Q1
**Files likely touched:**
- `internal/pkg/qr/registry.go`
- `internal/pkg/qr/registry_test.go`
**Estimated scope:** S

## Task Q3: QR_ error catalog and locale
**Description:** Add `internal/models/qr/errors.go` with the five `QR_` classes via `herrx.Def` and `locale.go` registering the Indonesian messages through `locale.Register`. Point Q1 and Q2 at these classes.
**Acceptance criteria:**
- [ ] The five codes from the spec exist with the spec's EN messages and Kinds.
- [ ] The shared completeness test passes with an `id` message for each.
**Verification:**
- [ ] `go test ./internal/pkg/herrx/ ./internal/pkg/qr/`
**Dependencies:** Q1, Q2, `platform: herrx.Def`, `platform: locale.Register`
**Files likely touched:**
- `internal/models/qr/errors.go`
- `internal/models/qr/locale.go`
**Estimated scope:** S

## Checkpoint: pure package
- [ ] `go test ./internal/pkg/qr/` green with round-trip, tamper, rotation, length and registry cases.
- [ ] `go build ./...` clean with the new config fields.
- [ ] No import of `events`, `cool`, `identity` or `permissions` inside `internal/pkg/qr`.

## Task Q4: Scan log table, repository and purge hook
**Description:** Migration `000024_qr_scan_logs` creating `qr_scan_logs` with the spec's columns and the two indexes. Add `internal/repositories/pgsql/qr/scan_log.go` with `Insert(ctx, row)`. If a tx is present, `Insert` uses it through `Atomic`. Add `qr_scan_logs` to the platform retention purge list keyed by `platform.retention_days.scan_logs`.
**Acceptance criteria:**
- [ ] Migration applies and rolls back cleanly.
- [ ] `Insert` writes one row with `kind`, `ref`, `scanner`, `action`, `result`, `client_ip`.
- [ ] Purge deletes rows older than the configured days and nothing newer.
**Verification:**
- [ ] `make migration name=qr_scan_logs` then edit the pair.
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/qr/ -v`
**Dependencies:** `platform: retention.Purge` list, `platform: configdb.GetInt`
**Files likely touched:**
- `tests/integration/db/migrations/000024_qr_scan_logs.up.sql`
- `tests/integration/db/migrations/000024_qr_scan_logs.down.sql`
- `internal/repositories/pgsql/qr/scan_log.go`
- `internal/platform/retention/purge.go`
- `tests/integration/qr/scan_log_test.go`
**Estimated scope:** M

## Task Q5: Resolve endpoint
**Description:** Add `internal/usecases/qr/resolve.go` and `internal/deliveries/http/v3/qr/handler.go` with `POST /v3/qr/resolve`. Verify the token, write one scan log row, return `{kind, ref, issued_at, expires_at, actions}` through `response.OK`. Failures go through `response.Fail` and still log with `result` set to the error code.
**Acceptance criteria:**
- [ ] A valid token returns the payload fields and the filtered action list.
- [ ] A member scanning a stranger's personal QR gets `actions: []` and 200.
- [ ] An invalid or expired token returns the `QR_` error and one scan log row with that code.
**Verification:**
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/qr/ -v`
- [ ] `make generate-docs`
**Dependencies:** Q2, Q3, Q4, `identity: v3 JWT middleware`, `identity: Actor(c)`, `platform: response.OK/Fail`
**Files likely touched:**
- `internal/usecases/qr/resolve.go`
- `internal/deliveries/http/v3/qr/handler.go`
- `internal/models/qr/dto.go`
- `tests/integration/qr/resolve_test.go`
**Estimated scope:** M

## Task Q6: Act endpoint
**Description:** Add `POST /v3/qr/act` with `{token, action, params}`. Verify the token again and run `Dispatch`. If the handler is transactional, write the scan log row in the same `Atomic` transaction. Return the handler output as `data`. Consumer errors pass through unchanged.
**Acceptance criteria:**
- [ ] A refused action returns 403 `QR_ACTION_NOT_ALLOWED` and `Handle` is never called.
- [ ] A handler error surfaces with the consumer's own code, not a `QR_` code.
- [ ] Exactly one scan log row per call, with `action` set to the action name.
**Verification:**
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/qr/ -v` using a fake action registered in the test.
- [ ] `make generate-docs`
**Dependencies:** Q5
**Files likely touched:**
- `internal/usecases/qr/act.go`
- `internal/deliveries/http/v3/qr/handler.go`
- `tests/integration/qr/act_test.go`
**Estimated scope:** S

## Checkpoint: endpoints
- [ ] Resolve and act each insert exactly one scan log row in the integration tests.
- [ ] Swagger regenerated and both routes appear under `/api/v3/qr`.
- [ ] Both routes reject an unauthenticated call.

## Task Q7: Personal QR PNG endpoint (needs dependency approval)
**Description:** Ask first: add `github.com/skip2/go-qrcode`, the only new dependency in this module. Then add `internal/pkg/qr/png.go` with `RenderPNG(url, size)` and `GET /v3/me/qr.png?size=` that renders the caller's own personal token, clamps size to 128 to 1024 and sets `Cache-Control: private, max-age=86400`.
**Acceptance criteria:**
- [ ] Two calls a day apart return identical bytes for the same member.
- [ ] The PNG decodes back to a URL that verifies with `Codec.Decode` (decoder used in the test only).
- [ ] `size=9999` is clamped, not rejected.
**Verification:**
- [ ] `go get github.com/skip2/go-qrcode@latest && go mod tidy` after approval.
- [ ] `go test ./internal/pkg/qr/`
- [ ] Manual: open `/api/v3/me/qr.png` in a browser and scan it with a phone camera.
**Dependencies:** Q1, Q5 (handler file and JWT middleware)
**Files likely touched:**
- `go.mod`, `go.sum`
- `internal/pkg/qr/png.go`
- `internal/pkg/qr/png_test.go`
- `internal/deliveries/http/v3/qr/handler.go`
**Estimated scope:** S

## Task Q8: Wire into the composition root
**Description:** In `internal/contract/contract.go`, build `qr.NewCodec(config)`, one `qr.Registry`, the scan log repository and the usecases. Mount the v3 qr handler under the `/api/v3` group in `main_handler.go`. Keep the registry reachable so `events` and `cool` call `RegisterQRActions(registry, ...)` from the same place.
**Acceptance criteria:**
- [ ] App boots with the qr routes mounted and no registered actions.
- [ ] `RegisterQRActions` from a consumer can be called from `contract.go` without changing `internal/pkg/qr`.
**Verification:**
- [ ] `go build ./... && ENV=local go run ./cmd/api/main.go` then `curl -X POST localhost:<port>/api/v3/qr/resolve` returns 401.
- [ ] `make generate-docs`
**Dependencies:** Q5, Q6, Q7
**Files likely touched:**
- `internal/contract/contract.go`
- `internal/deliveries/http/main_handler.go`
**Estimated scope:** S

## Checkpoint: module complete
- [ ] All `SPEC-qr.md` success criteria 1 to 9 pass or are covered by a test.
- [ ] `grep -rn 'events\|cool' internal/pkg/qr` returns nothing.
- [ ] Human review before `events` and `cool` register their actions.

## Exports
- `go-community/internal/pkg/qr`: `Codec`, `NewCodec`, `Payload`, `KindPersonal`, `KindSession`, `KindMeeting`, `KindAttendee`, `Registry`, `NewRegistry`, `Action`, `ActionCtx`, `RenderPNG`
- `go-community/internal/models/qr`: `ErrInvalidToken`, `ErrExpiredToken`, `ErrActionNotFound`, `ErrActionNotAllowed`, `ErrRenderFailed`
- `go-community/internal/config`: `QR{ActiveKid, Secret}`, `Application.PublicURL`

## Risks
| Risk | Impact | Mitigation |
|---|---|---|
| `skip2/go-qrcode` is unmaintained | Low | Pure Go, tiny, pinned. If it breaks, the swap is one file. |
| Public URL or payload field rename after posters are printed | High | Boundaries forbid it. Length and round-trip tests pin the format. |
| Scan log row written outside the handler's transaction on `act` | Medium | Q6 passes the `Atomic` tx into the repository. Integration test asserts one row on handler failure too. |

---

# Module: events

# Plan: events

Base: `docs/plans/2026-07-07-event-management-v3-implementation.md` (Tasks 1-17) re-shaped by `SPEC-events.md` deltas 1-14. Package layout: `internal/models/events`, `internal/usecases/events`, `internal/deliveries/http/v3/events`, `internal/repositories/pgsql/events`, `tests/integration/events`. Original Task 8 (QR package) moved to `qr`. Original Task 12 outbox and notifier moved to `platform`. Only the PDF ticket stays here. Original Task 2 keeps only the `EVENT_` classes.

Commands used below:

```bash
go build ./...
go test ./internal/pkg/eligibility/ ./internal/pkg/geo/ ./internal/pkg/form/ ./internal/pkg/recurrence/ ./internal/models/events/ ./internal/usecases/events/...
TEST_DATABASE_URL="postgres://postgres:$DB_PASSWORD@localhost:5888/community_db?sslmode=disable" go test ./tests/integration/events/ -v
make generate-docs
```

## Task E1: Migration 000025 and EVENT_ catalog
**Description:** Create the v3 event tables (delta 14 names), with `category`, `department_code`, `self_scan_protection` inside `geo_config`, `v3_attendance_logs.idempotency_key` unique nullable, and `v3_event_audit_logs`. No `v3_event_organizers`, no `v3_notification_outbox`. Define every `EVENT_` class from delta 1 with `herrx.Def` plus `id` messages. From plan Task 1 and Task 2, delta 1, 9, 12, 14. The migration also creates `v3_attendance_logs.label`, `offline`, `device_scanned_at` and `synced_at` (deltas 15 and 16), and the catalog includes `EVENT_OFFLINE_SYNC_EXPIRED` and `EVENT_BATCH_TOO_LARGE`.
**Acceptance criteria:**
- [ ] `migrate up` then `down` on a clean DB leaves no v3 table and no error.
- [ ] `go test ./internal/models/events/` passes the shared completeness test for all 21 `EVENT_` codes.
- [ ] `grep -c organizers tests/integration/db/migrations/000025_*` returns 0.
**Verification:**
- [ ] `go build ./...`
- [ ] `go test ./internal/models/events/`
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/events/ -run TestMigration -v`
**Dependencies:** `platform: herrx.Def`, `platform: locale.Register`, `identity: migration 000023`, `qr: migration 000024`.
**Files likely touched:** `tests/integration/db/migrations/000025_events_v3.up.sql`, `.down.sql`, `internal/models/events/errors.go`, `internal/models/events/locale.go`, `tests/integration/events/harness_test.go`.
**Estimated scope:** M

## Task E2: JSONB configs with category, conditions and self_scan_protection
**Description:** Port plan Task 3 config structs and strict validation into `internal/models/events/config.go`. Add the `category` constant set, which rejects `sunday_service` with `EVENT_CATEGORY_NOT_ALLOWED`. Add the AND-conditions eligibility shape of delta 4. Limit `attendance_modes` to `registration_qr` and `manual`. Add `geo_config.self_scan_protection` with `none|window|window_geo`, and `registration_config.channels`. From plan Task 3, deltas 3, 4, 6, 7, 13 (11a).
**Acceptance criteria:**
- [ ] `ParseEligibility` rejects `audience: members`, `community_ids`, `roles`, `user_types` and accepts the five condition kinds.
- [ ] `ValidateStrict` rejects `attendance_modes` containing `personal_qr` or `session_qr`, an empty `channels`, and an unknown `self_scan_protection`.
- [ ] Table-driven test covers every reject above and one accept.
**Verification:**
- [ ] `go test ./internal/models/events/`
**Dependencies:** E1.
**Files likely touched:** `internal/models/events/config.go`, `internal/models/events/config_test.go`, `internal/models/events/entities.go`.
**Estimated scope:** M

## Task E3: Eligibility package with AND conditions and reader interfaces
**Description:** Replace plan Task 4's pure `Check(User, Eligibility)` with `eligibility.Check(ctx, actor, e) (bool, string, error)`. Three reader interfaces live in `internal/pkg/eligibility` and are supplied from `contract.go`. `ProfileReader` gives campus and departments, `CoolReader` gives cool membership, `TitleReader` gives active titles. The string names the first failing condition. From plan Task 4, delta 4, spec Open Question 1.
**Acceptance criteria:**
- [ ] Unit tests with fakes: AND across conditions, OR inside one, each of the five kinds, `everyone` rejects a deactivated account, failing condition named.
- [ ] `go list -deps ./internal/pkg/eligibility` shows no `internal/usecases` or `internal/repositories` package.
**Verification:**
- [ ] `go test ./internal/pkg/eligibility/`
- [ ] `go list -deps ./internal/pkg/eligibility | grep -c 'go-community/internal/\(usecases\|repositories\)'` returns 0.
**Dependencies:** E2.
**Files likely touched:** `internal/pkg/eligibility/eligibility.go`, `internal/pkg/eligibility/eligibility_test.go`.
**Estimated scope:** S

## Checkpoint: after E1-E3
- [ ] `go build ./... && go test ./internal/models/events/ ./internal/pkg/eligibility/` green.
- [ ] Migration up and down clean on docker Postgres.
- [ ] Review the eligibility JSON shape with the frontend before E9 depends on it.

## Task E4: Geo, form and recurrence pure packages
**Description:** Port plan Tasks 5, 6, 7. One change: when `mode == session_qr`, `geo.Validate` reads `self_scan_protection`. Staff modes keep the per-mode map (delta 7). Recurrence keeps one session per occurrence (delta 5) and the per-session defaults carry the attendance fields (delta 13, 11j).
**Acceptance criteria:**
- [ ] `geo.Validate` table test: `session_qr` with `none`, `window`, `window_geo` inside and outside radius, missing coords. `personal_qr` with `off|warn|require` and override.
- [ ] `recurrence.Expand` weekly and monthly tests pass, and the expanded session input carries `attendance_modes`, `enable_checkout`, `enable_late_marking`, `self_scan_protection`.
- [ ] `form.Visible` and `form.Validate` tests from plan Task 6 pass.
**Verification:**
- [ ] `go test ./internal/pkg/geo/ ./internal/pkg/form/ ./internal/pkg/recurrence/`
**Dependencies:** E2.
**Files likely touched:** `internal/pkg/geo/geo.go`, `internal/pkg/form/form.go`, `internal/pkg/recurrence/recurrence.go`, and one `_test.go` each.
**Estimated scope:** M

## Task E5: Entities, repositories and audit writer
**Description:** Port plan Task 9 into `internal/repositories/pgsql/events`, without the organizer repository. Add `event_audit_logs` insert with a field diff helper, `department_code` and `campus_codes` loaders for `Resource.Parents`, and the idempotency key lookup on attendance logs. From plan Task 9, deltas 2, 9, 11.
**Acceptance criteria:**
- [ ] Every repository method used by later tasks exists and one integration test round-trips each table.
- [ ] `AuditWrite(ctx, tx, actor, targetType, targetRef, action, before, after)` stores only changed fields with `from` and `to`.
- [ ] `grep -rn organizer internal/repositories/pgsql/events` returns nothing.
**Verification:**
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/events/ -run TestRepositories -v`
**Dependencies:** E1, E2.
**Files likely touched:** `internal/repositories/pgsql/events/events.go`, `sessions.go`, `registrations.go`, `attendance.go`, `audit.go`, `tests/integration/events/repositories_test.go`.
**Estimated scope:** M

## Task E6: Event and session CRUD with permission guards and audit
**Description:** Plan Task 13 CRUD half: `CreateDraft`, `Patch`, `CreateSessions`, `PatchSession`, `DeleteSession`, categories, `Cancel`, `Archive`, `HardDelete`. Every admin write runs inside `Atomic` and writes one audit row. Handlers use `permission.Require("event.manage", resolver)` where the resolver loads the event and sets `Parents`. From plan Task 13, deltas 2, 9.
**Acceptance criteria:**
- [ ] Integration test: patch `total_seats` writes an audit row with `changes.total_seats.from` and `.to`.
- [ ] A `campus_admin` of BSD patches an event with `campus_codes = [BSD]` and gets 403 `PERM_FORBIDDEN` on `[PIK]`.
- [ ] `POST /v3/events` with `category = sunday_service` returns 400 `EVENT_CATEGORY_NOT_ALLOWED`.
**Verification:**
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/events/ -run 'TestEventCrud|TestAudit' -v`
- [ ] `make generate-docs`
**Dependencies:** E5, `permissions: permission.Require`, `permissions: permission.FromParam`, `identity: JWT middleware`.
**Files likely touched:** `internal/usecases/events/event_usecase.go`, `internal/deliveries/http/v3/events/event_handler.go`, `internal/deliveries/http/v3/events/resolver.go`, `tests/integration/events/event_lifecycle_test.go`.
**Estimated scope:** M

## Checkpoint: after E4-E6
- [ ] All unit and integration tests green.
- [ ] Swagger shows `/api/v3/events` routes.
- [ ] Human review of the audit diff shape before it is used by E7 onward.

## Task E7: Publish gate and templates without Sunday Service
**Description:** Plan Task 13 second half: `Publish` strict gate with `EVENT_PUBLISH_VALIDATION` field errors, `SaveAsTemplate`, `CreateFromTemplate`, `Duplicate`, `SeedSystemTemplates` seeding Conference, Weekly Class, Announcement and Monthly Volunteer Meeting. From plan Task 13, delta 6.
**Acceptance criteria:**
- [ ] `SeedSystemTemplates` twice yields exactly four rows and none named Sunday Service.
- [ ] Publish without sessions fails naming `sessions`, with incoherent times names the session code.
- [ ] Template routes are behind `event.templates.manage`.
**Verification:**
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/events/ -run 'TestPublish|TestTemplates' -v`
**Dependencies:** E6.
**Files likely touched:** `internal/usecases/events/event_usecase.go`, `internal/usecases/events/templates.go`, `internal/deliveries/http/v3/events/template_handler.go`, `tests/integration/events/templates_test.go`.
**Estimated scope:** M

## Task E8: Recurrence generation per occurrence
**Description:** `RecurrenceUsecase.GenerateSessions` with `ON CONFLICT DO NOTHING` on `(event_id, occurrence_key)`, publish-window flips, internal endpoint `POST /v3/internal/events/generate-sessions` behind `X-API-Key`. From plan Task 13, deltas 5, 13 (11j).
**Acceptance criteria:**
- [ ] Weekly event: first run creates N sessions, second run creates 0, an edited session survives a re-run.
- [ ] Each generated session carries the per-session attendance defaults.
- [ ] No endpoint or usecase accepts a registration without a session code.
**Verification:**
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/events/ -run TestRecurrence -v`
**Dependencies:** E4, E6.
**Files likely touched:** `internal/usecases/events/recurrence_usecase.go`, `internal/deliveries/http/v3/events/internal_handler.go`, `tests/integration/events/recurrence_test.go`.
**Estimated scope:** S

## Task E9: Registration with seat lock, eligibility and idempotent re-registration
**Description:** Plan Task 10 core: `Register`, `Cancel`, `EditAnswers`, `bookSeats` helper with `SELECT ... FOR UPDATE` inside `Atomic`, eligibility and form validation before `BEGIN`, `max_sessions_per_person`, `EVENT_ALREADY_REGISTERED`. From plan Task 10, ADR 0004, ADR 0007, deltas 4, 5, 11, 13 (11h).
**Acceptance criteria:**
- [ ] 200 concurrent registrations for one seat leave `booked_seats = 1` and one `confirmed` row.
- [ ] A MUSIC member from PIK is rejected by `campus in [BSD] AND department in [MUSIC]` with a message naming campus.
- [ ] Registering for April after March on a monthly event succeeds, and `max_sessions_per_person = 1` blocks it with `EVENT_SESSION_LIMIT_REACHED`.
**Verification:**
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/events/ -run 'TestRegister|TestSeatRace' -v`
**Dependencies:** E3, E5, E6.
**Files likely touched:** `internal/usecases/events/registration_usecase.go`, `internal/usecases/events/seats.go`, `internal/deliveries/http/v3/events/registration_handler.go`, `tests/integration/events/registration_test.go`.
**Estimated scope:** M

## Checkpoint: after E7-E9
- [ ] Seat race test green three runs in a row.
- [ ] Eligibility error messages read well in `id` and `en`.
- [ ] Human review before check-in work starts.

## Task E10: Registration channels and desk registration
**Description:** `registration_config.channels` enforcement with `EVENT_CHANNEL_CLOSED`, and `POST /v3/sessions/{code}/registrations/desk` with `source = admin`, `event.checkin` during the window and `event.manage` after, optional email, `EVENT_ANONYMOUS_NOT_ALLOWED` on restricted events. From delta 13 (11a, 11b), ADR 0004, ADR 0006.
**Acceptance criteria:**
- [ ] Web registration on a `channels = [desk]` session returns 403 `EVENT_CHANNEL_CLOSED`.
- [ ] Desk registration after `register_end_at` needs `event.manage`, and shares the seat lock (race test with web registrations).
- [ ] Reports later show `source = admin` for these rows.
**Verification:**
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/events/ -run TestDeskRegistration -v`
**Dependencies:** E9.
**Files likely touched:** `internal/usecases/events/registration_usecase.go`, `internal/deliveries/http/v3/events/registration_handler.go`, `tests/integration/events/desk_registration_test.go`.
**Estimated scope:** S

## Task E11: Attendee tokens and PDF ticket for windowed events
**Description:** Mint `qr.KindAttendee` tokens only when `registration_config.mode != none`. Generate the PDF ticket at send time and on `GET /v3/registrations/{id}/ticket.pdf`, one code per attendee row. Enqueue the registration email through the platform outbox. From plan Task 12 (PDF only), delta 8.
**Acceptance criteria:**
- [ ] A walk-in-only session produces no attendee token and the ticket endpoint returns 404 `EVENT_NOT_FOUND`.
- [ ] A party of four yields a PDF with four distinct decodable QR URLs.
- [ ] One outbox row per confirmed registration, inside the registration transaction.
**Verification:**
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/events/ -run 'TestTicket|TestOutboxEnqueue' -v`
- [ ] Manual: open the PDF from the endpoint and scan one code with a phone.
**Dependencies:** E9, `qr: qr.Codec`, `qr: qr.RenderPNG`, `platform: outbox producer`. Ask first: `gofpdf` or `maroto` is a new dependency.
**Files likely touched:** `internal/pkg/pdfticket/ticket.go`, `internal/usecases/events/ticket.go`, `internal/deliveries/http/v3/events/registration_handler.go`, `tests/integration/events/ticket_test.go`.
**Estimated scope:** M

## Task E12: Check-in core with personal_qr and session_qr
**Description:** Plan Task 11 first slice. `CheckinUsecase.Act` for `checkin` with `personal_qr` (staff, walk-in auto-create through `bookSeats`). Also `session_qr` (self), with `self_scan_protection` read from the session row. Enforce the check-in window. Write one `v3_attendance_logs` row per path. Store and replay `Idempotency-Key`. From plan Task 11, deltas 7, 11.
**Acceptance criteria:**
- [ ] `session_qr` from outside the radius with `window_geo` returns 403 `EVENT_GEO_OUT_OF_RANGE` and a `rejected_geo` log row exists. With `none` the same call checks in.
- [ ] 200 concurrent check-ins of one attendee produce one `checked_in` log row.
- [ ] Two calls with the same `Idempotency-Key` return the same result and one log row.
**Verification:**
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/events/ -run 'TestCheckinCore|TestCheckinRace' -v`
**Dependencies:** E4, E9.
**Files likely touched:** `internal/usecases/events/checkin_usecase.go`, `internal/deliveries/http/v3/events/checkin_handler.go`, `tests/integration/events/checkin_test.go`.
**Estimated scope:** M

## Checkpoint: after E10-E12
- [ ] Both race tests green three runs in a row.
- [ ] A staff member can check in a walk-in and a self-scan works inside the window on docker Postgres.
- [ ] Human review of the check-in log outcomes list.

## Task E13: registration_qr, manual lookup and eligibility recheck
**Description:** Plan Task 11 second slice. `registration_qr` by attendee id (staff). `manual` with an existing `attendee_id` for lookup check-in. `manual` with a form for walk-in. Category selection for categorised sessions. Eligibility recheck `off|note|block` stored on the log. From plan Task 11, deltas 13 (11c, 11d, 11g), ADR 0008, ADR 0009.
**Acceptance criteria:**
- [ ] Lookup check-in creates no registration and one log row.
- [ ] Walk-in on a categorised session without a category returns 400 `EVENT_CATEGORY_REQUIRED`.
- [ ] Recheck `block` on a now-ineligible attendee returns 403 `EVENT_ELIGIBILITY_RECHECK_FAILED`, `note` checks in and flags the log.
**Verification:**
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/events/ -run 'TestCheckinManual|TestRecheck' -v`
**Dependencies:** E12.
**Files likely touched:** `internal/usecases/events/checkin_usecase.go`, `tests/integration/events/checkin_manual_test.go`.
**Estimated scope:** M

## Task E14: Checkout, lateness, headcount and undo
**Description:** Plan Task 11 third slice. Checkout inside `[start_at, end_at]` when `enable_checkout` is on. `EVENT_ALREADY_CHECKED_IN` is an error with its own log row. Lateness is derived at read time when `enable_late_marking` is on. Headcount writes log rows, and undo writes a correcting row with outcome `headcount_undone`. From delta 13 (11e, 11f), ADR 0004, ADR 0005.
**Acceptance criteria:**
- [ ] Checkout before check-in returns 409 `EVENT_NOT_CHECKED_IN`. Checkout after `end_at` is rejected.
- [ ] Headcount 3 then undo 1 gives a net count of 2 in the session counts query.
- [ ] Attendee list shows `late = true` only when `enable_late_marking` is on and `attended_at > start_at`.
**Verification:**
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/events/ -run 'TestCheckout|TestHeadcount|TestLate' -v`
**Dependencies:** E12.
**Files likely touched:** `internal/usecases/events/checkin_usecase.go`, `internal/repositories/pgsql/events/attendance.go`, `tests/integration/events/checkout_test.go`.
**Estimated scope:** M

## Task E15: QR action registration for events
**Description:** Plan Task 14 without a separate QR usecase: register `event_checkin` and `event_checkout` on `personal`, `attendee` (staff only, `event.checkin` on the event) and `session` (self) kinds through `qr.Registry`, adapting `qr.ActionCtx` to `CheckinInput`. Session tokens expire at `end_at` when `enable_checkout` is on, else `checkin_close_at`. From plan Task 14, deltas 2, 7, 8, 13 (11i).
**Acceptance criteria:**
- [ ] Staff resolving a personal token during a session they may run sees `event_checkin`. A plain member sees an empty list.
- [ ] A member resolving their own attendee token sees no action.
- [ ] A session token minted for a session with checkout on decodes with `exp = end_at`.
**Verification:**
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/events/ -run TestQRActions -v`
**Dependencies:** E12, E13, E14, `qr: qr.Registry`, `qr: qr.Action`, `qr: qr.ActionCtx`, `permissions: permission.Can`.
**Files likely touched:** `internal/usecases/events/qr_actions.go`, `internal/deliveries/http/v3/events/session_handler.go` (poster PNG), `tests/integration/events/qr_actions_test.go`.
**Estimated scope:** S

## Checkpoint: after E13-E15
- [ ] End-to-end on docker Postgres: register, scan poster, desk scan, checkout, all through `/v3/qr/act`.
- [ ] Every attendance path leaves exactly one log row (assert with a count test).
- [ ] Human review before reports.

## Task E16: Attendee list, live counts and computed session state
**Description:** Plan Task 16 read side: `GET /v3/sessions/{code}/attendees` with search and status filter behind `event.attendees.view`, live counts in one `count(*) filter` query, `BuildSessionState` pure builder with reason constants, and public `GET /v3/events`, `GET /v3/events/{code}`, `GET /v3/sessions/{code}` with the 5 second in-process cache on session detail. From plan Task 16, delta 11.
**Acceptance criteria:**
- [ ] Unit tests for every availability case from plan Task 16 Step 1.
- [ ] Attendee list and counts issue one query each (assert with the GORM logger in the test).
- [ ] Two session detail reads inside 5 seconds hit Postgres once. A write invalidates nothing and the next read after 5 seconds is fresh.
**Verification:**
- [ ] `go test ./internal/models/events/ -run TestState`
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/events/ -run 'TestAttendees|TestPublicRead' -v`
**Dependencies:** E14.
**Files likely touched:** `internal/models/events/state.go`, `internal/models/events/state_test.go`, `internal/deliveries/http/v3/events/session_handler.go`, `internal/deliveries/http/v3/events/public_handler.go`, `tests/integration/events/read_test.go`.
**Estimated scope:** M

## Task E17: Reports CSV, XLSX and JSON
**Description:** Plan Task 15 with `format=json`, the `source` split in the Summary sheet, checkout and lateness columns only when enabled, `category` in the summary, all behind `event.reports.view`. From plan Task 15, delta 10.
**Acceptance criteria:**
- [ ] CSV header contains the custom form field, JSON keys equal the CSV header.
- [ ] Summary sheet has `web`, `admin`, `walk_in` counts and no checkout column when checkout is off.
- [ ] `GET .../report?category=volunteer` on an event report filters sessions by event category.
**Verification:**
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/events/ -run TestReports -v`
**Dependencies:** E14.
**Files likely touched:** `internal/usecases/events/report_usecase.go`, `internal/deliveries/http/v3/events/report_handler.go`, `tests/integration/events/report_test.go`.
**Estimated scope:** M

## Task E18: Attendance matrix report
**Description:** `GET /v3/events/{code}/attendance-matrix?format=csv|xlsx|json`: one row per person, one column per occurrence, cells `present|late|absent|checked_out|-`, totals and `attendance_pct`, the JSON shape from delta 10. From delta 10.
**Acceptance criteria:**
- [ ] A person registered for two of three occurrences gets `-` in the third cell and a percentage over two.
- [ ] `late` appears only when `enable_late_marking` is on for that session.
- [ ] JSON field names match delta 10 exactly.
**Verification:**
- [ ] `go test ./internal/usecases/events/ -run TestMatrix`
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/events/ -run TestMatrixReport -v`
**Dependencies:** E17.
**Files likely touched:** `internal/usecases/events/matrix.go`, `internal/usecases/events/matrix_test.go`, `internal/deliveries/http/v3/events/report_handler.go`, `tests/integration/events/matrix_test.go`.
**Estimated scope:** S

## Checkpoint: after E16-E18
- [ ] A three-occurrence volunteer event produces matrix CSV, XLSX and JSON that open and agree.
- [ ] Human review of the report column names with the office.

## Task E19: Staff dashboard, audit endpoint and permission scopes
**Description:** `GET /v3/staff/dashboard` listing events from `permission.Scopes(actor, "event.view_internal", "event")` with `parents` expanded through `department_code` and `campus_codes`, draft and archived visibility behind `event.view_internal`, `GET /v3/events/{code}/audit-logs` behind `event.audit.view`. From plan Task 16, deltas 2, 9.
**Acceptance criteria:**
- [ ] A `department_leader` of MUSIC sees MUSIC events on the dashboard and no others.
- [ ] Audit endpoint pages newest first with a cursor.
- [ ] `grep -rn 'UserMiddleware\|userTypes\|event_organizers' internal/**/events` returns nothing.
**Verification:**
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/events/ -run 'TestDashboard|TestAuditEndpoint' -v`
**Dependencies:** E6, E16, `permissions: permission.Scopes`.
**Files likely touched:** `internal/deliveries/http/v3/events/staff_handler.go`, `internal/usecases/events/event_usecase.go`, `tests/integration/events/staff_test.go`.
**Estimated scope:** S

## Task E20: Pool caps, wiring, seeds and v2 410
**Description:** Plan Tasks 16 mounting and 17: set GORM pool limits from YAML in `contract.go`, mount `events.NewHandler` under `/api/v3`, register `event.*` actions in `init()`, supply the three eligibility readers, call `SeedSystemTemplates` at boot, add the `events:` YAML block to the template, and make sure the platform 410 middleware covers every `/api/v2` event route. From plan Task 16, Task 17, deltas 2, 11, 12.
**Acceptance criteria:**
- [ ] `GET /api/v2/events` returns 410 `COMMON_DEPRECATED_ENDPOINT` in the HTTP smoke test.
- [ ] `make run_api` boots, logs four templates seeded, and `db.Stats().MaxOpenConnections == 10`.
- [ ] HTTP smoke test covers list, register, attendees, report through real handlers with an injected auth middleware.
**Verification:**
- [ ] `go build ./... && go test ./internal/... && TEST_DATABASE_URL=... go test ./tests/integration/events/ -v`
- [ ] `make generate-docs` then `make run_api` boots clean.
**Dependencies:** E7, E15, E18, E19, `platform: deprecate.Middleware`, `platform: configdb.EnsureDefaults`.
**Files likely touched:** `internal/contract/contract.go`, `internal/deliveries/http/main_handler.go`, `internal/deliveries/http/v3/events/handler.go`, `config/config.local.template.yaml`, `tests/integration/events/http_smoke_test.go`.
**Estimated scope:** M

## Checkpoint: complete
- [ ] All `SPEC-events.md` success criteria 1-10 pass by test or manual check.
- [ ] `go build ./... && go test ./...` green, docker Postgres integration green.
- [ ] Human review, then hand table names and `eligibility.Check` to `calendar`.

## Task E21: Scan labels on check-in

**Description:** Add `scan_labels` to the event check-in settings and copy the label for the used method onto each attendance log row as a snapshot. From `SPEC-events.md` delta 16. The `label` column ships in the E1 migration.

**Acceptance criteria:**
- [ ] Saving `scan_labels` with a 41-character value or an unknown method is rejected with `COMMON_INVALID_INPUT` field errors.
- [ ] A check-in by `personal_qr` on an event with a label for that method stores the label on the log row. A method with no label stores null.
- [ ] Changing `scan_labels` after a check-in does not change the stored label. Templates and recurrence carry `scan_labels`.

**Verification:**
- [ ] `go test ./internal/models/events/... ./internal/usecases/events/...`
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/events/ -run TestScanLabels -v`

**Dependencies:** E2, E12
**Files likely touched:** `internal/models/events/config.go`, `internal/usecases/events/checkin_usecase.go`, `tests/integration/events/checkin_test.go`
**Estimated scope:** S

## Task E22: Roster sync endpoint

**Description:** `GET /v3/sessions/{code}/roster?since=<cursor>` behind `event.checkin`. Full roster with no `since`, only changed attendee rows with `since`, and `changed: false` with zero rows when nothing moved. From delta 15.

**Acceptance criteria:**
- [ ] A full call, then one registration and one cancellation, then a `since` call returns exactly those two rows and a newer cursor.
- [ ] A `since` call with no changes returns `changed: false` and runs one cheap query (`max(updated_at)`), not the row query.
- [ ] A roster row has no phone, email or form answers. A caller without `event.checkin` gets 403 `PERM_FORBIDDEN`.

**Verification:**
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/events/ -run TestRosterSync -v`
- [ ] `go build ./...`

**Dependencies:** E9, E12, `permissions: permission.Require`
**Files likely touched:** `internal/repositories/pgsql/events/attendee_repository.go`, `internal/usecases/events/roster_usecase.go`, `internal/deliveries/http/v3/events/checkin_handler.go`, `tests/integration/events/roster_test.go`
**Estimated scope:** M

## Task E23: Offline batch check-in sync

**Description:** `POST /v3/sessions/{code}/checkin/batch` behind `event.checkin`, up to 200 items, one short transaction per item through the existing check-in usecase. Sets `offline`, `device_scanned_at`, `synced_at`. Applies the clock tolerance, the max sync age, `offline_unknown_policy` and the recheck-as-review rule. Writes one `qr_scan_logs` row per token item with `offline = true`. From delta 15. The three columns ship in the E1 migration.

**Acceptance criteria:**
- [ ] The same attendee from two batches gives one `ok`, one `duplicate` and one log row. Headcount 30 plus 25 gives 55.
- [ ] One tampered token returns `QR_INVALID_TOKEN` for that item only. The other items in the batch succeed.
- [ ] An item outside the window plus tolerance is saved as `needs_review` with reason `outside_window`. An unknown person follows `offline_unknown_policy`. `GET .../attendees?review=true` lists them.

**Verification:**
- [ ] `go test ./internal/usecases/events/... -run TestOfflineClassifier`
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/events/ -run TestBatchSync -v`

**Dependencies:** E12, E13, E14, E22, `qr: qr.Codec`, `qr: migration 000024`, `platform: configdb.GetInt`
**Files likely touched:** `internal/usecases/events/batch_checkin_usecase.go`, `internal/usecases/events/offline_classifier.go`, `internal/deliveries/http/v3/events/checkin_handler.go`, `tests/integration/events/batch_sync_test.go`
**Estimated scope:** M

## Task E24: Label counts in reports

**Description:** When an event has at least one scan label, add a `label` column and per-label counts to the session report and the attendance matrix, in CSV, XLSX and JSON (`label` per row, `label_counts` per occurrence). With no labels, output is unchanged. From delta 16.

**Acceptance criteria:**
- [ ] An event with two labels shows both counts in the session report and `label_counts` in JSON.
- [ ] An event with no labels produces byte-identical reports to before this task.
- [ ] Attendance percentage is identical with and without labels.

**Verification:**
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/events/ -run 'TestReport|TestMatrix' -v`

**Dependencies:** E17, E18, E21
**Files likely touched:** `internal/usecases/events/report_usecase.go`, `internal/usecases/events/matrix_usecase.go`, `tests/integration/events/report_test.go`
**Estimated scope:** S

## Checkpoint: offline and labels
- [ ] Roster sync, batch sync and label tests pass on real Postgres.
- [ ] A 200-item batch finishes without holding any transaction longer than one item.
- [ ] Review with the owner: try a full offline run with the web app before event day.

## Exports
- `internal/pkg/eligibility`: `Check(ctx, actor string, e Eligibility) (bool, string, error)`, `ProfileReader`, `CoolReader`, `TitleReader`, `Eligibility` type.
- `internal/repositories/pgsql/events`: `EventRepository` (`GetByCode`, `ListPublishedBetween`), `SessionRepository` (`ListBetween`, `GetByCode`), `RegistrationRepository` (`ListByCommunityID`).
- Table names: `v3_events`, `v3_sessions`, `v3_session_ticket_categories`, `v3_registrations`, `v3_registration_attendees`, `v3_attendance_logs`, `v3_event_templates`, `v3_event_audit_logs`.
- Actions registered: `event.manage`, `event.checkin`, `event.attendees.view`, `event.reports.view`, `event.view_internal`, `event.templates.manage`, `event.audit.view`.

## Risks
| Risk | Impact | Mitigation |
|---|---|---|
| Seat lock held while eligibility or config reads run, deadlocks under a 3,000 burst | High | E9 puts every read before `BEGIN`. Race tests in E9 and E12 run three times in CI. |
| `permission.Scopes` parents expansion not ready when E19 starts | Med | E19 depends on the permissions export. Stub `all = true` in the test until it lands. |
| PDF library choice (`gofpdf` archived, `maroto` heavier) | Med | E11 is Ask first. The ticket endpoint can return 501 until the dependency is approved. |
| In-process session cache wrong under two instances | Low | Free tier runs one instance. The cache is 5 seconds and read-only, so the worst case is a stale count. |

---

# Module: cool

# Plan: cool

Source: `SPEC-cool.md`. Upstream exports named as `module: identifier`. The existing `cools` table is in `internal/models/cool_model.go:11-28`. It has an integer `id`, `campus_code`, `category`, `gender`, `recurrence`, `location_type`, `location_name`, `status` and timestamps. Its three `*_community_ids` arrays are frozen.

## Task C1: Migration and COOL_ catalog

**Description:** Create the v3 COOL tables and the `COOL_` error catalog. The migration adds `area` and `region` to `cools`. It creates `cool_members`, `cool_meetings`, `cool_meeting_attendance` and `cool_meeting_guests` with the unique indexes from the spec. It inserts no membership rows. The catalog defines the 12 `COOL_` classes with EN messages and registers the ID map.

**Acceptance criteria:**
- [ ] `make migration_up` then `migration_down` leaves the schema as before, and up again passes. `cool_members` is empty after up.
- [ ] Inserting two active `cool_members` rows for one `community_id` fails on the partial unique index.
- [ ] Every `COOL_` class has an ID message, checked by the shared platform completeness test.

**Verification:**
- [ ] `go build ./...`
- [ ] `go test ./internal/models/cool/`
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/cool/ -run TestMigration -v`

**Dependencies:** `platform: herrx.Def`, `platform: locale.Register`.

**Files likely touched:**
- `tests/integration/db/migrations/0000NN_cool_v3.up.sql` and `.down.sql` (NN is 000026, the number after the events migration 000025)
- `internal/models/cool/errors.go`
- `internal/models/cool/locale.go`
- `tests/integration/cool/migration_test.go`

**Estimated scope:** S

## Task C2: COOL entities and DTOs

**Description:** Define the GORM entities for the four new tables and the two new `cools` columns. Define the request and response DTOs for every endpoint in the spec, in one models package. No logic. This gives every later task a stable type set.

**Acceptance criteria:**
- [ ] Every table column in the spec's entity table has a struct field with the right GORM tag.
- [ ] Status values for meetings and attendance are typed constants, and a validator rejects any other value with `COOL_INVALID_STATUS`.

**Verification:**
- [ ] `go build ./...`
- [ ] `go test ./internal/models/cool/`

**Dependencies:** C1.

**Files likely touched:**
- `internal/models/cool/cool_model.go`
- `internal/models/cool/meeting_model.go`
- `internal/models/cool/attendance_model.go`
- `internal/models/cool/dto.go`

**Estimated scope:** S

## Checkpoint: Schema and types

- [ ] `go build ./... && go test ./internal/models/cool/` pass.
- [ ] Migration up and down are clean on a real database.
- [ ] Review the entity table against `SPEC-cool.md` before any usecase is written.

## Task C3: COOL create, edit, list and detail

**Description:** Implement `POST /v3/cools`, `PATCH /v3/cools/{id}`, `GET /v3/cools` and `GET /v3/cools/{id}` end to end: repository, usecase, handler, route guards. `POST` needs `cool.manage` and creates in the caller's campus. `GET /v3/cools` uses `permission.Scopes(cool.view, cool)` plus the campus parent. Detail returns the roster and the titles read from grants.

**Acceptance criteria:**
- [ ] A `campus_admin` of BSD can create a COOL in BSD and gets 403 `PERM_FORBIDDEN` for PIK.
- [ ] A `cool_facilitator` of two COOLs sees exactly those two in the list.
- [ ] `PATCH` by a `cool_leader` edits name, description, area, region, location fields and writes one `user_audit_logs` row.

**Verification:**
- [ ] `go build ./... && make generate-docs`
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/cool/ -run TestCoolCRUD -v`
- [ ] Manual: call the four routes with curl as two different title holders.

**Dependencies:** C2, `permissions: permission.Require`, `permissions: permission.Scopes`, `permissions: permission.Actor`, `identity: v3 JWT middleware`, `identity: user_audit_logs writer`, `platform: response.OK/Fail`.

**Files likely touched:**
- `internal/repositories/pgsql/cool/cool_repository.go`
- `internal/usecases/cool/cool_usecase.go`
- `internal/deliveries/http/v3/cool/cool_handler.go`
- `tests/integration/cool/cool_crud_test.go`

**Estimated scope:** M

## Task C4: Members add and remove

**Description:** Implement `POST /v3/cools/{id}/members` and `DELETE /v3/cools/{id}/members/{communityId}`. If the person has an active row anywhere, add fails with `COOL_ALREADY_IN_COOL`. Remove sets `left_at` and revokes every `cool_leader` and `cool_core_team` grant on that COOL in the same `Atomic` transaction. When `cool.max_members` is not zero, add enforces it.

**Acceptance criteria:**
- [ ] Add of a person active in another COOL returns 409 `COOL_ALREADY_IN_COOL`. Remove there, then add here, succeeds.
- [ ] Two concurrent adds of the same person produce one active row.
- [ ] Removing a leader leaves zero active `cool_leader` grants for them on that COOL, and one audit row.

**Verification:**
- [ ] `go build ./...`
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/cool/ -run TestMembers -v`

**Dependencies:** C3, `permissions: grants usecase Revoke`, `platform: configdb.GetInt`.

**Files likely touched:**
- `internal/repositories/pgsql/cool/member_repository.go`
- `internal/usecases/cool/member_usecase.go`
- `internal/deliveries/http/v3/cool/cool_handler.go`
- `tests/integration/cool/members_test.go`

**Estimated scope:** M

## Task C5: Titles and facilitators endpoints

**Description:** Implement `POST` and `DELETE /v3/cools/{id}/titles` behind `cool.roles.change`, and `PATCH /v3/cools/{id}/facilitators` behind `cool.manage`. Both write `permission_grants` rows through the permissions grants usecase, scoped to the COOL, never a raw action. Titles are limited to `cool_leader` and `cool_core_team` and the target must be an active member.

**Acceptance criteria:**
- [ ] A facilitator grants `cool_leader` to a member and the member passes `permission.Can(cool.confirm_attendance)` on that COOL.
- [ ] `POST .../titles` with `cool_facilitator` returns 400 `COOL_TITLE_NOT_ALLOWED`. With a non-member returns 403 `COOL_NOT_A_MEMBER`.
- [ ] `PATCH .../facilitators` with `{add, remove}` adds and revokes `cool_facilitator` grants and writes one audit row per change.

**Verification:**
- [ ] `go build ./...`
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/cool/ -run TestTitles -v`

**Dependencies:** C4, `permissions: grants usecase Grant/Revoke`.

**Files likely touched:**
- `internal/usecases/cool/title_usecase.go`
- `internal/deliveries/http/v3/cool/cool_handler.go`
- `tests/integration/cool/titles_test.go`

**Estimated scope:** S

## Checkpoint: COOLs and people

- [ ] All `tests/integration/cool/` tests pass on a real database.
- [ ] `grep -rn 'FacilitatorCommunityIds\|LeaderCommunityIds\|CoreCommunityIds' internal/usecases/cool internal/repositories/pgsql/cool` returns nothing.
- [ ] Review with the user: roster and title flow end to end with curl.

## Task C6: Meeting create, edit and cancel

**Description:** Implement `POST /v3/cools/{id}/meetings`, `PATCH /v3/meetings/{code}`, `GET /v3/cools/{id}/meetings` and `GET /v3/meetings/{code}`. Creation inserts the meeting as `pending` and one `pending` attendance row per active member, in one `Atomic` transaction. It also mints the meeting token and stores it on the row, with `exp = ends_at + cool.meeting_qr_grace_minutes`. `PATCH` edits fields or sets `cancelled` while `pending`.

**Acceptance criteria:**
- [ ] Creating a meeting for a COOL of 5 yields 5 `pending` rows and a stored token.
- [ ] `PATCH` with `status = cancelled` on a `closed` meeting returns 409 `COOL_MEETING_CLOSED`.
- [ ] `location_type` outside `offline|online|hybrid` returns 400 with a field error.

**Verification:**
- [ ] `go build ./... && make generate-docs`
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/cool/ -run TestMeetingCreate -v`

**Dependencies:** C4, `qr: qr.Codec`, `qr: qr.KindMeeting`, `platform: configdb.GetInt`.

**Files likely touched:**
- `internal/repositories/pgsql/cool/meeting_repository.go`
- `internal/usecases/cool/meeting_usecase.go`
- `internal/deliveries/http/v3/cool/meeting_handler.go`
- `tests/integration/cool/meeting_create_test.go`

**Estimated scope:** M

## Task C7: Meeting QR image and scan handler

**Description:** Serve `GET /v3/meetings/{code}/qr.png` behind `cool.attendance.record` using `qr.RenderPNG` on the stored token, and register the `cool_mark_scanned` action for `qr.KindMeeting`. `Allowed` is "active member of that COOL or holder of `cool.attendance.record` on it". `Handle` sets `scanned_at` on the caller's row. If the row is missing, it creates the row as `pending`. It never changes `status`.

**Acceptance criteria:**
- [ ] A member's `POST /v3/qr/act` on the meeting token sets `scanned_at` and leaves `status = pending`. A second scan returns the same row.
- [ ] A member of another COOL gets 403 `QR_ACTION_NOT_ALLOWED` and a scan log row exists.
- [ ] A scan on a `closed` meeting returns 409 `COOL_MEETING_CLOSED`. A scan after `ends_at + grace` returns 401 `QR_EXPIRED_TOKEN`.

**Verification:**
- [ ] `go build ./...`
- [ ] `go test ./internal/usecases/cool/ -run TestMarkScanned`
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/cool/ -run TestScan -v`
- [ ] Manual: open the PNG in a browser and scan it with a phone against the resolve endpoint.

**Dependencies:** C6, `qr: qr.Registry`, `qr: qr.Action`, `qr: qr.ActionCtx`, `qr: qr.RenderPNG`, `permissions: permission.Can`.

**Files likely touched:**
- `internal/usecases/cool/qr_actions.go`
- `internal/usecases/cool/meeting_usecase.go`
- `internal/deliveries/http/v3/cool/meeting_handler.go`
- `tests/integration/cool/scan_test.go`

**Estimated scope:** M

## Task C8: Mark attendance and close

**Description:** Implement `GET /v3/meetings/{code}/attendance`, `PUT /v3/meetings/{code}/attendance` and `POST /v3/meetings/{code}/close` behind `cool.attendance.view` and `cool.confirm_attendance`. `PUT` accepts a partial list of marks and rejects a non-member with `COOL_NOT_A_MEMBER`. Close fails with `COOL_MEETING_PENDING` while any row is `pending`, and on success sets `closed_by`, `closed_at` and writes one audit row.

**Acceptance criteria:**
- [ ] Close with one `pending` row returns 409 `COOL_MEETING_PENDING`. After marking it, close returns 200.
- [ ] A mark for a person who is not an active member returns 403 `COOL_NOT_A_MEMBER`.
- [ ] A member who joined after meeting creation can be marked and gets a new row.

**Verification:**
- [ ] `go test ./internal/usecases/cool/ -run TestClose`
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/cool/ -run TestMarkAndClose -v`

**Dependencies:** C6.

**Files likely touched:**
- `internal/repositories/pgsql/cool/attendance_repository.go`
- `internal/usecases/cool/attendance_usecase.go`
- `internal/deliveries/http/v3/cool/attendance_handler.go`
- `tests/integration/cool/mark_close_test.go`

**Estimated scope:** M

## Checkpoint: Meeting lifecycle

- [ ] Success criteria 1 to 5 of `SPEC-cool.md` pass as integration tests.
- [ ] `go build ./... && go test ./...` pass.
- [ ] Review with the user: create, scan, mark, close with curl and a phone.

## Task C9: Edit attendance after close with revisions

**Description:** Implement `PATCH /v3/meetings/{code}/attendance` behind `cool.attendance.edit`. It only works on a `closed` meeting. It changes one or more rows to another marked status. It appends one revision entry per row with the old status, the new status, who and when. Nothing is deleted.

**Acceptance criteria:**
- [ ] One edit after close appends exactly one revision with `from`, `to`, `by`, `at`.
- [ ] `PATCH` on a `pending` meeting returns 409 (use `PUT` instead), on `cancelled` returns 409 `COOL_MEETING_CANCELLED`.

**Verification:**
- [ ] `go test ./internal/usecases/cool/ -run TestRevision`
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/cool/ -run TestEditAfterClose -v`

**Dependencies:** C8.

**Files likely touched:**
- `internal/usecases/cool/attendance_usecase.go`
- `internal/deliveries/http/v3/cool/attendance_handler.go`
- `tests/integration/cool/edit_after_close_test.go`

**Estimated scope:** S

## Task C10: Guests add and link

**Description:** Implement `POST /v3/meetings/{code}/guests` and `PATCH /v3/meetings/{code}/guests/{id}` behind `cool.attendance.record`. Add stores name plus optional phone, identifier and notes. Link sets `community_id`, `linked_by`, `linked_at` and fails with `COOL_GUEST_ALREADY_LINKED` on a second link. Guests never appear in rates.

**Acceptance criteria:**
- [ ] A guest without a phone can be added. A guest row appears in `GET .../attendance` under `guests`.
- [ ] Linking a guest to a `pending` account created through `POST /v3/admin/users` succeeds. A second link returns 409 `COOL_GUEST_ALREADY_LINKED`.

**Verification:**
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/cool/ -run TestGuests -v`

**Dependencies:** C8, `identity: pending account creator (POST /v3/admin/users)`.

**Files likely touched:**
- `internal/repositories/pgsql/cool/guest_repository.go`
- `internal/usecases/cool/guest_usecase.go`
- `internal/deliveries/http/v3/cool/attendance_handler.go`
- `tests/integration/cool/guests_test.go`

**Estimated scope:** S

## Task C11: Member "mine" endpoints

**Description:** Implement `GET /v3/cools/mine` and `GET /v3/cools/mine/meetings` for any signed-in member. `mine` returns the member's own COOL page or `COOL_NOT_A_MEMBER`. `mine/meetings` returns upcoming and past meetings of that COOL with the caller's own attendance status per meeting.

**Acceptance criteria:**
- [ ] A member with no active row gets 403 `COOL_NOT_A_MEMBER`.
- [ ] A member sees their own status per meeting and never another member's status.

**Verification:**
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/cool/ -run TestMine -v`

**Dependencies:** C8.

**Files likely touched:**
- `internal/usecases/cool/cool_usecase.go`
- `internal/deliveries/http/v3/cool/cool_handler.go`
- `tests/integration/cool/mine_test.go`

**Estimated scope:** S

## Checkpoint: Corrections, guests and members

- [ ] Success criteria 6 and 7 of `SPEC-cool.md` pass.
- [ ] `go test ./...` passes.

## Task C12: Stats package and reporting endpoints

**Description:** Implement the pure `coolstats` package. Monthly rate is `min(attended, 3) / 3`, quarterly is over 9, only `closed` meetings and `present` rows count, guests are excluded. Then implement `GET /v3/cools/{id}/stats?quarter=` behind `cool.stats.view`. Then implement `GET /v3/cools/{id}/attendance/members?from=&to=` behind `cool.attendance.view`, with the `cool.attendance_window_months` cap for callers without `cool.manage`. The per-member query is the single grouped SQL from the spec. JSON uses the same field names as the CSV.

**Acceptance criteria:**
- [ ] `coolstats` unit tests pass. A 2-meeting month gives 2/3. A 4-meeting month gives 3/3. A quarter of 3, 4 and 1 gives 7/9. A cancelled meeting and a `pending` row are ignored.
- [ ] A core team member asking for 8 months gets 403 `COOL_WINDOW_EXCEEDED`. A campus admin does not.
- [ ] `stats` returns member growth by month, the quarterly table and the COOL average in one round trip per section.

**Verification:**
- [ ] `go test ./internal/pkg/coolstats/`
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/cool/ -run TestStats -v`

**Dependencies:** C9, `platform: configdb.GetInt`, `permissions: permission.Can`.

**Files likely touched:**
- `internal/pkg/coolstats/coolstats.go`
- `internal/pkg/coolstats/coolstats_test.go`
- `internal/repositories/pgsql/cool/stats_repository.go`
- `internal/usecases/cool/stats_usecase.go`
- `internal/deliveries/http/v3/cool/cool_handler.go`

**Estimated scope:** M

## Task C13: Wiring and registration

**Description:** Wire the cool repositories, usecases and handlers into `contract.go` and `main_handler.go` under `/api/v3`, register the `cool.*` actions with `permission.Register` in `init()`, register the meeting QR action through `RegisterQRActions`, and add the `cool.*` configdb defaults to `EnsureDefaults`. Confirm no v3 code reads the frozen arrays or `cool_new_joiners`.

**Acceptance criteria:**
- [ ] The app starts with the `cool.*` defaults present in the `config` table.
- [ ] Every route in the spec's endpoint list responds (not 404) under `/api/v3`.
- [ ] `grep -rn 'cool_new_joiners\|_community_ids\|users.cool_id' internal/usecases/cool internal/repositories/pgsql/cool internal/deliveries/http/v3/cool` returns nothing.

**Verification:**
- [ ] `go build ./... && make generate-docs`
- [ ] `TEST_DATABASE_URL=... go test ./tests/integration/cool/ -v`
- [ ] Manual: `make run` and walk success criteria 1 to 10 from `SPEC-cool.md` with curl.

**Dependencies:** C3 to C12, `permissions: permission.Register`, `qr: qr.Registry`, `platform: configdb.EnsureDefaults`.

**Files likely touched:**
- `internal/contract/contract.go`
- `internal/deliveries/http/main_handler.go`
- `internal/usecases/cool/actions.go`

**Estimated scope:** S

## Checkpoint: Module complete

- [ ] All 10 success criteria of `SPEC-cool.md` verified.
- [ ] `go build ./... && go test ./...` pass, `golangci-lint run ./internal/...` clean on defaults.
- [ ] Review with the user before `calendar` starts.

## Exports

For `calendar`:

- `internal/repositories/pgsql/cool.MeetingRepository.ListInRange(ctx, coolIDs []int, from, to time.Time) ([]cool.Meeting, error)` returns `pending` and `closed` meetings only, with `starts_at`, `ends_at`, `location_type`, `code`, `topic`.
- `internal/repositories/pgsql/cool.MemberRepository.ActiveCoolID(ctx, communityID string) (int, bool, error)` returns the member's one active COOL.
- `internal/models/cool` types `Meeting`, `MeetingStatus`, `LocationType`.

For `qr` consumers: none. `cool` is a consumer, not a provider.

## Risks

| Risk | Impact | Mitigation |
|---|---|---|
| Membership starts empty, so leaders must rebuild rosters by hand before any meeting works | High | C4 first, a CSV-free bulk add is out of scope, so ship C3 to C5 early and let leaders fill rosters while C6 to C12 are built |
| Meeting creation inserts N attendance rows in one transaction; a COOL of 40 is fine, but the pattern must not be copied to events | Low | One insert statement with `unnest`, asserted in C6's test |
| Revoking titles on member removal crosses into the permissions usecase inside the cool transaction | Medium | Use the tx-scoped repository set from `Atomic`, and cover with the C4 concurrency test |

---

# Module: calendar

# Plan: calendar

Base: `docs/plans/2026-07-20-community-calendar-implementation.md` (Tasks 1-5), reshaped by `SPEC-calendar.md` deltas 1-8. Read-only module. Build only after `events` and `cool` tables and exports exist.

## Task L1: Pure calendar package: DTOs, query validation, status mapping, `CAL_` catalog

**Description:** Re-do plan Task 1 under deltas 2, 4 and 7. Create the entry, ref, flags and query DTOs. Add `Query.Validate` for the 92-day cap, the `Sources` set and the `category` CSV against the events category constant set. Add `SortEntries`. Add the meeting status mapper: `pending` before `ends_at` is `scheduled` or `ongoing` by clock, and `pending` after `ends_at` or `closed` is `completed`. Drop `NormalizeLocationType`, `location_type` is passed through as stored. Define `CAL_INVALID_RANGE` and `CAL_INVALID_FILTER` with `herrx.Def` and register the `id` messages.

**Acceptance criteria:**
- [ ] A 93-day range, a zero `From`, or `To` before `From` fails with `CAL_INVALID_RANGE`. An unknown `sources` or `category` value fails with `CAL_INVALID_FILTER`.
- [ ] `MapMeetingStatus` returns `scheduled`, `ongoing` and `completed` for the three delta 4 cases with an injected clock.
- [ ] Both `CAL_` classes have an `id` message and pass the shared platform completeness test.

**Verification:**
- [ ] `go test ./internal/models/calendar/ -v`
- [ ] `go test ./internal/pkg/herrx/ -v`
- [ ] `go build ./...`

**Dependencies:** `platform: herrx.Def`, `platform: locale.Register`, `events: category constant set`.

**Files likely touched:**
- `internal/models/calendar/calendar.go`
- `internal/models/calendar/errors.go`
- `internal/models/calendar/locale.go`
- `internal/models/calendar/calendar_test.go`

**Estimated scope:** M

## Task L2: `.ics` serializer with category line

**Description:** Plan Task 2 with delta 5. Pure `ical.Render([]Entry) []byte` with CRLF lines, `VTIMEZONE` for `Asia/Jakarta`, `DESCRIPTION` starting with `Category: <category>` for sessions and `Category: cool` for meetings, `STATUS:CANCELLED` only for cancelled sessions. Golden file test.

**Acceptance criteria:**
- [ ] Golden file matches byte for byte and contains the `Category:` line on every `VEVENT`.
- [ ] A cancelled session renders `STATUS:CANCELLED`. No cancelled meeting is ever passed in, and the serializer does not filter.

**Verification:**
- [ ] `go test ./internal/pkg/ical/ -v`
- [ ] Manual: open `testdata/golden.ics` in a calendar app and see two events in Jakarta time.

**Dependencies:** L1.

**Files likely touched:**
- `internal/pkg/ical/ical.go`
- `internal/pkg/ical/ical_test.go`
- `internal/pkg/ical/testdata/golden.ics`

**Estimated scope:** S

## Checkpoint: pure layer
- [ ] `go build ./... && go test ./internal/models/calendar/ ./internal/pkg/ical/` pass.
- [ ] No import of `internal/repositories` or `internal/usecases` from either package.
- [ ] `grep -rn 'meeting_date\|onsite' internal/models/calendar internal/pkg/ical` returns nothing.

## Task L3: Sessions query on v3 tables with organizer refs and parents

**Description:** Plan Task 3 sessions half, under delta 1 and resolved open question 1. Raw SQL over `v3_sessions` joined to `v3_events`, filtered by range, campus and category. Each row returns the event code, category, campus codes, department code, eligibility JSON and status. The `registered` flag comes from `v3_registrations` where `registered_by = $actor AND status = 'confirmed'`. The organizer filter takes three parameters: event code refs, department refs and campus refs from `permission.Scopes`, and an `all` flag. The query returns `published` sessions for everyone plus `draft` sessions that match the organizer predicate. `UserFacts` is not created. Wire `Calendar` into `PostgreRepositories`.

**Acceptance criteria:**
- [ ] With `all = false` and empty refs, only `published` sessions in range are returned. With a department ref, a `draft` session of an event with that `department_code` is also returned.
- [ ] If no confirmed `v3_registrations` row exists for the actor and session, `registered` is false.
- [ ] `archived` events never appear. Rows carry `category` and `eligibility` for the usecase.

**Verification:**
- [ ] `TEST_DATABASE_URL="postgres://postgres:$DB_PASSWORD@localhost:5888/community_db?sslmode=disable" go test ./tests/integration/calendar/ -run TestCalendarRepositorySessions -v`
- [ ] `go build ./...`

**Dependencies:** L1, `events: v3_events and v3_sessions migration`, `events: v3_registrations`, `permissions: permission.Scopes`.

**Files likely touched:**
- `internal/repositories/pgsql/calendar_repository.go`
- `internal/repositories/pgsql/calendar_pg_query.go`
- `internal/repositories/pgsql/main_pg_repository.go`
- `tests/integration/calendar/repository_test.go`

**Estimated scope:** M

## Task L4: Meetings query on cool tables

**Description:** Plan Task 3 meetings half, under delta 4. Raw SQL over `cool_meetings` joined to `cools`, filtered by range and a list of COOL ids, with `status IN ('pending', 'closed')`. Each row returns `starts_at`, `ends_at`, `location_type` as stored, `status`, `cool_id` and `meeting_code`. A helper returns the actor's own COOL id from `cool_members` where `left_at IS NULL`. No join to the dropped `cool_facilitators` table.

**Acceptance criteria:**
- [ ] A `cancelled` meeting in range is not returned.
- [ ] Only meetings of the given COOL ids are returned. `location_type` values `offline`, `online`, `hybrid` pass through unchanged.

**Verification:**
- [ ] `TEST_DATABASE_URL="postgres://postgres:$DB_PASSWORD@localhost:5888/community_db?sslmode=disable" go test ./tests/integration/calendar/ -run TestCalendarRepositoryMeetings -v`

**Dependencies:** L3 (shares the repository file), `cool: cool_meetings and cool_members migration`.

**Files likely touched:**
- `internal/repositories/pgsql/calendar_repository.go`
- `internal/repositories/pgsql/calendar_pg_query.go`
- `tests/integration/calendar/repository_test.go`

**Estimated scope:** S

## Task L5: Usecase tiering, eligibility, category filter, merge

**Description:** Plan Task 4 under deltas 1, 3 and 7. `List(ctx, actor, query)`: call `permission.Scopes(actor, "event.view_internal", "event")` once, fetch sessions, then per distinct event code run `eligibility.Check(ctx, actor, e)` from `events` for `published` events that are not in the organizer set. For admin tier, call `permission.Can(actor, "calendar.view_all", Resource{Kind: "global"})` once. If that is false, call it again per event with `Resource{Kind: "event", Ref: code, Parents: {campus...}}`, only for events that eligibility dropped. Meetings come from the union of the actor's COOL id and `permission.Scopes(actor, "cool.attendance.view", "cool")` refs. Apply the category filter to event entries only, map meeting status with L1, set `flags.organizer` and `flags.draft`, merge and sort. Wire into the usecase aggregate.

**Acceptance criteria:**
- [ ] A plain member gets published eligible sessions plus own COOL meetings and nothing else. Eligibility is called once per event code, not once per session.
- [ ] An `event_organizer` grant on a draft event adds its sessions with `flags.draft = true`. A `campus_admin` of BSD sees draft sessions at BSD and none at PIK.
- [ ] `category=volunteer` keeps volunteer sessions and all meetings. Meetings carry `category: null`.

**Verification:**
- [ ] `TEST_DATABASE_URL="postgres://postgres:$DB_PASSWORD@localhost:5888/community_db?sslmode=disable" go test ./tests/integration/calendar/ -run TestCalendarUsecase -v`
- [ ] `go test ./internal/usecases/calendar/ -v` (unit: category filter and merge with fakes)

**Dependencies:** L3, L4, `events: eligibility.Check`, `permissions: permission.Scopes`, `permissions: permission.Can`.

**Files likely touched:**
- `internal/usecases/calendar/calendar_usecase.go`
- `internal/usecases/calendar/calendar_usecase_test.go`
- `internal/usecases/main_usecase.go`
- `tests/integration/calendar/usecase_test.go`

**Estimated scope:** M

## Task L6: JSON and `.ics` endpoints, mount, docs

**Description:** Plan Task 5 under deltas 2 and 8. `GET /api/v3/calendar` and `GET /api/v3/calendar/export.ics` behind the identity JWT middleware only. Parse `from`, `to`, `sources`, `campus`, `category`, call `Query.Validate`, then the usecase. Success uses `response.OK` with `{ "entries": [...] }`. Errors go through `response.Fail`. `.ics` sets `Content-Type: text/calendar` and a filename. Regenerate swagger.

**Acceptance criteria:**
- [ ] When no `Accept-Language` is sent, a 100-day range returns 400 `CAL_INVALID_RANGE` with an Indonesian message.
- [ ] The `.ics` response for a range with one cancelled meeting does not contain it, and the JSON response does not either.
- [ ] `grep -rn 'event_organizers\|cool_facilitators\|meeting_date\|ErrorMapping\|UserMiddleware' internal/deliveries/http/v3/calendar internal/usecases/calendar internal/repositories/pgsql/calendar*' returns nothing.

**Verification:**
- [ ] `TEST_DATABASE_URL="postgres://postgres:$DB_PASSWORD@localhost:5888/community_db?sslmode=disable" go test ./tests/integration/calendar/ -v`
- [ ] `go build ./... && make generate-docs`
- [ ] Manual: `curl -H "Authorization: Bearer <token>" "localhost:<port>/api/v3/calendar?from=2026-10-01&to=2026-10-31"` returns `data.entries`.

**Dependencies:** L2, L5, `identity: v3 JWT middleware`, `platform: response.OK`, `platform: response.Fail`.

**Files likely touched:**
- `internal/deliveries/http/v3/calendar/calendar_handler.go`
- `internal/deliveries/http/main_handler.go`
- `tests/integration/calendar/handler_test.go`
- `docs/` (generated)

**Estimated scope:** M

## Checkpoint: calendar complete
- [ ] All six success criteria in `SPEC-calendar.md` pass against a seeded database.
- [ ] Integration seeds touch only `v3_events`, `v3_sessions`, `v3_registrations`, `cools`, `cool_members`, `cool_meetings`, `permission_grants`.
- [ ] Human review of the JSON shape before the web app builds against it.

## Risks

| Risk | Impact | Mitigation |
|---|---|---|
| `eligibility.Check` per event code does a DB read each, so a 92-day range with many events is slow | Med | Dedupe by event code in L5. If it is still slow, batch the reader interfaces in `events` instead of caching here |
| Organizer predicate on `campus_codes && parents.campus` needs the events array column type to match | Med | L3 integration test seeds an event with two campus codes and asserts the overlap operator works on the real column |
