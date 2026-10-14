# Capability Map: v3 Platform

Approved 2026-09-17. Module ids are stable and never renamed. Each module has its own spec at the project root named `SPEC-<module id>.md`. The confirmed intent is in `docs/intent/v3-platform.md`.

| Module id | Responsibility | Depends on |
|---|---|---|
| `platform` | herr error contract, global code catalog with EN/ID messages, `{data, meta}` envelope, 410 deprecation of v1/v2, configdb tunables, retention cleanup, email outbox with a channel field | none |
| `identity` | Users, community ID, OTP + PIN + trusted devices, user carry-over from v1/v2, user audit log | platform |
| `permissions` | Scoped `permission_grants`, five scope kinds, role titles, grant audit, replacement of `user_types` | identity |
| `qr` | One token package for personal, session and meeting QR, resolve and act, static personal QR from community ID | identity |
| `events` | Events, sessions, recurrence, combinable eligibility, per-occurrence registration, desk and self-scan check-in with three protection levels, Sunday service block, settings audit, attendance matrix reports and exports | permissions, qr |
| `cool` | COOLs, members, Facilitator/Leader/Core team roles, meetings (offline/online/hybrid), meeting QR, pending-until-all-confirmed attendance, guests | permissions, qr |
| `calendar` | Read-only aggregation of sessions and meetings, `.ics`, reconciled to v3 table names and `permission_grants` | events, cool |

Build order: platform, then identity, then permissions, then qr, then events and cool in parallel, then calendar.

Notes:

- Volunteer is not a module. It is an event category plus eligibility rules inside `events`.
- Custom registration forms belong to `events` (delta 17 in `SPEC-events.md`). The engine in `internal/pkg/form` imports only the standard library. `cool` or a future `forms` module (standalone forms, quizzes) can reuse it without a change to `events`.
- The existing event v3 design, implementation plan, check-in spec and ADRs 0001 to 0009 remain the spec for `events`. `SPEC-events.md` is a reconciliation delta on top of them, not a rewrite.
- `calendar` gets a fix-up delta against `docs/plans/2026-07-20-community-calendar-design.md`.
- `platform`, `identity`, `permissions`, `qr` and `cool` get fresh specs.
