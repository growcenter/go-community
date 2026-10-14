# Task List: v3 Platform

Ordered by dependency. Full task text with acceptance criteria is in `tasks/plan.md` under the task id.

## Phase 1: platform foundation

- [x] P1: Go 1.26 bump and herr dependency compile
- [x] P2: herrx registry with COMMON_ classes and completeness test
- [ ] P3: response.OK/Fail envelope with tests
- [ ] P5: configdb typed getters and boot-time defaults
- [ ] P4: deprecation middleware returns 410 on v1 and v2
- [ ] P6: outbox table, repository and Notifier interface
- [ ] P7: retention purge and scheduler endpoint
- [ ] P8: dispatch endpoint and email Notifier

### Checkpoint: Phase 1
- [ ] `go build ./... && go test ./internal/pkg/... ./internal/platform/...` pass
- [ ] Every v1 and v2 route returns 410 and `/health` returns 200
- [ ] Outbox concurrency test sends each row once
- [ ] Review with the owner before the next phase

## Phase 2: identity core (sign-in)

- [ ] I1: Migration 000023 with account status backfill
- [ ] I2: `USER_` error catalog and locale
- [ ] I3: Models, repositories and audit writer
- [ ] I4: OTP package with limits and cooldown
- [ ] I12: v3 JWT issuer and middleware
- [ ] I5: User check, OTP send and verification, create user with claim
- [ ] I6: Login, new device verification, device limit
- [ ] I7: Refresh and logout with device revoke
- [ ] I8: PIN reset
- [ ] I9: Me endpoints and the email-required gate

### Checkpoint: Phase 2
- [ ] Register, login, new device, refresh and logout pass integration tests
- [ ] `community_id` unchanged for every backfilled row
- [ ] Three sends in two hours trigger the cooldown
- [ ] Review with the owner before the next phase

## Phase 3: permissions core

- [ ] M1: Grants table migration and `PERM_` catalog
- [ ] M2: Action registry, title table and matrix test
- [ ] M3: Evaluator `Can` and `Scopes` with parents
- [ ] M4: `Require` middleware, resolvers and `Actor`

### Checkpoint: Phase 3
- [ ] Matrix test covers every title times every action
- [ ] `campus_admin` of BSD passes `Can` on an event with parent BSD and fails on PIK
- [ ] `permission.Require` returns 403 `PERM_FORBIDDEN` on a miss
- [ ] Review with the owner before the next phase

## Phase 4: identity admin and permissions endpoints

- [ ] I10: Admin create-pending, status change, admin reset
- [ ] I11: Admin list, detail, patch, delete
- [ ] I13: Dashboard and audit-log endpoints
- [ ] I14: OTP on/off switch
- [ ] M5: Grant and revoke endpoints with audit
- [ ] M6: List, me, check and titles endpoints
- [ ] M7: Backfill migration from frozen tables
- [ ] M8: `EnsureSeeds` and boot wiring in `contract.go`
- [ ] I15: Wire v3 identity into the composition root and configdb defaults

### Checkpoint: Phase 4
- [ ] Admin endpoints are guarded by `permission.Require`, no allow-list exists
- [ ] Grant, revoke and backfill integration tests pass
- [ ] `grep -rn 'UserMiddleware(' internal/deliveries/http/v3` returns nothing
- [ ] Review with the owner before the next phase

## Phase 5: qr

- [ ] Q1: Token codec with kid map and config validation
- [ ] Q2: Action registry with Allowed and Dispatch
- [ ] Q3: QR_ error catalog and locale
- [ ] Q4: Scan log table, repository and purge hook
- [ ] Q5: Resolve endpoint
- [ ] Q6: Act endpoint
- [ ] Q7: Personal QR PNG endpoint (needs dependency approval)
- [ ] Q8: Wire into the composition root

### Checkpoint: Phase 5
- [ ] Personal token is byte-identical across calls and verifies after key rotation
- [ ] `act` on a not-allowed action returns 403 and never runs the handler
- [ ] Every URL is under 160 characters in the length test
- [ ] Review with the owner before the next phase

## Phase 6a: events (can run in parallel with 6b)

- [ ] E1: Migration 000025 and EVENT_ catalog
- [ ] E2: JSONB configs with category, conditions and self_scan_protection
- [ ] E3: Eligibility package with AND conditions and reader interfaces
- [ ] E4: Geo, form and recurrence pure packages
- [ ] E5: Entities, repositories and audit writer
- [ ] E6: Event and session CRUD with permission guards and audit
- [ ] E7: Publish gate and templates without Sunday Service
- [ ] E8: Recurrence generation per occurrence
- [ ] E9: Registration with seat lock, eligibility and idempotent re-registration
- [ ] E10: Registration channels and desk registration
- [ ] E11: Attendee tokens and PDF ticket for windowed events
- [ ] E12: Check-in core with personal_qr and session_qr
- [ ] E13: registration_qr, manual lookup and eligibility recheck
- [ ] E14: Checkout, lateness, headcount and undo
- [ ] E21: Scan labels on check-in
- [ ] E22: Roster sync endpoint
- [ ] E23: Offline batch check-in sync
- [ ] E15: QR action registration for events
- [ ] E16: Attendee list, live counts and computed session state
- [ ] E17: Reports CSV, XLSX and JSON
- [ ] E18: Attendance matrix report
- [ ] E24: Label counts in reports
- [ ] E19: Staff dashboard, audit endpoint and permission scopes
- [ ] E20: Pool caps, wiring, seeds and v2 410

### Checkpoint: Phase 6a
- [ ] 200 concurrent registrations for one seat leave `booked_seats = 1`
- [ ] Poster scan from home with `window_geo` returns `EVENT_GEO_OUT_OF_RANGE`
- [ ] Matrix report has one column per occurrence in CSV, XLSX and JSON
- [ ] 50 queued offline check-ins sync in one batch with `offline = true`, and two desks never double-count a person
- [ ] Review with the owner before the next phase

## Phase 6b: cool (can run in parallel with 6a)

- [ ] C1: Migration and COOL_ catalog
- [ ] C2: COOL entities and DTOs
- [ ] C3: COOL create, edit, list and detail
- [ ] C4: Members add and remove
- [ ] C5: Titles and facilitators endpoints
- [ ] C6: Meeting create, edit and cancel
- [ ] C7: Meeting QR image and scan handler
- [ ] C8: Mark attendance and close
- [ ] C9: Edit attendance after close with revisions
- [ ] C10: Guests add and link
- [ ] C11: Member "mine" endpoints
- [ ] C12: Stats package and reporting endpoints
- [ ] C13: Wiring and registration

### Checkpoint: Phase 6b
- [ ] Close with one `pending` row returns `COOL_MEETING_PENDING`
- [ ] A scan never changes `status`, only `scanned_at`
- [ ] Removing a member revokes their COOL titles in the same transaction
- [ ] Review with the owner before the next phase

## Phase 7: calendar

- [ ] L1: Pure calendar package: DTOs, query validation, status mapping, `CAL_` catalog
- [ ] L2: `.ics` serializer with category line
- [ ] L3: Sessions query on v3 tables with organizer refs and parents
- [ ] L4: Meetings query on cool tables
- [ ] L5: Usecase tiering, eligibility, category filter, merge
- [ ] L6: JSON and `.ics` endpoints, mount, docs

### Checkpoint: Phase 7
- [ ] A plain member sees only eligible published sessions plus their own COOL meetings
- [ ] A cancelled meeting is absent from JSON and `.ics`
- [ ] `grep -rn 'event_organizers\|cool_facilitators\|ErrorMapping' internal/**/calendar` returns nothing
- [ ] Review with the owner before the next phase

