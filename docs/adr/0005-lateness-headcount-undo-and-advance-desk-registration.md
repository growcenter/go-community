---
status: accepted
---

# Lateness is derived, headcount taps are undoable, and post-deadline desk registration is organizer-only

Continuation of the 2026-09-09 design session (see ADR 0004). These decisions extend check-in beyond the large-event case toward small recurring gatherings — a monthly volunteer meeting, where the question is not gatekeeping but who turned up on time.

## Lateness is derived from arrival time, never stored

An attendee is late when their `attended_at` is after the session's `start_at`. This is computed at read time, not recorded at check-in. Storing a flag would freeze a judgment that costs nothing to recompute, and would prevent changing the definition later without losing data.

"Late" deliberately means *after the session started*, not *after check-in closed* — it matches the everyday meaning of the word, needs no configuration, and works for sessions that set no check-in window at all. Defining it against `checkin_close_at` was rejected because it would overload that field into a grace-period setting, which is not what it means elsewhere.

**No "allow late check-in" toggle was added.** How forgiving a session is falls out of `checkin_close_at` alone: close it at `start_at` for a strict door, at `end_at` to accept arrivals throughout, or fifteen minutes past the start for a grace period. A separate toggle could only ever duplicate or contradict what that timestamp already says.

**A late mark cannot be cleared.** Arriving at 19:20 is a fact, not an opinion; if there is a good reason (the volunteer was parking cars), that context belongs in a note alongside the attendance, not in erasing the record. Allowing an override would also force lateness to become stored state — plus who granted the exception and why, plus rules about who may — turning a free derived value into a real feature. Keeping the fact immutable also keeps the number interpretable months later: "arrived after we started," not "arrived late and nobody vouched for them."

**Whether lateness is surfaced at all is a per-session toggle** (proposed `enable_late_marking`, default off), sitting alongside `attendance_modes` and `enable_checkout` rather than on the event. When off, nothing anywhere mentions lateness — no badge on the door scanner, no report column — consistent with design §2.3's rule that unconfigured capabilities stay invisible, and with how `enable_checkout` already behaves. Session-level rather than event-level for two reasons: it belongs next to the other two attendance-mechanics settings, and it allows suppressing lateness for a single occurrence (the week the meeting is held at a retreat) without disturbing the other fifty-one. Recurrence `session_defaults` means this is still configured once for a recurring event.

## Headcount taps are undoable while check-in is open

A mis-tap on the anonymous `+1` counter would otherwise inflate attendance permanently, since `attendance_logs` is append-only — and double-taps are likely from an unsure usher in a busy doorway. Staff may undo a tap by appending a **correcting entry**; the original tap row is never removed, and the headcount total is the net of taps and undos.

Undo is available to any staff member at that session while check-in is open — deliberately not restricted to the original tapper or a short timer. A tight window was considered and rejected: it frustrates the realistic case (an usher notices the miscount a few minutes later) without providing the protection it appears to, since every original tap remains permanently visible for review regardless. One undo cancels exactly one tap, so correcting three mistakes takes three undos — preventing a single action from wiping out a large count.

## Staff may register people past the deadline, but only organizers

Following the same "allow the action, record what happened" principle as lateness: staff can create advance registrations (`source=admin`, the lobby-table endpoint from ADR 0004) after `register_end_at` has passed, up until check-in closes. Blocking them would not prevent the outcome — staff would simply tell the person to arrive on the day and be registered as a walk-in, which admits them anyway while giving the organizer *less* warning and filing them under the wrong source. Seat limits still apply, so this cannot overfill a venue.

**The two acts are permission-split**, because they carry different weight:

- **While registration is open** — any staff member who can run check-in (`event.checkin`). The lobby table is typically staffed by volunteers, and they are only doing for someone what that person could have done themselves online.
- **After `register_end_at`** — event organizers only (`event.manage`). This overrides a cutoff the organizer deliberately set and silently changes the catering and seating numbers they believed were final, so it should be their call.

Permissions are evaluated through `permission_grants`/`Can(actor, action, resource)` per ADR 0001.
