---
status: accepted
---

# Eligibility rules run on every registration path, and restricted events reject anonymous registrations

Continuation of the 2026-09-09 design session (see ADR 0004–0006). This closes three ways an event's `eligibility` rules can currently be bypassed.

## The gap

`eligibility` (design §2.2) restricts who may register — by role, user type, campus, or explicit community IDs. It is enforced on the online self-registration path (§4). It is **not** enforced anywhere else:

1. **QR walk-in auto-creation** — the most serious, because it needs no staff involvement. Design §5.2's walk-in branch checks only two conditions before creating a registration: is the mode enabled, and are seats available. The implementation plan's Task 11 restates the same two. So on a volunteers-only event, a non-volunteer can present their personal QR at the door, or self-scan the session poster, and the system creates a registration and checks them in — eligibility never runs, because the online path that enforces it was skipped entirely.
2. **Anonymous desk registration** at the lobby table — there is no account, so there is nothing for the rules to evaluate against.
3. **Anonymous walk-in registration** at the door via `manual` — the same hole, later in the process.

## The decision

**Eligibility runs wherever a registration is created** — online self-registration, staff advance registration (`source=admin`), staff walk-in creation at the door, and QR-triggered walk-in auto-creation. A person who does not qualify is rejected with a message naming the reason, not a generic failure. **Delta:** Task 11's walk-in branch gains an eligibility check alongside its existing mode and seat checks; the desk/advance registration endpoint (ADR 0004) carries the same check.

**Restricted events reject anonymous registrations entirely.** When `eligibility.audience` is anything other than `everyone`, staff must identify the person from their account rather than taking a name — at the lobby table and at the door alike. This costs nothing real: if an event is restricted to people the system can identify, then everyone legitimately eligible has an account by definition, and a person with no account is by definition not on the list. Events open to everyone are unaffected and keep accepting anonymous registrations as before, per ADR 0002.

Two alternatives were rejected. Relying on the staff member present as the check works until the day a coordinator asks how fifteen ineligible people reached the list, and the honest answer is that the rule they configured was never enforced at the desk — a bad surprise, and one that only surfaces after the fact. Flagging such registrations as unverified for later review was rejected because flagged queues rot: nobody reviews them, and the restriction is still bypassed in the meantime.

Consistency is the point — closing the online path while leaving the desk and QR paths open would not restrict anything, it would just move where people get in.
