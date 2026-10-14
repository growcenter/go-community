# Intent: v3 Church Platform Backend

Confirmed with the product owner on 2026-09-17. This is the shared statement of what v3 is, before any capability map or spec.

## Outcome

All six features ship on `/api/v3`. v3 is the last API version. Breaking changes after launch are handled by additive changes inside v3, never by a v4. v1 and v2 stay compiled, and every old endpoint returns HTTP 410 with the code `DEPRECATED_ENDPOINT`.

The existing event v3 design, implementation plan, check-in spec and ADRs 0001 to 0009 are kept and reconciled. User management and COOL are re-specified for v3 instead of "v2 in place". New specs cover the unified QR system, the herr error contract, deprecation, calendar fixes, and one scoped permission model.

## Users

One first-party web app first, a mobile app next. Members range from children to people in their 60s and 70s.

## Why now

The six features are half on paper and half inconsistent: the calendar design uses different table names from the events plan, two features target v2 while the rest target v3, and "volunteer" and "QR" have no spec of their own. There is no v4 to fix this later.

## Decisions

### Permissions

- Scoped `permission_grants` replace `user_types` and the `user_types` to `roles` mapping table.
- Five scope kinds: global, campus, department, COOL, event.
- Admins assign a ready-made title (for example "COOL Leader") plus a scope. One person can hold many titles. Raw per-action grants are the exception, not the normal path.

### Events and volunteers

- A volunteer event is an event with a `volunteer` category and eligibility rules. No volunteer hours, no approval step, no separate tables.
- Eligibility combines conditions. All listed conditions must be true. One condition can list several values, for example two campuses or one COOL.
- Every occurrence of a recurring event needs its own registration. There is no "register once for all occurrences".
- Sunday service attendance is out of scope and must be blocked.
- Two check-in modes are always available on every event: desk scans the member, and member scans a poster.
- Self-scan protection is a per-event setting with three levels: none, check-in window (default), and check-in window plus location.
- Admin changes to sessions and settings (time, capacity, protection level) are audit-logged with who, what and when.
- Desk check-in works with no network. The desk device syncs the session roster while online, downloading only what changed. Offline it checks scans against that roster and queues them. The server accepts the queue in a batch, keeps the first check-in, adds up headcount taps from several desks, and flags every such row as offline with the device scan time. Clock tolerance is two minutes. Member self-scan of the poster stays online only. What happens to a person who is not on the roster is a per-event setting: let in and review at sync (default), or turn away.
- An organizer can give each check-in method a free-text label for the event, for example personal QR means "on site" and poster QR means "online". The label is optional, is saved on each check-in, is counted in reports, is carried by templates and recurring events, and never changes the attendance percentage.

### COOL

- Meetings can be offline, online or hybrid.
- One meeting QR, which the leader can print as a poster, show on their phone, or share on a video call screen. No separate link.
- A scan only pre-fills "present". The leader decides.
- A meeting stays pending until the leader marks every member as one of: confirmed present, absent with notice, absent without notice.
- Roles in order of strength: Facilitator, Leader, Core team. A facilitator can oversee several COOLs and still hold a lower role in their own.
- Guests are typed in by the leader with a name and an optional phone or other identifier. A guest can register an account by themselves, or a leader or core team member can create it for them.

### QR

- Personal QR is static and derived from the community ID.
- Session QR, meeting QR and registration QR tokens come from one shared QR package.
- A registration QR exists only for events with a registration window. Each attendee row gets its own code, so a family of four gets four codes. It can be printed. The attendee shows it and desk staff scan it. It is never scanned by the attendee.

### Errors and responses

- Errors follow herr, plus a top-level `message` on validation errors: `{ "code", "message", "errors": [ { "field", "code", "message" } ] }`. `errors` appears only for validation failures.
- Success is `{ "data", "meta" }`. A success body never has an `errors` key.
- Error codes are globally unique and module-prefixed (`EVENT_QUOTA_FULL`, `COOL_MEETING_PENDING`), in one catalog with an EN and ID message each, guarded by a completeness test.

### Reports

- Recurring event attendance: one row per person, one column per occurrence (present, late, absent), totals and percentage.
- Export as CSV, XLSX and clean JSON with stable field names.

### Platform

- Notifications: email only. The sender has a channel field so SMS and WhatsApp can be added later.
- Sign-in: OTP plus PIN plus trusted devices, unchanged from the user management design. Every account needs an email, because email is the only OTP channel. An account without an email must add one before it can use member features. No device is ever trusted without an OTP or an admin-assisted code. OTP sends are rate limited per identifier and per client IP.
- Short-lived data (OTP challenges, outbox rows, scan logs) is deleted after a configurable number of days. Attendance records are kept.
- One church, several campuses, one database, one user list. A member can attend any campus.
- Free-tier hosting (GCP free, Supabase free Postgres, possibly Cloudflare free). 200 to 350 users on a normal day, 1,000 to 3,000 on event days. Seat races at registration open and check-in bursts at the door are both load cases.

### Data carry-over

- Users carry over into v3.
- Old events and attendance stay in the old tables as history.
- COOL membership starts empty.

## Out of scope

- Sunday service attendance
- Volunteer hours, approvals, or a volunteer HR module
- Rotating personal QR
- Migrating old events into v3 tables
- Multi-church
- AI features inside the backend
- The move to go-echo-boilerplate and wlog
