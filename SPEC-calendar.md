# Spec: calendar (fix-up delta)

Module id `calendar` in `CAPABILITY-MAP.md`. Intent: `docs/intent/v3-platform.md`. Depends on `events` and `cool`.

This file is a delta, not a rewrite. The spec for `calendar` stays:

- `docs/plans/2026-07-20-community-calendar-design.md` (cited as "design §N")
- `docs/plans/2026-07-20-community-calendar-implementation.md` (cited as "Task N")

Where this file and those documents disagree, this file wins. Everything not named here is unchanged.

## Objective

Make the calendar read the v3 tables that `events` and `cool` actually create, use the v3 permission model, and speak the platform error contract. The calendar stays a read-only, query-time merge of event sessions and COOL meetings with a 92-day cap.

## Deltas

### 1. Table names and "my events"

Changes: design §2, design §6, Task 3 (both SQL queries and `UserFacts`).

- `event_sessions` becomes `v3_sessions`. `events` becomes `v3_events`. `registrations` becomes `v3_registrations`.
- `event_organizers` and `cool_facilitators` do not exist. Remove both joins.
- If `v3_registrations` has a row with the session id, `registered_by = community_id` and `status = confirmed`, the `registered` flag is true. Rows in `v3_registration_attendees` are not needed for the flag. They stay out of the query.
- The `organizer` flag comes from `permission.Scopes(ctx, actor, "event.view_internal", "event")`. If `all` is true, every event is organized. Otherwise the returned refs are the organized event codes. The sessions query takes that list as a parameter in place of the organizer join.
- COOL membership is one row in `cool_members` with `left_at IS NULL`. A holder of `cool_facilitator` sees their COOLs through `permission.Scopes(ctx, actor, "cool.attendance.view", "cool")`. The meetings query takes the union of the member's COOL id and those refs.
- `UserFacts` is removed. The eligibility checker in `events` loads what it needs (delta 3).

### 2. Errors and envelope

Changes: design §7, Task 1 `errors.go`, Task 5 responder.

- Design §7 said sentinel errors in `error_model.go` plus `ErrorMapping`. That is wrong. The plan already said herr. Resolve to herr.
- The two classes become `CAL_INVALID_RANGE` and `CAL_INVALID_FILTER`, created with `herrx.Def` in `internal/models/calendar/errors.go`, with `id` messages in `locale.go`. The old codes `CALENDAR_INVALID_RANGE` and `CALENDAR_INVALID_FILTER` are not used.
- Handlers call `response.OK` and `response.Fail` from `platform`. The success body is `{ "data": { "entries": [...] } }`. `response.Success` is never used in v3.
- `models.ErrorMapping` is never touched.

### 3. Visibility

Changes: design §4, Task 4 rules 2 and 3.

- Member tier: sessions of `published` events where `eligibility.Check(ctx, actor, e)` from `events` returns true, evaluated once per event code in the request. The calendar never re-implements eligibility and never reads `roles` or `user_types`.
- Member tier also includes meetings of the member's own COOL and of every COOL the member holds a COOL title on.
- Staff overlay: events returned by `permission.Scopes(actor, "event.view_internal", "event")` are shown in `draft` status and outside eligibility, with `flags.organizer` and `flags.draft` set as before.
- Admin tier: `permission.Can(ctx, actor, "calendar.view_all", Resource{Kind: "global"})`. `superadmin` passes by the evaluator's own bypass. `campus_admin` and `event_admin` hold `calendar.view_all` per `SPEC-permissions.md`. A `campus_admin` grant is campus-scoped, so the calendar passes each event's campus codes as `Resource.Parents` and checks per event, not once per request.
- `sunday_service` never appears, because `events` rejects that category on create and patch.

### 4. COOL meetings

Changes: design §3, design §6, Task 3 meetings query, Task 4 rule 4.

- `cool_meetings` has `starts_at` and `ends_at` as `timestamptz`. The `meeting_date` plus `start_time` composition and the two-hour default are removed.
- Only meetings with `status` in (`pending`, `closed`) appear. `cancelled` meetings never appear. Design §7's rule that cancelled entries stay visible now applies to event sessions only.
- Entry `status` for a meeting maps from the meeting status: `pending` before `ends_at` is `scheduled` or `ongoing` by clock, `pending` after `ends_at` and `closed` are `completed`.
- `location_type` is exposed as stored: `offline`, `online` or `hybrid`. The normalization of `offline` to `onsite` is removed. Event sessions keep their own vocabulary. The JSON field documents both value sets.
- `ref` for a meeting is `{"cool_id": ..., "meeting_code": ...}` as before.

### 5. `.ics` export

Changes: design §5 iCal, Task 2.

- Unchanged, except the entries come from the tables in delta 1.
- `DESCRIPTION` gains a first line `Category: <category>` for event sessions. Meetings use `Category: cool`.
- `STATUS:CANCELLED` is emitted for cancelled sessions only, since cancelled meetings are not exported.

### 6. Range cap and time zone

Changes: none.

The 92-day cap and `Asia/Jakarta` serialization stay exactly as in design §3 and §5.

### 7. Category filter

Changes: design §5 query params, Task 1 `Query`, Task 4 rule 6, Task 5 param parsing.

- New optional param `category`, a CSV of event category values from `SPEC-events.md` delta 3. Unknown values return `CAL_INVALID_FILTER`.
- The filter drops event entries whose `v3_events.category` is not in the list. Meeting entries pass through, like the campus filter.
- Each event entry gains a `category` field in the JSON shape. Meeting entries carry `category: null`.

### 8. Mount

Changes: design §5, Task 5 wiring.

- Routes are `GET /api/v3/calendar` and `GET /api/v3/calendar/export.ics`.
- Both use the identity JWT middleware only. No `permission.Require`, because tiering happens inside the usecase. `middleware.UserMiddleware` is never used.
- v2 has no calendar. Nothing to deprecate.

## Testing Strategy additions

Unit, no DB:

- Meeting status mapping for the three cases in delta 4.
- Category filter keeps matching events and passes meetings through.
- `.ics` golden file gains the `Category:` line and has no cancelled meeting.

Integration, real Postgres via `TEST_DATABASE_URL`:

- Seed `v3_events`, `v3_sessions`, `v3_registrations`, `cools`, `cool_members`, `cool_meetings` and `permission_grants` only. No legacy table is seeded.
- A member of COOL `C-1` sees `C-1` meetings and none from `C-2`. A `cool_facilitator` of `C-2` also sees `C-2`.
- A `cancelled` meeting in range is absent. A `cancelled` session in range is present with `status: cancelled`.
- An `event_organizer` grant on a draft event shows its session with `flags.draft`. Revoking the grant hides it on the next request.
- A `campus_admin` of BSD sees a draft event with `campus_codes = [BSD]` and does not see one with `[PIK]`.
- Every `CAL_` code has an `id` message (shared platform test).

## Boundaries

Always:

- Read only. No write to any table from this module.
- Reuse `eligibility.Check` and `permission.Scopes`. Never copy their logic.
- Keep the 92-day cap.

Ask first:

- A third source (room booking, spec B).
- Pagination on the JSON endpoint.
- Any change to the `UID` format in `.ics`.

Never:

- Read `event_sessions`, `registrations`, `event_organizers`, `cool_facilitators`, `roles` or `user_types`.
- Show a `cancelled` COOL meeting or an `archived` event.
- Add to `models.ErrorMapping`.

## Success Criteria

1. `GET /api/v3/calendar?from=&to=` for a plain member returns only published, eligible sessions plus their own COOL's meetings.
2. The same call with `category=volunteer` returns only volunteer sessions and all the member's meetings.
3. A `cancelled` COOL meeting in range does not appear in JSON or `.ics`.
4. A `campus_admin` of BSD sees draft sessions at BSD and none at PIK.
5. A 100-day range returns 400 `CAL_INVALID_RANGE` with an Indonesian message by default.
6. `grep -rn 'event_organizers\|cool_facilitators\|meeting_date\|ErrorMapping' internal/**/calendar` returns nothing.

## Open Questions

1. Resolved 2026-09-17: `permission.Scopes` now also returns `parents` (campus and department refs). The sessions query filters `v3_events.code IN refs OR department_code IN parents.department OR campus_codes && parents.campus`.
2. Resolved 2026-09-17: a `pending` meeting after `ends_at` shows as `completed` by clock. The leader closing it later does not change what the calendar shows.
