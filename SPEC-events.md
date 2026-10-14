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
EVENT_OFFLINE_SYNC_EXPIRED EVENT_BATCH_TOO_LARGE        EVENT_FORM_UNSAFE_CHANGE
EVENT_FILES_IN_USE         EVENT_DESK_ANSWERS_REQUIRED
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
action       created | updated | published | cancelled | archived | deleted | settings_changed | files_deleted
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

### 15. Offline desk check-in

Changes: design §5 check-in, Task 1 (`v3_attendance_logs` columns), Task 11, delta 11 (`Idempotency-Key`). New in this delta, approved 2026-09-18.

Why: with 1,000 to 3,000 people in one place the mobile network fails. The desk must keep working with no signal, and the server stays the final judge once the device is back online.

- Offline scope: staff actions only. Those are `personal_qr` scans, `registration_qr` scans, `manual` lookup, checkout, headcount taps and headcount undo. `session_qr` self-scan by a member needs a connection and is never queued.
- Roster sync: `GET /v3/sessions/{code}/roster?since=<cursor>` behind `event.checkin`. With no `since`, it returns the full roster and a `cursor`. With `since`, it returns only attendee rows that changed after that cursor, cancelled ones included, and a new `cursor`. If nothing changed, it returns `{ "changed": false, "cursor": <same> }` and no rows. The cursor is the highest `updated_at` of the session's attendee rows, so no new column is needed.
- A roster row holds `attendee_id`, `community_id` (null for a companion), `name`, `status`, `category_id`, `checked_in_at` and `checked_out_at`. It never holds phone, email, form answers or any other personal data.
- The personal QR payload is readable without the signing key (see `SPEC-qr.md`). The offline device reads `community_id` from it and matches the roster. The signature is checked later by the server at sync.
- Batch sync: `POST /v3/sessions/{code}/checkin/batch` behind `event.checkin`. Body is up to 200 items. Each item has `idempotency_key` (a UUID made on the device), `action` (`checkin`, `checkout`, `headcount`, `headcount_undo`), `method` (`personal_qr`, `registration_qr`, `manual`), `token` or `attendee_id`, and `device_scanned_at`. The response has one result per item: `ok`, `duplicate`, `needs_review` or `rejected`, each with an `EVENT_` or `QR_` code when not `ok`.
- Every item runs through the same check-in usecase as an online scan, one short transaction per item. A failed item never fails the batch.
- First check-in wins. If two offline desks scanned the same person, the later item returns `duplicate` and changes nothing. Nobody sees an error.
- Headcount taps add up. Each tap is its own item with its own key, so 30 taps from one desk and 25 from another give 55. An undo is its own item and subtracts one.
- Each log row written from the batch has `offline = true`, `device_scanned_at` from the device and `synced_at` from the server. `checked_in_at` is set to `device_scanned_at`, so lateness is judged by when the person arrived, not when the signal came back.
- Clock: `device_scanned_at` is accepted if it is inside the check-in window, widened by `event.offline_clock_tolerance_seconds` (configdb, default `120`). An item outside that range is still saved, with result `needs_review` and reason `outside_window`, because the person is already inside.
- Items older than `event.offline_sync_max_hours` (configdb, default `24`) after the session end are `rejected` with `EVENT_OFFLINE_SYNC_EXPIRED`.
- Person not on the roster: the event setting `offline_unknown_policy` decides. With `admit_review` (default) the device shows amber and lets the person in. At sync the server checks them in if they are registered. If they are not registered, walk-ins are open and a seat remains, it creates the walk-in. Otherwise the item is `needs_review`. With `reject` the device shows red and queues nothing.
- The eligibility recheck from ADR 0008 runs at sync. A `block` result becomes `needs_review` with reason `eligibility`, not a rejection, because the person is already inside.
- `GET /v3/sessions/{code}/attendees` accepts `review=true` and lists every `needs_review` row with its reason, so the desk lead can settle them after the event.
- What offline cannot promise: walk-in seat limits and `max_sessions_per_person` are only enforced at sync. The spec accepts that an offline desk can let in a few people over the limit. Reports show them as `needs_review`.
- The web app owns storing the roster and the queue on the device. That is out of scope for this backend.

### 16. Scan labels

Changes: design §2.1 `attendance_modes`, design §7.4 reports, Task 3, Task 11, Task 15, delta 10. New in this delta, approved 2026-09-18.

- An organizer can give each check-in method a label that says what that scan means for this event. Example for a hybrid volunteer briefing: `personal_qr` means "Hadir di lokasi", `session_qr` means "Hadir online".
- The setting is `scan_labels` in the event's check-in settings, next to `attendance_modes` and `self_scan_protection`. It is a map from method (`personal_qr`, `session_qr`, `registration_qr`, `manual`) to free text, 40 characters at most. Every entry is optional. An empty map is the default.
- At check-in the server copies the label for the used method onto the log row as `label`. It is a snapshot, so a later change of the setting never rewrites history.
- The label never changes the attendance count or percentage. It is data about how the person attended, nothing more.
- If at least one method has a label, the session report and the attendance matrix gain a `label` column and a count per label, for example `Hadir di lokasi: 40, Hadir online: 12`. The JSON export carries `label` per row and `label_counts` per occurrence. If no method has a label, none of that appears.
- Templates and recurring events carry `scan_labels` like every other check-in setting (Task 11j).
- Labels are free text, so totals across different events are only as clean as what organizers type. Inside one event or one recurring series they are consistent.

### 17. Custom registration form

Changes: design §2.2 `registration_config.form` and design §7.4 reports. Also Task 3 (`FormField`, `Condition`), Task 6 (form engine) and Task 10 (registration). Also Task 11b (desk), Task 11d (walk-in), Task 13 (patch and publish), Task 15 (reports) and plan task E4. Decided in the interview of 2026-09-30.

Why: some events need more than a name. The organizer builds the registration form like a Google Form, with question types, validation rules and conditions. The same engine can later serve COOL and standalone forms.

Unchanged from design §2.2: `answered_by`, the `show_if` combinators and leaf operators, `edit_policy`, `answer_revisions` and `companion_detail`. `name` stays the only default field, and `name`, `email`, `phone` and `nik` stay built-in types.

#### 17.1 The engine stands alone

- The engine lives in `internal/pkg/form` and imports only the standard library. It owns the types `Field`, `Option`, `Condition`, `Facts` and `Problem`.
- `v3.FormField` and `v3.Condition` from Task 3 are removed. `RegistrationConfig.Form` becomes `[]form.Field`.
- The caller gives the engine a `Facts` value. Events fill it with category, session and person facts. A later COOL form or standalone form fills it with its own facts, and the engine stays the same.
- The engine returns plain `Problem` values (field key, attendee index, rule). The events layer turns them into herrx `FieldError` entries with localized text.
- Quiz mode (answer key, points, score) is out of scope. It comes with standalone forms. An answer key is a new optional attribute later, so nothing here prepares for it.

#### 17.2 Field types

| Type | Answer shape | Notes |
|---|---|---|
| `name`, `email`, `phone`, `nik` | string | Unchanged. Format rules as in Task 6. |
| `text` | string | Short text. Default limit 500 characters. |
| `textarea` | string | Long text. Default limit 5,000 characters. |
| `number` | number | Unchanged. |
| `radio` | option value, or `{"other": "..."}` | New. Single choice, every option visible. |
| `select` | option value, or `{"other": "..."}` | Single choice as a dropdown. |
| `multiselect` | array of option values, plus at most one `{"other": "..."}` | Several choices ("checkboxes" in Google Forms). |
| `checkbox` | boolean | One yes/no box, for example consent. Unchanged. |
| `date` | `"2026-10-01"` | Unchanged. |
| `time` | `"14:30"` | New. 24-hour wall time. |
| `datetime` | `"2026-10-01T14:30"` | New. Wall time with no zone. |
| `scale` | integer | New. Linear scale. `display: stars` shows it as a rating. |
| `grid` | object from row value to column value. With `multiple` on, each row holds an array of column values. | New. Rows in `rows`, columns in `options`. |
| `file` | array of file URLs | New. See 17.6. |

- Options change from strings to objects: `{ "value": "veg", "label": "Vegetarian" }`. `value` is stable, unique inside the field, and matches `^[a-z0-9_]{1,40}$`.
- Answers, `show_if` and `edit_policy` refer to option values. Reports print labels. So an organizer can rename a label and no answer breaks.
- `allow_other: true` on `radio`, `select` and `multiselect` adds an "Other" choice with free text, 200 characters at most. The answer is `{"other": "Halal only"}`, and reports print `Other: Halal only`.
- `scale` has `min` (0 or 1), `max` (2 to 10), optional `min_label` and `max_label`, and `display` (`number` by default, or `stars`).
- A `grid` with `required: true` needs an answer in every row.
- `date`, `time` and `datetime` answers are wall time as the person typed it. They are form data, not system timestamps, so they carry no zone.
- Every field can have `help`, a text under the label, 500 characters at most.

#### 17.3 Response validation

Each field can carry `validation`:

```json
"validation": { "min_length": 20, "max_length": 300, "min": 13, "max": 25,
                "min_picks": 1, "max_picks": 3, "pattern": "[0-9]{10}",
                "message": "Tulis 10 angka." }
```

| Rule | Applies to |
|---|---|
| `min_length`, `max_length` | `text`, `textarea`, and the Other text |
| `min`, `max` | `number` |
| `min_picks`, `max_picks` | `multiselect`, and each row of a `grid` with `multiple` on |
| `pattern` | `text` |
| `message` | Any rule. At most 120 characters. It replaces the default error text for that field. |

- `pattern` uses Go `regexp` syntax (RE2), so it runs in linear time. The engine wraps the pattern in `^(?:...)$`, so it matches the whole answer. A pattern is at most 200 characters.
- Save rejects a pattern that does not compile, a rule on a type that does not support it, and `min` above `max`.
- An organizer can set `max_length` below the default, but never above 500 for `text` or 5,000 for `textarea`.

#### 17.4 Conditions on category, session and person

A `show_if` leaf gets an optional `source`. A leaf with no `source` reads the same attendee's answers, as today.

```json
{ "source": "category", "op": "in", "value": ["VOL"] }
{ "source": "session",  "op": "eq", "value": "S-02" }
{ "source": "person", "of": "attendee", "field": "age", "op": "lt", "value": 17 }
```

- `category` is the ticket category code of the registration. `session` is the session code. Both accept `eq`, `neq`, `in` and `not_in`.
- `person` facts are `campus`, `department`, `cool`, `title`, `account_status`, `age`, `gender`, `marital_status` and `baptized`. `department` and `title` are lists, so `eq` on them means "contains". `age` is whole years on the day of submission.
- `of` is `attendee` (default) or `registrant`. `attendee` reads the account of the row being answered. `registrant` reads the account of the person who registered the party.
- A companion with no account has no facts. Every `person` leaf is false for them, except `not_answered`, which is true. The organizer combines leaves with `any` and `all` to decide what such companions see.
- Example: `any[ age lt 17, age not_answered ]` shows a consent question to minors and to companions with no account.
- Events build `Facts` through the identity, cool and permissions readers from open question 1. `internal/pkg/eligibility` and the events usecase share those readers. The identity reader adds `date_of_birth`, `gender`, `marital_status` and `baptized`.
- The server is the final judge. It evaluates every condition again at submit and at answer edit, and drops answers to hidden fields, as design §2.2 says.
- The web app evaluates conditions live while the person types. For that, `GET /v3/events/{code}` returns `viewer_facts`. It holds the signed-in person's own facts, and only the keys that a `person` condition in this form uses. It never holds facts of another person.
- If the server result for a companion with an account differs from the web app, the submit returns `COMMON_INVALID_INPUT` with a `FieldError` for that companion. The person fixes it and submits again.
- `internal/pkg/form/testdata/visible_cases.json` holds the evaluator cases as data. The Go test runs them, and the web app can run the same file.

#### 17.5 Desk and walk-in

- Each field gets `required_at_desk`, `false` by default. On desk registration (`source = admin`) and on a manual walk-in, `required_at_desk` replaces `required`. Every other field is optional there, and staff fill it in or leave it blank.
- Walk-in auto-creation from a personal QR scan (Task 11d) has no form screen. If a field is visible with no answers and has `required_at_desk = true`, the scan fails with `EVENT_DESK_ANSWERS_REQUIRED`. The desk then uses desk registration.
- An offline batch walk-in under `admit_review` (delta 15) gets `needs_review` with reason `form` in the same case.

#### 17.6 File questions

- The web app uploads to UploadThing, as it does for profile files. The backend stores only the URL and never stores bytes (`SPEC-identity.md`).
- Every file is public. Anyone with the link can open it. The owner accepted this on 2026-09-30, see 17.9.
- A `file` field has `file: { "accept": ["image", "pdf"], "max_files": 1, "max_mb": 5 }`. `accept` takes `image` and `pdf`. `max_files` is 1 to 5. `max_mb` is 1 to 10.
- The web app enforces `accept` and `max_mb` in its UploadThing route, because the backend never sees the bytes.
- If a URL does not start with `upload.public_url_prefix`, the backend rejects it with a `FieldError`. That YAML value is the file host of this UploadThing app. This blocks links to other sites and to files of other apps.
- Files stay until an organizer deletes them. `POST /v3/events/{code}/files/delete` is behind `event.manage`. If the event is not `completed`, `cancelled` or `archived`, it returns `EVENT_FILES_IN_USE`.
- The delete removes every file URL in the event's current answers and `answer_revisions` through the UploadThing server API. It sets those answers to `[]` and writes one `event_audit_logs` row with action `files_deleted` and the count.
- A second delete call finds no files, calls nothing and returns 200.
- The UploadThing secret is `upload.uploadthing_secret` in YAML. It never goes into configdb or logs.

#### 17.7 Changing a form after answers exist

Save compares the new form with the old one through `form.DiffSafe(old, new, used)`. `used` holds the field keys and option values that at least one stored answer of the event uses.

Allowed at any time:

- Change `label`, `help`, `validation.message`, option labels and row labels.
- Reorder fields, options and rows.
- Add a field with `required = false` and `required_at_desk = false`.
- Add an option or a row.
- Change `validation` and `show_if`. New rules apply to new submissions and to edits, never to stored answers.
- Retire a field with `retired: true`. New registrations and edits no longer show it. Its stored answers stay, and reports keep its column.

Blocked once the event has a registration, with `EVENT_FORM_UNSAFE_CHANGE` and one `FieldError` per field:

- Delete a field that has an answer. Retire it instead.
- Change the `key`, `type` or `answered_by` of a field.
- Remove an option or a row that an answer uses.
- Turn on `required` or `required_at_desk` for an existing field.

Before the first registration, every change is allowed, as in a draft. Every form change writes the `event_audit_logs` row that delta 9 already requires for `registration_config`.

#### 17.8 Pre-fill, reports and limits

- For the registrant's own row, `GET /v3/events/{code}` returns `prefill` with `name`, `email` and `phone` from the profile, for the fields of those types.
- The person can change a pre-filled value. The registration stores what the person submitted. The profile stays the same.
- Reports (delta 10 and design §7.4) print labels, not values. `multiselect` joins labels with `; `. A `grid` gets one column per row, named `<field label> - <row label>`.
- A `file` column joins its URLs with a space. Retired fields keep their column. The JSON export uses the same columns as the CSV.
- The roster never holds answers or file URLs. Delta 15 stays as it is.
- Limits, enforced at save and at submit: 100 fields per form, 50 options per field, 20 rows per grid, 200 characters per `label`, and 32 KB of answers JSON per attendee.

#### 17.9 Known risks

- Public file links. A forwarded report or export opens every file in it, payment proofs included. Private files need a paid UploadThing plan or a second store, and Supabase Storage is the free option.
- Storage. The UploadThing free plan has 2 GB, and nothing deletes files automatically. 3,000 photos at 300 KB each take about 900 MB. An organizer must run the delete after large events.
- Orphan files. A file uploaded for a registration that the person never submits stays on UploadThing. No event refers to it, so the delete endpoint cannot find it.
- Personal data. Answers and files stay until an organizer deletes them. Under UU PDP, personal data with no retention limit is a compliance risk. The church owns that decision.
- Flexibility cost. Each switch in this delta is one more thing an organizer must understand. The admin UI must show `required_at_desk`, `validation`, condition sources and `of` as advanced options with safe defaults.

#### 17.10 Errors

| Code | Kind | EN message |
|---|---|---|
| `EVENT_FORM_UNSAFE_CHANGE` | conflict | This change breaks answers that people already gave. |
| `EVENT_DESK_ANSWERS_REQUIRED` | invalid | This event needs a few answers first. Please use desk registration. |
| `EVENT_FILES_IN_USE` | conflict | You can delete files only after the event ends. |

Answer validation failures stay `COMMON_INVALID_INPUT` with `FieldError` entries. Form schema errors at publish stay `EVENT_PUBLISH_VALIDATION`.

#### 17.11 Where the code goes

```
internal/pkg/form/form.go                      → Field, Option, Condition, Facts, Problem, Validate, Visible
internal/pkg/form/schema.go                    → ValidateSchema (save and publish), DiffSafe
internal/pkg/form/form_test.go                 → table tests per type, rule and diff case
internal/pkg/form/testdata/visible_cases.json  → evaluator cases, shared with the web app
internal/pkg/uploadthing/client.go             → DeleteFiles(ctx, keys) over net/http
internal/config/config.go                      → Upload{PublicURLPrefix, UploadThingSecret}
```

The shape of a field, as a style reference:

```go
// internal/pkg/form/form.go
type Field struct {
	Key            string      `json:"key"`
	Type           string      `json:"type"`
	Label          string      `json:"label"`
	Help           string      `json:"help,omitempty"`
	Required       bool        `json:"required"`
	RequiredAtDesk bool        `json:"required_at_desk,omitempty"`
	Retired        bool        `json:"retired,omitempty"`
	Options        []Option    `json:"options,omitempty"`
	Rows           []Option    `json:"rows,omitempty"`
	AllowOther     bool        `json:"allow_other,omitempty"`
	Multiple       bool        `json:"multiple,omitempty"`
	Scale          *Scale      `json:"scale,omitempty"`
	File           *FileRule   `json:"file,omitempty"`
	Validation     *Validation `json:"validation,omitempty"`
	AnsweredBy     string      `json:"answered_by,omitempty"`
	ShowIf         *Condition  `json:"show_if,omitempty"`
}

// Facts is what the caller knows about the attendee being answered.
// A nil Attendee means a companion with no account.
type Facts struct {
	Category   string
	Session    string
	Attendee   map[string]any
	Registrant map[string]any
}

// Validate returns the answers with hidden fields dropped, and every problem found.
// Desk makes RequiredAtDesk replace Required.
func Validate(fields []Field, in Input) (map[string]any, []Problem)
```

### Deferred: registration approval

Not in this spec. The interview of 2026-09-30 decided these points, so a future spec starts from them:

- The event sets a default for approval, and each ticket category can override it.
- A pending registration holds its seat. A rejection frees it.
- A new action `event.registrations.approve` exists and can be granted alone, for example to a treasurer.
- The approver can send a note. The person fixes their answers while the seat stays held.
- A pending person at the door is blocked. A staff member who also holds the approve action can approve on the spot and then check the person in.
- Still open: whether desk and walk-in registrations skip approval.

## Testing Strategy additions

Unit, no DB:

- Eligibility: AND across conditions, OR inside one condition, every condition kind, first failing condition named, `everyone` with a deactivated account rejected.
- `self_scan_protection`: each of the three values against inside window, outside window, inside radius, outside radius, missing coordinates.
- Matrix builder: a person registered for two of three occurrences yields `-` in the third cell and a percentage over two.
- Category constant set rejects `sunday_service`.
- Offline item classifier: inside window, inside tolerance, outside tolerance, older than the max sync hours, unknown person under each `offline_unknown_policy`.
- `scan_labels`: label copied for the used method, no label when the method has none, 41 characters rejected.
- Form types: a valid and an invalid answer for each new type. `{"other": ...}` fails with `allow_other` off and passes with it on. Other text of 201 characters fails.
- Grid: a missing row with `required` fails. `multiple` on and off. `max_picks` counts per row.
- Validation rules: each rule at its limit and one past it. Pattern `[0-9]{2}` rejects `123`, because the match covers the whole answer. The custom `message` comes back.
- Schema: a pattern that does not compile, `max_length` 501 on `text`, and `min` above `max` all fail at save.
- Conditions: each `source`, `of: registrant`, `department` as "contains", and a companion with no account. All cases live in `testdata/visible_cases.json`.
- `DiffSafe`: each allowed change passes. Each blocked change fails and names its field. Deleting a field with no answers passes.
- Desk rule: with `Desk` set, `required_at_desk` applies and `required` does not.
- `go list -deps ./internal/pkg/form | grep go-community` prints only `go-community/internal/pkg/form`.

Integration, real Postgres:

- Each row of the action table in delta 2: the listed title passes, a member without the title gets 403 `PERM_FORBIDDEN`.
- A `campus_admin` of BSD manages an event with `campus_codes = [BSD]` and gets 403 on one with `[PIK]`.
- An admin patch to `checkin_close_at` writes one `event_audit_logs` row with old and new values.
- Idempotency: two check-ins with the same `Idempotency-Key` produce one log row.
- Two hundred concurrent registrations for one seat produce one registration with status `confirmed`.
- Roster sync: a full call, then a registration and a cancellation, then a `since` call returns exactly those two rows. A second `since` call returns `changed: false`.
- Batch sync: two batches from two desks with the same attendee give one `ok` and one `duplicate`, and one log row with `offline = true`. Headcount 30 plus 25 gives 55.
- A batch item with a tampered token returns `QR_INVALID_TOKEN` for that item only, and the other items succeed.
- Every `EVENT_` code has an `id` message (shared platform test).
- A file answer with a URL from another host returns 400 with a `FieldError` on that field.
- After one registration with answers, a patch that deletes an answered field returns 409 `EVENT_FORM_UNSAFE_CHANGE`. Retiring the field returns 200, and the report keeps its column.
- An option label changed after answers exist shows the new label in the report.
- A personal QR walk-in on an event with a `required_at_desk` field returns `EVENT_DESK_ANSWERS_REQUIRED`. Desk registration with that field filled returns 201.
- File delete on a published event returns 409 `EVENT_FILES_IN_USE`. On a completed event, an `httptest` fake of UploadThing receives every file key, the answers become `[]`, and one audit row exists.
- A second file delete returns 200 and sends no request to the fake.
- A member aged 16 must answer a consent field shown to `age lt 17`. A member aged 20 submits without it, and an answer sent for the hidden field is dropped.

## Boundaries

Always:

- Pass `Resource.Parents` with campus codes and department on every event and session route.
- Run eligibility before `BEGIN` on every path that creates a registration.
- Write the audit row in the same `Atomic` transaction as the admin change.
- Evaluate `show_if` and validation again on the server at submit and at answer edit.
- Accept a file URL only with the `upload.public_url_prefix` prefix.
- Return only the viewer's own facts in `viewer_facts`, and only the keys the form uses.

Ask first:

- Adding a category value.
- Adding a condition kind to eligibility.
- Adding a fourth `self_scan_protection` value, for example a rotating poster.
- Any new column on `v3_attendance_logs` beyond `label`, `offline`, `device_scanned_at` and `synced_at` from deltas 15 and 16.
- Letting `session_qr` self-scan work offline.
- A new form field type, a new person fact, or a new condition source.
- Private files, or a second file store.
- Automatic file deletion.
- Registration approval, quiz mode or standalone forms.

Never:

- Read `user_types`, `roles` or `event_organizers`.
- Accept a session token's window from the client. The server reads the session row.
- Create a registration that spans more than one session.
- Purge `v3_attendance_logs` or `event_audit_logs`.
- Put phone, email or form answers in a roster row.
- Let a scan label change an attendance count or percentage.
- Fail a whole batch because one item failed.
- Store file bytes in the backend or in Postgres.
- Import a package from `go-community/internal/...` inside `internal/pkg/form`.
- Rewrite stored answers because the form changed. The file delete in 17.6 is the only exception.

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
11. A desk downloads the roster, goes offline, queues 50 check-ins, and one batch call later all 50 rows exist with `offline = true` and `checked_in_at` equal to the device scan time.
12. The same attendee sent from two desks yields one check-in and one `duplicate` result, with no error shown.
13. A roster `since` call with no changes returns `changed: false` and zero rows.
14. An event with `scan_labels = {personal_qr: "Hadir di lokasi", session_qr: "Hadir online"}` shows both counts in the session report and `label_counts` in JSON. An event with no labels shows no label column.
15. An organizer publishes a form with `radio`, `scale`, `time`, `datetime`, `grid` and `file` fields. A member registers with valid answers. The CSV has one labeled column per field and one per grid row.
16. A `radio` field with `allow_other` accepts `{"other": "Halal only"}`, and the report prints `Other: Halal only`. Without `allow_other`, the same answer returns `COMMON_INVALID_INPUT` for that field.
17. A `text` field with pattern `[0-9]{10}` and message `Tulis 10 angka.` rejects `123` with that message.
18. A consent field shown to `age lt 17` is required for a member aged 16 and absent for a member aged 20.
19. After one registration, deleting an answered field returns 409 `EVENT_FORM_UNSAFE_CHANGE`. Retiring it returns 200.
20. A personal QR walk-in on an event with a `required_at_desk` field returns `EVENT_DESK_ANSWERS_REQUIRED`. Desk registration with the field filled succeeds.
21. A file answer with a URL outside `upload.public_url_prefix` is rejected.
22. File delete on a completed event empties every file answer and writes one `files_deleted` audit row.
23. `go list -deps ./internal/pkg/form | grep go-community` prints only the form package itself.

## Open Questions

1. The eligibility checker now reads from `identity`, `cool` and `permissions`. Make sure that the three reader interfaces live in `internal/pkg/eligibility` and each module supplies its implementation in `contract.go`, so `eligibility` imports none of them.
2. The 5 second in-memory cache of session detail (delta 11) is per process. If the app ever runs more than one instance, the cache is per instance, which is still correct but less effective. Make sure that the free tier runs one instance.
3. Delta 7 removes `personal_qr` and `session_qr` from `attendance_modes`. Make sure that the admin UI shows them as always-on and only exposes `registration_qr` and `manual` as toggles.
4. Resolved 2026-09-17: `training` and `class` stay separate categories. The list is `general`, `volunteer`, `training`, `class`, `announcement`.
5. Before the delete task starts, make sure that the UploadThing delete API matches its current docs. That covers the path, the auth header, and the file key as the last URL segment.
6. Make sure that `max_mb` up to 10 fits the upload limits of the web app's UploadThing route on the free plan.
7. The church decides whether "keep files until an organizer deletes them" is acceptable under UU PDP for payment proofs and documents.
