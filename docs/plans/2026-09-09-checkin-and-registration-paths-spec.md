# Check-in & Registration Paths — Spec

**Date:** 2026-09-09
**Status:** Ready for implementation
**Builds on:** `docs/plans/2026-07-07-event-management-v3-design.md` (§4, §5, §6), its implementation plan (Tasks 10, 11, 13, 15), and ADRs 0002–0009
**Scope:** `/api/v3` only — v2 is untouched

---

## Problem Statement

Event v3 was designed around large ticketed events, where people sign themselves up online in advance and are scanned in at a door. Real church operations are broader than that, and several everyday situations currently have no correct path through the system.

**Staff cannot sign anyone up.** After Sunday service there is a table in the lobby where people put their names down for the Christmas service five weeks away. Today the only way into an event is for the person to register themselves online. A first-time visitor with no smartphone, no email and no account simply cannot be signed up at all.

**People who did register can be turned away.** An Anonymous Registration created at that lobby table has no account and possibly no email, so no Ticket QR ever reaches the person. When they arrive, there is nothing to scan. The only staff-driven option creates a *second* Registration, double-counting them and consuming another seat. The same trap catches anyone whose phone died or who lost their printed ticket.

**Restricted events are not actually restricted.** An event limited to volunteers enforces that rule only on the online sign-up path. Anyone can bypass it — most easily by presenting a Personal QR at the door, which creates a Registration on the spot without ever consulting the eligibility rules, and needs no staff involvement whatsoever.

**Small recurring gatherings are not served.** A monthly volunteer meeting does not need gatekeeping; it needs to know who turned up and who was late. There is no concept of lateness anywhere in the system, and check-in closing time is a hard wall rather than a line people can cross late.

**Anonymous counting is unforgiving.** The Headcount tap writes to an append-only record, so an unsure usher double-tapping in a busy doorway inflates the evening's attendance permanently, with no way to correct it.

**Check-out barely works.** It expires with the check-in window, so someone leaving a two-hour meeting an hour after arrivals closed cannot record it — and the Session QR poster they would scan to do so has already stopped working.

**Multi-room sessions have undefined behaviour.** A session split into Main Hall and Family Room quotas has no defined answer for which pool a walk-up scan should book into.

---

## Solution

Registrations can be created through three doors, all of which enforce the same rules: people signing themselves up online, staff signing someone up in advance at a Desk Registration, and QR Registration creating one automatically when somebody scans in without a booking. Organizers choose which doors are open per event.

Staff gain a way to register someone for a Session happening later, and a way to find an already-registered person by name and mark them present without creating anything new. Both are permission-controlled: any check-in-capable staff member can run the lobby table while registration is open, but only an organizer can add someone after the deadline they set.

Eligibility rules run wherever a Registration is created, closing the bypass. Events restricted to identifiable people stop accepting Anonymous Registrations, since there would be nothing to check the rules against. Organizers can additionally choose to re-check eligibility at the door, either noting or blocking people who no longer qualify.

Lateness is derived by comparing when somebody arrived against when the Session started — nothing new is stored, and how forgiving an event is falls out of when check-in closes. Sessions that do not care about lateness never mention it.

Headcount taps become correctable, check-out works for the whole Session, and walk-ups on a multi-room Session are asked which room rather than guessed at.

---

## User Stories

**Signing people up at a Desk Registration**

1. As a welcome-desk volunteer, I want to sign someone up at the lobby table for an event happening next month, so that people without smartphones are not excluded from events.
2. As a welcome-desk volunteer, I want to sign up a visitor who has no account, so that first-time guests can attend without being forced to create one.
3. As a welcome-desk volunteer, I want the person's seat reserved the moment I sign them up, so that I never promise a place that is already gone.
4. As a welcome-desk volunteer, I want to be told immediately if the event is full, so that I can tell the person honestly rather than discovering it later.
5. As an event organizer, I want only my organizers to be able to add people after the registration deadline, so that the catering numbers I signed off do not change without my knowledge.
6. As an event organizer, I want ordinary volunteers to run the lobby table while registration is open, so that I do not have to stand there myself.
7. As an event organizer, I want to see how many sign-ups the lobby table produced, separately from online sign-ups, so that I can judge whether staffing it was worthwhile.
8. As a welcome-desk volunteer, I want to sign someone up even after online registration has closed, so that a person asking two days before the event is not sent away.
9. As a visitor with no email address, I want to be signed up without one, so that a missing email does not stop me attending.
10. As a visitor signed up at a desk, I want to be told simply to give my name at the door, so that I know what to do without carrying anything.

**Checking people in**

11. As an usher, I want to find a registered person by name and mark them present, so that somebody whose ticket never arrived is not turned away.
12. As an usher, I want marking a person present by name to use their existing booking, so that I do not double-count them or take a second seat.
13. As an usher, I want to scan a member's Personal QR and have them registered and checked in together if they never booked, so that walk-ups are handled in one action.
14. As an attendee, I want to check myself in by scanning the poster at the venue, so that I do not have to queue.
15. As an usher, I want a clear message when somebody has already been checked in, so that I know it was a duplicate scan rather than a problem.
16. As an event organizer, I want duplicate scans recorded even though they are rejected, so that I can tell afterwards whether people were confused by the process.
17. As an usher, I want to be told which room to place a walk-up in when the Session has separate seating areas, so that I do not put a family with a toddler in the main hall.
18. As an event organizer, I want walk-ups never to be placed in a room automatically, so that reserved Family Room seats are not silently consumed by people who did not need them.
19. As an usher, I want to add somebody at the door who has no booking, so that a genuine attendee is not sent home.
20. As an event organizer, I want to close the door to unbooked arrivals while still allowing advance sign-ups, so that a catered dinner has final numbers.
21. As an event organizer, I want to close advance sign-ups while still admitting unbooked arrivals at the door, so that staff cannot quietly add people over the preceding month.

**Counting anonymously**

22. As an usher, I want to tap a counter for somebody I am not registering at all, so that overflow crowds are still counted.
23. As an usher, I want to undo a counter tap I made by mistake, so that the evening's numbers are not permanently wrong.
24. As an event organizer, I want undone taps to remain visible in the record, so that nobody can quietly rewrite an evening's attendance.
25. As an event organizer, I want anonymous counts reported separately from named attendees, so that I know how much of my total is people I know nothing about.
26. As an event organizer, I want counter taps to keep counting past the seat limit, so that I find out when more people came than the venue was meant to hold.

**Lateness**

27. As a volunteer coordinator, I want to see who arrived after the meeting started, so that I can follow up on persistent lateness.
28. As a volunteer coordinator, I want to accept people arriving late rather than locking them out, so that somebody delayed is still recorded as present.
29. As an event organizer, I want to decide how long after the start people can still check in, so that I control how forgiving each occasion is.
30. As an event organizer, I want lateness never mentioned on events where I have not asked for it, so that a Christmas service does not shame late arrivals.
31. As a volunteer coordinator, I want a late arrival to stay recorded as late permanently, so that the record reflects what happened rather than who argued about it.
32. As a volunteer coordinator, I want to add a note beside somebody's attendance, so that a good reason for lateness sits with the fact rather than erasing it.
33. As an event organizer, I want to turn lateness off for one occurrence of a recurring meeting, so that a retreat week is treated differently from the other fifty-one.

**Checking out**

34. As an attendee, I want to record that I left, so that presence duration is accurate for events that need it.
35. As a parent, I want the person collecting my child recorded, so that there is a record of who picked them up.
36. As an attendee, I want to check out any time before the event ends, so that leaving early is possible after arrivals have closed.
37. As an attendee, I want the venue poster still to work when I leave, so that I can check out the same way I checked in.
38. As an attendee, I want checking out to be a deliberate choice rather than a consequence of scanning twice, so that I do not accidentally mark myself as gone when I was checking that my arrival registered.

**Restricted events**

39. As a volunteer coordinator, I want people who do not qualify to be refused wherever they try to register, so that a restriction I set is actually enforced.
40. As a volunteer coordinator, I want somebody scanning in without a booking to be checked against the rules, so that the door is not an unguarded way in.
41. As a volunteer coordinator, I want nameless sign-ups refused on restricted events, so that there is always somebody the rules can be applied to.
42. As a volunteer coordinator, I want somebody who has since stopped qualifying to be noted when they arrive, so that I learn about it without a scene at the door.
43. As a volunteer coordinator, I want the option to turn away somebody who no longer qualifies, so that a confidential meeting can exclude a former member.
44. As an attendee refused entry, I want to be told why, so that I am not left guessing.
45. As an attendee who booked legitimately, I want not to be re-judged at the door by default, so that a change since I booked does not cost me my place unexpectedly.

**Configuring events**

46. As an event organizer, I want to run an event with no public sign-up link, so that everyone is registered by staff who can speak to them first.
47. As an event organizer, I want to require people to be physically at the venue to sign themselves up, so that remote sign-ups are prevented.
48. As an event organizer, I want the location check not applied to staff signing people up, so that the lobby table works for an event held elsewhere.
49. As an event organizer, I want to configure the location check for staff sign-ups when I need it, so that onsite-only registration is possible.
50. As an event organizer, I want to run a sign-up sheet with no attendance tracking, so that a trip needing bus numbers does not report that nobody came.
51. As an event organizer, I want to limit somebody to one Session of my event, so that people choose between the 8am and 10am service rather than booking both.
52. As an event organizer, I want that limit off by default, so that a weekly service does not stop people attending more than one week.
53. As an event organizer, I want somebody exceeding the limit told which Session they already booked, so that they understand the refusal.
54. As an event organizer, I want my settings applied to every generated week of a recurring event, so that I configure it once rather than fifty-two times.

---

## Implementation Decisions

### Registration channels

`registration_config` gains **`channels`**, a list of `web` (online self-registration) and `desk` (staff-created advance registration), defaulting to both. `mode` continues to say *whether* registration happens; `channels` says *through which door*. An empty list is rejected — `mode: none` already expresses "no registration," and two ways to say it invites drift. (ADR 0006)

`channels` and `attendance_modes` remain independent: the lobby table and the door are separately switchable, and both asymmetric combinations are supported deliberately. Admin UI must label them by *when* — "in advance" versus "at the door" — since both otherwise read as "staff can register people."

With `["desk"]`, the public event page must render alternative wording in place of a sign-up button.

### Desk Registration in advance

A new capability on the **existing registration usecase** (not a new usecase) creates a Registration on behalf of another person, writing `source=admin`. It is deliberately not a separate module, to keep the test seam count at one.

- Governed by the same `desk_registration` setting as the door (ADR 0002): `account_required`, `anonymous_only`, `anonymous_allowed`.
- Books seats through the same atomic seat-lock path as online registration — capacity is one shared pool across all three doors, hard-rejecting when full.
- Permission split (ADR 0005): while registration is open, any actor holding `event.checkin`; after `register_end_at`, `event.manage` only. Evaluated via `permission_grants` / `Can(actor, action, resource)` per ADR 0001.
- Allowed up until check-in closes. Blocking it would not prevent the outcome — staff would tell the person to arrive on the day and be registered as a walk-in, which admits them anyway while giving the organizer less warning and filing them under the wrong `source`.
- Email confirmation is sent when an address was given and silently skipped otherwise; no outbox row is created that can never be delivered. Nothing physical is produced — the person is found by name at the door.

`desk_registration_outcome` was considered and dropped: it made configuration out of what is simply which action a staff member took, and `source` already distinguishes the two cases. (ADR 0004)

### Check-in by attendee lookup

`manual` accepts an optional existing attendee identifier alongside its current typed-form input. Given one, it checks that attendee in **without** creating a Registration or booking a seat; staff locate the person through the existing attendee-list search. The log's `mode` remains `manual`, which is accurate — it was done by hand at the desk either way — and the created-versus-existing distinction stays derivable from `registrations.source`. No separate mode value was added; the reporting need it would serve is met by the source split below. (ADR 0004)

### Eligibility on every path

`eligibility` is evaluated wherever a Registration is **created**: online self-registration, advance Desk Registration, staff walk-in creation at the door, and QR Registration's walk-in auto-creation. The last is the significant gap — the current walk-in branch checks only that the mode is enabled and seats are available, so a person can scan into a restricted event with no staff involvement at all. Refusals name the reason rather than failing generically. (ADR 0007)

When `eligibility.audience` is anything other than `everyone`, **Anonymous Registration is refused** at both the lobby table and the door — there is no account for the rules to evaluate, and anyone legitimately eligible has one by definition. Open events are unaffected and keep accepting Anonymous Registrations per ADR 0002.

### Eligibility re-check at check-in

An **event-level** setting alongside `eligibility`, three-way like `geo_config.modes`: `off` (default, no re-check), `note` (admit, record that they no longer qualify), `block` (turn away). The result is **evaluated and stored at check-in time** — unlike lateness, it is judged against account state that keeps changing, so the answer that matters is the one true when the person walked in. Existing Registrations are never retroactively invalidated. (ADR 0008)

### Lateness

Derived: an attendee is late when `attended_at` is after the Session's `start_at`. Computed at read time, never stored, never clearable. Defined against the Session start rather than `checkin_close_at`, matching the everyday meaning and requiring no configuration.

**No "allow late check-in" setting exists.** How forgiving a Session is falls out of `checkin_close_at` alone — set it at `start_at` for a strict door, at `end_at` to accept arrivals throughout, or fifteen minutes past for a grace period.

**`enable_late_marking`** (per-Session, default off) governs whether lateness is surfaced at all — no door badge, no report column when off. Session-level, alongside `attendance_modes` and `enable_checkout`, so a single occurrence can be treated differently from its siblings. (ADR 0005)

### Check-out

Remains an **explicit action** (`action: checkout`) surfaced by `qr/resolve`'s per-caller `allowed_actions`, not a toggle on repeat scan — auto-toggling would silently check out an unsure attendee who rescans to confirm their arrival registered.

**Delta to Task 11:** the shared check-in/check-out code path must branch its window check on `Action` — `checkin` uses `[checkin_open_at, checkin_close_at]`, `checkout` uses the Session period up to `end_at`. An attendee who never checks out stays `attended` with a null `checked_out_at`. (ADR 0005)

**Delta to the QR subsystem:** when `enable_checkout` is true, the `session` token expires at `end_at` rather than `checkin_close_at`; otherwise unchanged. Tying this to the setting preserves the short expiry's purpose for the majority of Sessions that have no use for the poster once doors close. (ADR 0009)

### Duplicate check-in

**Delta to Task 11:** returns a distinct `ALREADY_CHECKED_IN` error class rather than `nil`. The `attendance_logs` row with `outcome=already_checked_in` is still appended, preserving the design's "every path appends exactly one log" invariant — the log records that a scan was *attempted*, not that attendance occurred, and reports counting `outcome='checked_in'` are unaffected. Scanner UI should render it distinctly from geo and forgery rejections. (ADR 0004)

### Headcount undo

Staff may undo a Headcount tap by appending a **correcting entry**; the original row is never removed and the total is the net. Available to any staff member at that Session while check-in is open — deliberately not restricted to the original tapper or a short timer, since every tap remains permanently visible for review regardless. One undo cancels exactly one tap. Requires its own recorded outcome value (e.g. `headcount_undone`) so the headcount count can net taps against undos; it cannot reuse `checked_in`. (ADR 0005)

### Ticket categories

Walk-in auto-creation on a Session with ticket categories **must be supplied a category** by the scanner — an usher picks with the attendee, a self-scanner chooses from a list — and creation is refused with a clear message when none is given. Never guessed: each category is a separately-sized quota, so a wrong guess either misplaces the person or silently consumes a scarce reserved seat. Sessions without categories are unaffected. (ADR 0009)

### Cross-session limit

`registration_config` gains **`max_sessions_per_person`**: `0` unlimited (default), `1` one Session of that Event only, `N` up to N. Distinct from the existing `allow_multiple_registrations`, which governs registering twice for the *same* Session; the numeric shape mirrors the neighbouring `max_per_registration`. Exceeding it is a hard reject naming the Session already registered for — never a silent auto-cancel-and-switch. **Anonymous Registrations are exempt**, consistent with ADR 0002: with no account, name matching fails in both directions, and enforcing it would require making phone mandatory for people who may have none to give. (ADR 0004)

### Geo policy

`geo_config.modes` gains a row for desk registration, same `off | warn | require` shape and `staff_override` flag, **defaulting to `off`** — staff registering someone in a church lobby weeks before an event held elsewhere are legitimately nowhere near the venue, and the phone being checked belongs to a trusted staff member rather than the attendee whose presence the check exists to verify. Configurable because onsite-only registration is anticipated.

`geo_config` **stays event-level**. Per-session geo was deferred: the config carries the venue coordinates, so moving it per Session really means "this Session is somewhere else," which deserves deliberate design. (ADR 0006)

### Attendance modes may be empty

An earlier decision to require at least one attendance mode is **reversed**. Empty is valid and meaningful — registration happens, attendance is not tracked (a trip sign-up needing bus numbers). Forcing a mode onto such an event produces misleading "0 of 40 attended" statistics. Distinct from `registration_config.mode: none`, which is an information-only Event with neither registration nor check-in.

### Reporting

**Delta to Task 15:** the Summary sheet splits advance registrations by `source` into three counts — `web`, `admin`, `walk_in` — so the lobby table's contribution is visible rather than merged into a single "registered" total. Lateness and eligibility-recheck columns appear only when their respective settings are enabled, following the existing pattern where checkout columns appear only when `enable_checkout`.

### Recurrence and templates

**Delta to Task 13:** recurrence `session_defaults` must carry the per-Session attendance settings (`attendance_modes`, `enable_checkout`, `enable_late_marking`) alongside its existing fields. Without this, every generated week of a recurring Event silently comes out with attendance tracking off. Event templates need the same treatment.

---

## Testing Decisions

### What makes a good test here

Tests assert **observable outcomes**, not how the code reached them: that a seat was consumed, that a Registration exists with the expected `source`, that an `attendance_logs` row was appended with the expected `outcome`, that a refusal carries the right error class. They never assert on call order, internal helper invocation, or the shape of intermediate values.

Because several decisions in this spec are specifically about **concurrent seat integrity** — two people racing for the last Family Room seat, walk-in creation booking atomically under load — tests exercise a real Postgres rather than a mocked repository. A mock cannot demonstrate that `SELECT ... FOR UPDATE` actually serialises anything, which is the property under test.

### Seams

**Primary seam: the usecase layer against a real database.** Every behaviour in this spec is observable by driving the check-in usecase and the registration usecase directly, with migrations applied and tables truncated between tests. This is the existing seam established by the v3 implementation plan's integration tests, and no new seam is introduced.

The new advance Desk Registration capability is deliberately placed on the **existing registration usecase** rather than a new module, so the seam count stays at one.

**Secondary seam: pure functions.** The eligibility evaluator, geo policy and QR token codec are already covered by table-driven unit tests with no database or HTTP. This spec barely touches them — it changes *who calls* eligibility, not the rule logic — so existing coverage stands, with additions only where a new caller needs a new case.

**Deliberately not added: HTTP-level tests.** Handler tests are not a seam in the v3 plan, and everything worth asserting is reachable one level down. Adding them would duplicate assertions across two layers for no additional confidence.

### Modules under test

- **Check-in usecase** — eligibility on walk-in creation, category selection on categorised Sessions, attendee-lookup check-in, checkout window branching, duplicate-scan error class plus its log row, headcount undo netting, lateness derivation, eligibility re-check outcomes.
- **Registration usecase** — advance Desk Registration with each `desk_registration` value, permission split across the `register_end_at` boundary, `channels` enforcement, `max_sessions_per_person` including the anonymous exemption, shared-pool capacity across all three sources.
- **QR subsystem** — `session` token expiry following `enable_checkout`.
- **Report usecase** — source split, and conditional appearance of lateness and eligibility columns.

### Prior art

The v3 implementation plan's existing integration tests are the model: a shared harness providing a migrated database, per-flow test files, table-driven cases, and seeding helpers that build an Event and Session through the real usecases rather than inserting rows directly. Seat-exhaustion races are already tested this way for online registration; walk-in and Desk Registration races follow the same pattern.

---

## Out of Scope

- **Per-Session geo policy.** Deferred deliberately — carrying venue coordinates per Session means "this Session is somewhere else," which touches how Sessions and venues relate throughout the system and deserves its own design. (ADR 0006)
- **Personal QR rotation.** Personal QRs are generated once at account registration and are permanent, with no self-service or staff-initiated regeneration. A deliberate trade for a congregation that skews older and less technical. (ADR 0003)
- **Anti-replay protection on the Session QR poster.** It stays a static, non-rotating code; the trust boundary is accepted for this audience. If abuse appears, `session_qr` can be disabled for that Session. (ADR 0003)
- **Clearing a late mark.** Lateness is a fact, not a judgment; excuses belong in a note beside the attendance. Allowing an override would force lateness to become stored state with its own permission rules. (ADR 0005)
- **Enforcing `max_sessions_per_person` for Anonymous Registrations.** Requires making phone mandatory for people who may have none to give, and name matching produces both false blocks and false passes. (ADR 0004)
- **Printing at the lobby table.** Printers, paper and maintenance for a slip likely lost over five weeks; the attendee-lookup check-in solves the actual problem.
- **Revising how user types and roles are modelled.** Flagged by the organizer as wanting attention, but explicitly left as-is for now.
- **Payments, v2 migration, and additional notification channels** — unchanged non-goals from the v3 design.

---

## Further Notes

**Four of the changes here are bug fixes to already-approved work**, not new features, and would otherwise have shipped: eligibility never running on QR walk-ins; pre-registered people with no working QR having no check-in path; the Session poster expiring before check-out can use it; and walk-in creation having no defined behaviour on categorised Sessions.

**The two items with real scope** are the advance Desk Registration capability and the eligibility enforcement fix. Everything else adjusts work already planned.

**Admin UI wording carries unusual weight in this spec.** Three pairs of settings are genuinely distinct but sound alike: `channels` versus `attendance_modes` (staff registering people *in advance* versus *at the door*); `max_sessions_per_person` versus `allow_multiple_registrations` (different Sessions of one Event versus the same Session twice); and `enable_late_marking` versus the check-in window (whether lateness is *shown* versus how long the door stays *open*). Each pair needs labelling that names the distinction, or they will be set wrongly.

**`block` on the eligibility re-check should be presented as the unusual choice** in the admin UI, with `note` recommended. It means an usher turning somebody away in a queue with no ability to judge or fix the situation; it exists for genuinely sensitive gatherings, not as a default posture.
