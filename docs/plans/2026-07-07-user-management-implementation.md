# MyGrow User Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build admin user management (permissioned CRUD, scoped views, audit, dashboard) and humanized OTP+PIN+trusted-device auth (spec: `docs/plans/2026-07-07-user-management-design.md`), on the shared permission engine and notification infrastructure.

**Architecture:** Evolves the v2 user domain: one migration adds columns + `user_departments`/`kom_certifications`/`user_devices`/`auth_challenges`. The COOL plan's permission engine is extended with `user`-type resources (target descriptors) so `own_departments`/`own_cool` scoping and FieldSet visibility work for user actions. OTP/reset delivery rides the shared notification outbox. Errors: herr, bilingual EN/ID.

**Tech Stack:** Go ≥1.26, Echo v4, GORM/Postgres, herr, `golang.org/x/crypto/bcrypt` (already transitively present via existing hash pkg — verify; else `go get`).

## Global Constraints

- **Hard prerequisite:** COOL plan (`2026-07-07-cool-and-unified-permissions-implementation.md`) **Tasks 1–6 must be complete first** — they provide herr+locale registry, `permission_grants` + `user_audit_logs` tables, the `permission.Engine`, `LoadActor`, and the grants API. This plan modifies those, never duplicates them.
- **Soft dependency:** events-v3 plan Task 12 provides the outbox dispatcher + `notify` pkg. Task 5 here creates the same pieces **idempotently** (identical DDL/paths, `IF NOT EXISTS` / skip-if-present) so either build order works. Whichever plan runs second skips what exists.
- Module path `go-community`; `Atomic`-only transactions; herr bilingual errors registered via `locale.Register("id", ...)` + completeness gate; default locale `id`.
- All auth secrets hashed at rest: PINs bcrypt, OTP codes + reset tokens SHA-256. Never logged.
- Forgot-* endpoints always return 200 (no account enumeration).
- Every user-management mutation writes a field-level diff to `user_audit_logs` **in the same transaction**.
- Configdb keys (seeded on boot, exact names): `usermgmt.max_departments_per_user`="0", `usermgmt.max_active_devices`="1", `usermgmt.otp_length`="6", `usermgmt.otp_expiry_minutes`="5", `usermgmt.otp_max_attempts`="5", `usermgmt.pin_lockout_attempts`="5", `usermgmt.pin_lockout_minutes`="15", `usermgmt.otp_fallback`="admin_assist", `usermgmt.password_login_enabled`="true".
- DB tests: `TEST_DATABASE_URL` or skip; base migrations + 000023/000024 applied.
- Conventional Commits `feat(usermgmt): ...`.

## Cross-Plan Sync Map (who owns what)

| Piece | Owner plan | This plan's relationship |
|---|---|---|
| herr + `internal/pkg/locale` registry | COOL T1 (or events T2) | consumes; registers `usermgmt` ID catalog |
| `permission_grants`, `user_audit_logs`, seeds mechanism | COOL T2 | consumes; **adds `user.*` seed rows** (Task 1) |
| `permission.Engine`, `Actor`, `LoadActor` | COOL T3/T5 | **extends** (Task 3): user targets, FieldSets, `LedDepartments` loading |
| Grants API action-catalog validation | COOL T6 | modifies: registers `user.*` actions in the catalog list |
| `v3_notification_outbox` + `internal/pkg/notify` + dispatch endpoint | events T12 | Task 5 creates idempotently if absent |
| `eligibility` extensions (cools rules) | events plan (amended) | not touched here |
| `users.cool_id` legacy mirror | COOL T5/T7 | untouched; admin CRUD here never writes cool fields (COOL module owns membership) |

## File Structure

```
tests/integration/db/migrations/000025_user_management.up.sql / .down.sql
internal/models/usermgmt/errors.go / entities.go / dto.go
internal/pkg/permission/permission.go        MODIFY: UserRes target + FieldSets
internal/pkg/otp/otp.go                      challenge create/verify (pure given store+clock)
internal/pkg/notify/notify.go                create if absent (events T12 owns shape)
internal/repositories/pgsql/usermgmt_repositories.go   departments, kom, devices, challenges (+ wiring)
internal/repositories/pgsql/coolmgmt_repositories.go   MODIFY: LoadActor adds LedDepartments
internal/usecases/usermgmt/admin_usecase.go / auth_usecase.go / dashboard_usecase.go / usecases.go
internal/deliveries/http/v2/usermgmt_handler.go / auth_v2mgmt_handler.go
internal/deliveries/http/middleware/authorization.go   MODIFY: login block for deactivated + device check hook
tests/integration/usermgmt/harness_test.go + *_test.go
```

---

### Task 1: Migration 000025 — user domain evolution + user.* grant seeds

**Files:**
- Create: `tests/integration/db/migrations/000025_user_management.up.sql`, `.down.sql`
- Create: `tests/integration/usermgmt/harness_test.go` (same pattern as coolmgmt harness; applies 000023–000025; cleanup deletes fixtures `community_id LIKE '98%'`, non-seed grants, truncates new tables)

- [ ] **Step 1: Up migration**

```sql
SET TIME ZONE 'Asia/Jakarta';

ALTER TABLE "users" ADD COLUMN IF NOT EXISTS "sub_district" VARCHAR(100);
ALTER TABLE "users" ADD COLUMN IF NOT EXISTS "district" VARCHAR(100);
ALTER TABLE "users" ADD COLUMN IF NOT EXISTS "volunteer_status" VARCHAR(10)
    CHECK (volunteer_status IN ('active','sabat') OR volunteer_status IS NULL);
ALTER TABLE "users" ADD COLUMN IF NOT EXISTS "kkj_file_url" VARCHAR(512);
ALTER TABLE "users" ADD COLUMN IF NOT EXISTS "baptism_file_url" VARCHAR(512);
ALTER TABLE "users" ADD COLUMN IF NOT EXISTS "pin_hash" VARCHAR(70);
ALTER TABLE "users" ADD COLUMN IF NOT EXISTS "pin_failed_attempts" INT NOT NULL DEFAULT 0;
ALTER TABLE "users" ADD COLUMN IF NOT EXISTS "pin_locked_until" TIMESTAMPTZ;
ALTER TABLE "users" ADD COLUMN IF NOT EXISTS "deactivated_at" TIMESTAMPTZ;
ALTER TABLE "users" ADD COLUMN IF NOT EXISTS "deactivated_by" VARCHAR(15);

CREATE TABLE IF NOT EXISTS "user_departments" (
    "id" BIGSERIAL PRIMARY KEY,
    "community_id" VARCHAR(15) NOT NULL REFERENCES users(community_id),
    "department_code" VARCHAR(50) NOT NULL,
    "position" VARCHAR(10) NOT NULL DEFAULT 'member' CHECK (position IN ('leader','core_team','member')),
    "is_primary" BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE (community_id, department_code)
);
CREATE INDEX IF NOT EXISTS idx_user_departments_dept ON user_departments(department_code);

CREATE TABLE IF NOT EXISTS "kom_certifications" (
    "id" BIGSERIAL PRIMARY KEY,
    "community_id" VARCHAR(15) NOT NULL REFERENCES users(community_id),
    "level" INT NOT NULL,
    "kom_id" VARCHAR(50),
    "file_url" VARCHAR(512),
    "completed_at" DATE,
    UNIQUE (community_id, level)
);

CREATE TABLE IF NOT EXISTS "user_devices" (
    "id" UUID PRIMARY KEY,
    "community_id" VARCHAR(15) NOT NULL REFERENCES users(community_id),
    "device_id" VARCHAR(64) NOT NULL,
    "platform_name" VARCHAR(100) NOT NULL DEFAULT '',
    "last_seen_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "revoked_at" TIMESTAMPTZ,
    UNIQUE (community_id, device_id)
);

CREATE TABLE IF NOT EXISTS "auth_challenges" (
    "id" UUID PRIMARY KEY,
    "community_id" VARCHAR(15) NOT NULL,
    "purpose" VARCHAR(20) NOT NULL CHECK (purpose IN
        ('register','login_new_device','pin_reset','password_reset','admin_reset')),
    "channel" VARCHAR(10) NOT NULL DEFAULT 'email',
    "code_hash" VARCHAR(64) NOT NULL,
    "attempts" INT NOT NULL DEFAULT 0,
    "expires_at" TIMESTAMPTZ NOT NULL,
    "used_at" TIMESTAMPTZ,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_auth_challenges_lookup ON auth_challenges(community_id, purpose);

-- Backfills (idempotent):
INSERT INTO kom_certifications (community_id, level)
SELECT community_id, 100 FROM users WHERE is_kom100 = TRUE
ON CONFLICT (community_id, level) DO NOTHING;

INSERT INTO user_departments (community_id, department_code, position, is_primary)
SELECT community_id, department, 'member', TRUE FROM users
WHERE department IS NOT NULL AND department <> ''
ON CONFLICT (community_id, department_code) DO NOTHING;

-- user.* office-standard grant seeds (engine + grants table from migration 000023):
INSERT INTO permission_grants (subject_type, subject_ref, action, resource_type, scope, granted_by) VALUES
 ('role','gco-admin','user.manage','global','all','seed'),
 ('role','gco-admin','user.view_full','global','all','seed'),
 ('role','gco-admin','user.activate','global','all','seed'),
 ('role','gco-admin','user.dashboard.view','global','all','seed'),
 ('role','gco-view','user.view_basic','global','all','seed'),
 ('role','pastoral','user.view_basic','global','all','seed'),
 ('role','pastoral','user.activate','global','all','seed'),
 ('role','pastoral','user.dashboard.view','global','all','seed'),
 ('role','dept-leader','user.view_basic','global','own_departments','seed'),
 ('role','dept-leader','user.activate','global','own_departments','seed'),
 ('role','community-leader','user.view_basic','global','own_cool','seed'),
 ('role','community-leader','user.activate','global','own_cool','seed'),
 ('role','community-facilitator','user.view_basic','global','own_cool','seed'),
 ('role','community-department','user.view_basic','global','all','seed'),
 ('role','mis','user.view_basic','global','all','seed'),
 ('role','mis','user.activate','global','all','seed'),
 ('role','mis','user.audit.view','global','all','seed'),
 ('role','mis','user.dashboard.view','global','all','seed')
ON CONFLICT DO NOTHING;
INSERT INTO roles (role, description) VALUES
 ('gco-admin','GC Office admin'), ('gco-view','GC Office viewer'), ('pastoral','Pastoral team'),
 ('dept-leader','Department leader'), ('community-leader','COOL leader (admin console)'),
 ('community-facilitator','COOL facilitator (admin console)'), ('community-department','Community department'),
 ('mis','MIS team')
ON CONFLICT (role) DO NOTHING;
```

Down: drop the four new tables; ALTER users DROP the new columns; `DELETE FROM permission_grants WHERE granted_by='seed' AND action LIKE 'user.%'`.

- [ ] **Step 2:** Harness compiles; with DB: backfill assertions (a seeded `is_kom100` user gains a level-100 row; a user with legacy `department='MUSIC'` gains a primary membership; running twice is a no-op).
- [ ] **Step 3:** Commit `feat(usermgmt): user domain migration, backfills and user.* grant seeds`.

---

### Task 2: Bilingual herr catalog for usermgmt

**Files:** Create `internal/models/usermgmt/errors.go` + `errors_test.go` (identical structure to coolmgmt Task 1: `def()` registrar, `AllClasses`, completeness gate, leak test, `Localize`).

Classes (EN shown; ID registered via `locale.Register`):

| Code | Kind | EN |
|---|---|---|
| `USER_NOT_FOUND` | NotFound | "We couldn't find that person." |
| `OTP_INVALID` | Unauthorized | "That code doesn't match. Please check and try again." |
| `OTP_EXPIRED` | Unauthorized | "That code has expired. We can send you a new one." |
| `OTP_TOO_MANY_ATTEMPTS` | RateLimited | "Too many tries. Please wait a moment, then request a new code." |
| `PIN_INCORRECT` | Unauthorized | "That PIN isn't right. Please try again." |
| `PIN_LOCKED` | RateLimited | "Too many PIN attempts. Please try again in a few minutes." |
| `DEVICE_NOT_TRUSTED` | Unauthorized | "We don't recognize this phone yet. We'll send a code to verify it's you." |
| `DEVICE_LIMIT_REACHED` | Conflict | "You're signed in on another phone. Continue here and we'll sign that one out." |
| `ACCOUNT_DEACTIVATED` | Forbidden | "This account has been deactivated. Please contact the GC Office." |
| `RESET_TOKEN_INVALID` | Unauthorized | "That reset link isn't valid anymore. Please request a new one." |
| `CHANNEL_UNAVAILABLE` | Unprocessable | "We can't send a code to this account yet. Please contact the GC Office for help." |
| `TOO_MANY_DEPARTMENTS` | Invalid | "This person already has the maximum number of departments." |
| `INVALID_INPUT` | Invalid | "Something in your submission doesn't look right. Please review and try again." |
| `FORBIDDEN` | Forbidden | "You don't have access to do this." |

Steps: failing tests → implement (+ Indonesian catalog for every code) → PASS → commit `feat(usermgmt): bilingual error catalog`.

---

### Task 3: Permission engine extension — user targets + FieldSets

**Files:**
- Modify: `internal/pkg/permission/permission.go` + `permission_test.go`
- Modify: `internal/repositories/pgsql/coolmgmt_repositories.go` (`LoadActor` gains `LedDepartments` from `user_departments WHERE position='leader'`)
- Modify: COOL grants usecase's action catalog list to include the seven `user.*` actions and `permission.manage`

**Interfaces (added to package `permission`):**

```go
type UserTarget struct {
	CommunityID     string
	DepartmentCodes []string
	CoolID          int64
}
func UserRes(t UserTarget) Resource // Resource{Type:"user", Ref:t.CommunityID, target:&t}

type FieldSet string
const (
	FieldsAll   FieldSet = "all"
	FieldsBasic FieldSet = "basic" // name, date_of_birth, address(+sub_district,district), phone_number
	FieldsNone  FieldSet = "none"
)
// Highest visibility the actor has on target: view_full→all, else view_basic→basic, else none.
func (e *Engine) VisibleFields(ctx context.Context, a Actor, t UserTarget) (FieldSet, error)
// SQL predicate builder for list endpoints: returns (whereSQL, args) limiting rows to the
// actor's visibility scope, or ok=false when the actor may list everyone.
func (e *Engine) UserListScope(ctx context.Context, a Actor) (where string, args []interface{}, restricted bool, err error)
```

Resource matching additions: when `res.Type == "user"` and grant scope is `own_departments` → intersect `a.LedDepartments` with `res.target.DepartmentCodes`; `own_cool` → `res.target.CoolID != 0 && (res.target.CoolID == a.CoolID || contains(a.FacilitatedCoolIDs, res.target.CoolID))`; `all` matches any user target. `UserListScope` mirrors the same rules as SQL: dept-scoped → `community_id IN (SELECT community_id FROM user_departments WHERE department_code = ANY(?))`; cool-scoped → `community_id IN (SELECT community_id FROM cool_members WHERE cool_id = ?)`.

- [ ] **Step 1: Failing tests** (extend the matrix):

```go
func TestUserTargetScoping(t *testing.T) {
	grants := []Grant{
		{SubjectType: "role", SubjectRef: "dept-leader", Action: "user.view_basic", ResourceType: "global", Scope: "own_departments"},
		{SubjectType: "role", SubjectRef: "gco-admin", Action: "user.view_full", ResourceType: "global", Scope: "all"},
		{SubjectType: "role", SubjectRef: "community-leader", Action: "user.activate", ResourceType: "global", Scope: "own_cool"},
	}
	e := NewEngine(fakeStore(grants), time.Minute)
	deptLead := Actor{CommunityID: "1", Roles: []string{"dept-leader"}, LedDepartments: []string{"MUSIC"}}
	admin := Actor{CommunityID: "2", Roles: []string{"gco-admin"}}
	coolLead := Actor{CommunityID: "3", Roles: []string{"community-leader"}, CoolID: 7, CoolRole: "leader"}

	inDept := UserTarget{CommunityID: "9", DepartmentCodes: []string{"MUSIC", "USHER"}}
	outDept := UserTarget{CommunityID: "8", DepartmentCodes: []string{"MEDIA"}}
	inCool := UserTarget{CommunityID: "7b", CoolID: 7}

	if ok, _ := e.Can(ctx(), deptLead, "user.view_basic", UserRes(inDept)); !ok {
		t.Fatal("dept leader sees own-dept member")
	}
	if ok, _ := e.Can(ctx(), deptLead, "user.view_basic", UserRes(outDept)); ok {
		t.Fatal("dept leader must not see other depts")
	}
	if ok, _ := e.Can(ctx(), coolLead, "user.activate", UserRes(inCool)); !ok {
		t.Fatal("cool leader activates own member")
	}
	if fs, _ := e.VisibleFields(ctx(), admin, outDept); fs != FieldsAll {
		t.Fatal("admin sees all fields")
	}
	if fs, _ := e.VisibleFields(ctx(), deptLead, inDept); fs != FieldsBasic {
		t.Fatal("dept leader sees basic")
	}
	if fs, _ := e.VisibleFields(ctx(), deptLead, outDept); fs != FieldsNone {
		t.Fatal("out of scope sees none")
	}
}
```

- [ ] Steps: FAIL → implement → all permission tests (old + new) PASS → commit `feat(perm): user-target resources, field sets and list scoping`.

---

### Task 4: Entities + repositories for usermgmt

**Files:** Create `internal/models/usermgmt/entities.go`, `internal/repositories/pgsql/usermgmt_repositories.go`; modify `main_pg_repository.go` (add `UserMgmt *UserMgmtRepositories`). Test: `tests/integration/usermgmt/repositories_test.go`.

**Interfaces:** entities `UserDepartment`, `KomCertification`, `UserDevice`, `AuthChallenge` (tables per Task 1); repos:

```go
Departments: ListByUser / ReplaceForUser(ctx, communityID, []UserDepartment) / LeadersOf(dept) 
Kom:         ListByUser / ReplaceForUser(ctx, communityID, []KomCertification)  // MaxLevel derived in Go
Devices:     ListActive(ctx, communityID) / Get(ctx, communityID, deviceID) / Trust(ctx, *UserDevice) /
             Revoke(ctx, id uuid.UUID) / RevokeAllForUser(ctx, communityID) / CountActive(ctx, communityID) /
             OldestActive(ctx, communityID)
Challenges:  Create(ctx, *AuthChallenge) / Latest(ctx, communityID, purpose) / IncrementAttempts(ctx, id) /
             MarkUsed(ctx, id)
```

Steps: failing test (Trust + CountActive + Revoke round-trip; ReplaceForUser is transactional replace) → implement thin GORM → PASS → commit `feat(usermgmt): entities and repositories`.

---

### Task 5: Shared notification foundation (idempotent with events-v3 Task 12)

**Files:** Create **only if absent**: migration `tests/integration/db/migrations/000026_notification_outbox.up.sql` (exact `v3_notification_outbox` DDL from the events plan Task 1, wrapped in IF NOT EXISTS; down = no-op comment "owned jointly, dropped by events plan"), `internal/pkg/notify/notify.go` (the events plan Task 12 `Notifier` interface + SMTP impl, verbatim shape), outbox repo methods on `V3Repositories` **or**, when the events plan hasn't run, a minimal `OutboxRepo` in `usermgmt_repositories.go` with the same `Enqueue/ClaimBatch/MarkSent/MarkFailed` signatures, plus internal endpoint `POST /v2/internal/notifications/dispatch` guarded by API key — **skipped entirely when `/v3/internal/notifications/dispatch` exists**.

**Dispatcher template additions (this plan owns these regardless of build order):** templates `otp_code` (subject "Kode verifikasi MyGrow", body renders the 6-digit code, bilingual body by user preference defaulting id) and `password_reset` (link `https://<frontend>/reset?token=...`).

Steps: integration test (enqueue `otp_code` → fake notifier receives rendered code) → implement → PASS → commit `feat(usermgmt): shared outbox foundation and auth message templates`.

---

### Task 6: OTP/challenge package (pure logic)

**Files:** Create `internal/pkg/otp/otp.go` + `otp_test.go`.

**Interfaces:**

```go
type Params struct{ Length, MaxAttempts int; Expiry time.Duration }
type Clock func() time.Time
func Generate(p Params) (code string, hash string)          // numeric code, sha256 hex hash
func HashCode(code string) string
// Verify returns nil on success; usermgmt.ErrOTPExpired / ErrOTPInvalid / ErrOTPTooManyAttempts otherwise.
// Caller persists attempt increments / used_at via the Challenges repo.
func Verify(ch usermgmt.AuthChallenge, code string, p Params, now Clock) error
func NewResetToken() (token string, hash string)             // 32 random bytes, base64url + sha256
```

- [ ] **Step 1: Failing test:**

```go
func TestOTPLifecycle(t *testing.T) {
	p := Params{Length: 6, MaxAttempts: 5, Expiry: 5 * time.Minute}
	code, hash := Generate(p)
	if len(code) != 6 || HashCode(code) != hash {
		t.Fatal("code/hash pair")
	}
	now := func() time.Time { return time.Now() }
	ch := usermgmt.AuthChallenge{CodeHash: hash, ExpiresAt: time.Now().Add(p.Expiry)}
	if err := Verify(ch, code, p, now); err != nil {
		t.Fatal("fresh code verifies")
	}
	if err := Verify(ch, "000000", p, now); !usermgmt.ErrOTPInvalid.Is(err) {
		t.Fatal("wrong code → OTP_INVALID")
	}
	ch.Attempts = 5
	if err := Verify(ch, code, p, now); !usermgmt.ErrOTPTooManyAttempts.Is(err) {
		t.Fatal("attempt cap")
	}
	ch.Attempts = 0
	ch.ExpiresAt = time.Now().Add(-time.Minute)
	if err := Verify(ch, code, p, now); !usermgmt.ErrOTPExpired.Is(err) {
		t.Fatal("expired")
	}
	used := time.Now()
	ch.UsedAt = &used
	ch.ExpiresAt = time.Now().Add(time.Minute)
	if err := Verify(ch, code, p, now); !usermgmt.ErrOTPInvalid.Is(err) {
		t.Fatal("single use")
	}
}
```

- [ ] Steps: FAIL → implement (crypto/rand digits; constant-time hash compare) → PASS → commit `feat(usermgmt): otp challenge primitives`.

---

### Task 7: Admin user usecase — CRUD, scoped views, status, audit

**Files:** Create `internal/usecases/usermgmt/admin_usecase.go`, `internal/models/usermgmt/dto.go`. Test: `tests/integration/usermgmt/admin_test.go`.

**Interfaces:**

```go
type AdminUsecase struct{ r *pgsql.PostgreRepositories; eng *permission.Engine; cfg *config.Configuration }
Create(ctx, actor, CreateUserInput) (*UserFull, error)          // user.manage; embeds departments[], kom[], relations reuse existing v2 logic
Get(ctx, actor, communityID) (*UserView, error)                 // VisibleFields decides UserFull vs UserBasic vs NOT_FOUND-as-forbidden
List(ctx, actor, ListParams) ([]UserView, Cursor, error)        // UserListScope SQL predicate + existing cursor pagination
Update(ctx, actor, communityID, UpdateUserInput) (*UserFull, error) // user.manage; dept count vs configdb max → TOO_MANY_DEPARTMENTS
Delete(ctx, actor, communityID) error                           // user.manage; soft delete + RevokeAllForUser
SetStatus(ctx, actor, communityID, StatusInput) error           // user.activate scoped via UserRes(target);
   // StatusInput{Account *string("active"|"deactivated"), Volunteer *string("active"|"sabat"), Reason string}
   // deactivated → deactivated_at/by set + RevokeAllForUser (login paths check this)
EventHistory(ctx, actor, communityID) ([]EventHistoryEntry, error) // v2 event_registration_records; feature-detect v3 registrations table and union when present
```

Every mutation: load-before, apply, diff (`map[field]{old,new}`, files/PIN excluded from values — logged as "(changed)"), audit append, all in one `Atomic`. Target descriptors for `UserRes` loaded via a helper `loadTarget(ctx, communityID)` (departments + cool_id).

- [ ] Steps: failing tests (admin full view vs dept-leader basic view of in-scope user vs FORBIDDEN out-of-scope; list as dept leader returns only dept members — assert SQL-level by seeding 3 users; update writing dept over max → TOO_MANY_DEPARTMENTS; deactivate revokes devices + audit row has old/new status; event history returns seeded v2 record) → implement → PASS → commit `feat(usermgmt): permissioned admin crud with field visibility, scoping, status and audit`.

---

### Task 8: Auth usecase — register, PIN login, device step-up, resets

**Files:** Create `internal/usecases/usermgmt/auth_usecase.go`. Test: `tests/integration/usermgmt/auth_test.go`.

**Interfaces:**

```go
type AuthUsecase struct{ r *pgsql.PostgreRepositories; cfg *config.Configuration; out OutboxEnqueuer }
RegisterStart(ctx, identifier string) error                 // existing user w/o pin OR new shell? spec: registration
   // creates users row (status pending_verification) when identifier unknown; enqueues otp_code; always nil error (no enumeration)
RegisterVerify(ctx, identifier, code string) (setupToken string, err error)   // verifies challenge purpose=register; returns 15-min JWT-style setup token (HMAC, existing auth pkg secret)
RegisterComplete(ctx, RegisterCompleteInput) (*Tokens, error) // {SetupToken, PIN, Name, DeviceID, Platform}
   // bcrypt PIN, activate user, Trust device, issue existing v2 token pair
Login(ctx, LoginInput) (*Tokens, error)                       // {Identifier, PIN, DeviceID}
   // deactivated → ACCOUNT_DEACTIVATED; pin lockout via pin_failed_attempts/pin_locked_until (configdb thresholds);
   // wrong PIN → PIN_INCORRECT (+increment; lock at threshold → PIN_LOCKED);
   // untrusted device → DEVICE_NOT_TRUSTED + auto-enqueue login_new_device OTP (or CHANNEL_UNAVAILABLE fallback per usermgmt.otp_fallback)
DeviceVerify(ctx, DeviceVerifyInput) (*Tokens, error)         // {Identifier, Code, DeviceID, Platform, ReplaceOldest bool}
   // verify challenge; if CountActive >= max_active_devices: ReplaceOldest=false → DEVICE_LIMIT_REACHED (client shows friendly prompt);
   // true → Revoke(OldestActive) then Trust; all in Atomic
PinForgot(ctx, identifier) error                              // always nil; enqueues pin_reset OTP when account+channel exist
PinReset(ctx, identifier, code, newPIN string) error
ForgotPassword(ctx, identifier) error                         // PRD email-link flow: NewResetToken, challenge purpose=password_reset, template password_reset
ResetPassword(ctx, token, newPassword string) error
AdminReset(ctx, actor, communityID string) (oneTimeCode string, err error) // permission.manage; purpose=admin_reset; audit-logged; code returned once
MyDevices(ctx, communityID) ([]UserDevice, error)
RevokeMyDevice(ctx, communityID string, id uuid.UUID) error
```

Refresh-token/device binding: the existing v2 refresh JWT gains a `did` claim (device_id); `RefreshMiddleware` (modify) rejects when the device row is revoked/missing — guarded by configdb `usermgmt.password_login_enabled`-independent check only when `did` present (legacy tokens without `did` keep working).

- [ ] Steps: failing tests (full happy path register→verify→complete→login on same device; login from second device → DEVICE_NOT_TRUSTED then DeviceVerify without ReplaceOldest → DEVICE_LIMIT_REACHED then with → old device revoked + old refresh rejected; 5 wrong PINs → PIN_LOCKED with locked_until set; PinReset round-trip; ForgotPassword token single-use; AdminReset writes audit + code verifies; deactivated user login → ACCOUNT_DEACTIVATED) → implement → PASS → commit `feat(usermgmt): humanized otp+pin auth with trusted devices and resets`.

---

### Task 9: Demographic dashboard

**Files:** Create `internal/usecases/usermgmt/dashboard_usecase.go`. Test: `tests/integration/usermgmt/dashboard_test.go`.

**Interfaces:** `Dashboard(ctx, actor, DashboardFilter{Campus, Department string}) (*DashboardOut, error)` — requires `user.dashboard.view` (global). Output groups (each `[]{Key string; Count int}`): `ByVolunteerStatus`, `ByDepartment` (user_departments), `ByCool` (cool_members joined to cools.name), `ByKKJ` (has kkj_number yes/no), `ByBaptism`, `ByKomLevel` (MAX(level) per user, bucketed, `none` for absent), `ByCampus`, plus `TotalUsers`, `TotalActive`. Single SQL per group, `deleted_at IS NULL` everywhere.

- [ ] Steps: failing test (seed 3 users across 2 campuses/2 kom levels; assert exact counts incl. `none` bucket) → implement → PASS → commit `feat(usermgmt): demographic dashboard aggregates`.

---

### Task 10: HTTP delivery, boot seeding, docs

**Files:** Create `internal/deliveries/http/v2/usermgmt_handler.go`, `internal/deliveries/http/v2/auth_v2mgmt_handler.go`; modify `v2_handler.go` (mount), `middleware/authorization.go` (deactivated-account check + `did` claim validation in RefreshMiddleware), `internal/usecases/main_usecase.go` (add `UserMgmt`), `contract.go` (configdb seeding of the nine `usermgmt.*` keys + `locale.Install()` ordering), `config/config.local.template.yaml` (frontend reset URL base `usermgmt.frontend_reset_url`), `AGENTS.md` pointer. Test: `tests/integration/usermgmt/http_smoke_test.go`.

Routes exactly spec §6 (auth block public with rate limiter middleware; `/v2/me/devices`; `/v2/admin/users*`; `/v2/admin/audit-logs` via the shared audit repo; grants routes already exist from COOL T6 — only the action catalog union matters, done in Task 3). Responder: locale from `Accept-Language`, `usermgmt.Localize`. `WithAuth` export for smoke tests.

- [ ] Steps: smoke test (register→login flow through HTTP; admin list as dept-leader scoped; dashboard 200; audit list as MIS; Indonesian message on wrong PIN with `Accept-Language: id`) → implement → `go build ./...` + full `go test` green → commit `feat(usermgmt): http delivery, middleware hooks, boot seeding`.

---

## Build Order Across the Three Plans

```
1. COOL plan T1–T6            → herr/locale, grants+audit tables, engine, grants API
2. THIS PLAN (all tasks)      → extends engine for user targets; auth; admin mgmt
   (its Task 5 creates outbox infra if events hasn't run)
3. COOL plan T7–T12           → membership/meetings/attendance/stats (uses LoadActor incl. LedDepartments — additive, safe either order vs step 2)
4. Events-v3 plan             → skips herr/outbox pieces that exist; per amendments: no event_organizers,
                                staff checks via Engine.Can(EventRes), eligibility gains cools/requires_cool_membership/cool_roles
                                and dept cross-tab in COOL T11 flips from users.department to user_departments automatically (HasTable check)
```

## Self-Review Notes

- **Spec coverage:** permission matrix seeds (T1), sheets-grants reuse (COOL T6 + T3 catalog union), all PRD attributes incl. files-as-URLs/departments/KOM (T1,T4,T7), view field-limiting + row scoping (T3,T7), activate/deactivate/sabat + device revocation (T7,T8), audit (T7,T8 via shared table), dashboard filters (T9), forgot password email link (T8 ForgotPassword/ResetPassword), OTP register/login + single-device configurable + forgot PIN + admin-assist (T8), event history (T7), configdb keys (Global Constraints + T10).
- **Type consistency:** `permission.UserTarget/UserRes/FieldSet` defined in T3, consumed T7/T9/T10; `usermgmt.AuthChallenge` shared between T4 repo and T6 `otp.Verify`; outbox signatures identical to events plan T12.
- **Deliberate simplifications:** WhatsApp channel = enum value + fallback config only (no sender); `RegisterStart` on an existing fully-registered identifier silently re-sends nothing (returns nil — no enumeration); relations editing reuses the existing v2 update logic rather than reimplementing.
