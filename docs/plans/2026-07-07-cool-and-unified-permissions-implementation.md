# COOL Management & Unified Permissions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the COOL module (normalized membership, meetings, audited attendance, analytics) on the unified resource-scoped permission engine shared by users/cools/events (spec: `docs/plans/2026-07-07-cool-and-unified-permissions-design.md`).

**Architecture:** Evolves the v2 COOL domain in place. New `permission_grants` (subjects: role/user/cool_role; resources: global/cool/event) evaluated by a pure engine in `internal/pkg/permission`. `cool_members` replaces the arrays+`users.cool_id` dual-write (backfilled by idempotent SQL, legacy mirrors kept in sync during transition). Meetings/attendance/guests are new tables. Errors via herr, bilingual EN/ID.

**Tech Stack:** Go (≥1.26 once herr lands), Echo v4, GORM/Postgres, `github.com/jeremygprawira/herr`.

## Global Constraints

- Module path `go-community`. `Atomic`-only transactions (never `Transaction` — known bug).
- Errors: herr classes in package `coolmgmt`; every class has EN message + ID translation under `errors.<lowercase_code>.message`; completeness gate test enforces it. Default locale `id`.
- herr requires `go 1.26` in `go.mod` — Task 1 handles it idempotently (the events-v3 plan's Task 2 may have already done both the dependency and the bump; skip whatever is already present).
- DB integration tests read `TEST_DATABASE_URL` and `t.Skip` when unset; the target database must already have base migrations 000001–000021 applied (`make migration_up` against the test DB).
- Permission evaluation: superadmin user_type bypasses everything, hardcoded in the engine. Grants only ADD; no deny rules.
- Legacy mirrors during transition: every membership write also updates `users.cool_id` and the three arrays on `cools`, inside the same `Atomic` tx. Nothing else may write those mirrors.
- Configdb keys added by this plan: `cool.attendance_window_months` (default "6").
- Commit after every green cycle, Conventional Commits `feat(cool): ...` / `feat(perm): ...`.

## File Structure

```
tests/integration/db/migrations/000023_permission_grants.up.sql / .down.sql
tests/integration/db/migrations/000024_cool_management.up.sql / .down.sql
internal/models/coolmgmt/errors.go        herr classes + EN/ID catalog (self-registering with shared v3-style pattern)
internal/models/coolmgmt/entities.go      CoolMember, CoolFacilitator, CoolMeeting, MeetingAttendance, MeetingGuest, AuditLog
internal/models/coolmgmt/dto.go           request/response DTOs
internal/pkg/permission/permission.go     Actor, Resource, Grant, Engine.Can (pure eval + cache)
internal/pkg/coolstats/coolstats.go       quarterly-report math (pure)
internal/repositories/pgsql/coolmgmt_repositories.go   CoolMgmtRepositories (members, facilitators, meetings, attendance, guests, grants, audit) + ActorLoader; wired into PostgreRepositories
internal/usecases/coolmgmt/grant_usecase.go
internal/usecases/coolmgmt/member_usecase.go
internal/usecases/coolmgmt/cool_usecase.go       (new admin CRUD; existing usecases.CoolUsecase untouched)
internal/usecases/coolmgmt/meeting_usecase.go
internal/usecases/coolmgmt/attendance_usecase.go
internal/usecases/coolmgmt/stats_usecase.go
internal/usecases/coolmgmt/usecases.go           aggregate + New(r, cfg)
internal/deliveries/http/v2/coolmgmt_handler.go  all new routes + locale-aware herr responder
internal/usecases/main_usecase.go                add CoolMgmt field (modify)
internal/deliveries/http/v2/v2_handler.go        mount (modify)
tests/integration/coolmgmt/harness_test.go + *_test.go
```

---

### Task 1: Foundations — herr + bilingual COOL error catalog

**Files:**
- Modify: `go.mod` (idempotent: `go 1.26` + herr dep if absent)
- Create: `internal/models/coolmgmt/errors.go`
- Test: `internal/models/coolmgmt/errors_test.go`

**Interfaces:**
- Produces (package `coolmgmt`): classes `ErrAlreadyInCool` (metadata carries current cool name), `ErrNotCoolMember`, `ErrCoolNotFound`, `ErrMeetingNotFound`, `ErrAttendanceWindowExceeded`, `ErrMeetingAlreadyRecorded`, `ErrForbidden`, `ErrLastLeader` (cannot remove the only leader), `ErrInvalidInput(detail)`; `InstallLocale()` (registers ID messages — merges with any localizer already installed by other modules via a shared `mapl` map builder); `Localize(err, locale) (int, []byte)`; `AllClasses`/`englishMessages` for the completeness gate.

- [ ] **Step 1:** If missing: `go get github.com/jeremygprawira/herr@latest`; set `go 1.26` in go.mod; `go mod tidy && go build ./...`.
- [ ] **Step 2: Failing test** — same four tests as the events plan's Task 2 pattern, adapted:

```go
// internal/models/coolmgmt/errors_test.go
package coolmgmt

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMain(m *testing.M) { InstallLocale(); m.Run() }

func TestBilingualComplete(t *testing.T) {
	for _, cls := range AllClasses {
		_, en := Localize(cls.New(), "en")
		_, id := Localize(cls.New(), "id")
		var me, mi map[string]any
		json.Unmarshal(en, &me)
		json.Unmarshal(id, &mi)
		if me["message"] == "" || mi["message"] == "" || me["message"] == mi["message"] {
			t.Errorf("%v: EN and ID must both exist and differ", me["code"])
		}
	}
}

func TestAlreadyInCoolCarriesCoolName(t *testing.T) {
	err := ErrAlreadyInCool.New().WithPublic("current_cool", "COOL Alpha")
	status, body := Localize(err, "en")
	if status != 409 || !strings.Contains(string(body), "COOL Alpha") {
		t.Fatalf("409 with public metadata: %d %s", status, body)
	}
}

func TestInternalNeverLeaks(t *testing.T) {
	_, body := Localize(ErrCoolNotFound.New().Internal("row 42 shard x"), "id")
	if strings.Contains(string(body), "shard") {
		t.Fatal("internal leaked")
	}
}
```

- [ ] **Step 3:** Run → FAIL. **Step 4: Implement** — `def(code, kind, en)` helper identical in shape to the events plan Task 2 (registers into `AllClasses`/`englishMessages`); classes:

| Code | Kind | EN message (ID translation in `InstallLocale`) |
|---|---|---|
| `ALREADY_IN_COOL` | Conflict | "This person is already part of another COOL. Remove them there first." / "Orang ini sudah tergabung di COOL lain. Keluarkan dari sana terlebih dahulu." |
| `NOT_COOL_MEMBER` | NotFound | "This person isn't a member of this COOL." / "Orang ini bukan anggota COOL ini." |
| `COOL_NOT_FOUND` | NotFound | "We couldn't find that COOL." / "Kami tidak dapat menemukan COOL tersebut." |
| `MEETING_NOT_FOUND` | NotFound | "We couldn't find that meeting." / "Kami tidak dapat menemukan pertemuan tersebut." |
| `ATTENDANCE_WINDOW_EXCEEDED` | Forbidden | "That's beyond the attendance history you can view. Ask the GC Office for older records." / "Data kehadiran itu di luar rentang yang dapat Anda lihat. Hubungi GC Office untuk data lama." |
| `MEETING_ALREADY_RECORDED` | Conflict | "Attendance for this meeting is already recorded. You can edit it instead." / "Kehadiran pertemuan ini sudah dicatat. Anda dapat mengubahnya." |
| `FORBIDDEN` | Forbidden | "You don't have access to do this in this COOL." / "Anda tidak memiliki akses untuk melakukan ini di COOL ini." |
| `LAST_LEADER` | Conflict | "A COOL needs at least one leader — assign another leader first." / "COOL harus memiliki minimal satu pemimpin — tunjuk pemimpin lain terlebih dahulu." |
| `INVALID_INPUT` | Invalid | "Something in your submission doesn't look right. Please review and try again." / "Ada data yang belum sesuai. Mohon periksa kembali lalu coba lagi." |

`InstallLocale()` must MERGE with other modules' catalogs: maintain a package-level shared registry pattern — each module contributes its `id` map through `herr.SetLocalizer(mapl.New(mergedMaps))`; implement a tiny `internal/pkg/locale` helper `locale.Register(lang string, msgs map[string]string)` + `locale.Install()` that all modules (v3, coolmgmt, usermgmt later) call, so the last installer never clobbers earlier catalogs. (If `internal/pkg/locale` already exists from another plan's execution, just Register.)

- [ ] **Step 5:** Run → PASS. Commit `feat(cool): bilingual herr catalog + shared locale registry`.

---

### Task 2: Migration 000023 — permission_grants + audit + office-standard seeds

**Files:**
- Create: `tests/integration/db/migrations/000023_permission_grants.up.sql`, `.down.sql`
- Create: `tests/integration/coolmgmt/harness_test.go`

**Interfaces:**
- Produces: tables `permission_grants`, `user_audit_logs` (IF NOT EXISTS — shared with the user-mgmt plan); seeded office-standard grant rows; test helper `func testDB(t *testing.T) *gorm.DB` (applies 000023+000024, truncates coolmgmt tables + grants seeded by tests only — seeded standards preserved via `granted_by='seed'` guard).

- [ ] **Step 1: Up migration**

```sql
SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE IF NOT EXISTS "permission_grants" (
    "id" BIGSERIAL PRIMARY KEY,
    "subject_type" VARCHAR(10) NOT NULL CHECK (subject_type IN ('role','user','cool_role')),
    "subject_ref"  VARCHAR(50) NOT NULL,
    "action"       VARCHAR(60) NOT NULL,
    "resource_type" VARCHAR(10) NOT NULL DEFAULT 'global' CHECK (resource_type IN ('global','cool','event')),
    "resource_ref" VARCHAR(50),
    "scope"        VARCHAR(20) NOT NULL DEFAULT 'all',
    "scope_refs"   TEXT[] NOT NULL DEFAULT '{}',
    "granted_by"   VARCHAR(15) NOT NULL DEFAULT '',
    "created_at"   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (subject_type, subject_ref, action, resource_type, resource_ref)
);
CREATE INDEX IF NOT EXISTS idx_grants_action ON permission_grants(action);

CREATE TABLE IF NOT EXISTS "user_audit_logs" (
    "id" UUID PRIMARY KEY,
    "actor" VARCHAR(15) NOT NULL,
    "target_user" VARCHAR(15),
    "action" VARCHAR(60) NOT NULL,
    "changes" JSONB NOT NULL DEFAULT '{}',
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_audit_target ON user_audit_logs(target_user);

-- Office standard seeds (granted_by='seed'); idempotent via ON CONFLICT.
INSERT INTO permission_grants (subject_type, subject_ref, action, resource_type, granted_by) VALUES
 ('cool_role','leader','cool.members.view','global','seed'),
 ('cool_role','leader','cool.members.add','global','seed'),
 ('cool_role','leader','cool.members.remove','global','seed'),
 ('cool_role','leader','cool.roles.change','global','seed'),
 ('cool_role','leader','cool.edit','global','seed'),
 ('cool_role','leader','cool.meetings.create','global','seed'),
 ('cool_role','leader','cool.attendance.record','global','seed'),
 ('cool_role','leader','cool.attendance.edit','global','seed'),
 ('cool_role','leader','cool.attendance.view','global','seed'),
 ('cool_role','leader','cool.stats.view','global','seed'),
 ('cool_role','core_team','cool.members.view','global','seed'),
 ('cool_role','core_team','cool.meetings.create','global','seed'),
 ('cool_role','core_team','cool.attendance.record','global','seed'),
 ('cool_role','core_team','cool.attendance.view','global','seed'),
 ('cool_role','facilitator','cool.view','global','seed'),
 ('cool_role','facilitator','cool.members.view','global','seed'),
 ('cool_role','facilitator','cool.members.add','global','seed'),
 ('cool_role','facilitator','cool.members.remove','global','seed'),
 ('cool_role','facilitator','cool.roles.change','global','seed'),
 ('cool_role','facilitator','cool.edit','global','seed'),
 ('cool_role','facilitator','cool.meetings.create','global','seed'),
 ('cool_role','facilitator','cool.attendance.record','global','seed'),
 ('cool_role','facilitator','cool.attendance.edit','global','seed'),
 ('cool_role','facilitator','cool.attendance.view','global','seed'),
 ('cool_role','facilitator','cool.stats.view','global','seed'),
 ('role','gco-admin','cool.manage_all','global','seed'),
 ('role','gco-admin','cool.create','global','seed'),
 ('role','gco-admin','event.manage_all','global','seed'),
 ('role','event-admin','event.manage_all','global','seed')
ON CONFLICT DO NOTHING;
```

Down: `DROP TABLE IF EXISTS permission_grants;` (audit table intentionally NOT dropped — shared).

- [ ] **Step 2: Harness** — same shape as the events plan Task 1 harness: connect via `TEST_DATABASE_URL` or skip; execute both new migration files statement-by-statement (errors on already-exists ignored by IF NOT EXISTS/ON CONFLICT); per-test cleanup: `DELETE FROM permission_grants WHERE granted_by <> 'seed'`, `TRUNCATE user_audit_logs, cool_meeting_guests, cool_meeting_attendance, cool_meetings, cool_facilitators, cool_members CASCADE`, `DELETE FROM cools WHERE name LIKE 'test-%'`, `DELETE FROM users WHERE community_id LIKE '99%'` (test fixtures use `99…` community ids and `test-` cool names).
- [ ] **Step 3:** `go vet ./tests/integration/coolmgmt/` → OK; with DB: seeds present (`SELECT count(*) FROM permission_grants WHERE granted_by='seed'` ≥ 29).
- [ ] **Step 4:** Commit `feat(perm): permission_grants schema, shared audit table and office-standard seeds`.

---

### Task 3: Permission engine (pure evaluation + cache)

**Files:**
- Create: `internal/pkg/permission/permission.go`
- Test: `internal/pkg/permission/permission_test.go`

**Interfaces:**
- Produces:

```go
type Actor struct {
	CommunityID        string
	Roles, UserTypes   []string
	CoolID             int64  // 0 = none
	CoolRole           string // member|core_team|leader|""
	FacilitatedCoolIDs []int64
	LedDepartments     []string // reserved for user-mgmt actions
}
type Resource struct{ Type, Ref string } // "global"|"cool"|"event"
func GlobalRes() Resource
func CoolRes(id int64) Resource   // {Type:"cool", Ref: strconv...}
func EventRes(code string) Resource
type Grant struct {
	ID int64; SubjectType, SubjectRef, Action, ResourceType string
	ResourceRef *string; Scope string; ScopeRefs []string
}
type Store interface{ LoadAll(ctx context.Context) ([]Grant, error) }
func NewEngine(store Store, ttl time.Duration) *Engine
func (e *Engine) Can(ctx context.Context, a Actor, action string, res Resource) (bool, error)
func (e *Engine) Invalidate()
```

- Umbrellas: `cool.*` actions also satisfied by `cool.manage_all`; `event.*` by `event.manage_all`; `user.*` by nothing (explicit). Superadmin user_type ⇒ always true.
- Matching rules (exactly the spec §2.1): cool_role subject matches when the actor holds that cool_role — leader/core_team via `CoolRole`+`CoolID`, facilitator via `FacilitatedCoolIDs`; a global-resource cool_role grant applies only to `CoolRes` of the actor's own cool(s); a pinned grant (`resource_ref` set) applies only to exactly that resource AND still requires subject match; role/user subjects with scope `all` match any resource of the grant's implied domain, `own_cool` restricts to the actor's cool, `list` to `ScopeRefs`.

- [ ] **Step 1: Failing matrix test** (in-memory fake Store):

```go
func TestEngineMatrix(t *testing.T) {
	grants := []Grant{
		{SubjectType: "cool_role", SubjectRef: "leader", Action: "cool.members.add", ResourceType: "global"},
		{SubjectType: "cool_role", SubjectRef: "core_team", Action: "cool.attendance.record", ResourceType: "global"},
		{SubjectType: "cool_role", SubjectRef: "core_team", Action: "cool.members.add", ResourceType: "cool", ResourceRef: sptr("12")},
		{SubjectType: "cool_role", SubjectRef: "facilitator", Action: "cool.members.add", ResourceType: "global"},
		{SubjectType: "role", SubjectRef: "gco-admin", Action: "cool.manage_all", ResourceType: "global"},
		{SubjectType: "user", SubjectRef: "777", Action: "event.checkin", ResourceType: "event", ResourceRef: sptr("XMAS25")},
	}
	e := NewEngine(fakeStore(grants), time.Minute)
	leader5 := Actor{CommunityID: "1", CoolID: 5, CoolRole: "leader"}
	core5 := Actor{CommunityID: "2", CoolID: 5, CoolRole: "core_team"}
	core12 := Actor{CommunityID: "3", CoolID: 12, CoolRole: "core_team"}
	fac := Actor{CommunityID: "4", FacilitatedCoolIDs: []int64{5, 9}}
	admin := Actor{CommunityID: "5", Roles: []string{"gco-admin"}}
	super := Actor{CommunityID: "6", UserTypes: []string{"superadmin"}}
	scanner := Actor{CommunityID: "777"}

	cases := []struct {
		name   string
		a      Actor
		action string
		res    Resource
		want   bool
	}{
		{"leader adds in own cool", leader5, "cool.members.add", CoolRes(5), true},
		{"leader cannot add in other cool", leader5, "cool.members.add", CoolRes(9), false},
		{"core cannot add by default", core5, "cool.members.add", CoolRes(5), false},
		{"core CAN record attendance", core5, "cool.attendance.record", CoolRes(5), true},
		{"per-cool override: cool 12 core adds", core12, "cool.members.add", CoolRes(12), true},
		{"override does not leak to cool 5", core5, "cool.members.add", CoolRes(5), false},
		{"facilitator adds in facilitated cool", fac, "cool.members.add", CoolRes(9), true},
		{"facilitator not elsewhere", fac, "cool.members.add", CoolRes(12), false},
		{"admin umbrella covers everything cool", admin, "cool.attendance.edit", CoolRes(12), true},
		{"superadmin bypasses all", super, "user.manage", GlobalRes(), true},
		{"per-event scanner at XMAS25", scanner, "event.checkin", EventRes("XMAS25"), true},
		{"scanner not at other events", scanner, "event.checkin", EventRes("EASTER"), false},
	}
	for _, c := range cases {
		got, err := e.Can(context.Background(), c.a, c.action, c.res)
		if err != nil || got != c.want {
			t.Errorf("%s: got %v want %v (err %v)", c.name, got, c.want, err)
		}
	}
}

func TestCacheAndInvalidate(t *testing.T) {
	fs := &countingStore{}
	e := NewEngine(fs, time.Minute)
	e.Can(context.Background(), Actor{}, "cool.view", GlobalRes())
	e.Can(context.Background(), Actor{}, "cool.view", GlobalRes())
	if fs.loads != 1 {
		t.Fatalf("second call must hit cache, loads=%d", fs.loads)
	}
	e.Invalidate()
	e.Can(context.Background(), Actor{}, "cool.view", GlobalRes())
	if fs.loads != 2 {
		t.Fatal("invalidate must force reload")
	}
}
```

- [ ] **Step 2:** Run → FAIL. **Step 3: Implement:**

```go
package permission

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"
)

// (Actor, Resource, Grant, Store declared as in Interfaces block)

func GlobalRes() Resource          { return Resource{Type: "global"} }
func CoolRes(id int64) Resource    { return Resource{Type: "cool", Ref: strconv.FormatInt(id, 10)} }
func EventRes(code string) Resource { return Resource{Type: "event", Ref: code} }

type Engine struct {
	store   Store
	ttl     time.Duration
	mu      sync.Mutex
	grants  []Grant
	loaded  time.Time
}

func NewEngine(store Store, ttl time.Duration) *Engine { return &Engine{store: store, ttl: ttl} }

func (e *Engine) Invalidate() { e.mu.Lock(); e.loaded = time.Time{}; e.mu.Unlock() }

func (e *Engine) load(ctx context.Context) ([]Grant, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if time.Since(e.loaded) < e.ttl && e.grants != nil {
		return e.grants, nil
	}
	g, err := e.store.LoadAll(ctx)
	if err != nil {
		return nil, err
	}
	e.grants, e.loaded = g, time.Now()
	return g, nil
}

func umbrellas(action string) []string {
	out := []string{action}
	if strings.HasPrefix(action, "cool.") {
		out = append(out, "cool.manage_all")
	}
	if strings.HasPrefix(action, "event.") {
		out = append(out, "event.manage_all")
	}
	return out
}

func (e *Engine) Can(ctx context.Context, a Actor, action string, res Resource) (bool, error) {
	for _, ut := range a.UserTypes {
		if ut == "superadmin" {
			return true, nil
		}
	}
	grants, err := e.load(ctx)
	if err != nil {
		return false, err
	}
	acts := umbrellas(action)
	for _, g := range grants {
		if !containsStr(acts, g.Action) {
			continue
		}
		if e.subjectMatches(a, g) && e.resourceMatches(a, g, res) {
			return true, nil
		}
	}
	return false, nil
}

func (e *Engine) subjectMatches(a Actor, g Grant) bool {
	switch g.SubjectType {
	case "user":
		return g.SubjectRef == a.CommunityID
	case "role":
		return containsStr(a.Roles, g.SubjectRef)
	case "cool_role":
		if g.SubjectRef == "facilitator" {
			return len(a.FacilitatedCoolIDs) > 0
		}
		return a.CoolRole == g.SubjectRef && a.CoolID != 0
	}
	return false
}

func (e *Engine) actorCoolRefs(a Actor, coolRole string) []string {
	if coolRole == "facilitator" {
		out := make([]string, 0, len(a.FacilitatedCoolIDs))
		for _, id := range a.FacilitatedCoolIDs {
			out = append(out, strconv.FormatInt(id, 10))
		}
		return out
	}
	if a.CoolID != 0 {
		return []string{strconv.FormatInt(a.CoolID, 10)}
	}
	return nil
}

func (e *Engine) resourceMatches(a Actor, g Grant, res Resource) bool {
	// pinned grant: exact resource, regardless of subject kind
	if g.ResourceRef != nil {
		return g.ResourceType == res.Type && *g.ResourceRef == res.Ref
	}
	switch g.SubjectType {
	case "cool_role":
		// global cool_role grants apply only within the actor's own cool(s)
		return res.Type == "cool" && containsStr(e.actorCoolRefs(a, g.SubjectRef), res.Ref)
	default: // role | user with global resource
		switch g.Scope {
		case "", "all":
			return true
		case "own_cool":
			return res.Type == "cool" && containsStr(e.actorCoolRefs(a, ""), res.Ref)
		case "list":
			return containsStr(g.ScopeRefs, res.Ref)
		}
	}
	return false
}

func containsStr(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4:** Run → PASS. Commit `feat(perm): resource-scoped permission engine with umbrellas, cool_role subjects and cache`.

---

### Task 4: Migration 000024 — cool tables, backfill, v_cool_roster

**Files:**
- Create: `tests/integration/db/migrations/000024_cool_management.up.sql`, `.down.sql`
- Test: `tests/integration/coolmgmt/backfill_test.go`

- [ ] **Step 1: Up migration**

```sql
SET TIME ZONE 'Asia/Jakarta';

ALTER TABLE "cools" ADD COLUMN IF NOT EXISTS "area" VARCHAR(50);
ALTER TABLE "cools" ADD COLUMN IF NOT EXISTS "region" VARCHAR(100);

CREATE TABLE IF NOT EXISTS "cool_members" (
    "id" BIGSERIAL PRIMARY KEY,
    "community_id" VARCHAR(15) UNIQUE NOT NULL REFERENCES users(community_id),
    "cool_id" BIGINT NOT NULL REFERENCES cools(id),
    "role" VARCHAR(10) NOT NULL CHECK (role IN ('member','core_team','leader')),
    "joined_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "added_by" VARCHAR(15) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_cool_members_cool ON cool_members(cool_id);

CREATE TABLE IF NOT EXISTS "cool_facilitators" (
    "id" BIGSERIAL PRIMARY KEY,
    "cool_id" BIGINT NOT NULL REFERENCES cools(id),
    "community_id" VARCHAR(15) NOT NULL,
    UNIQUE (cool_id, community_id)
);

CREATE TABLE IF NOT EXISTS "cool_meetings" (
    "id" BIGSERIAL PRIMARY KEY,
    "code" VARCHAR(20) UNIQUE NOT NULL,
    "cool_id" BIGINT NOT NULL REFERENCES cools(id),
    "name" VARCHAR(255),
    "topic" VARCHAR(255) NOT NULL,
    "meeting_date" DATE NOT NULL,
    "start_time" TIME NOT NULL,
    "end_time" TIME NOT NULL,
    "location_type" VARCHAR(7) NOT NULL CHECK (location_type IN ('online','offline','hybrid')),
    "location_name" VARCHAR(255),
    "created_by" VARCHAR(15) NOT NULL,
    "status" VARCHAR(10) NOT NULL DEFAULT 'scheduled',
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_cool_meetings_cool ON cool_meetings(cool_id, meeting_date);

CREATE TABLE IF NOT EXISTS "cool_meeting_attendance" (
    "id" BIGSERIAL PRIMARY KEY,
    "meeting_id" BIGINT NOT NULL REFERENCES cool_meetings(id),
    "community_id" VARCHAR(15) NOT NULL,
    "status" VARCHAR(7) NOT NULL CHECK (status IN ('present','absent')),
    "reason" TEXT,
    "recorded_by" VARCHAR(15) NOT NULL,
    "revisions" JSONB NOT NULL DEFAULT '[]',
    UNIQUE (meeting_id, community_id)
);

CREATE TABLE IF NOT EXISTS "cool_meeting_guests" (
    "id" BIGSERIAL PRIMARY KEY,
    "meeting_id" BIGINT NOT NULL REFERENCES cool_meetings(id),
    "name" VARCHAR(100) NOT NULL,
    "phone" VARCHAR(15),
    "notes" TEXT
);

-- Backfill: leaders win over core over plain members (insert order + DO NOTHING on the UNIQUE).
INSERT INTO cool_members (community_id, cool_id, role, added_by)
SELECT unnest(leader_community_ids), id, 'leader', 'backfill' FROM cools
ON CONFLICT (community_id) DO NOTHING;
INSERT INTO cool_members (community_id, cool_id, role, added_by)
SELECT unnest(core_community_ids), id, 'core_team', 'backfill' FROM cools
WHERE core_community_ids IS NOT NULL
ON CONFLICT (community_id) DO NOTHING;
INSERT INTO cool_members (community_id, cool_id, role, added_by)
SELECT u.community_id, u.cool_id, 'member', 'backfill' FROM users u
WHERE u.cool_id IS NOT NULL AND u.cool_id <> 0 AND u.deleted_at IS NULL
ON CONFLICT (community_id) DO NOTHING;
INSERT INTO cool_facilitators (cool_id, community_id)
SELECT id, unnest(facilitator_community_ids) FROM cools
ON CONFLICT (cool_id, community_id) DO NOTHING;

-- Guard against orphan community_ids in legacy arrays that no longer exist in users:
DELETE FROM cool_members cm WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.community_id = cm.community_id);

CREATE OR REPLACE VIEW v_cool_roster AS
SELECT c.name AS cool_name, u.name AS member_name, cm.role, u.phone_number, cm.joined_at
FROM cool_members cm
JOIN cools c ON c.id = cm.cool_id
JOIN users u ON u.community_id = cm.community_id
ORDER BY c.name, cm.role, u.name;
```

Note: the backfill INSERT of leaders can violate the users FK for stale array entries — the FK is why the `DELETE ... NOT EXISTS` guard exists; to survive insert-time FK failures on dirty data, the leader/core inserts must be wrapped as `INSERT ... SELECT x FROM (SELECT unnest(...) AS cid, id FROM cools) t JOIN users u ON u.community_id = t.cid` — implement them in that JOIN form (the plan's simple form above is the intent; the JOIN form is the required implementation).

Down: drop view + the five new tables; `ALTER cools DROP COLUMN IF EXISTS area, DROP COLUMN IF EXISTS region;`

- [ ] **Step 2: Backfill integration test** — seed 2 users (`990000000001` leader in arrays, `990000000002` with `users.cool_id` only) + a `test-alpha` cool holding those arrays; re-run harness migration; assert `cool_members` has leader row for ...01 and member row for ...02; assert `v_cool_roster` returns both with names; running the migration a second time changes nothing (idempotent).
- [ ] **Step 3:** Run → PASS. Commit `feat(cool): normalized membership schema, idempotent backfill and v_cool_roster view`.

---

### Task 5: Entities, repositories, ActorLoader, aggregate wiring

**Files:**
- Create: `internal/models/coolmgmt/entities.go`, `internal/repositories/pgsql/coolmgmt_repositories.go`
- Modify: `internal/repositories/pgsql/main_pg_repository.go` (add `CoolMgmt *CoolMgmtRepositories`)
- Test: `tests/integration/coolmgmt/repositories_test.go`

**Interfaces:**
- Entities (GORM, table names matching Task 4): `CoolMember`, `CoolFacilitator`, `CoolMeeting`, `MeetingAttendance` (field `Revisions []byte gorm:"type:jsonb"`), `MeetingGuest`, `AuditLog` (table `user_audit_logs`), `GrantRow` (table `permission_grants`).
- `CoolMgmtRepositories` fields + key methods:

```go
Members interface {
	Get(ctx, communityID string) (*coolmgmt.CoolMember, error)      // gorm.ErrRecordNotFound
	ListByCool(ctx, coolID int64) ([]coolmgmt.CoolMember, error)
	Insert(ctx, m *coolmgmt.CoolMember) error                        // unique violation → caller maps to ErrAlreadyInCool
	UpdateRole(ctx, communityID, role string) error
	Delete(ctx, communityID string) error
	CountLeaders(ctx, coolID int64) (int64, error)
	GrowthByMonth(ctx, coolID int64) ([]coolmgmt.MonthCount, error)  // joined_at buckets; coolID 0 = all
}
Facilitators interface {
	ListCoolIDs(ctx, communityID string) ([]int64, error)
	Set(ctx, coolID int64, communityIDs []string) error              // replace-all, admin path
	ListByCool(ctx, coolID int64) ([]string, error)
}
Meetings interface {
	Create(ctx, m *coolmgmt.CoolMeeting) error
	GetByCode(ctx, code string) (*coolmgmt.CoolMeeting, error)
	ListByCool(ctx, coolID int64, from, to time.Time) ([]coolmgmt.CoolMeeting, error)
}
Attendance interface {
	BulkInsert(ctx, rows []coolmgmt.MeetingAttendance) error
	ListByMeeting(ctx, meetingID int64) ([]coolmgmt.MeetingAttendance, error)
	HasAny(ctx, meetingID int64) (bool, error)
	Update(ctx, row *coolmgmt.MeetingAttendance) error
	MemberCounts(ctx, coolID int64, from, to time.Time) ([]coolmgmt.MemberAttendanceCount, error) // attended, total meetings per member
	MonthlyAttended(ctx, coolID int64, from, to time.Time) ([]coolmgmt.MemberMonthAttended, error) // per member per month, for quarterly math
}
Guests interface{ BulkInsert(ctx, rows []coolmgmt.MeetingGuest) error; ListByMeeting(ctx, meetingID int64) ([]coolmgmt.MeetingGuest, error) }
Grants interface {
	permission.Store                                   // LoadAll for the engine
	List(ctx, resourceType, resourceRef string) ([]coolmgmt.GrantRow, error)
	Insert(ctx, g *coolmgmt.GrantRow) error
	Delete(ctx, id int64) error
}
Audit interface{ Append(ctx, l *coolmgmt.AuditLog) error; List(ctx, filter coolmgmt.AuditFilter) ([]coolmgmt.AuditLog, error) }
LoadActor(ctx, communityID string) (permission.Actor, error)  // joins users (roles/user_types) + cool_members + cool_facilitators
LegacyMirror(ctx, tx *PostgreRepositories, coolID int64) error // recompute users.cool_id + the three arrays for one cool from cool_members/facilitators
```

- [ ] **Step 1: Failing integration test** — insert member via repo; duplicate insert returns error mapped from unique violation; `LoadActor` for a seeded leader yields `CoolRole=leader, CoolID=<id>`; for a facilitator yields `FacilitatedCoolIDs`; `LegacyMirror` rewrites `users.cool_id` and cool arrays to match `cool_members` exactly.
- [ ] **Step 2:** FAIL. **Step 3: Implement** (thin GORM; `LegacyMirror` = three UPDATEs built from `SELECT array_agg(...) FILTER (WHERE role='leader') ...` over cool_members + facilitators). **Step 4:** PASS. **Step 5:** Commit `feat(cool): repositories, actor loader and legacy mirror`.

---

### Task 6: Grants usecase + API (the future permissions UI backend)

**Files:**
- Create: `internal/usecases/coolmgmt/grant_usecase.go`, `internal/usecases/coolmgmt/usecases.go`
- Test: `tests/integration/coolmgmt/grants_test.go`

**Interfaces:**
- `GrantUsecase{r, engine}`: `List(ctx, actor, resourceType, resourceRef)` (requires `permission.manage` OR — sheets-style — `cool.manage_all` for cool resources / `event.manage_all` for event resources), `Add(ctx, actor, GrantInput) (*GrantRow, error)`, `Remove(ctx, actor, id int64) error`. Every Add/Remove: validates action against the code catalogs (unknown action → `ErrInvalidInput`), appends audit (`permission.grant_added`/`removed` with the grant as `changes`), calls `engine.Invalidate()`.
- `usecases.go`: `type CoolMgmt struct{ Engine *permission.Engine; Grant *GrantUsecase; Member *MemberUsecase; Cool *CoolUsecase; Meeting *MeetingUsecase; Attendance *AttendanceUsecase; Stats *StatsUsecase }` + `New(r *pgsql.PostgreRepositories, cfg *config.Configuration) *CoolMgmt` (engine TTL 60s over `r.CoolMgmt.Grants`). Modify `internal/usecases/main_usecase.go`: add `CoolMgmt *coolmgmt.CoolMgmt` built in `New`.

- [ ] Steps: failing test (add grant as gco-admin for a cool resource succeeds + audit row + a previously-denied core actor now allowed — proving Invalidate works; add with unknown action fails INVALID_INPUT; remove restores denial) → implement → PASS → commit `feat(perm): grants usecase with catalog validation, audit and cache invalidation`.

---

### Task 7: Member management (add / remove / promote, mirrors, audit)

**Files:**
- Create: `internal/usecases/coolmgmt/member_usecase.go`
- Test: `tests/integration/coolmgmt/members_test.go`

**Interfaces:**
- `MemberUsecase`: `Add(ctx, actor, coolID int64, targetCommunityID, role string) error` (role ∈ member|core_team|leader; permission: `cool.members.add`, and `cool.roles.change` additionally required when role != member), `Remove(ctx, actor, coolID, target) error` (`cool.members.remove`; removing a leader requires `CountLeaders > 1` else `ErrLastLeader`), `ChangeRole(ctx, actor, coolID, target, newRole) error` (`cool.roles.change`, same last-leader guard when demoting), `Roster(ctx, actor, coolID) ([]RosterEntry, error)` (`cool.members.view`).
- Every mutation inside `Atomic`: membership write + `LegacyMirror(coolID)` (+ old cool's mirror on moves — moves are NOT allowed implicitly: Add on someone with an existing row returns `ErrAlreadyInCool.WithPublic("current_cool", name)`; the client removes first, per the checklist rule) + audit row (`cool.member_added` etc. with `{cool_id, role}` changes).

- [ ] Steps: failing tests (add member happy + audit row + `users.cool_id` mirror updated; add someone already in another cool → `ALREADY_IN_COOL` carrying the other cool's name; core actor denied add under office standard but allowed after Task-6 per-cool override grant; remove last leader → `LAST_LEADER`; promote member→leader works and mirrors arrays) → implement → PASS → commit `feat(cool): member management with one-cool rule, mirrors and audit`.

---

### Task 8: Cool CRUD + scoped lists

**Files:**
- Create: `internal/usecases/coolmgmt/cool_usecase.go`
- Test: `tests/integration/coolmgmt/cools_test.go`

**Interfaces:**
- `CoolUsecase`: `Create(ctx, actor, CreateCoolInput) (*Cool, error)` (`cool.create`; input: name, description, campus, category, gender, recurrence, location type/name, area, region, facilitators[], leaders[], core[] — members created via Task 7 logic reused internally, facilitators via `Facilitators.Set`), `Patch(ctx, actor, coolID, patch map[string]any) error` (`cool.edit` for descriptive fields incl. area/region; the `facilitators` key requires `cool.manage_all` — the §4.1 carve-out), `List(ctx, actor) ([]CoolSummary, error)` (admin: all; facilitator: facilitated; leader/core/member: own — resolved from actor, no post-filtering), `Detail(ctx, actor, coolID)` (`cool.members.view` for roster inclusion), `Mine(ctx, actor)` (member self-view: cool info page).

- [ ] Steps: failing tests (create with leaders seeds cool_members + mirrors; leader patches description OK but patching facilitators → FORBIDDEN; facilitator list sees only theirs; member Mine returns own cool) → implement → PASS → commit `feat(cool): cool crud with facilitator carve-out and engine-scoped lists`.

---

### Task 9: Meetings

**Files:**
- Create: `internal/usecases/coolmgmt/meeting_usecase.go`
- Test: `tests/integration/coolmgmt/meetings_test.go`

**Interfaces:**
- `MeetingUsecase`: `Create(ctx, actor, CreateMeetingInput) (*CoolMeeting, error)` — input: coolID, optional name, **required topic** (else `ErrInvalidInput("topic is required")`), date, start/end (end after start), **required locationType** ∈ online|offline|hybrid, optional locationName, plus `recordAttendanceNow bool` + optional inline attendance payload (delegates to Task 10's `Record` in the same request when provided — "mau attendancenya skrg atau nanti"); permission `cool.meetings.create` on that cool. `List(ctx, actor, coolID, from, to)` (`cool.attendance.view` OR own-cool member for the member-facing upcoming/past lists), `MineUpcoming/MinePast(ctx, actor)`.
- Meeting `code` generated with the existing `generator.GenerateHashCode` (7 chars, `mtg-` prefix).

- [ ] Steps: failing tests (create without topic fails with detail; create with attendance-now creates meeting + attendance rows atomically; member sees upcoming list) → implement → PASS → commit `feat(cool): meetings with attendance-now-or-later`.

---

### Task 10: Attendance — record, edit with revisions, views, window rule

**Files:**
- Create: `internal/usecases/coolmgmt/attendance_usecase.go`
- Test: `tests/integration/coolmgmt/attendance_test.go`

**Interfaces:**
- `AttendanceUsecase`:

```go
type RecordInput struct {
	MeetingCode string
	Entries     []Entry // {CommunityID, Status "present"|"absent", Reason string}
	Guests      []GuestInput // {Name, Phone, Notes}
}
Record(ctx, actor, RecordInput) error       // cool.attendance.record; every roster member must appear exactly once
                                            // (missing/extra → ErrInvalidInput naming the community_id);
                                            // second Record on same meeting → ErrMeetingAlreadyRecorded
Edit(ctx, actor, meetingCode, entries []Entry) error  // cool.attendance.edit; per changed row append
                                            // {at, by, from, to, reason_from, reason_to} to revisions JSONB
ByMeeting(ctx, actor, meetingCode) (*MeetingAttendanceView, error) // rows + guests + totals {present, absent, rate}
MemberRates(ctx, actor, coolID, from, to) ([]MemberRate, error)    // attended/total per member + overall;
                                            // WINDOW RULE: unless Can(actor, "cool.manage_all", global),
                                            // `from` is clamped to now - configdb "cool.attendance_window_months" (default 6);
                                            // a request explicitly beyond it → ErrAttendanceWindowExceeded
```

- [ ] Steps: failing tests (record full roster + guests atomically; roster mismatch names the missing id; double record → MEETING_ALREADY_RECORDED; edit flips absent→present and appends one revision entry with recorder; non-admin range 8 months back → ATTENDANCE_WINDOW_EXCEEDED, admin same range OK; ByMeeting totals correct) → implement → PASS → commit `feat(cool): audited attendance with roster integrity and history window`.

---

### Task 11: Stats — quarterly math (pure) + analytics endpoints

**Files:**
- Create: `internal/pkg/coolstats/coolstats.go`, `internal/usecases/coolmgmt/stats_usecase.go`
- Test: `internal/pkg/coolstats/coolstats_test.go`, `tests/integration/coolmgmt/stats_test.go`

**Interfaces:**
- Pure math:

```go
package coolstats
const NormPerMonth = 3
func MonthlyRate(attended int) float64            // min(attended,3)/3
func QuarterRate(attendedByMonth [3]int) float64  // (Σ min(m,3)) / 9
```

- `StatsUsecase` (`cool.stats.view` scoped / `cool.manage_all` for global):
  - `CoolStats(ctx, actor, coolID, quarterStart time.Time)` → member growth (MonthCount series), per-member quarterly report (`MemberQuarter{CommunityID, Name, Months [3]MonthDetail{Attended, MeetingsHeld, Rate}, QuarterRate}`) using `MonthlyAttended` repo data + pure math.
  - `GlobalStats(ctx, actor)` → cool count over time (from cools.created_at), member spread by category, member growth overall.
  - `DepartmentCrossTab(ctx, actor, from, to)` → attendance rate grouped by users' department (uses `user_departments` when that table exists — feature-detect via `HasTable`; falls back to `users.department` string until the user-mgmt plan lands, with a code comment marking the switch).
  - `NewJoinerFunnel(ctx, actor)` → union of `cool_new_joiners` (source `form`) and `cool_meeting_guests` joined to meetings/cools (source `meeting`), newest first.

- [ ] **Step 1: Pure math failing test:**

```go
func TestQuarterMath(t *testing.T) {
	if MonthlyRate(2) != 2.0/3 || MonthlyRate(3) != 1 || MonthlyRate(5) != 1 || MonthlyRate(0) != 0 {
		t.Fatal("monthly: min(attended,3)/3")
	}
	if got := QuarterRate([3]int{2, 4, 3}); got != (2.0+3+3)/9 {
		t.Fatalf("quarter caps each month at 3: got %f", got)
	}
}
```

- [ ] Steps: pure test → implement → PASS → integration test (seed one cool, 4 meetings in one month, member attends all 4 → month rate 100% not 133%; a 2-meeting month with 2 attends → 66.7%; funnel returns both sources with discriminator) → implement usecase → PASS → commit `feat(cool): analytics with min(attended,3)/3 quarterly normalization and joiner funnel`.

---

### Task 12: HTTP delivery, config, docs

**Files:**
- Create: `internal/deliveries/http/v2/coolmgmt_handler.go`
- Modify: `internal/deliveries/http/v2/v2_handler.go` (register `NewCoolMgmtHandler(v2, u, c)`), `internal/contract/contract.go` (call `coolmgmt` + other modules' `locale.Install()` once), `config/config.local.template.yaml` (none needed beyond existing), `AGENTS.md` (one-paragraph pointer)
- Test: `tests/integration/coolmgmt/http_smoke_test.go`

**Interfaces:**
- Routes exactly the spec §6 table (`/v2/cools/mine`, `/v2/cools...`, `/v2/meetings/{code}/attendance`, `/v2/cools/{id}/stats`, `/v2/cools/stats`, `/v2/new-joiners`, `/v2/permissions/grants`). Handler shape: bind → `callerActor(ctx)` (builds `permission.Actor` via `r.CoolMgmt.LoadActor` from the JWT community id, request-cached) → usecase → respond. Responder: locale from `Accept-Language` (`en`/`id`, default `id`), `status, body := coolmgmt.Localize(err, locale); ctx.JSONBlob(status, body)`; success `{"code":"SUCCESS","data":...}`.
- Configdb: on boot (contract.go), upsert default `cool.attendance_window_months = "6"` if absent (existing configdb usecase).
- Export `NewCoolMgmtHandlerWithAuth(g, u, c, authMW echo.MiddlewareFunc)` for the smoke test to inject a fake-identity middleware (same pattern as the events plan Task 16).

- [ ] Steps: smoke test (as seeded leader: GET /v2/cools → own cool only; POST members add → 200 then duplicate → 409 ALREADY_IN_COOL with Indonesian message when `Accept-Language: id`; GET stats route 200) → implement handlers + mounting + boot seeding → `go build ./...` + all tests green → commit `feat(cool): http delivery, locale-aware responder and boot config seeding`.

---

## Cross-plan notes (do NOT implement here)

- **Events-v3 plan amendments** (execute with that plan): drop its `event_organizers` table/endpoints; its `requireStaff` and QR `Allowed` callbacks call `Engine.Can(actor, "event.checkin"|..., EventRes(code))`; its `eligibility.Check` gains `CoolID int64`, `CoolRole string`, and rules `cools []int64`, `requires_cool_membership bool`, `cool_roles []string` (loader: `cool_members`).
- **User-mgmt plan**: its grants table IS this one (Task 2 creates it with the resource columns already); its seeds add the `user.*` grants; `user_audit_logs` already exists from Task 2 here.

## Self-Review Notes

- **Spec coverage:** engine+layers (T2,T3,T6), membership normalization+backfill+view+mirrors (T4,T5,T7), facilitator carve-out (T8), meetings incl. now-or-later (T9), attendance+guests+revisions+window (T10), analytics incl. quarterly normalization+cross-tab+funnel (T11), scoped lists+member self-views+grants API+locale (T6,T8,T12). Cross-module amendments deliberately deferred with explicit notes.
- **Type consistency:** `permission.Actor/Resource/Grant` defined once (T3), consumed by T5 LoadActor, T6 engine wiring, T7–T12 checks; `coolstats` consumed only by T11.
- **Known simplification:** member growth uses `joined_at` (backfilled rows all carry migration time — growth charts become meaningful going forward; documented, acceptable).
