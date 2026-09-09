---
status: accepted
---

# Event v3 authorizes staff via `permission_grants`, not a standalone `event_organizers` table

The Event Management v3 design (`docs/plans/2026-07-07-event-management-v3-design.md`, §2.1/§9) specifies a standalone `event_organizers` table for per-event staff (`event_id`, `community_id`, `role: owner|staff`). The COOL & Unified Permissions design (`docs/plans/2026-07-07-cool-and-unified-permissions-design.md`, §5.2), dated and approved the same day, explicitly amends this: it removes `event_organizers` and replaces per-event staff with `permission_grants` rows (`subject=user, action=event.checkin|event.manage, resource=event:<code>`).

We decided to build Event v3's authorization directly against `permission_grants`/`Can(actor, action, resource)` from the start, and never build `event_organizers` at all — rather than building the table as originally specced and migrating it away later. Both docs represent one approved program, not independently-evolving specs, so there's no value in shipping the throwaway table first.

Consequence: Event v3 cannot be implemented in full isolation — it needs the `permission_grants` schema and `Can()` evaluator (COOL spec §2.1) to exist first, or be built alongside it. The `POST /v3/events/{code}/organizers` endpoint in the Event v3 API surface (§9) is replaced by the generic grants API scoped to `resource=event:<code>`.
