---
status: accepted
---

# Check-out is an explicit action, desk registration splits into two staff actions, and cross-session limits skip anonymous registrations

A design-grilling session (2026-09-09) produced conclusions that conflicted with `docs/plans/2026-07-07-event-management-v3-design.md` and its implementation plan. This ADR records how each conflict was resolved, and the resulting deltas to the implementation plan.

## Check-out stays an explicit action, but its window differs from check-in

The grilling session initially concluded that any repeat scan should silently toggle a person between checked-in and checked-out. We rejected that in favour of the existing design (§5.3): the client sends an explicit `action: checkout`, surfaced by `qr/resolve`'s per-caller `allowed_actions`. Auto-toggling has a bad failure mode with this congregation — someone unsure whether their self-scan worked scans again "to be safe" and silently checks themselves out, walking away believing the opposite of what the record says. The resolve step already computes valid actions per state, so the explicit form costs nothing extra.

Two refinements to existing behaviour follow:

**Check-out is bounded by the session's `end_at`, not `checkin_close_at`.** The implementation plan's Task 11 applies the check-in window as a shared gate across one code path switching on `Action`; as written, someone leaving at 10:15 could not check out of a service whose check-in closed at 09:30. The check-in window exists to stop late arrivals and has no bearing on someone already inside recording their departure. **Delta:** Task 11's window check must branch on `Action` — `checkin` uses `[checkin_open_at, checkin_close_at]`, `checkout` uses the session period up to `end_at`. Someone who never checks out simply stays `attended` with a null `checked_out_at`, indistinguishable from any other non-checkout, which is fine.

**A duplicate check-in returns an error to the scanner, while still appending its log row.** Task 11 currently specifies `already_checked_in` as a non-error (`nil error`). The organizer wants the scanner to treat it as a failure state showing "already checked in," not a success. **Delta:** Task 11 returns a distinct `ALREADY_CHECKED_IN` error class rather than `nil`; the `attendance_logs` row with `outcome=already_checked_in` is still appended, unchanged. This preserves the design's "every path appends exactly one log" invariant — the log records that a scan was *attempted*, not that attendance occurred, and reports counting `outcome='checked_in'` are unaffected. The scanner UI should render this distinctly from geo/forgery rejections so staff can tell a harmless duplicate from a real problem at a glance.

## `desk_registration_outcome` is dropped; the two desk scenarios are separate staff actions

The session proposed a per-session `desk_registration_outcome: qr_only | auto_checkin | flexible`. We dropped it. It was trying to express as configuration something that is simply *which action a staff member took*, and the `registrations.source` enum already distinguishes the two cases:

- **`source=admin` — the lobby table.** Staff sign someone up for a session happening later (e.g. a Christmas sign-up table after Sunday service); the person checks in weeks afterwards. This is a real need at this church and **has no endpoint today** — `admin` exists in the enum but nothing writes it. **Delta:** a new staff "register on behalf of someone" endpoint is required. It is governed by the *same* `desk_registration` setting as the door (ADR 0002) — one setting covers both, since both are staff creating a registration for a person in front of them.
- **`source=walk_in` — the door.** Staff register and check someone in in one action, already fully covered by `manual` mode.

No organizer-facing toggle is needed for either; a toggle could only ever forbid staff from doing something reasonable.

**`manual` mode gains the ability to check in an already-registered attendee.** Today `CheckinInput.AttendeeID` is only populated from a scanned `registration_qr`, and `manual` always *creates* a walk-in — so a pre-registered person arriving without a working QR cannot be checked in at all. That is not an edge case: an anonymous lobby-table registrant has no account and possibly no email, so no ticket ever reached them; dead phones and lost paper tickets do the same. Creating a walk-in for them would double-count the person and burn a second seat. **Delta:** `manual` accepts an optional existing `AttendeeID` (staff find the person via the existing attendee-list search) and checks that attendee in without creating a registration or booking a seat. The log's `mode` stays `manual` — it was done by hand at the desk either way — and the created-versus-existing distinction remains derivable from `registrations.source`. We deliberately did **not** add a separate mode value for lookup-based check-in; the reporting need it would serve is already met by the source split below.

**Reports split advance registrations by source.** The Summary sheet currently reports walk-ins (`source=walk_in`) against an undifferentiated "registered" total, which would hide how many sign-ups the lobby table actually produced. **Delta:** Task 15's Summary sheet breaks registrations into three counts — `web` (online self-registration), `admin` (staff-registered in advance), and `walk_in` (registered at the door).

## Cross-session limit is `max_sessions_per_person`, and skips anonymous registrations

A new event-level setting limits how many Sessions of one Event a person may register for (the "pick the 8am or the 10am service, not both" case). It lives in `registration_config` as **`max_sessions_per_person`**: `0` = unlimited (default, and correct for a recurring weekly service where every week is a Session of one Event), `1` = one session only, `N` = up to N. The name and numeric shape deliberately mirror the neighbouring `max_per_registration`, and avoid confusion with the existing `allow_multiple_registrations`, which governs the different question of registering twice for the *same* session. Exceeding the limit is a hard reject naming the session already registered for — never a silent auto-cancel-and-switch, which would be a surprising side effect of pressing register.

**Anonymous registrations are exempt**, consistent with ADR 0002's exemption of them from the same-session duplicate guard. Enforcing it for them was considered and rejected: with no account, the only identifiers are a name and possibly a phone, and name matching fails in both directions (two real "Budi Santoso"s get one wrongly turned away; one person entering "Budi" then "Budi S." passes twice). Making it work would require forcing phone to be mandatory for anonymous registrations at such events — unacceptable for a congregation where some visitors have no phone to give. Since these registrations only ever happen face-to-face, the staff member present is a better and more forgiving check than fuzzy matching, and the cost of a false block (turning away a genuine first-time visitor at a Christmas service) far exceeds the cost of a miss.
