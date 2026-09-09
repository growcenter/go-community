# Spec: `attendance`

**Module id:** `attendance` · **Depends on:** `registration`, `qr`, `policy-kernel`, *external* `permissions` · **Build order:** 5th (parallel with `notifications`)
**Parent:** [SPEC.md](SPEC.md)

---

## Objective

Record who came, without excluding anyone who legitimately turned up. Four scanning modes plus a bare anonymous tally, an explicit check-out, and a lateness signal for the small recurring gatherings the original design never served. Attendance is entirely optional and independent of registration — an event that never asked for it must never report that nobody came.

Four of this module's requirements are **bug fixes to already-approved work**, not new features: eligibility never running on QR walk-ins, pre-registered people with no working QR having no check-in path, the session poster expiring before check-out can use it, and walk-in creation having no defined behaviour on categorised sessions.

## Scope

**In:** Check-in and check-out across four modes; walk-in creation (delegated to `registration.BookAndCreate`); check-in by attendee lookup; headcount and headcount undo; lateness derivation; eligibility re-check; the append-only `attendance_logs` invariant.

**Out:** Creating registrations (that primitive belongs to `registration`); QR token mechanics (`qr`); report columns (`reporting`).

---

## Contracts

```go
type CheckinUsecase struct{ r *pgsql.PostgreRepositories; cfg *config.Configuration }

type CheckinInput struct {
	SessionCode  string
	Mode         string  // personal_qr | session_qr | registration_qr | manual
	Action       string  // checkin | checkout
	IsHeadcount  bool    // rides on `manual`; NOT a fifth mode
	AttendeeID   *uuid.UUID       // registration_qr, or manual lookup check-in
	CommunityID  string           // personal_qr target, or session_qr caller
	ManualForm   map[string]any   // manual walk-in creation
	CategoryCode string           // required when the session has categories and this creates
	Coords       *geo.Coords
	Override     bool
	ActedBy      eligibility.User // staff for staff modes; the member for session_qr
}

type CheckinResult struct {
	Outcome  string   // checked_in | checked_out | headcount_recorded | headcount_undone
	Attendee *v3.AttendeeV3
	IsLate   bool     // populated only when enable_late_marking
}

Act(ctx, in CheckinInput) (*CheckinResult, error)
UndoHeadcount(ctx, sessionCode string, by eligibility.User) error
```

---

## Acceptance Criteria

### The invariant

1. **Every path appends exactly one `attendance_logs` row — success and rejection alike.** The log records that a scan was *attempted*; it is not a claim of attendance. `attendance_logs` is never updated and never deleted; corrections are new rows.
2. `checked_by` is populated for the staff-actuated modes (`personal_qr`, `registration_qr`, `manual`) and **null for `session_qr`**, which is self-scanned. This is the existing schema and is not to be re-litigated.
3. `mode` is one of **four** values. Headcount is a sub-capability of `manual` flagged by `is_headcount`, not a fifth mode (ADR 0005).

### Mode gating and windows

4. The requested mode must be present in the session's `attendance_modes`. `attendance_modes: []` activates no check-in endpoint at all — a test asserts every mode returns `FORBIDDEN` and no log row is written.
5. Staff modes require the caller to satisfy `Can(actor, "event.checkin", "event:<code>")` (ADR 0001). Superadmin bypass applies.
6. **The window check branches on `Action`** (ADR 0004) — this is the delta the original plan gets wrong:

   | Action | Window |
   |---|---|
   | `checkin` | `[checkin_open_at, checkin_close_at]` |
   | `checkout` | the session period, up to `end_at` |

   The check-in window exists to stop late arrivals and has no bearing on someone already inside recording their departure. A test checks somebody out at a time after `checkin_close_at` and before `end_at` and asserts success.
7. Geo runs per mode via `geo.Validate(cfg, mode, coords, override)`. On a `require` failure, append a log with `outcome=rejected_geo` **and then** return the error — the rejection is logged, not swallowed.

### Duplicate check-in (ADR 0004)

8. A second check-in of an already-attended attendee returns a distinct **`ALREADY_CHECKED_IN` error class**, not `nil`. The organizer wants the scanner to show a failure state, not a success.
9. The `attendance_logs` row with `outcome=already_checked_in` is **still appended**, preserving the invariant.
10. A report counting `outcome='checked_in'` is unaffected by any number of duplicates. A test asserts the count after three duplicate scans is 1.
11. Scanner-facing note for the API consumer: render this distinctly from geo and forgery rejections, so staff can tell a harmless duplicate from a real problem at a glance.

### Walk-in creation — the QR registration path

12. `personal_qr` or `session_qr` presented by someone with no confirmed registration for the session creates one (`source=walk_in`) **and** checks them in, in one step — gated by the mode being enabled and seats being available.
13. **Creation goes through `registration.BookAndCreate`**, inside the same `Atomic` transaction, so capacity is one shared pool and eligibility cannot be skipped.
14. **Eligibility runs** (ADR 0007). This closes the most serious gap in the current design: today a non-volunteer can present a personal QR at a volunteers-only event, or self-scan the poster, and be admitted with **no staff involvement at all**. Refusal is `NOT_ELIGIBLE` with a log row `outcome=rejected_not_eligible`.
15. Anonymous walk-in creation is refused on restricted events (`ANONYMOUS_NOT_ALLOWED`) and permitted on open ones per the event's `desk_registration` setting.
16. **On a session with ticket categories, the category must be supplied** (ADR 0009). Absent one, creation is refused with `CATEGORY_REQUIRED` and a clear message; the log records `outcome=rejected_no_category`.
    - Never guessed. Each category is a separately-sized quota, so defaulting to the largest pool puts a family with a small child in the main hall, and defaulting to first-available silently consumes one of fifty reserved Family Room seats for someone who did not need it.
    - An usher scanning a personal QR is already speaking to the person, so "main hall or family room?" costs one question; a self-scanner picks from a list before check-in completes.
    - Sessions with no categories are unaffected and prompt for nothing.
17. Full and finite → `QUOTA_FULL`, hard reject.

### Check-in by attendee lookup — `manual` with an existing attendee (ADR 0004)

18. `manual` accepts an **optional existing attendee id** alongside its typed-form input. Given one, it checks that attendee in **without creating a registration and without booking a seat**.
19. The log's `mode` stays `manual` — it was done by hand at the desk either way — and the created-versus-existing distinction remains derivable from `registrations.source`. **Do not add a fifth mode value for this**; the reporting need it would serve is met by `reporting`'s source split.
20. This is the path for anyone whose ticket never arrived: an anonymous lobby-table registrant with no account and possibly no email, a dead phone, a lost paper ticket. A test asserts that checking such a person in creates **no** second registration and consumes **no** second seat — the trap being avoided is double-counting them.

### Check-out (ADR 0004)

21. Check-out is an **explicit action** (`action: checkout`), surfaced by `qr/resolve`'s per-caller `allowed_actions`. It is **never** a toggle on a repeat scan.
    - The failure mode being avoided is specific to this congregation: someone unsure whether their self-scan worked scans again "to be safe" and silently checks themselves out, walking away believing the opposite of what the record says.
22. Requires `enable_checkout` (`FORBIDDEN` otherwise) and a prior check-in (`NOT_CHECKED_IN` otherwise).
23. Any enabled mode can perform it. Attendee status becomes `checked_out` with `checked_out_at` set.
24. Someone who never checks out simply stays `attended` with a null `checked_out_at` — indistinguishable from any other non-checkout, which is fine.
25. Guarded-pickup case: when a guardian's personal QR is scanned at collection, the log records who collected the child via `checked_by`.

### Headcount and undo (ADR 0005)

26. A headcount tap is a bare `+1` with **no identity captured at all** — a log row with null `attendee_id` and null `community_id`, `mode=manual`, `is_headcount=true`. Not a registration, no attendee row.
27. **Headcount keeps counting past the seat limit.** Its purpose is to find out when more people came than the venue was meant to hold, so a full session must not block a tap.
28. `UndoHeadcount` appends a **correcting entry** with `outcome=headcount_undone`. **The original tap row is never removed**; the total is the net.
29. Available to **any** staff member at that session while check-in is open — deliberately not restricted to the original tapper and not on a short timer. A tight window frustrates the realistic case (an usher notices the miscount a few minutes later) without providing the protection it appears to, since every original tap remains permanently visible for review regardless.
30. **One undo cancels exactly one tap** — correcting three mistakes takes three undos, so a single action can never wipe out a large count.
31. An undo when the net is already zero is refused with `NOTHING_TO_UNDO`.
32. `CountHeadcount` nets taps against undos. It **cannot** reuse `checked_in` for the undo outcome — that would corrupt the attended count.

### Lateness (ADR 0005)

33. Late means **`attended_at > start_at`**. Derived at read time, never stored, never clearable.
34. Deliberately defined against the session start, **not** `checkin_close_at`. It matches the everyday meaning of the word, needs no configuration, and works for sessions that set no check-in window at all. Defining it against `checkin_close_at` would overload that field into a grace-period setting, which is not what it means elsewhere.
35. **There is no "allow late check-in" toggle, and none is to be added.** How forgiving a session is falls out of `checkin_close_at` alone — set it at `start_at` for a strict door, at `end_at` to accept arrivals throughout, fifteen minutes past for a grace period.
36. **`enable_late_marking`** (per-session, default off) governs whether lateness is surfaced at all. When off, nothing anywhere mentions it — no badge on the door scanner, no field in the response, no report column. A test asserts `IsLate` is never populated and no lateness key appears in any response when off.
37. Session-level rather than event-level, so a single occurrence can be treated differently from its siblings — the week the meeting is held at a retreat — without disturbing the other fifty-one.
38. **A late mark cannot be cleared.** Arriving at 19:20 is a fact, not an opinion. A good reason belongs in a note alongside the attendance, not in erasing the record. Allowing an override would force lateness to become stored state with its own permission rules, turning a free derived value into a real feature.

### Eligibility re-check at check-in (ADR 0008)

39. An **event-level** setting alongside `eligibility` itself, three-way like `geo_config.modes`:

    | Value | Behaviour |
    |---|---|
    | `off` (default) | No re-check. Nothing stored, no column, no badge. |
    | `note` | Admitted, but recorded as no longer qualifying. |
    | `block` | Turned away with `ELIGIBILITY_REVOKED`. |

40. **The result is evaluated and stored at check-in time** in `attendance_logs.eligibility_result`. This differs from lateness, which is recomputed from two immutable timestamps: eligibility is judged against account state that keeps changing, so a report run in March asking "did she qualify?" would answer for March, not for the December night in question. The answer that matters is the one that was true when the person walked in.
41. **Existing registrations are never retroactively invalidated.** This setting affects only what happens at the door.
42. Someone who never registered at all and scans in still gets a **full** eligibility check regardless of this setting, because that scan *creates* a registration (criterion 14).
43. Placement is on the event, not the session — setting *who qualifies* in one place and *how strictly we hold them to it* in another would be split-brained, and the second is meaningless without the first.
44. API-consumer note: present `block` as the **unusual** choice, with `note` recommended. `block` means an usher turning somebody away in a queue with no ability to judge or fix the situation; it exists for genuinely sensitive gatherings — a confidential meeting an ex-member should not sit in on — not as a default posture.

---

## Verification

```bash
TEST_DATABASE_URL=... go test ./tests/integration/v3/ -run TestCheckin -race -v
```

One test per behaviour, all asserting observable outcomes — the log row's `outcome`, the seat count, the error class:

registration_qr happy path · duplicate scan returns the error class *and* appends the log *and* leaves `checked_in` count at 1 · personal_qr for a registered member · walk-in auto-creation increments `booked_seats` · walk-in refused on a restricted event with `rejected_not_eligible` logged · walk-in refused on a categorised session with `rejected_no_category` logged · lookup check-in creates no registration and no seat · checkout after `checkin_close_at` succeeds · checkout without check-in → `NOT_CHECKED_IN` · checkout with `enable_checkout: false` → `FORBIDDEN` · geo `require` failure errors *and* logs `rejected_geo` · headcount tap then undo nets zero with both rows present · second undo → `NOTHING_TO_UNDO` · headcount past a full session still records · lateness derived when `enable_late_marking`, absent when off · re-check `note` admits and stores, `block` refuses, `off` stores nothing · `attendance_modes: []` refuses every mode.

---

## Boundaries (module-specific)

- **Always** append exactly one log row per attempt, including rejections.
- **Always** create registrations through `registration.BookAndCreate` — never write a registration row directly from this module.
- **Never** update or delete an `attendance_logs` row.
- **Never** add a fifth `attendance_modes` value, including for headcount or lookup check-in.
- **Never** auto-toggle check-out on a repeat scan.
- **Never** store lateness or allow it to be cleared.
- **Never** guess a ticket category.
- **Never** retroactively invalidate a registration when rules change.
- **Ask first** before adding an `outcome` value.

---

## Open Question (blocking)

**Headcount undo linkage.** This spec nets by counting outcomes (`checked_in` minus `headcount_undone` among `is_headcount` rows) and refuses an undo below zero. A `undoes_log_id` self-reference would be stricter but needs a column added in `foundation`, not here. See parent Open Question 2.
