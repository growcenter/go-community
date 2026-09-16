# Spec: permissions

Module id `permissions` in `CAPABILITY-MAP.md`. Intent: `docs/intent/v3-platform.md`. Depends on `identity`. Replaces section 2 of `docs/plans/2026-07-07-cool-and-unified-permissions-design.md` and the permission section of `docs/plans/2026-07-07-user-management-design.md`. ADR 0001 (events use grants, not `event_organizers`) stays in force.

## Objective

One table answers "can this person do this action on this thing". Admins assign a ready-made title plus a scope, for example "COOL Leader of C-12". One person can hold many titles. No v3 code reads `user_types`, `roles` or the mapping table.

What this module delivers:

1. The `permission_grants` table with five scope kinds: `global`, `campus`, `department`, `cool`, `event`.
2. Titles as named bundles of actions, seeded at boot and fixed in code.
3. A per-request evaluator and one Echo middleware that replace `middleware.UserMiddleware`.
4. Admin endpoints to grant and revoke a title or a raw action.
5. An action registry that other modules add their actions to.
6. Audit rows for every grant change, and a one-time backfill from `user_types` and `user_departments.position`.

Users are admins who hand out titles, and every v3 handler that guards a route. Success is that an admin never types an action name, and a handler never lists roles.

## Tech Stack

- Go 1.26, Echo v4, GORM over PostgreSQL (existing)
- `github.com/jeremygprawira/herr` through `internal/pkg/herrx` (from `platform`)
- `user_audit_logs` table (from `identity`)

## Commands

```bash
go build ./...                                                    # compile
go test ./internal/pkg/permission/...                             # matrix and evaluator unit tests, no DB
TEST_DATABASE_URL="postgres://postgres:$DB_PASSWORD@localhost:5888/community_db?sslmode=disable" \
  go test ./tests/integration/permissions/ -v                     # DB tests, skipped when the URL is unset
make migration name=permission_grants_v3                          # migration pair, numbered after identity's
make generate-docs                                                # swagger, after any handler change
```

## Project Structure

```
internal/pkg/permission/             → registry, titles, Can(), Echo middleware, pure core
internal/pkg/permission/actions.go   → Register(action string), the action registry
internal/pkg/permission/titles.go    → seed titles, one table in code
internal/models/permissions/         → entities, DTOs, errors.go (PERM_ codes), locale.go (id messages)
internal/repositories/pgsql/permissions → grants repository, actor loader
internal/usecases/permissions/       → grant, revoke, list, backfill
internal/deliveries/http/v3/permissions → grants_handler.go, titles_handler.go
tests/integration/permissions/       → grant, revoke, evaluate, backfill flows
tests/integration/db/migrations/0000NN_permission_grants_v3.up.sql
```

Other modules register actions in their own `init()` and call `permission.Require(action)` on their routes. They never touch the grants table.

## Code Style

### Table

```
permission_grants
├─ id             uuid
├─ community_id   who holds the grant
├─ title          nullable, one of the seed titles, e.g. cool_leader
├─ action         nullable, one raw action, e.g. cool.confirm_attendance
├─ scope_kind     global | campus | department | cool | event
├─ scope_ref      nullable for global, else campus code, department code, cool id, event code
├─ granted_by     community_id of the actor
├─ created_at
├─ revoked_at     nullable, set on revoke, row never deleted
```

Rules:

- Exactly one of `title` or `action` is set. A title row means every action in that title. An action row is a raw grant and is the exception.
- A unique index covers `(community_id, title, action, scope_kind, scope_ref)` where `revoked_at IS NULL`.
- A title has one fixed `scope_kind`. The database rejects a `cool_leader` row with `scope_kind = campus` through a check against the title table in code, verified at grant time.
- Grants only add. There are no deny rows.
- Scope match: a `global` grant matches every resource. A `campus`, `department`, `cool` or `event` grant matches a resource with the same kind and ref, or a resource whose parent has that kind and ref.
- Parents: an event belongs to one campus and at most one department. A COOL belongs to one campus. The consuming module passes those parents in `Resource.Parents`, so a `campus_admin` of BSD matches every event and COOL at BSD, and a `department_leader` of MUSIC matches every event owned by MUSIC. A COOL title never matches an event and an event title never matches a COOL.

### Titles

Seeded at boot from `titles.go` through `EnsureSeeds`. The title table is code, so it is under review and test. Strength inside COOL is `cool_facilitator`, then `cool_leader`, then `cool_core_team`. A stronger title includes every action of the weaker one.

| Title | Scope kind | Actions |
|---|---|---|
| `superadmin` | global | every action, hardcoded bypass before any grant is read |
| `campus_admin` | campus | `user.view_full`, `user.manage`, `user.activate`, `user.audit.view`, `user.dashboard.view`, `event.manage`, `event.checkin`, `event.attendees.view`, `event.reports.view`, `cool.view`, `cool.manage`, `cool.stats.view`, `calendar.view_all` |
| `event_admin` | global | `event.manage`, `event.checkin`, `event.attendees.view`, `event.reports.view`, `event.view_internal`, `calendar.view_all` |
| `department_leader` | department | `user.view_basic`, `event.manage`, `event.checkin`, `event.attendees.view`, `event.reports.view`, `event.view_internal` |
| `department_core_team` | department | `user.view_basic`, `event.checkin`, `event.attendees.view`, `event.view_internal` |
| `cool_facilitator` | cool | everything in `cool_leader`, plus `cool.view`, `cool.stats.view`, `cool.roles.change` |
| `cool_leader` | cool | everything in `cool_core_team`, plus `cool.members.remove`, `cool.edit`, `cool.confirm_attendance`, `cool.attendance.edit` |
| `cool_core_team` | cool | `cool.members.view`, `cool.members.add`, `cool.meetings.create`, `cool.attendance.record`, `cool.attendance.view`, `user.create_pending` |
| `event_organizer` | event | `event.manage`, `event.checkin`, `event.attendees.view`, `event.reports.view`, `event.view_internal` |
| `event_desk` | event | `event.checkin`, `event.attendees.view` |
| `permission_admin` | global | `permission.manage`, `permission.view` |

`user.create_pending` is the action behind `POST /v3/admin/users` for a guest. It is on `cool_core_team`, so the two stronger COOL titles have it too. `campus_admin` has it through `user.manage`.

### Action registry

An action is `<module>.<verb>` in snake case, for example `cool.confirm_attendance`, `event.manage`, `calendar.view_all`. Each module registers its actions once:

```go
func init() {
	permission.Register("cool.confirm_attendance", "cool")
}
```

`Register` panics on a duplicate name or a name whose module prefix is not the registering module. A title in `titles.go` that names an unregistered action fails the completeness test. An action that no title names is allowed, since raw grants can use it.

Actions owned by this module: `permission.manage`, `permission.view`.

### Evaluator

```go
type Resource struct {
	Kind    string            // global | campus | department | cool | event
	Ref     string            // "" for global
	Parents map[string]string // e.g. {"campus": "BSD", "department": "MUSIC"}, set by the consuming module
}

// Can reports whether the actor may do action on res.
func Can(ctx context.Context, actor string, action string, res Resource) (bool, error)

// Scopes returns every scope_ref of the given kind the actor may do action on,
// plus the refs of parent kinds (campus, department) that also grant it.
// A list endpoint builds: code IN refs OR campus_code IN parents["campus"] OR department_code IN parents["department"].
// Returns all=true for a global grant.
func Scopes(ctx context.Context, actor string, action string, kind string) (refs []string, parents map[string][]string, all bool, err error)
```

Evaluation order:

1. If the actor holds an active `superadmin` row, return true.
2. Load every active grant row for the actor once per request and keep it on the Echo context.
3. Expand title rows into actions through the title table.
4. If any row has the action and its scope matches `res` or one of `res.Parents`, return true.

No cache across requests. A revoke takes effect on the next request. A deactivated account is stopped by `identity` before this runs.

### Middleware

```go
// internal/deliveries/http/v3/cool/meeting_handler.go
g.POST("/cools/:coolId/meetings/:meetingId/confirm",
	h.Confirm,
	permission.Require("cool.confirm_attendance", permission.FromParam("cool", "coolId")),
)

func (h *MeetingHandler) Confirm(c echo.Context) error {
	var req cool.ConfirmRequest
	if err := c.Bind(&req); err != nil {
		return response.Fail(c, herrx.ErrInvalidInput.Wrap(err))
	}
	if err := h.validate.Struct(req); err != nil {
		return response.Fail(c, cool.FieldErrors(err))
	}
	actor := permission.Actor(c)
	out, err := h.meetings.Confirm(c.Request().Context(), actor, c.Param("coolId"), c.Param("meetingId"), req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out, nil)
}
```

`permission.Require(action, resolver)` runs after the identity JWT middleware. The resolver builds the `Resource` from a path param, a query param or a constant. On a miss it returns `PERM_FORBIDDEN`. Routes with no action use only the identity middleware. `middleware.UserMiddleware(roles...)` is never used in v3.

### Endpoints

All under `/api/v3`. "Permission admin" means a person with `permission.manage`. "Viewer" means a person with `permission.view` or the target user.

```
GET    /v3/permissions/titles                       any signed-in user, the seed title table
GET    /v3/permissions/me                           own active grants
GET    /v3/permissions/grants?community_id=&title=&scope_kind=&scope_ref=   viewer, cursor pagination
POST   /v3/permissions/grants                       permission admin, {community_id, title | action, scope_kind, scope_ref}
DELETE /v3/permissions/grants/{id}                  permission admin, sets revoked_at
GET    /v3/permissions/check?action=&kind=&ref=     any signed-in user, own answer only, for the web app
```

A `POST` with `action` instead of `title` is a raw grant. It needs `permission.manage` and writes an audit row with `raw: true`. A `POST` for a title with the wrong `scope_kind` returns `PERM_SCOPE_MISMATCH`. A `POST` that duplicates an active row returns `PERM_ALREADY_GRANTED`. Nobody can grant or revoke `superadmin` through the API. That row is set by migration only.

### Audit

Every `POST` and `DELETE` writes one `user_audit_logs` row in the same `Atomic` transaction. The row has four parts. `actor` is the caller. `target_user` is the `community_id` of the grant. `action` is `permission.grant_added` or `permission.grant_removed`. `changes` holds the full grant row, plus `raw` for an action grant.

### Backfill

One migration reads the frozen tables once and inserts grants with `granted_by = 'migration'`. The old rows are left in place and never read again.

| Old value | New grant |
|---|---|
| `users.user_types` contains `superadmin` | `superadmin`, global |
| `users.user_types` contains `admin` | `event_admin`, global. The old admin was never tied to a campus |
| `users.roles` contains `event-internal-view`, `event-internal-edit`, `event-verify-record` or `event-view-event-viewer` | `event_admin`, global |
| `users.roles` contains `event-view-volunteer` | nothing, it is eligibility, handled by `events` |
| `users.user_types` contains `volunteer` | nothing, `identity` keeps `volunteer_status` |
| `user_types.category = cool` rows | nothing here, `cool` seeds titles from its own membership tables |
| `user_departments.position = leader` | `department_leader` on that department |
| `user_departments.position = core_team` | `department_core_team` on that department |
| `user_departments.position = member` | nothing, membership is the `user_departments` row itself |

### Errors

All through `herrx.Def` with `PERM_` prefix, EN message here, ID message in `locale.go`.

| Code | Kind | EN message |
|---|---|---|
| `PERM_FORBIDDEN` | forbidden | You do not have access to do this. |
| `PERM_TITLE_UNKNOWN` | invalid | That title does not exist. |
| `PERM_ACTION_UNKNOWN` | invalid | That action does not exist. |
| `PERM_SCOPE_MISMATCH` | invalid | That title cannot be given at this scope. |
| `PERM_SCOPE_REF_REQUIRED` | invalid | Please choose which campus, department, COOL or event this applies to. |
| `PERM_ALREADY_GRANTED` | conflict | This person already has that title here. |
| `PERM_GRANT_NOT_FOUND` | not found | We did not find that grant. |
| `PERM_SUPERADMIN_LOCKED` | forbidden | Superadmin cannot be changed from the app. |

### Configdb keys

None. Titles and actions are code. Grants are rows.

## Testing Strategy

Unit, no DB, under `internal/pkg/permission`:

- Matrix: every seed title, every registered action, in scope and out of scope. Generated from `titles.go`, so a new title is covered without a new test.
- Strength: every action of `cool_core_team` is in `cool_leader`, and every action of `cool_leader` is in `cool_facilitator`.
- Registry: duplicate action panics, wrong module prefix panics, title naming an unregistered action fails.
- `superadmin` returns true with zero grant rows loaded.
- A revoked row is ignored.

Integration, real Postgres via `TEST_DATABASE_URL`:

- Grant a title, `Can` is true on that ref and false on another ref of the same kind.
- `campus_admin` of BSD passes `Can` for an event resource with parent campus BSD and fails for one with parent campus PIK.
- Revoke, the next `Can` is false, no cache.
- Duplicate active grant is rejected. Grant again after revoke succeeds.
- Raw action grant works and its audit row has `raw: true`.
- `Scopes` returns the exact ref list for a `cool_facilitator` of two cools, and `all = true` for `event_admin`.
- `Scopes(actor, "event.view_internal", "event")` for a `department_leader` of MUSIC returns no event refs and `parents["department"] = [MUSIC]`.
- Backfill on a fixture copy of the frozen tables produces the mapped rows and leaves the old tables unchanged.
- Catalog completeness for every `PERM_` class (shared test from `platform`).

No database mocks.

## Boundaries

Always:

- Guard every v3 route that is not public with `permission.Require` or the identity middleware.
- Write the audit row in the same `Atomic` transaction as the grant change.
- Add a new action to the registry and to the matrix test in the same commit.

Ask first:

- Adding or renaming a seed title, or moving an action between titles.
- Adding a sixth scope kind.
- Any cache across requests.

Never:

- Read `user_types`, `roles` or `users.roles` from v3 code after the backfill migration.
- Put a role or title name inside a handler. Handlers name an action only.
- Add a deny rule.
- Delete a grant row. Revoke sets `revoked_at`.
- Grant or revoke `superadmin` through the API.

## Success Criteria

1. An admin sends `POST /v3/permissions/grants` with `title = cool_leader` and `scope_ref = C-12`. The person can confirm attendance for C-12 and gets 403 `PERM_FORBIDDEN` for C-13.
2. The same person also gets `department_core_team` on `MUSIC` and `event_desk` on `EV-2026-CAMP`. `GET /v3/permissions/me` lists all three.
3. `DELETE` on the cool grant makes the next confirm call return 403 with no restart and no wait.
4. `POST` with `title = cool_leader` and `scope_kind = campus` returns 400 `PERM_SCOPE_MISMATCH`.
5. A `superadmin` row holder passes every `Can` call with the grants table otherwise empty.
6. `grep -rn 'UserMiddleware(' internal/deliveries/http/v3` returns nothing.
7. After backfill, every user who had `superadmin` in `user_types` has an active `superadmin` grant, and every `user_departments.position = leader` row has a matching `department_leader` grant.
8. Every `PERM_` code has an ID message, or `go test` fails.
9. The matrix test covers every title times every action with no manual list.

## Open Questions

1. Resolved 2026-09-17: old `user_types = admin` backfills to `event_admin`, global. Nobody gets `campus_admin` from the backfill.
2. `user_types` rows are data, not code, so production can hold type names this spec has not seen. Run `SELECT DISTINCT unnest(user_types) FROM users` on production before writing the backfill and add any unknown value to the table above.
3. `campus_admin` action list above is a first draft. Ask the office which user and event actions a campus admin really needs.
4. Whether `permission_admin` must be a title at all, or `permission.manage` stays on `superadmin` only until a second admin needs it.
