# Spec: cool

Module id `cool` in `CAPABILITY-MAP.md`. Intent: `docs/intent/v3-platform.md`. Depends on `permissions` and `qr`. Replaces sections 3 to 9 of `docs/plans/2026-07-07-cool-and-unified-permissions-design.md`. Section 2 of that design is replaced by `SPEC-permissions.md`.

## Objective

Track who belongs to which COOL and who came to each meeting, with the leader as the only judge of attendance. A member's scan is a hint. The leader's mark is the fact.

What this module delivers:

1. COOL records, carried over from the existing `cools` table, with membership starting empty in v3.
2. Meetings that are offline, online or hybrid, each with one meeting QR.
3. Attendance that stays `pending` until the leader marks every member, then closes.
4. Guests typed in per meeting, linkable to an account later.
5. Per-member, per-meeting and quarterly attendance stats.

Users are members (scan the QR), core team (create meetings, add guests), leaders (mark and close), facilitators (oversee several COOLs) and campus admins. Success: after a meeting, the leader sees a roster with the scanners pre-filled as present. The leader marks the rest and closes it. The quarterly report is then right without a spreadsheet.

## Tech Stack

- Go 1.26, Echo v4, GORM over PostgreSQL (existing)
- `internal/pkg/herrx`, `internal/pkg/response`, `internal/pkg/configdb` (from `platform`)
- `internal/pkg/permission` (from `permissions`)
- `internal/pkg/qr` (from `qr`)

## Commands

```bash
go build ./...
go test ./internal/pkg/coolstats/ ./internal/usecases/cool/...                 # unit, no DB
TEST_DATABASE_URL="postgres://postgres:$DB_PASSWORD@localhost:5888/community_db?sslmode=disable" \
  go test ./tests/integration/cool/ -v                                           # DB tests, skipped when unset
make migration name=cool_v3                                                      # numbered after qr (000024)
make generate-docs
```

## Project Structure

```
internal/models/cool/                 → entities, DTOs, errors.go (COOL_ codes), locale.go
internal/pkg/coolstats/               → quarterly math, pure functions
internal/repositories/pgsql/cool/     → cools, members, meetings, attendance, guests
internal/usecases/cool/               → cools, members, meetings, attendance, stats, qr_actions.go
internal/deliveries/http/v3/cool/     → cool_handler.go, meeting_handler.go, attendance_handler.go
tests/integration/cool/               → membership, meeting, scan, close, revision and stats flows
tests/integration/db/migrations/000025_cool_v3.up.sql
```

## Code Style

### Entities

`cools` is the existing table. Add `area` and `region` (nullable text). The three `*_community_ids` arrays and `users.cool_id` are frozen. v3 never reads or writes them. No `cool_facilitators` table. A facilitator is a `cool_facilitator` grant scoped to the COOL.

| Table | Columns |
|---|---|
| `cool_members` | `id`, `cool_id`, `community_id`, `joined_at`, `added_by`, `left_at` nullable, `removed_by` nullable. Unique index on `community_id` where `left_at IS NULL`. One COOL per person, by the database. |
| `cool_meetings` | `id`, `cool_id`, `code` unique, `topic` required, `name` nullable. `starts_at`, `ends_at` timestamptz. `location_type` is `offline`, `online` or `hybrid`. `location_name` nullable. `status` is `pending`, `closed` or `cancelled`. `created_by`, `closed_by` nullable, `closed_at` nullable. |
| `cool_meeting_attendance` | `id`, `meeting_id`, `community_id`. `status` is `pending`, `present`, `absent_with_notice` or `absent_without_notice`. `scanned_at`, `reason`, `marked_by`, `marked_at`, all nullable. `revisions` jsonb, append-only. Unique on `(meeting_id, community_id)`. |
| `cool_meeting_guests` | `id`, `meeting_id`, `name` required, `added_by`. `phone`, `identifier` (any other contact the guest gives), `notes`, all nullable. `community_id`, `linked_by`, `linked_at`, all nullable, set when linked. |

Membership starts empty. The migration creates the tables and adds the two columns. It inserts no `cool_members` rows. Leaders and admins rebuild rosters through the API.

### Meeting states

```
create ──▶ pending ──close (zero pending rows)──▶ closed ──edit (revision)──▶ closed
              │
              └──cancel──▶ cancelled
```

A meeting is `pending` from creation. On creation, one `pending` attendance row is inserted for every active member of the COOL. If a member joins the COOL after creation, the leader can still mark them, and that creates their row. Close fails with `COOL_MEETING_PENDING` while any row is still `pending`. Nothing closes a meeting by itself. A `cancelled` meeting keeps its rows and counts in no report.

### Attendance states

```
pending ──scan──▶ pending (scanned_at set)
pending ──leader marks──▶ present | absent_with_notice | absent_without_notice
marked  ──leader re-marks before close──▶ another marked status
marked  ──edit after close──▶ another marked status, one revision appended
```

A scan never changes `status`. It sets `scanned_at`, so the roster screen can pre-fill `present`. Only a leader or facilitator turns a row into a fact. `reason` is optional and is meant for `absent_with_notice`.

### Endpoints

All under `/api/v3`. The second column is the `permission.Require` action. The resource is the COOL, with `Parents = {campus: cools.campus_code}`, so a campus admin passes too.

```
Member
GET    /v3/cools/mine                              signed in, own COOL page, or COOL_NOT_A_MEMBER
GET    /v3/cools/mine/meetings                     signed in, upcoming and past meetings of own COOL

Any COOL title, or campus admin
GET    /v3/cools                                   cool.view, list scoped by Scopes(cool.view, cool)
GET    /v3/cools/{id}                              cool.members.view, detail with roster and titles
GET    /v3/cools/{id}/meetings                     cool.attendance.view
GET    /v3/meetings/{code}                         cool.attendance.view
GET    /v3/meetings/{code}/attendance              cool.attendance.view, roster with status and scanned_at
GET    /v3/meetings/{code}/qr.png?size=            cool.attendance.record, the meeting QR for print or screen
GET    /v3/cools/{id}/attendance/members?from=&to= cool.attendance.view, per-member rates, window rule
GET    /v3/cools/{id}/stats?quarter=               cool.stats.view

Core team and up
POST   /v3/cools/{id}/members                      cool.members.add, {community_id}
POST   /v3/cools/{id}/meetings                     cool.meetings.create
PATCH  /v3/meetings/{code}                         cool.meetings.create, edit fields or cancel while pending
POST   /v3/meetings/{code}/guests                  cool.attendance.record, {name, phone?, identifier?, notes?}
PATCH  /v3/meetings/{code}/guests/{id}             cool.attendance.record, {community_id} links a guest to an account

Leader and up
PUT    /v3/meetings/{code}/attendance              cool.confirm_attendance, {marks: [{community_id, status, reason?}]}
POST   /v3/meetings/{code}/close                   cool.confirm_attendance
PATCH  /v3/meetings/{code}/attendance              cool.attendance.edit, after close, appends a revision
DELETE /v3/cools/{id}/members/{communityId}        cool.members.remove
PATCH  /v3/cools/{id}                              cool.edit, name, description, topic defaults, area, region, location

Facilitator and up
POST   /v3/cools/{id}/titles                       cool.roles.change, {community_id, title} where title is cool_leader or cool_core_team
DELETE /v3/cools/{id}/titles/{grantId}             cool.roles.change

Campus admin
POST   /v3/cools                                   cool.manage, create in own campus
PATCH  /v3/cools/{id}/facilitators                 cool.manage, {add: [], remove: []} community ids
```

Rules:

- `PUT .../attendance` accepts a partial list. Rows not in the list stay as they are. A mark on a person who is not on the roster and not an active member returns `COOL_NOT_A_MEMBER`.
- `POST .../titles` writes a `permission_grants` row through the `permissions` usecase, scoped to that COOL, and never a raw action. The target must be an active member of that COOL.
- If the person has an active row anywhere, `POST .../members` fails with `COOL_ALREADY_IN_COOL`. Moving a person is remove then add.
- Removing a member revokes every `cool_leader` and `cool_core_team` grant they hold on that COOL, in the same transaction.
- Every membership, title and close write also writes one `user_audit_logs` row in the same `Atomic` transaction.
- `GET .../attendance/members` caps `from` at `cool.attendance_window_months` ago for callers without `cool.manage`.

### Meeting QR

The meeting token is minted by `cool` at meeting creation with `qr.KindMeeting`, `ref = meeting code`, `exp = ends_at + cool.meeting_qr_grace_minutes`. The same token is served by `GET /v3/meetings/{code}/qr.png` for a poster, the leader's phone, or a shared screen.

```go
// internal/usecases/cool/qr_actions.go, called from contract.go
func RegisterQRActions(r qr.Registry, cu CoolUsecase, perm permission.Evaluator) {
	r.Register(qr.KindMeeting, qr.Action{
		Name: "cool_mark_scanned",
		Allowed: func(c qr.ActionCtx) bool {
			coolID, err := cu.CoolIDByMeetingCode(c.Ctx, c.Payload.Ref)
			if err != nil {
				return false
			}
			return cu.IsActiveMember(c.Ctx, coolID, c.Caller) ||
				perm.Can(c.Ctx, c.Caller, "cool.attendance.record", permission.Resource{Kind: "cool", Ref: coolID})
		},
		Handle: func(c qr.ActionCtx) (any, error) {
			// upsert the caller's row as pending with scanned_at = now, never a status change
			return cu.MarkScanned(c.Ctx, c.Payload.Ref, c.Caller)
		},
	})
}
```

`MarkScanned` fails with `COOL_MEETING_CLOSED` on a closed or cancelled meeting. A second scan is a no-op that returns the same row. A scan by a member of another COOL fails `Allowed`, so the client sees `QR_ACTION_NOT_ALLOWED` and the scan log keeps the attempt.

### Stats

`internal/pkg/coolstats` is pure. Monthly rate is `min(attended_in_month, 3) / 3`. Quarterly rate is the sum of the three capped numerators divided by 9. Only `closed` meetings count. Only `present` counts as attended. Guests never count in rates.

Query shape for the per-member view, one round trip:

```sql
SELECT a.community_id,
       date_trunc('month', m.starts_at) AS month,
       count(*) FILTER (WHERE a.status = 'present') AS attended,
       count(*) AS marked
FROM cool_meeting_attendance a
JOIN cool_meetings m ON m.id = a.meeting_id
WHERE m.cool_id = $1 AND m.status = 'closed'
  AND m.starts_at >= $2 AND m.starts_at < $3
GROUP BY 1, 2
```

The Go side folds months into quarters. `GET /v3/cools/{id}/stats` returns member growth by month from `joined_at`, the quarterly table per member, and the COOL average. The report is also served as JSON with the same field names the CSV uses.

### Errors

All through `herrx.Def`, `COOL_` prefix, EN here, ID in `locale.go`.

| Code | Kind | EN message |
|---|---|---|
| `COOL_NOT_FOUND` | not found | We did not find that COOL. |
| `COOL_MEETING_NOT_FOUND` | not found | We did not find that meeting. |
| `COOL_NOT_A_MEMBER` | forbidden | This person is not a member of this COOL. |
| `COOL_ALREADY_IN_COOL` | conflict | This person is already part of another COOL. Remove them there first. |
| `COOL_MEETING_PENDING` | conflict | Some members are still unmarked. Mark everyone before closing. |
| `COOL_MEETING_CLOSED` | conflict | This meeting is already closed. |
| `COOL_MEETING_CANCELLED` | conflict | This meeting was cancelled. |
| `COOL_INVALID_STATUS` | invalid | That attendance status does not exist. |
| `COOL_WINDOW_EXCEEDED` | forbidden | You can only view the last few months. Ask an admin for older data. |
| `COOL_TITLE_NOT_ALLOWED` | invalid | Only leader or core team can be given here. |
| `COOL_GUEST_ALREADY_LINKED` | conflict | This guest is already linked to an account. |
| `COOL_INVALID_INPUT` | invalid | Something in your submission does not look right. |

### Configdb keys

| Key | Default | Meaning |
|---|---|---|
| `cool.meeting_qr_grace_minutes` | `60` | scans work this long after `ends_at`. One hour covers a meeting that runs late and is short enough that a scan the next day fails. |
| `cool.attendance_window_months` | `6` | history cap for callers without `cool.manage` |
| `cool.max_members` | `0` | 0 means unlimited. If set, `POST .../members` enforces it. |

## Testing Strategy

Unit, no DB:

- `coolstats`: a 2-meeting month gives 2/3, a 4-meeting month gives 3/3, a quarter of 3, 4 and 1 gives 7/9. A cancelled meeting is ignored. A `pending` row is not attended.
- Meeting state: close with one `pending` row fails, close with zero passes, edit after close appends exactly one revision.
- Catalog completeness for every `COOL_` class (shared test from `platform`).

Integration, real Postgres via `TEST_DATABASE_URL`:

- Add a member who is active elsewhere fails with `COOL_ALREADY_IN_COOL`. Remove then add succeeds. The unique index holds under two concurrent adds.
- Create a meeting with 5 members, 5 `pending` rows exist. Two members scan, `scanned_at` set, `status` still `pending`. Leader marks 5, close succeeds, one audit row.
- Scan on a closed meeting fails with `COOL_MEETING_CLOSED` and writes a scan log row.
- Scan after `ends_at + grace` fails with `QR_EXPIRED_TOKEN` at the `qr` layer.
- Remove a leader from a COOL, their `cool_leader` grant is revoked in the same transaction.
- Guest link: a guest gets an account through `POST /v3/admin/users`, the leader links the guest row, a second link fails.
- Window: a core team member asking for 8 months gets `COOL_WINDOW_EXCEEDED`, a campus admin does not.

No database mocks. Every rule in Code Style has one test.

## Boundaries

Always:

- Insert the `pending` rows in the same `Atomic` transaction as the meeting.
- Keep a scan free of status changes. `scanned_at` only.
- If a member leaves, revoke their COOL titles in the same transaction.

Ask first:

- Auto-closing or auto-marking a meeting by time.
- Adding a fourth attendance status.
- Backfilling `cool_members` from the frozen arrays.
- Changing the quarterly formula.

Never:

- Read `cools.*_community_ids`, `users.cool_id` or `cool_new_joiners` from v3 code.
- Let a member mark their own status.
- Count a `pending` row or a guest in a rate.
- Delete an attendance row or a guest row. Corrections are revisions.

## Success Criteria

1. Leader creates a hybrid meeting for a COOL of 8. The roster shows 8 `pending` rows and one QR PNG.
2. Three members scan the QR on the video call screen. Their rows show `scanned_at` and `status = pending`.
3. `POST .../close` returns 409 `COOL_MEETING_PENDING` with 5 rows unmarked.
4. Leader marks 3 `present`, 4 `absent_with_notice` with reasons, 1 `absent_without_notice`. Close returns 200, `closed_by` is the leader.
5. A member of another COOL scanning the same QR gets 403 `QR_ACTION_NOT_ALLOWED` and a scan log row exists.
6. Leader edits one row after close. The row has one revision with the old status, the new status, who and when.
7. Adding a person who is active in another COOL returns 409 `COOL_ALREADY_IN_COOL`.
8. The quarterly report for a member with 3, 4 and 1 closed meetings attended shows 7/9.
9. After migration, `cool_members` is empty and every old `cools` row still exists with the same id.
10. Every `COOL_` code has an ID message, or `go test` fails.

## Open Questions

1. Resolved 2026-09-17: `cool.manage` is on `campus_admin` in `SPEC-permissions.md`.
2. Resolved 2026-09-17: facilitators grant `cool_leader` and `cool_core_team` inside their own COOLs through `POST .../titles`. The intent says a facilitator oversees several COOLs, and this is the one thing they do most.
3. Resolved 2026-09-17: the new-joiner funnel report is not in v3. Guests are listed per meeting and that is enough.
4. Resolved 2026-09-17: a removed member keeps every past attendance row, and past meeting reports still show them. Roster views and new meetings exclude them from `left_at` on. Nothing is deleted.
