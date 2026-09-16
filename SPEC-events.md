# Spec: events (reconciliation delta)

Module id `events` in `CAPABILITY-MAP.md`. Intent: `docs/intent/v3-platform.md`. Depends on `permissions` and `qr`.

This file is a delta, not a rewrite. The spec for `events` stays:

- `docs/plans/2026-07-07-event-management-v3-design.md` (the design, cited as "design §N")
- `docs/plans/2026-07-07-event-management-v3-implementation.md` (the plan, cited as "Task N")
- `docs/plans/2026-09-09-checkin-and-registration-paths-spec.md` (the check-in spec)
- `docs/adr/0001` to `docs/adr/0009`

Where this file and those documents disagree, this file wins. Everything not named here is unchanged.

## Objective

Bring the events design in line with the platform, identity, permissions and qr specs, and with the decisions in the intent document. The result is one event engine that serves general events, volunteer events and ticketed events, with no attendance loophole left undefined.

## Deltas

### 1. Errors and envelope

Changes: Task 2, Task 16 `respond.go`, design §10 convention 1.

- Every class moves from `internal/models/v3/errors.go` to `internal/models/events/errors.go` and is created with `herrx.Def`. Codes get the `EVENT_` prefix. The full list:

```
EVENT_NOT_FOUND            EVENT_NOT_ELIGIBLE           EVENT_REGISTRATION_CLOSED
EVENT_QUOTA_FULL           EVENT_MARKED_SOLD_OUT        EVENT_ALREADY_REGISTERED
EVENT_GEO_REQUIRED         EVENT_GEO_OUT_OF_RANGE       EVENT_ALREADY_CHECKED_IN
EVENT_NOT_CHECKED_IN       EVENT_NOT_REGISTERED         EVENT_REGISTRATION_CANCELLED
EVENT_EDIT_WINDOW_CLOSED   EVENT_INVALID_ACCESS_CODE    EVENT_PUBLISH_VALIDATION
EVENT_CATEGORY_NOT_ALLOWED EVENT_CATEGORY_REQUIRED      EVENT_SESSION_LIMIT_REACHED
EVENT_CHANNEL_CLOSED       EVENT_ANONYMOUS_NOT_ALLOWED  EVENT_ELIGIBILITY_RECHECK_FAILED
```
- `INVALID_QR_TOKEN` and `EXPIRED_QR_TOKEN` are removed. The `qr` module owns `QR_INVALID_TOKEN` and `QR_EXPIRED_TOKEN`.
- `INVALID_INPUT` is removed. Validation failures use `herrx.ErrInvalidInput` with `FieldError` entries, so the body has `code`, a top-level `message` and `errors[]`.
- `respond.go` is deleted. Handlers call `response.OK` and `response.Fail` from `platform`. The success body is `{ "data", "meta" }`. `{"code":"SUCCESS"}` never appears in v3.
- `models.ErrorMapping` is never touched.
- The Go toolchain bump and the locale registration move to `platform`. Task 2 keeps only the `EVENT_` classes and their `id` messages.

### 2. Permissions replace organizers and roles

Changes: design §5.2 "Staff authorization", design §7.4 "Access", design §9 auth groups, Task 1 (`v3_event_organizers`), Task 9, Task 13 "organizers", Task 16 auth wiring. ADR 0001 already removed `event_organizers` and stays in force.

- Drop the `v3_event_organizers` table, the organizer repository, `IsOrganizer`, `requireStaff`, `callerFrom`, and `POST /v3/events/{code}/organizers`. Giving someone the `event_organizer` or `event_desk` title on an event is done through `POST /v3/permissions/grants`.
- `middleware.UserMiddleware` and the role names `event-admin`, `roles`, `userTypes` are never used in v3. Every non-public route is guarded by the identity JWT middleware plus `permission.Require(action, resolver)`.
- `Resource.Parents` must carry the event's campus codes and department code, so `campus_admin` and `department_leader` titles match. Add `department_code` (nullable) to `v3_events`. An event with several `campus_codes` passes each one as a parent.
- Actions registered by this module, in `init()`:

| Action | Routes |
|---|---|
| `event.manage` | Every create, patch, publish, cancel and delete on events, sessions and categories. Save as template. Duplicate. Desk registration after `register_end_at`. |
| `event.checkin` | Staff check-in and checkout. Headcount and undo. Desk registration during the registration window. The staff side of the `event_checkin` and `event_checkout` QR actions. |
| `event.attendees.view` | `GET /v3/sessions/{code}/attendees`, attendee search for lookup check-in |
| `event.reports.view` | `GET /v3/events/{code}/summary`, both report endpoints, the matrix report |
| `event.view_internal` | `GET /v3/staff/dashboard`, draft and archived events in list and detail |
| `event.templates.manage` | `GET /v3/templates`, `POST /v3/events/from-template/{id}` |
| `event.audit.view` | `GET /v3/events/{code}/audit-logs` |

- The resolver for session routes loads the session, then builds `Resource{Kind: "event", Ref: event.Code, Parents: {campus..., department}}`. Member self-registration and self check-in need no action, only a signed-in identity.
- The staff dashboard lists events where `permission.Scopes(actor, "event.view_internal", "event")` returns the event, or `all` is true.

### 3. Volunteer is a category

Changes: design §2.1 `events` table, Task 1, Task 3, design §7.4.

- Add `category` (text, not null) to `v3_events`. Values: `general`, `volunteer`, `training`, `class`, `announcement`. The list is a Go constant set. Save rejects any other value. `sunday_service` is not in the list (see delta 6).
- A volunteer event is an event with `category = volunteer`. It has no extra tables, no approval step and no hours. The eligibility rule in delta 4 decides who can join.
- `GET /v3/events` accepts `category=` as a filter. Both report endpoints and the matrix report accept the same filter and print the category in the summary.

### 4. Eligibility is a combined rule

Changes: design §2.2 `eligibility`, Task 3 `Eligibility` struct, Task 4 `eligibility.Check`, ADR 0007 and ADR 0008 (unchanged in intent, changed in shape).

The old shape was `audience` plus lists joined with OR. The new shape is a list of conditions joined with AND. Inside one condition, the listed values are joined with OR.

```json
{
  "audience": "everyone | rules",
  "conditions": [
    { "kind": "campus",     "values": ["BSD", "PIK"] },
    { "kind": "department", "values": ["MUSIC"] },
    { "kind": "cool",       "values": ["C-12"] },
    { "kind": "title",      "values": ["cool_leader", "department_core_team"] },
    { "kind": "account_status", "values": ["active"] }
  ]
}
```

- `audience = everyone` means any signed-in account. `members` is removed as an alias. `community_ids` is removed, since one person is a `cool` or `title` case in practice and a list of ids in JSONB is a maintenance trap.
- `roles` and `user_types` conditions are removed. `title` replaces them and is checked against active `permission_grants` through `permissions`.
- `eligibility.Check(u User, e Eligibility) bool` becomes `eligibility.Check(ctx, actor string, e Eligibility) (bool, string, error)`. The string names the first failing condition for the `EVENT_NOT_ELIGIBLE` message. The checker reads campus and departments from `identity`, COOL membership from `cool`, and titles from `permissions`. It does so through three small reader interfaces, so unit tests use fakes.
- Phases keep `eligibility_override` with the same new shape.
- ADR 0007 (eligibility on every registration path) and ADR 0008 (recheck at check-in, `off | note | block`) are unchanged. `block` returns `EVENT_ELIGIBILITY_RECHECK_FAILED`.

### 5. Register per occurrence

Changes: design §8.3, Task 7, Task 13. No code change, a rule made explicit.

- Recurrence generates one `v3_sessions` row per occurrence. A registration belongs to one session. There is no "register once for all occurrences". No endpoint can create one.
- `max_sessions_per_person` from ADR 0004 stays and defaults to `0` (unlimited), so a monthly volunteer meeting lets the same person register every month.
- The per-session recurrence defaults must carry `attendance_modes`, `enable_checkout`, `enable_late_marking` and `self_scan_protection` (delta 7), per the check-in spec delta to Task 13.

### 6. Sunday services are out

Changes: design §7.1 templates, Task 13 `SeedSystemTemplates`, Task 17.

- Remove the seeded "Sunday Service" template. Seeded templates are Conference, Weekly Class, Announcement and a new "Monthly Volunteer Meeting" (category `volunteer`, recurrence monthly, `personal_qr` and `session_qr`, `enable_late_marking` on).
- `category = sunday_service` is rejected on create and patch with `EVENT_CATEGORY_NOT_ALLOWED`. The value is reserved so nobody re-adds it under another name by accident.

### 7. Check-in modes and self-scan protection

Changes: design §2.1 `attendance_modes`, design §2.2 `geo_config.modes`, design §5.1, Task 5, Task 11, ADR 0006 (desk geo mode), ADR 0003 (poster anti-replay, superseded).

- `personal_qr` and `session_qr` are always available on every session that tracks attendance. They are no longer entries in `attendance_modes`. `attendance_modes` keeps only the optional modes: `registration_qr`, `manual`. An empty list still means "registration only, no check-in" as the check-in spec decided, and in that case the two QR modes are off too.
- A new event-level setting `self_scan_protection` on `geo_config` replaces the per-mode `session_qr` row:

| Value | What `session_qr` self check-in enforces |
|---|---|
| `none` | Nothing. The poster works from anywhere at any time inside the session period. |
| `window` (default) | The check-in window `[checkin_open_at, checkin_close_at]`, and `[start_at, end_at]` for checkout. |
| `window_geo` | The window, plus the venue radius. Missing or out-of-range coordinates are rejected with `EVENT_GEO_OUT_OF_RANGE`. No staff override, since no staff is present. |

- ADR 0003 said the poster has no anti-replay protection and the trust boundary is accepted. This is superseded. The protection is the setting above, chosen per event.
- Staff modes keep their per-mode geo rows from ADR 0006. Those rows are `personal_qr`, `registration_qr`, `manual`, `desk_registration` and `web_registration`, each with `off | warn | require` and `staff_override`. `session_qr` leaves that map. When the mode is `session_qr`, Task 5 `geo.Validate` reads `self_scan_protection` instead.
- The window check for `window` and `window_geo` is enforced on the server from the session row, never from the token.

### 8. Registration QR

Changes: design §2.1 `registration_attendees.id "QR ref"`, design §6.1 `attendee`, design §8.2, Task 12.

- If `registration_config.mode` is `none`, no `attendee` token is minted. Only events with a registration window get attendee tokens. Information-only and walk-in-only sessions get no attendee tokens and no PDF ticket.
- The PDF ticket keeps one code per attendee row. A party of four gets four codes on one ticket.
- The attendee token is desk-scanned only. The `event_checkin` action for kind `attendee` sets `Allowed` to "caller holds `event.checkin` on this event". A member scanning their own ticket sees no action.

### 9. Admin audit log

Changes: Task 1, Task 9, Task 13, design §7.2.

- New table `event_audit_logs`. Never purged. Columns:

```
id           uuid
actor        community_id of the admin
target_type  event | session | category | template
target_ref   event code, session code, category id or template id
action       created | updated | published | cancelled | archived | deleted | settings_changed
changes      jsonb, field-level diff with old and new values
created_at   timestamptz
```

- This is a separate table, not `user_audit_logs` with a `target_type`. Organizers read the events audit per event. The office reads the user audit per person. The two never join.
- Every admin write in Task 13 writes one row in the same `Atomic` transaction. Changes to `checkin_open_at`, `checkin_close_at`, `total_seats`, `attendance_modes`, `self_scan_protection`, `geo_config`, `eligibility` and `registration_config` are always recorded with old and new values.
- `GET /v3/events/{code}/audit-logs` lists rows for one event, newest first, cursor paginated, behind `event.audit.view`. `event_organizer` and `event_admin` hold that action.

### 10. Reports

Changes: design §7.4, Task 15, the check-in spec "Reporting" delta.

- Add the matrix report for recurring events: `GET /v3/events/{code}/attendance-matrix?format=csv|xlsx|json`.
- One row per person who registered for at least one occurrence. One column per occurrence, ordered by `start_at`. Cell values: `present`, `late`, `absent`, `checked_out`, `-` (not registered for that occurrence). If `enable_late_marking` is off for that session, the cell says `present` instead of `late`. Then `total_present`, `total_late`, `total_absent`, `attendance_pct` (present plus late over registered occurrences).
- JSON export shape, with stable snake_case field names:

```json
{ "event_code": "EV-2026-VOL", "category": "volunteer",
  "occurrences": [ { "session_code": "S-01", "start_at": "2026-03-01T09:00:00+07:00" } ],
  "rows": [ { "community_id": "1234567890", "name": "Budi",
              "cells": { "S-01": "present", "S-02": "absent" },
              "total_present": 1, "total_late": 0, "total_absent": 1, "attendance_pct": 50.0 } ] }
```

- The existing event and session reports also accept `format=json` with the same column names as the CSV header.
- The Summary sheet splits registrations by `source` (`web`, `admin`, `walk_in`) per the check-in spec. If lateness or recheck is off, its column is absent.
- Every report is behind `event.reports.view`.

### 11. Load

Changes: design §11, Task 10, Task 11, Task 16.

Already in the design: `SELECT ... FOR UPDATE` on the session or category row inside `Atomic`, idempotent re-scan, outbox claims with `SKIP LOCKED`, pooled Supavisor connection string.

Added:

- Inside the seat lock, the transaction does only the read, the check, the two inserts and the counter update. Configuration reads, eligibility, form validation and geo run before `BEGIN`. No configdb read and no HTTP call inside the lock.
- `POST /v3/sessions/{code}/checkin` and `POST /v3/qr/act` accept an `Idempotency-Key` header. The check-in usecase stores the key in `v3_attendance_logs.idempotency_key` (unique, nullable) and returns the earlier result for a repeated key. A scanner that retries on a timeout never double-logs.
- GORM pool limits are set in `contract.go` from YAML: `MaxOpenConns` 10, `MaxIdleConns` 5, `ConnMaxLifetime` 5 minutes. A burst of 3,000 people then queues on the app side. The Supabase free tier pool is never exhausted.
- The session detail response is cached in memory for 5 seconds per session code, computed state included. A refresh storm at registration open then does not hit Postgres per request. Writes are never cached.
- `GET /v3/sessions/{code}/attendees` and the live counts use one query with `count(*) filter (where ...)`, never N+1.
- One integration test drives 200 concurrent registrations at one seat and asserts one booked seat. Another drives 200 concurrent check-ins of one attendee and asserts one `checked_in` log.

### 12. Legacy data and v2 routes

Changes: design §1 non-goals, design §13 rollout.

- Old events, instances and registration records stay in the v2 tables. No migration, no read from v3.
- Every `/api/v2` event route returns 410 `COMMON_DEPRECATED_ENDPOINT` from the platform middleware on the day v3 events go live. The rollout in design §13 step 2 (run v2 in parallel as fallback) is dropped. There is no parallel run.
- The plan's migration number `000022` collides with the existing `000022_configs_setup`. Renumber the events migration after `identity` (000023) and `qr` (000024), so events is `000025`.

### 13. Check-in spec deltas not yet in the plan

Changes: Task 10, Task 11, Task 13, Task 15, Task 16. Add these tasks to the plan, in this order, after Task 11:

- Task 11a: `registration_config.channels` with values `web` and `desk`. An empty list is rejected. If the used channel is off, the request fails with `EVENT_CHANNEL_CLOSED` (ADR 0006).
- Task 11b: desk registration endpoint `POST /v3/sessions/{code}/registrations/desk` with `source = admin`. It needs `event.checkin` while registration is open and `event.manage` after `register_end_at`. It uses the same seat lock. Email is optional. Restricted events reject it with `EVENT_ANONYMOUS_NOT_ALLOWED` (ADR 0004, ADR 0007).
- Task 11c: `manual` accepts an existing `attendee_id` for lookup check-in with no new registration (ADR 0004).
- Task 11d: eligibility runs on walk-in auto-creation and on desk registration, and the recheck setting `off | note | block` is stored at check-in (ADR 0007, ADR 0008).
- Task 11e: checkout window is `[start_at, end_at]`, `EVENT_ALREADY_CHECKED_IN` is an error with its log row, `enable_late_marking` per session, lateness derived at read time (ADR 0004, ADR 0005).
- Task 11f: headcount undo as a correcting log row with outcome `headcount_undone`, net count in reports (ADR 0005).
- Task 11g: walk-in on a categorised session requires a category, else `EVENT_CATEGORY_REQUIRED` (ADR 0009).
- Task 11h: `max_sessions_per_person`, anonymous exempt, `EVENT_SESSION_LIMIT_REACHED` names the session (ADR 0004).
- Task 11i: if `enable_checkout` is on, the session token expires at `end_at`. Otherwise it expires at `checkin_close_at`. This lives in the events code that mints the token through `qr` (ADR 0009).
- Task 11j: recurrence and template defaults carry the per-session attendance fields (check-in spec, Task 13 delta).

### 14. Table names

Changes: none in events. A note for `calendar`.

The v3 tables are `v3_events`, `v3_sessions`, `v3_session_ticket_categories`, `v3_registrations`, `v3_registration_attendees`, `v3_attendance_logs`, `v3_event_templates`, `v3_event_audit_logs`. The `v3_notification_outbox` table is replaced by the platform `notification_outbox`. `v3_event_organizers` is dropped. The calendar design must read `v3_sessions` and `v3_events`, never `event_sessions`, `registrations` or `event_organizers`.

## Testing Strategy additions

Unit, no DB:

- Eligibility: AND across conditions, OR inside one condition, every condition kind, first failing condition named, `everyone` with a deactivated account rejected.
- `self_scan_protection`: each of the three values against inside window, outside window, inside radius, outside radius, missing coordinates.
- Matrix builder: a person registered for two of three occurrences yields `-` in the third cell and a percentage over two.
- Category constant set rejects `sunday_service`.

Integration, real Postgres:

- Each row of the action table in delta 2: the listed title passes, a member without the title gets 403 `PERM_FORBIDDEN`.
- A `campus_admin` of BSD manages an event with `campus_codes = [BSD]` and gets 403 on one with `[PIK]`.
- An admin patch to `checkin_close_at` writes one `event_audit_logs` row with old and new values.
- Idempotency: two check-ins with the same `Idempotency-Key` produce one log row.
- Two hundred concurrent registrations for one seat produce one registration with status `confirmed`.
- Every `EVENT_` code has an `id` message (shared platform test).

## Boundaries

Always:

- Pass `Resource.Parents` with campus codes and department on every event and session route.
- Run eligibility before `BEGIN` on every path that creates a registration.
- Write the audit row in the same `Atomic` transaction as the admin change.

Ask first:

- Adding a category value.
- Adding a condition kind to eligibility.
- Adding a fourth `self_scan_protection` value, for example a rotating poster.
- Any new column on `v3_attendance_logs`.

Never:

- Read `user_types`, `roles` or `event_organizers`.
- Accept a session token's window from the client. The server reads the session row.
- Create a registration that spans more than one session.
- Purge `v3_attendance_logs` or `event_audit_logs`.

## Success Criteria

1. `POST /v3/events` with `category = sunday_service` returns 400 `EVENT_CATEGORY_NOT_ALLOWED`.
2. An event with conditions `campus in [BSD]` and `department in [MUSIC]` rejects a MUSIC member from PIK with `EVENT_NOT_ELIGIBLE`. The message names campus.
3. A member scans the poster from home for a session with `self_scan_protection = window_geo` and gets 403 `EVENT_GEO_OUT_OF_RANGE`. With `none`, the same scan checks in.
4. A member registered for a monthly volunteer event in March is not registered for April until they register again.
5. The matrix report for a three-occurrence event has one column per occurrence and a percentage per person, in CSV, XLSX and JSON.
6. A person with `event_desk` on `EV-1` can check in at `EV-1` and gets 403 on `EV-2`.
7. Patching `total_seats` writes an `event_audit_logs` row with `changes.total_seats.from` and `.to`.
8. `GET /api/v2/events` returns 410 `COMMON_DEPRECATED_ENDPOINT`.
9. Two hundred concurrent registrations for one seat leave `booked_seats = 1`.
10. `grep -rn 'event_organizers\|UserMiddleware\|userTypes' internal/**/events` returns nothing.

## Open Questions

1. The eligibility checker now reads from `identity`, `cool` and `permissions`. Make sure that the three reader interfaces live in `internal/pkg/eligibility` and each module supplies its implementation in `contract.go`, so `eligibility` imports none of them.
2. The 5 second in-memory cache of session detail (delta 11) is per process. If the app ever runs more than one instance, the cache is per instance, which is still correct but less effective. Make sure that the free tier runs one instance.
3. Delta 7 removes `personal_qr` and `session_qr` from `attendance_modes`. Make sure that the admin UI shows them as always-on and only exposes `registration_qr` and `manual` as toggles.
4. Resolved 2026-09-17: `training` and `class` stay separate categories. The list is `general`, `volunteer`, `training`, `class`, `announcement`.
