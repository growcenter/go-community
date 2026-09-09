# go-community

Backend for GROW IT Team's Church Community Dashboard — events, COOL (small groups), users, and community calendar.

## Language

**Event**:
The overall thing being organized — a conference, a recurring weekly service, an announcement page. Holds shared config (eligibility, form, geo policy) that applies across all its occurrences.

**Session**:
A single scheduled occurrence of an Event, with its own start/end time, registration window, check-in window, and seat count. Deliberately generic — the same shape covers a one-off gathering, a weekly class meeting, a conference talk slot, or a recurring service instance; the type of occurrence isn't baked into the model.
_Avoid_: Instance (v2's term for this concept — v3 renamed it since v3 occurrences carry materially more structure than v2's; v2 and v3 are separate, non-migrating domains so terms don't need to match across them).

**Desk Registration**:
A registration created by staff on the spot at the event (as opposed to the attendee registering themselves online beforehand) — the registration-creation side of the `manual` attendance mode (`docs/plans/2026-07-07-event-management-v3-design.md` §5.1/5.2). Configurable per event via a `desk_registration` setting (ADR 0002): `account_required` (staff look up an existing member), `anonymous_only`, or `anonymous_allowed` (both).
_Avoid_: Walk-in registration (still valid as a casual term, but "Desk Registration" is the precise one to use in specs/config — "walk-in" is also used for the separate no-registration **Headcount** tap, which is a different thing, see below).

**Anonymous Registration**:
A Registration with no owning account — used only for Desk Registration, never for online self-registration or QR Registration. Never auto-creates a user account for that person; if they want one later, that is a separate, self-initiated, consent-driven signup, disconnected from the registration record.

**QR Registration**:
Descriptive term (not a new field) for existing behavior in `attendance_modes[]` §5.2: scanning `personal_qr` or `session_qr` auto-creates a walk-in Registration (if the scanning account has none yet for that Session) and checks them in, in one step — gated only by that mode being enabled and seats being available. Always account-required — no anonymous path. **Personal QR** (staff scan the attendee's permanent, account-owned code) and **Session QR** (attendee self-scans a poster code owned by one Session) are the two scan actors for this same mechanism.

**Ticket QR** (canonical field/enum value: `registration_qr`):
Not a registration path — a one-time check-in credential issued for one specific Registration (produced by online self-registration or Desk Registration), scanned by staff on the attendee. No account required to check in with it. "Ticket QR" is the friendlier name used in conversation/specs; `registration_qr` is the value stored in `attendance_modes[]`/`attendance_logs.mode`.

**Headcount**:
A bare `+1` tally with **no identity captured at all** — not a Registration, no attendee row, just an anonymous count logged against the Session (`attendance_logs` with `attendee_id = null`). Rides on the `manual` attendance mode being enabled (not a separate mode). For overflow counting / no-phone attendees where even Desk Registration's minimal form is too much.

**Attendance Mode**:
One of `personal_qr`, `session_qr`, `registration_qr`, `manual` (design §5.1) — a Session's `attendance_modes[]` (subset of these four) governs which check-in methods are accepted. Headcount is a sub-capability of `manual`, not a fifth mode.
**Empty is valid and meaningful**: registration happens, attendance is not tracked (a trip sign-up sheet needing bus numbers; an information evening counting chairs). An earlier decision to make this a required non-empty field was reversed — forcing a check-in method onto an event that never intended to track attendance produces misleading "0 of 40 attended" statistics. Distinct from `registration_config.mode: none`, which is an information-only event with no registration *and* no check-in.

## Relationships

- An **Event** has 1..N **Sessions**.
- A **Session**'s own `start_at`/`end_at` (event dates), `register_start_at`/`register_end_at` (registration window), `checkin_open_at`/`checkin_close_at` (check-in window), seat count (capacity), and `attendance_modes[]` are independent per-session — an Event's overall period is derived (earliest start → latest end), never stored separately.
- An **Anonymous Registration** is only possible as a **Desk Registration** — online self-registration and QR Registration always require an account.
- Capacity (`total_seats`/`booked_seats`) is one shared pool across every registration path for a Session (self-registration, Desk Registration, QR Registration) — no separate per-path quotas. Enforced once, at the moment of registration (a hard reject when full, via `SELECT ... FOR UPDATE`) — including the auto-registration moment triggered by a QR Registration scan. A later Ticket QR (`registration_qr`) check-in never re-checks capacity, since it was already enforced when that registration was created.
- **Session-scoped duplicate registration** (`registration_config.allow_multiple_registrations`, default false) is a *different* concept from **cross-session restriction within one Event** — the former blocks registering twice for the *same* Session; the latter (name TBD — do not reuse "multi_session_registration" or "allow_multiple_registrations," both are already taken/confusing) would block registering for a *second* Session of the *same* Event. The cross-Session rule is a new, not-yet-implemented Event-level setting (default: unrestricted) and still needs a distinct name — see Open Decisions.
- A **Session QR** belongs to exactly one Session — a recurring Event with multiple Sessions has one Session QR per occurrence, not one shared across all of them, since check-in windows and capacity are already per-Session.
- **Personal QR** is auto-generated once, at account registration, and is permanent — no self-service or staff-initiated regeneration/rotation exists or is planned. Accessible any time via an in-app menu.
- Every check-in/checkout event logs (`attendance_logs`): timestamp, mode (`personal_qr`/`session_qr`/`registration_qr`/`manual`, or headcount as a `manual` sub-case), the Session, and — for staff-actuated modes only (`personal_qr`, `registration_qr`, `manual`) — the acting staff member (`checked_by`). `session_qr` has no staff actor (self-scanned).

## Resolved refinements (see ADR 0004 for reasoning)

- **Check-out is an explicit action**, not a toggle — the client sends `action: checkout`, offered by `qr/resolve`'s `allowed_actions`. Bounded by the session's `end_at`, **not** `checkin_close_at` (someone leaving mid-service can still record it after arrivals have closed). Consequently the **session poster token expires at `end_at` when `enable_checkout` is true** (design §6.1 otherwise expires it at `checkin_close_at`, which would leave self-scanners unable to check out) — see ADR 0009.
- **A duplicate check-in returns an error** to the scanner (distinct `ALREADY_CHECKED_IN` class), while still appending its `attendance_logs` row with `outcome=already_checked_in`. The log records that a scan was *attempted*; it is not a claim of attendance, and reports counting `outcome='checked_in'` are unaffected.
- **Two distinct staff registration actions**, distinguished by `registrations.source` — no organizer-facing toggle: `admin` (the lobby table — staff sign someone up for a later session; **new endpoint required**) and `walk_in` (the door — register + check in together, already covered by `manual` mode). Both governed by the same `desk_registration` setting (ADR 0002).
- **`manual` mode can check in an already-registered attendee** found via attendee-list search, without creating a registration or booking a seat — for people who never received a ticket (anonymous lobby-table registrants) or arrived without it. Log `mode` stays `manual`.
- **Walk-in auto-creation on a session with ticket categories must be given a category** — the scanner asks (usher picks with the attendee; self-scanners choose from a list), and the check-in is refused with a clear message if none is supplied. Never guessed: each category is a separate quota, so a wrong guess either misplaces the person or silently consumes a scarce reserved seat (e.g. one of 50 Family Room places). Sessions without categories are unaffected and prompt for nothing.
- **`max_sessions_per_person`** (in `registration_config`): `0` = unlimited (default), `1` = one Session of that Event only, `N` = up to N. Distinct from `allow_multiple_registrations`, which governs registering twice for the *same* Session. Hard reject naming the already-registered Session; never an auto-switch. **Anonymous registrations are exempt**, consistent with ADR 0002.
- **`attendance_logs.checked_by`** is populated for staff-actuated modes (`personal_qr`, `registration_qr`, `manual`) and null for `session_qr` (self-scanned) — already the existing schema; re-confirmed, do not re-litigate.
- **Reports split advance registrations by source** — `web`, `admin`, and `walk_in` counted separately, so the lobby table's contribution is visible rather than merged into a single "registered" total.

## Lateness, headcount undo, advance desk registration (see ADR 0005)

- **Late** means `attended_at > start_at` — derived at read time, never stored, never clearable. Deliberately *not* defined against `checkin_close_at`. How forgiving a session is comes from `checkin_close_at` alone; there is no separate "allow late check-in" toggle.
- **`enable_late_marking`** (per-session, default off) governs whether lateness is surfaced at all — no door badge, no report column when off. Sits alongside `attendance_modes` and `enable_checkout`.
- **Headcount taps are undoable** while check-in is open, by any staff at that session. An undo appends a correcting entry (the original tap row is never removed); the total is the net. One undo cancels one tap.
- **Recurrence `session_defaults` must carry the per-session attendance settings** (`attendance_modes`, `enable_checkout`, `enable_late_marking`) alongside its existing `start_time`/`duration_min`/`total_seats`. Without this, every auto-generated week of a recurring event comes out with attendance tracking off, silently — the weekly volunteer meeting would generate 52 sessions that record nobody. Event **templates** need the same treatment for the same reason.
- **A headcount undo needs its own recorded outcome** (e.g. `headcount_undone`) so `CountHeadcount` can net taps against undos; it cannot reuse `checked_in`.
- **Report columns** gain lateness (when `enable_late_marking`) and the eligibility-recheck result (when that setting is not `off`), following the existing pattern where checkout columns appear only when `enable_checkout` — unconfigured capabilities stay invisible.
- **Advance desk registration (`source=admin`) past `register_end_at`** is allowed but organizer-only (`event.manage`); while registration is open, any check-in-capable staff (`event.checkin`) may use the lobby table. Seat limits apply throughout.
- A lobby-table registrant receives **nothing physical** — email confirmation if they gave an address, otherwise silently skipped (no stuck outbox row). They are found by name at the door via `manual` mode's attendee lookup.

## Registration channels and desk geo (see ADR 0006)

- **`registration_config.channels`** — list of `web` (self-service online) and `desk` (staff-created in advance), default both, empty rejected. `mode` says *whether* registration happens; `channels` says *through which door*. `["desk"]` means no public sign-up link — the event page must say "sign up at the welcome desk" or read as broken.
- **The lobby table (`channels`) and the door (`attendance_modes` walk-in path) are independent switches** — both asymmetric combinations are deliberately supported. Label them by *when* ("in advance" vs "at the door") in admin UI, or they will be confused.
- **`geo_config.modes` gains a desk-registration row** (`off | warn | require`, default `off`). Off by default because staff legitimately register people from a church lobby weeks before an event held elsewhere — the check would measure the wrong person in the wrong place. `geo_config` stays **event-level**; per-session geo is deferred as really being "this session is somewhere else."

## Eligibility (see ADR 0007)

- **Eligibility runs wherever a registration is created** — online, staff advance (`source=admin`), staff walk-in at the door, and **QR-triggered walk-in auto-creation**. The last is a gap in the current design/Task 11, which check only mode-enabled and seats-available: today a non-volunteer could scan into a volunteers-only event with no staff involvement at all.
- **Restricted events (`eligibility.audience != everyone`) reject anonymous registrations** at both the lobby table and the door — there is no account for the rules to evaluate. Anyone legitimately eligible has an account by definition. Open events are unaffected and keep accepting anonymous registrations per ADR 0002.
- **Eligibility re-check at check-in** (ADR 0008) is an **event-level** setting alongside `eligibility` itself, three-way like `geo_config.modes`: `off` (default, no re-check) / `note` (admit, but record that they no longer qualify) / `block` (turn away). The result is **evaluated and stored at check-in time**, unlike lateness — eligibility is judged against account state that keeps changing, so the answer must be captured when it was true. Existing registrations are never retroactively invalidated.

## Flagged ambiguities

- v2 uses "Instance" (`EventInstance`, `event_instances` table) for the same concept v3 calls "Session" (`event_sessions` table). Resolved: v3 uses "Session" going forward; this is a deliberate rename, not drift — do not reuse "Instance" in new v3 docs/code.
