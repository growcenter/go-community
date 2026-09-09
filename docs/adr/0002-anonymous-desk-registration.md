---
status: accepted
---

# Allow anonymous registrations, desk-only, via a per-event `desk_registration` setting

The Event v3 design (`docs/plans/2026-07-07-event-management-v3-design.md`, §2.1) documents `registrations.registered_by` as `"community_id, always set"` — every registration requires an existing account. In practice, church walk-ins are often first-time visitors with no account, and staff should be able to register them on the spot without forcing an account to exist first.

We decided to make `registered_by` nullable and add a per-event `desk_registration` config value: `account_required` (today's assumption), `anonymous_only`, or `anonymous_allowed` (both). This only applies to desk/staff-created registrations — online self-registration (`POST /v3/sessions/{code}/registrations`) always requires auth, unchanged. Anonymous registrations are exempt from the session's duplicate-guard (`one confirmed registration per session_id, registered_by`), the same way guest companions already are.

Anonymous registration **never** auto-creates a user account. If that person wants an account later, that's a separate, self-initiated, consent-driven signup — not linked automatically to the anonymous registration record.
