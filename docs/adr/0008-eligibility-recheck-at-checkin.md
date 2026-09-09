---
status: accepted
---

# Eligibility may be re-checked at check-in, with the result recorded at that moment

Continuation of the 2026-09-09 design session (see ADR 0004–0007). ADR 0007 established that eligibility runs wherever a registration is *created*. This covers the separate question of someone who registered legitimately and later stopped qualifying.

## The situation

A volunteer registers for a volunteers-only meeting in November. By December they have stepped down. They arrive at the door having planned to come for weeks. Nothing in the design currently re-evaluates eligibility at check-in, so they are admitted.

Silently admitting them loses information the organizer may want. Turning them away is a bad moment — a public rejection in a queue, for a change the person may not connect to their attendance, in front of an usher with no ability to judge or fix it.

## The decision

An **event-level** setting controls this, reusing the three-way shape already used by `geo_config.modes` so the two read consistently:

| Value | Behaviour at check-in |
|---|---|
| **off** (default) | No re-check. Existing behaviour, unchanged. |
| **note** | Admitted, but recorded as no longer qualifying. |
| **block** | Turned away. |

Default `off`, so events that do not care are unaffected in every respect — no door badge, no report column.

`block` was offered reluctantly and deliberately kept available: for a genuinely sensitive gathering (a confidential team meeting an ex-member should not sit in on), admitting-and-noting is not sufficient. For everything else, `note` is the right setting and the one to recommend in the admin UI.

**The result must be evaluated and stored at check-in time**, not derived later. This differs from lateness (ADR 0005), which is recomputed from two immutable timestamps. Eligibility is evaluated against account state that keeps changing, so a report run in March asking "did she qualify?" would answer for March, not for the December night in question. The answer that matters is the one that was true when the person walked in, so it is captured then.

**Placement is on the event, alongside `eligibility` itself**, not on the session with the other door settings (`attendance_modes`, `enable_checkout`, `enable_late_marking`). Setting *who qualifies* in one place and *how strictly we hold them to it* in another would be split-brained, and the second is meaningless without the first. This matches `geo_config`, where the rules and their strictness already sit together on the event. The cost — no per-session variation in strictness — is acceptable, since the rules being re-checked are identical across an event's sessions anyway.

## Not re-litigated

Registrations that predate a rule change are **not** retroactively invalidated, and this setting does not do so — it only affects what happens at the door. Someone who never registered at all and scans in still gets a full eligibility check regardless of this setting, because that scan *creates* a registration (ADR 0007).
