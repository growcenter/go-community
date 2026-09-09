---
status: accepted
---

# QR-triggered walk-in registration is existing `attendance_modes` behavior, not a new path

A grilling session (see chat log, 2026-09-09) initially proposed a new "QR Registration" path alongside online self-registration and Desk Registration, for scanning `personal_qr`/`session_qr` to auto-register + check in in one step. On reconciling against `docs/plans/2026-07-07-event-management-v3-design.md` (§5.1/§5.2), this turned out to already be exactly the existing behavior: enabling `personal_qr` in a Session's `attendance_modes[]` alone permits walk-in registration creation on scan (gated by seats available/unlimited), and the same applies to `session_qr` for self-scan. No new field or mechanism is needed — "QR Registration" is a descriptive label for this existing behavior, not a new path.

This session also confirmed/clarified points not spelled out in the base design:

- QR-triggered walk-in registration is **account-required only** — no anonymous variant, unlike Desk Registration (ADR 0002).
- **Session QR belongs to exactly one Session**, not shared across a recurring Event's Sessions — consistent with check-in windows and capacity already being per-Session, not per-Event.
- Session QR is a **static, non-rotating poster** — accepted as a trust boundary for this (church congregation) audience; no anti-replay/geofencing is planned. If abuse becomes real, disable `session_qr` for that Session in favor of `personal_qr`/`registration_qr` only.
- **Personal QR is generated once at account registration, permanent, non-rotatable** — no self-service or staff-initiated regeneration exists or is planned, by deliberate choice (this church's user base skews older/less technical; a leaked-QR risk was accepted as a trade for zero user-facing complexity).
- If `session_qr` is disabled for a Session and someone has no other valid method, they're handled out-of-band by an usher via Desk Registration — no automated fallback is built for this case.

See `CONTEXT.md`'s "Open decisions" section for related conflicts this session surfaced but did **not** resolve (checkout mechanism, Desk-Registration register-only mode, cross-Session-of-one-Event restriction) — those remain open, unimplemented, and un-adopted pending a follow-up decision.
