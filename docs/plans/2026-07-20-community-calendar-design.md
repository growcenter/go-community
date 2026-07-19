# Community Calendar — Design

**Date:** 2026-07-20
**Status:** Approved design, pre-implementation
**Repo:** go-community (`/api/v3`, new `calendar` domain)
**Depends on:** Event Management v3 (`2026-07-07-event-management-v3-design.md`), COOL & Unified Permissions (`2026-07-07-cool-and-unified-permissions-design.md`)

---

## 1. Goal & Scope

One centralized, role-aware calendar: every user opens it and sees the church happenings relevant to *them* — event sessions they're eligible for, sessions they registered for (highlighted), and their COOL's meetings. Staff and admins see progressively more.

**Decomposition decision:** the original request bundled three systems. This spec covers only **(A) the calendar view**. **(B) Calendly-style public booking** (Google Calendar free/busy, booking slugs, tokenized reschedule/cancel, Meet links) is a separate future spec. **(C) Room booking** is future-reference only; the unified entry shape (§3) is designed so a new source slots in without breaking the API.

**Non-goals (v1):**

- No booking, no slot picking, no Google Calendar write integration (that's spec B).
- No manual/standalone calendar entries (holidays, announcements) — only derived sources.
- No live iCal subscription feed — `.ics` is a one-shot authenticated download.
- No department filter — neither source table carries a department column (see §5).
- No new tables, no migrations, no new permission actions.

**Scale context:** ~1,000 users. Query-time aggregation is comfortably sufficient; no materialization.

## 2. Approach

**Query-time aggregation** (chosen over a materialized `calendar_entries` table and over per-source endpoints merged client-side):

- The calendar usecase runs two range queries — `event_sessions ⋈ events` and `cool_meetings ⋈ cools` — merges the rows into one unified `CalendarEntry` list, sorts by `start_at`, and returns it.
- No sync machinery, no duplicate source of truth. Role logic lives server-side once.
- Future sources (room bookings, spec-B bookings) are added as a third query into the same merge.

New v3 domain following repo layering conventions:

| Layer | File |
|---|---|
| Models | `internal/models/calendar_model.go` |
| Repository | `internal/repositories/pgsql/calendar_pg.go` + `calendar_pg_query.go` (raw SQL, `Build*` assemblers) |
| Usecase | `internal/usecases/calendar_usecase.go` |
| Handler | `internal/deliveries/http/v3/calendar_handler.go` |

Read-only domain: repository exposes only range-scan reads; no writes, so no transaction helper involvement.

## 3. Unified entry shape

```json
{
  "source": "event_session | cool_meeting",
  "ref": {
    "event_code": "EVT123", "session_code": "SES456"
  },
  "title": "Session 2 — Worship Night",
  "parent_title": "Christmas Celebration 2026",
  "start_at": "2026-12-24T19:00:00+07:00",
  "end_at": "2026-12-24T21:00:00+07:00",
  "location_type": "onsite | online | hybrid",
  "location_name": "Main Hall",
  "online_url": "https://... (null when not applicable / not entitled)",
  "campus_codes": ["BKS"],
  "status": "scheduled | ongoing | completed | cancelled",
  "flags": {
    "registered": false,
    "organizer": false,
    "draft": false
  }
}
```

- For `cool_meeting`, `ref` is `{"cool_id": 1, "meeting_code": "..."}`; `title` = meeting `topic`, `parent_title` = cool name; `campus_codes` is `null` (COOLs carry `area`/`region`, not campus).
- `flags.registered` — user holds a `confirmed` registration on that session (`registrations.registered_by = community_id`). This is the "my registrations" highlight.
- `flags.organizer` / `flags.draft` — staff overlay markers (§4).
- COOL `location_type` values `online|offline|hybrid` are normalized to the event vocabulary (`offline` → `onsite`).
- All timestamps serialized in `Asia/Jakarta` (single-timezone church; matches existing v3 convention).

## 4. Visibility tiers (role-based view)

Evaluated per request, server-side, in the usecase. Tiers are additive.

1. **Member (everyone authenticated):**
   - Event sessions of `published` events where the user passes the event's `eligibility` config — reuse the events-v3 eligibility evaluator as-is, evaluated **once per event** in the range (not per session), memoized in request scope.
   - Meetings of the COOL where the user is a `cool_members` row, plus meetings of every COOL the user facilitates (`cool_facilitators`).
2. **Staff overlay (per-event, via `event_organizers`):**
   - Additionally: sessions of the user's organized events in `draft` status or outside their eligibility, flagged `"draft": true` and/or `"organizer": true`.
3. **Admin (unified permission engine):**
   - `superadmin` bypass, or a grant check through the permission engine (action `calendar.view_all`, added to the fixed action catalog — a catalog entry, not a new mechanism): all events in any status except `archived`, all COOL meetings.

No visibility rule lives in the handler; the handler passes the authenticated `community_id` and the usecase asks the permission engine.

## 5. API surface

```
GET /v3/calendar            ?from=&to=&sources=&campus=&mine=
GET /v3/calendar/export.ics ?from=&to=&sources=&campus=&mine=
```

Both bearer-auth required (standard `UserMiddleware` token-holder group; tiering happens inside, per §4).

**Query params (shared):**

| Param | Rule |
|---|---|
| `from`, `to` | Required, RFC3339 date or datetime. `to > from`. Max span **92 days** → guard against unbounded scans. |
| `sources` | Optional CSV of `events,cool`. Default: both. |
| `campus` | Optional CSV of campus codes; filters event entries by `campus_codes` array overlap. COOL entries pass through unaffected (no campus column). Codes validated against config `Campus` map. |
| `mine` | Optional bool. `true` → only sessions the user registered for + own/facilitated COOL meetings. |

**JSON endpoint** returns `{"entries": [...]}` sorted ascending by `start_at`, then `source`, then code — deterministic order. No pagination: 92-day cap × ~1,000-user scale keeps payloads small; revisit only if measured otherwise.

**iCal endpoint** runs the identical pipeline, serializes to `text/calendar` with `Content-Disposition: attachment; filename="community-calendar.ics"`. One `VEVENT` per entry:

- `UID`: `<source>-<code>@go-community` (stable across downloads).
- `DTSTART`/`DTEND` in `Asia/Jakarta` (`TZID` included).
- `SUMMARY` = `parent_title — title`; `LOCATION` = `location_name`; `URL` = `online_url` when present.
- Cancelled entries emit `STATUS:CANCELLED`.

**Department filter — deliberately dropped.** Neither `events` nor `cools` carries a department column. If it becomes real, amend events v3 with `department_codes text_arr` and add the param then. Documented here so the omission reads as a decision, not a miss.

## 6. Data access

Two raw-SQL queries in `calendar_pg_query.go` (repo convention for joins):

1. **Sessions query:** `event_sessions` ⋈ `events` where `start_at < :to AND end_at > :from`, `deleted_at IS NULL` on both, event status ∈ tier-dependent set; selects event `eligibility` JSONB (evaluated in Go), `campus_codes`, plus a `LEFT JOIN registrations` existence flag on (`session_id`, `registered_by = :community_id`, status `confirmed`) and a `LEFT JOIN event_organizers` flag.
2. **Meetings query:** `cool_meetings` ⋈ `cools` where `meeting_date` within range, joined against `cool_members`/`cool_facilitators` for the member tier (admin tier drops the membership predicate).

Time composition for COOL rows: `meeting_date + start_time`/`end_time` composed in `Asia/Jakarta`. `end_time` NULL → `start + 2h` default (keeps `.ics` `DTEND` valid; JSON gets the same default for consistency).

## 7. Errors & edge cases

Sentinel errors in `error_model.go` + `ErrorMapping`, bilingual EN/ID per the herr convention:

- Missing/invalid `from`/`to`, `to <= from`, span > 92 days → 400.
- Unknown `sources` value or campus code → 400.
- Cancelled sessions/meetings **stay visible** within range with `status: cancelled` — users must see cancellations, not silent disappearance. (Archived events are excluded for everyone.)
- Empty range → `{"entries": []}`, 200.
- `online_url` for events is included only when the member tier already entitles the user to register/attend; draft-overlay and admin tiers always see it.

## 8. Testing

- **Usecase unit tests:** visibility tier matrix (member / organizer / admin × event / cool), `registered` flag, `mine=true` narrowing, source & campus filters, range-guard rejections, COOL time composition incl. NULL `end_time`.
- **Repository integration tests** (against migration-built DB, existing harness): range boundary overlap (session straddling `from`), soft-delete exclusion, campus array overlap, registration-flag join correctness.
- **iCal serializer:** golden-file test — fixed entry list in, byte-exact `.ics` out (UID stability, TZID, CANCELLED status).

## 9. Build order

Implement after events v3 and COOL modules land (calendar reads both). Within its own slice: models → repo queries + integration tests → usecase + unit tests → handler + swagger annotations → iCal serializer. Single PR-sized module; no migration step.

## 10. Future extensions (documented, not built)

- **Spec B — booking system:** own spec; its confirmed bookings would surface here as a third source query.
- **Room booking:** same — new source + `"source": "room_booking"` value; entry shape already accommodates it.
- **Live iCal subscription feed:** would need per-user feed tokens; revisit on demand.
- **`department_codes` on events** → enables the department filter param.
