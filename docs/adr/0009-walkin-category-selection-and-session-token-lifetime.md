---
status: accepted
---

# Walk-in creation asks for a ticket category, and the session poster QR outlives check-in when checkout is on

Continuation of the 2026-09-09 design session (see ADR 0004–0008). Two gaps in walk-in and checkout mechanics.

## Walk-in auto-creation must be told which ticket category

A session may have multiple quota pools (`session_ticket_categories` — Main Hall, Family Room, Balcony), each with its own seat count. A registration targets exactly one. But the walk-in auto-creation path (design §5.2, Task 11) books a seat without any category input — it is only defined for the single-pool case, and its behaviour on a categorised session is currently undefined.

Guessing is not acceptable: each category is a separate, deliberately-sized quota. Defaulting to the largest pool puts a family with a small child in the main hall; defaulting to first-available silently consumes one of fifty reserved Family Room seats for someone who did not need it.

**The scanner supplies the category, and creation is refused with a clear message when it does not.** An usher scanning a personal QR is already speaking to the person, so "main hall or family room?" costs one question; a self-scanner picks from a list before the check-in completes. Sessions with no categories are unaffected — nothing is asked, and the session's single pool is used as today.

Disabling walk-in entirely on categorised sessions was considered and rejected: it would mean on-the-spot registration fails precisely at the largest and most complex events, which is backwards.

## The session poster token stays valid until `end_at` when `enable_checkout` is true

Design §6.1 expires the `session` QR token at `checkin_close_at`, which was correct when its only job was check-in. ADR 0005 extended checkout to run until the session's `end_at`, and design §5.3 allows any enabled mode to perform it — so a self-scanner leaving after check-in closed would find the poster dead and be unable to check out. A volunteer meeting whose check-in closes at 20:00 and ends at 21:00 has a full hour where checkout via the poster is impossible.

**When `enable_checkout` is true, the session token expires at `end_at` instead of `checkin_close_at`.** When checkout is off, the existing `checkin_close_at` expiry is unchanged.

Tying this to the checkout setting rather than extending every session token preserves what the short expiry is for — limiting how long a photographed poster stays useful. Most sessions have no checkout and no use for the poster once the doors close, so extending theirs would weaken that for nothing. For sessions where it is extended, the exposure is narrow: a late scan can only produce a checkout, and only for someone already checked in.
