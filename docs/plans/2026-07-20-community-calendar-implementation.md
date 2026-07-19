# Community Calendar Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the v3 community calendar (spec: `docs/plans/2026-07-20-community-calendar-design.md`) — role-aware, query-time aggregation of event sessions + COOL meetings into one JSON endpoint and one `.ics` download.

**Architecture:** Read-only domain, no tables, no migrations. Two raw-SQL range queries (`event_sessions ⋈ events`, `cool_meetings ⋈ cools`) merged in the usecase; visibility tiers (member eligibility / organizer overlay / admin) decided server-side via the events-v3 eligibility checker and the unified permission engine. A pure `ical` package serializes the same entries for download.

**Tech Stack:** Go 1.26, Echo v4, GORM raw SQL, `github.com/jeremygprawira/herr`. No new dependencies.

## Global Constraints

- Module path is `go-community`. All imports use it.
- **Build order:** this module compiles against events-v3 (`internal/models/v3`, `internal/pkg/eligibility`, migration 000022) and COOL (`internal/pkg/permission`, `LoadActor`, migrations 000023/000024). Do not start until both are merged.
- Errors use `github.com/jeremygprawira/herr` classes; matching via `.Is(err)`, never `==`; never touch `models.ErrorMapping`. **Bilingual:** every class has English `Public.Message` + Indonesian `mapl` entry under `errors.<lowercase_code>.message` (missing `id` translation = review-rejectable defect).
- Read-only domain: no writes, no `Atomic`, no `Transaction`.
- Time zone: business time composes in `Asia/Jakarta` via `common.GetLocation()`; API JSON timestamps are RFC3339.
- Range cap: **92 days** (spec §5). Exact value lives in one constant `calendar.MaxRangeDays`.
- DB integration tests read `TEST_DATABASE_URL`; unset → `t.Skip`. Local run: `TEST_DATABASE_URL="postgres://postgres:<pw>@localhost:5888/community_db?sslmode=disable"`.
- Unit tests: `go test ./internal/... ./tests/...`. Build check: `go build ./...`.
- Commit after every green cycle. Conventional Commits, `feat(calendar): ...` scope.

## File Structure

```
internal/models/calendar/calendar.go        Entry/Ref/Flags/Query DTOs + Query.Validate + NormalizeLocationType
internal/models/calendar/errors.go          herr classes (INVALID_RANGE, INVALID_FILTER) + localizer entries
internal/pkg/ical/ical.go                   pure serializer: []Event → .ics bytes
internal/repositories/pgsql/calendar_repository.go   CalendarRepository (raw SQL, read-only) + row structs
internal/repositories/pgsql/calendar_pg_query.go     package-level SQL string vars
internal/repositories/pgsql/main_pg_repository.go    (modify) add Calendar field
internal/usecases/calendar/calendar_usecase.go       tiering, merge, sort, filters
internal/usecases/main_usecase.go                    (modify) add Calendar field
internal/deliveries/http/v3/calendar_handler.go      GET /v3/calendar + /v3/calendar/export.ics
internal/deliveries/http/main_handler.go             (modify) mount calendar handler
<permission action catalog file from COOL Task 3/6>  (modify) add "calendar.view_all"
tests/integration/calendar/calendar_test.go          repo + usecase integration tests (reuses v3/COOL harness pattern)
```

---

### Task 1: Models — entry shape, query validation, bilingual errors

**Files:**
- Create: `internal/models/calendar/calendar.go`, `internal/models/calendar/errors.go`
- Test: `internal/models/calendar/calendar_test.go`

**Interfaces:**
- Produces (later tasks depend on these exact names):

```go
package calendar

type Source string
const (
	SourceEventSession Source = "event_session"
	SourceCoolMeeting  Source = "cool_meeting"
)
const MaxRangeDays = 92

type Ref struct {
	EventCode   string `json:"event_code,omitempty"`
	SessionCode string `json:"session_code,omitempty"`
	CoolID      int64  `json:"cool_id,omitempty"`
	MeetingCode string `json:"meeting_code,omitempty"`
}
type Flags struct {
	Registered bool `json:"registered"`
	Organizer  bool `json:"organizer"`
	Draft      bool `json:"draft"`
}
type Entry struct {
	Source       Source    `json:"source"`
	Ref          Ref       `json:"ref"`
	Title        string    `json:"title"`
	ParentTitle  string    `json:"parent_title"`
	StartAt      time.Time `json:"start_at"`
	EndAt        time.Time `json:"end_at"`
	LocationType string    `json:"location_type"` // onsite|online|hybrid
	LocationName string    `json:"location_name,omitempty"`
	OnlineURL    string    `json:"online_url,omitempty"`
	CampusCodes  []string  `json:"campus_codes,omitempty"`
	Status       string    `json:"status"`
	Flags        Flags     `json:"flags"`
}
type Query struct {
	From, To time.Time
	Sources  []Source // empty = both
	Campus   []string
	Mine     bool
}
func (q Query) Validate() error                 // ErrInvalidRange / ErrInvalidFilter
func (q Query) WantsSource(s Source) bool       // empty Sources ⇒ true
func NormalizeLocationType(s string) string     // "offline" → "onsite", else passthrough
func SortEntries(entries []Entry)               // start_at asc, then source, then ref code — deterministic
```

- `errors.go` produces herr classes `ErrInvalidRange` (code `CALENDAR_INVALID_RANGE`, EN "The requested date range is invalid or longer than 92 days.") and `ErrInvalidFilter` (code `CALENDAR_INVALID_FILTER`, EN "One of the calendar filters is not recognized."), plus `mapl` Indonesian entries `errors.calendar_invalid_range.message` = "Rentang tanggal tidak valid atau lebih dari 92 hari." and `errors.calendar_invalid_filter.message` = "Salah satu filter kalender tidak dikenali." — registered in the same localizer wiring the v3 error catalog uses (follow the pattern from events-v3 Task 2; if a completeness test enumerates classes, add these two to it).

- [ ] **Step 1: Write the failing test**

```go
// internal/models/calendar/calendar_test.go
package calendar

import (
	"testing"
	"time"
)

func d(s string) time.Time { t, _ := time.Parse(time.RFC3339, s); return t }

func TestQueryValidate(t *testing.T) {
	ok := Query{From: d("2026-08-01T00:00:00+07:00"), To: d("2026-08-31T00:00:00+07:00")}
	if err := ok.Validate(); err != nil {
		t.Fatalf("valid query rejected: %v", err)
	}
	cases := []struct {
		name string
		q    Query
	}{
		{"zero from", Query{To: d("2026-08-31T00:00:00+07:00")}},
		{"to before from", Query{From: d("2026-08-31T00:00:00+07:00"), To: d("2026-08-01T00:00:00+07:00")}},
		{"span over cap", Query{From: d("2026-01-01T00:00:00+07:00"), To: d("2026-06-01T00:00:00+07:00")}},
	}
	for _, c := range cases {
		if err := c.q.Validate(); !ErrInvalidRange.Is(err) {
			t.Errorf("%s: want ErrInvalidRange, got %v", c.name, err)
		}
	}
	bad := Query{From: ok.From, To: ok.To, Sources: []Source{"room_booking"}}
	if err := bad.Validate(); !ErrInvalidFilter.Is(err) {
		t.Errorf("unknown source: want ErrInvalidFilter, got %v", err)
	}
}

func TestNormalizeLocationType(t *testing.T) {
	if NormalizeLocationType("offline") != "onsite" {
		t.Error("offline should normalize to onsite")
	}
	if NormalizeLocationType("hybrid") != "hybrid" {
		t.Error("hybrid should pass through")
	}
}

func TestSortEntries(t *testing.T) {
	a := Entry{Source: SourceCoolMeeting, Ref: Ref{MeetingCode: "mtg-b"}, StartAt: d("2026-08-02T10:00:00+07:00")}
	b := Entry{Source: SourceEventSession, Ref: Ref{SessionCode: "SES-a"}, StartAt: d("2026-08-02T10:00:00+07:00")}
	c := Entry{Source: SourceEventSession, Ref: Ref{SessionCode: "SES-z"}, StartAt: d("2026-08-01T10:00:00+07:00")}
	list := []Entry{a, b, c}
	SortEntries(list)
	if list[0].Ref.SessionCode != "SES-z" || list[1].Source != SourceEventSession || list[2].Source != SourceCoolMeeting {
		t.Errorf("bad order: %+v", list)
	}
}

func TestWantsSource(t *testing.T) {
	if !(Query{}).WantsSource(SourceEventSession) {
		t.Error("empty sources must want everything")
	}
	q := Query{Sources: []Source{SourceCoolMeeting}}
	if q.WantsSource(SourceEventSession) || !q.WantsSource(SourceCoolMeeting) {
		t.Error("explicit sources must filter")
	}
}
```

- [ ] **Step 2: Run** `go test ./internal/models/calendar/ -v` → FAIL (package does not exist).

- [ ] **Step 3: Implement** `calendar.go` exactly per the Interfaces block. `Validate`: zero `From`/`To` or `!To.After(From)` or `To.Sub(From) > MaxRangeDays*24*time.Hour` → `ErrInvalidRange`; any `Sources` value outside the two constants → `ErrInvalidFilter` (campus codes are validated in the usecase against config, not here). `SortEntries` uses `sort.Slice` on `(StartAt.Unix, Source, Ref.SessionCode+Ref.MeetingCode)`. `errors.go`: two `herr.Define` classes following the v3 errors file layout, + localizer entries.

- [ ] **Step 4: Run** `go test ./internal/models/calendar/ -v` → PASS. Also run the v3 error-catalog completeness test if it globs all herr classes.

- [ ] **Step 5: Commit** `git add internal/models/calendar && git commit -m "feat(calendar): entry model, query validation and bilingual errors"`

---

### Task 2: iCal serializer (pure)

**Files:**
- Create: `internal/pkg/ical/ical.go`
- Test: `internal/pkg/ical/ical_test.go`, `internal/pkg/ical/testdata/golden.ics`

**Interfaces:**
- Produces:

```go
package ical

type Event struct {
	UID       string
	Summary   string
	Location  string
	URL       string
	Start, End time.Time
	Cancelled bool
}
// Render emits a VCALENDAR with one VEVENT per input, DTSTART/DTEND with
// TZID=Asia/Jakarta, CRLF line endings, STATUS:CANCELLED when Cancelled.
func Render(events []Event) []byte
```

- Fixed header: `BEGIN:VCALENDAR`, `VERSION:2.0`, `PRODID:-//go-community//calendar//EN`, `CALSCALE:GREGORIAN`, then a static `VTIMEZONE` block for `Asia/Jakarta` (UTC+7, no DST):

```
BEGIN:VTIMEZONE
TZID:Asia/Jakarta
BEGIN:STANDARD
DTSTART:19700101T000000
TZOFFSETFROM:+0700
TZOFFSETTO:+0700
TZNAME:WIB
END:STANDARD
END:VTIMEZONE
```

- Time format `20060102T150405` after converting with `common.GetLocation()`. Text fields escaped per RFC 5545 (`\`, `;`, `,`, newline).

- [ ] **Step 1: Write the failing golden test**

```go
// internal/pkg/ical/ical_test.go
package ical

import (
	"os"
	"testing"
	"time"
)

func TestRenderGolden(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	events := []Event{
		{
			UID:     "event_session-SES1@go-community",
			Summary: "Christmas Celebration — Worship Night",
			Location: "Main Hall",
			Start:   time.Date(2026, 12, 24, 19, 0, 0, 0, loc),
			End:     time.Date(2026, 12, 24, 21, 0, 0, 0, loc),
		},
		{
			UID:       "cool_meeting-mtg-abc@go-community",
			Summary:   "COOL Jakut 1 — Faith & Work",
			URL:       "https://meet.example.com/x",
			Start:     time.Date(2026, 12, 26, 10, 0, 0, 0, loc),
			End:       time.Date(2026, 12, 26, 12, 0, 0, 0, loc),
			Cancelled: true,
		},
	}
	got := Render(events)
	want, err := os.ReadFile("testdata/golden.ics")
	if err != nil {
		t.Fatalf("write got to testdata/golden.ics on first run: %v\n%s", err, got)
	}
	if string(got) != string(want) {
		t.Errorf("golden mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestEscaping(t *testing.T) {
	e := []Event{{UID: "u@x", Summary: "a;b,c\nd", Start: time.Now(), End: time.Now()}}
	out := string(Render(e))
	if !containsLine(out, "SUMMARY:a\\;b\\,c\\nd") {
		t.Errorf("summary not escaped: %s", out)
	}
}

func containsLine(haystack, needle string) bool {
	for _, l := range splitCRLF(haystack) {
		if l == needle {
			return true
		}
	}
	return false
}
```

(`splitCRLF` = small helper in the test file: `strings.Split(s, "\r\n")`.)

- [ ] **Step 2: Run** `go test ./internal/pkg/ical/ -v` → FAIL. Inspect the printed `got`, verify by eye (header, VTIMEZONE, both VEVENTs, `STATUS:CANCELLED` on the second, `URL:` line present), then save it as `testdata/golden.ics`.

- [ ] **Step 3: Implement** `Render` with a `strings.Builder`, CRLF (`\r\n`) after every line, one `escape(s string)` helper, `DTSTART;TZID=Asia/Jakarta:20061224T190000` style stamps.

- [ ] **Step 4: Run** `go test ./internal/pkg/ical/ -v` → PASS (golden now exists and matches).

- [ ] **Step 5: Commit** `git add internal/pkg/ical && git commit -m "feat(calendar): pure ics serializer with golden test"`

---

### Task 3: Repository — range queries + user facts

**Files:**
- Create: `internal/repositories/pgsql/calendar_repository.go`, `internal/repositories/pgsql/calendar_pg_query.go`
- Modify: `internal/repositories/pgsql/main_pg_repository.go` (add `Calendar CalendarRepository` to `PostgreRepositories`, wire in `New`)
- Test: `tests/integration/calendar/repository_test.go`

**Interfaces:**
- Produces:

```go
type CalendarSessionRow struct {
	EventCode, SessionCode, EventTitle, SessionTitle string
	StartAt, EndAt                                   time.Time
	LocationType, LocationName, OnlineURL            string
	CampusCodes                                      pq.StringArray `gorm:"type:text[]"`
	EventStatus, SessionStatus                       string
	Eligibility                                      []byte // raw jsonb, decoded by usecase
	Registered, Organizer                            bool
}
type CalendarMeetingRow struct {
	CoolID                                   int64
	MeetingCode, Topic, CoolName             string
	MeetingDate                              time.Time
	StartTime, EndTime                       *string // "15:04:05" or nil
	LocationType, LocationName, MeetingStatus string
}
type CalendarRepository interface {
	// admin=true drops the published/organizer visibility predicate (archived always excluded)
	SessionsInRange(ctx context.Context, communityID string, from, to time.Time, admin bool) ([]CalendarSessionRow, error)
	// admin=true drops the membership predicate
	MeetingsInRange(ctx context.Context, communityID string, fromDate, toDate string, admin bool) ([]CalendarMeetingRow, error)
	// one SELECT over users: campus + roles + user_types, for eligibility.Check
	UserFacts(ctx context.Context, communityID string) (eligibility.User, error)
}
```

- SQL in `calendar_pg_query.go` as package-level vars per repo convention:

```sql
-- calendarSessionsQuery ($1 community_id, $2 from, $3 to, $4 admin)
SELECT e.code AS event_code, s.code AS session_code,
       e.title AS event_title, s.title AS session_title,
       s.start_at, s.end_at, s.location_type,
       COALESCE(s.location_name,'') AS location_name,
       COALESCE(s.online_url,'') AS online_url,
       e.campus_codes, e.status AS event_status, s.status AS session_status,
       e.eligibility,
       (r.id IS NOT NULL) AS registered,
       (o.id IS NOT NULL) AS organizer
FROM event_sessions s
JOIN events e ON e.id = s.event_id AND e.deleted_at IS NULL
LEFT JOIN registrations r
       ON r.session_id = s.id AND r.registered_by = $1 AND r.status = 'confirmed'
LEFT JOIN event_organizers o
       ON o.event_id = e.id AND o.community_id = $1
WHERE s.deleted_at IS NULL
  AND s.start_at < $3 AND s.end_at > $2
  AND e.status <> 'archived'
  AND ($4 OR e.status = 'published' OR o.id IS NOT NULL)

-- calendarMeetingsQuery ($1 community_id, $2 from_date, $3 to_date, $4 admin)
SELECT m.cool_id, m.code AS meeting_code, m.topic, c.name AS cool_name,
       m.meeting_date, m.start_time::text AS start_time, m.end_time::text AS end_time,
       m.location_type, COALESCE(m.location_name,'') AS location_name,
       m.status AS meeting_status
FROM cool_meetings m
JOIN cools c ON c.id = m.cool_id
WHERE m.meeting_date BETWEEN $2::date AND $3::date
  AND ($4 OR m.cool_id IN (
        SELECT cool_id FROM cool_members      WHERE community_id = $1
        UNION
        SELECT cool_id FROM cool_facilitators WHERE community_id = $1))
```

- `UserFacts`: `SELECT campus_code, roles, user_types FROM users WHERE community_id = $1` mapped to `eligibility.User{CommunityID, CampusCode, Roles, UserTypes}` (adjust column names to the actual `users` schema at implementation time — check `internal/models/user_model.go`).

- [ ] **Step 1: Write the failing integration test** — reuse the harness pattern from `tests/integration/v3/harness_test.go` (skip without `TEST_DATABASE_URL`, apply migrations, truncate). Seed via plain SQL INSERTs:
  - event `E1` published (campus `["BKS"]`, eligibility `{"audience":"everyone"}`) with session `S1` inside range and session `S2` outside range;
  - event `E2` draft with organizer row for user `111`, session `S3` in range;
  - event `E3` archived, session in range;
  - a confirmed registration for user `111` on `S1`;
  - cool 1 with member `111` and meeting `M1` in range; cool 2 (user not member) with meeting `M2` in range.

  Assertions:

```go
rows, _ := repo.Calendar.SessionsInRange(ctx, "111", from, to, false)
// S1 (registered=true, organizer=false) and S3 (organizer=true, event_status=draft) present; S2 and archived absent
rowsAdmin, _ := repo.Calendar.SessionsInRange(ctx, "999", from, to, true)
// S1 and S3 present for a non-organizer admin; archived still absent
mts, _ := repo.Calendar.MeetingsInRange(ctx, "111", "2026-08-01", "2026-08-31", false)
// M1 present, M2 absent
mtsAdmin, _ := repo.Calendar.MeetingsInRange(ctx, "999", "2026-08-01", "2026-08-31", true)
// M1 and M2 present
facts, _ := repo.Calendar.UserFacts(ctx, "111")
// facts.CommunityID == "111", campus/roles populated from seed
```

- [ ] **Step 2: Run** `TEST_DATABASE_URL=... go test ./tests/integration/calendar/ -run TestCalendarRepository -v` → FAIL.

- [ ] **Step 3: Implement** repository with `db.WithContext(ctx).Raw(query, args...).Scan(&rows)`; wire `Calendar` into `PostgreRepositories` in `pgsql.New`.

- [ ] **Step 4: Run** same command → PASS. Also `go build ./...`.

- [ ] **Step 5: Commit** `git add internal/repositories/pgsql tests/integration/calendar && git commit -m "feat(calendar): read-only range queries for sessions and cool meetings"`

---

### Task 4: Usecase — tiers, merge, filters

**Files:**
- Create: `internal/usecases/calendar/calendar_usecase.go`
- Modify: `internal/usecases/main_usecase.go` (add `Calendar *calendar.Usecase` — package alias `calendaruc` if the model package import collides; construct in `New` with repos, permission engine from the CoolMgmt aggregate, and config)
- Modify: the permission action catalog file created by the COOL plan (Task 3/6 — the fixed in-code catalog the grants usecase validates against): add `"calendar.view_all"`.
- Test: `tests/integration/calendar/usecase_test.go`

**Interfaces:**
- Consumes: `CalendarRepository` (Task 3), `calendar` models (Task 1), `eligibility.Check` (events-v3 Task 4), `permission.Engine.Can` + `permission.Actor` + `GlobalRes()` and `LoadActor` (COOL Tasks 3/5), `config.Configuration.Campus` map.
- Produces:

```go
package calendar // internal/usecases/calendar

type Usecase struct { /* repo aggregate, engine, actor loader, cfg */ }
func New(r *pgsql.PostgreRepositories, engine *permission.Engine, cfg *config.Configuration) *Usecase
// Range validates q, resolves the tier, merges both sources, sorts, returns entries.
func (u *Usecase) Range(ctx context.Context, communityID string, q calmodel.Query) ([]calmodel.Entry, error)
```

Rules (all from spec §§3–7):

1. `q.Validate()` first; then every `q.Campus` code must exist in `cfg.Campus` else `ErrInvalidFilter`.
2. `actor := LoadActor(communityID)`; `admin := actor superadmin || engine.Can(actor, "calendar.view_all", permission.GlobalRes())`.
3. Sessions (when `q.WantsSource(SourceEventSession)`): fetch rows with `admin`; per row decode `Eligibility` once **per event code** (memo map); for non-admin rows of published events where `!row.Organizer`, drop the row unless `eligibility.Check(facts, elig)` (facts from `UserFacts`, loaded once). `Draft` flag = `EventStatus != "published"`. `OnlineURL` included only when `admin || row.Organizer || row.Registered || eligible`; otherwise blanked.
4. Meetings (when wanted): fetch with `admin`; compose `StartAt = meeting_date + start_time` in `common.GetLocation()`; nil/empty `end_time` → `StartAt.Add(2 * time.Hour)`; `NormalizeLocationType` on `location_type`; `Title = topic`, `ParentTitle = cool_name`, `Status = meeting_status`.
5. `q.Mine`: keep only session entries with `Registered` + all meeting entries returned by the **non-admin** membership query (when admin and `mine=true`, run meetings with `admin=false` so "mine" means mine even for admins).
6. `q.Campus`: drop event entries whose `CampusCodes` has no intersection with the filter; meeting entries pass through.
7. Merge, `SortEntries`, return. Empty result → empty slice, not nil.

- [ ] **Step 1: Write the failing integration test** — same harness/seeds as Task 3 plus: user `222` seeded as plain member failing `E2`'s eligibility... use a fourth event `E4` published with `eligibility {"audience":"rules","rules":{"roles":["usher"]}}` and no matching role on `222`; a grant row `{subject_type: "role", subject_ref: "gco-admin", action: "calendar.view_all", resource_type: "global"}` and user `333` with role `gco-admin`. Cases:

```go
// member tier: user 111 sees S1 (registered flag true), S3 (organizer+draft flags), M1; NOT E4's session, NOT M2
// eligibility: user 222 does not see E4's session; sees S1
// online_url: 222's S1 entry carries online_url (eligible); E4 hidden entirely
// admin tier: 333 sees E2 draft session + M2 via calendar.view_all grant
// mine=true for 111: only S1 + M1
// campus filter ["JKT"] for 111: S1 dropped (BKS), M1 kept
// sources=["cool"] : only meetings
// invalid campus code → ErrInvalidFilter; 100-day span → ErrInvalidRange
// cool meeting with NULL end_time → EndAt = StartAt + 2h
```

Write these as table-driven subtests asserting entry presence by `Ref` and flag values.

- [ ] **Step 2: Run** `TEST_DATABASE_URL=... go test ./tests/integration/calendar/ -run TestCalendarUsecase -v` → FAIL.

- [ ] **Step 3: Implement** per rules 1–7. Add `"calendar.view_all"` to the action catalog so the grants API accepts it.

- [ ] **Step 4: Run** same command → PASS. Run full suite `go test ./internal/... ./tests/...` → PASS.

- [ ] **Step 5: Commit** `git add internal/usecases tests/integration/calendar <catalog file> && git commit -m "feat(calendar): role-tiered aggregation usecase with calendar.view_all"`

---

### Task 5: HTTP delivery — JSON + .ics endpoints, mounting, docs

**Files:**
- Create: `internal/deliveries/http/v3/calendar_handler.go`
- Modify: `internal/deliveries/http/main_handler.go` (mount within the existing v3 group)
- Test: extend `tests/integration/calendar/usecase_test.go` with a handler-level test using `httptest` (or the repo's existing handler test pattern if one exists by then)

**Interfaces:**
- Consumes: `Usecase.Range` (Task 4), `ical.Render` (Task 2), `middleware.UserMiddleware` (token-holder group, no role restriction — tiering is inside the usecase).
- Routes registered by `NewCalendarHandler(group *echo.Group, uc *calendar.Usecase)`:

```
GET /api/v3/calendar
GET /api/v3/calendar/export.ics
```

Param parsing (shared helper `parseCalendarQuery(ctx echo.Context) (calmodel.Query, error)`):

- `from`/`to`: try RFC3339, fall back to `2006-01-02` interpreted midnight `Asia/Jakarta` (a date-only `to` means end of that day: add 24h). Missing → `ErrInvalidRange`.
- `sources`: CSV, map `events`→`SourceEventSession`, `cool`→`SourceCoolMeeting`, anything else → `ErrInvalidFilter`.
- `campus`: CSV passthrough (validated in usecase). `mine`: `strconv.ParseBool`, absent = false.
- Community ID comes from the JWT claims exactly the way other v3 handlers read it.

JSON handler: `response.Success` with `{"entries": [...]}`. Errors go through the v3 herr error responder (never `models.ErrorMapping`).

ICS handler: same query pipeline, then map entries → `[]ical.Event` (`UID = fmt.Sprintf("%s-%s@go-community", e.Source, refCode(e))` where `refCode` = SessionCode or MeetingCode; `Summary = ParentTitle + " — " + Title`; `Cancelled = e.Status == "cancelled"`), then:

```go
res := ctx.Response()
res.Header().Set(echo.HeaderContentType, "text/calendar; charset=utf-8")
res.Header().Set(echo.HeaderContentDisposition, `attachment; filename="community-calendar.ics"`)
return ctx.Blob(http.StatusOK, "text/calendar; charset=utf-8", ical.Render(events))
```

Swagger annotations (`@Summary`, `@Router /api/v3/calendar [get]`, `@Param from query string true ...`, etc.) above both handlers; regenerate `docs/` with `make generate-docs`.

- [ ] **Step 1: Write the failing handler test** — spin the Echo group with a stubbed auth context (repo's existing pattern), seeded DB from Task 4:

```go
// GET /api/v3/calendar?from=2026-08-01&to=2026-08-31 as user 111 → 200, entries JSON contains S1 with flags.registered=true
// GET without from → 400, body carries code CALENDAR_INVALID_RANGE (and Accept-Language: id yields the Indonesian message)
// GET /api/v3/calendar/export.ics same range → 200, Content-Type text/calendar, body contains "BEGIN:VCALENDAR" and S1's UID line
```

- [ ] **Step 2: Run** `TEST_DATABASE_URL=... go test ./tests/integration/calendar/ -run TestCalendarHandler -v` → FAIL.

- [ ] **Step 3: Implement** handler + mounting per Interfaces block.

- [ ] **Step 4: Run** handler test → PASS. Full suite `go test ./internal/... ./tests/...` → PASS. `go build ./...` → OK. `make generate-docs` → regenerates cleanly.

- [ ] **Step 5: Commit** `git add internal/deliveries docs && git commit -m "feat(calendar): json and ics endpoints mounted under /api/v3"`

---

## Self-Review Notes

- Spec coverage: §2 approach → Tasks 3–4; §3 entry shape → Task 1; §4 tiers → Task 4 (rules 2–3, `calendar.view_all` catalog entry); §5 API/params/92-day cap → Tasks 1 & 5; §6 data access → Task 3; §7 errors & edge cases → Tasks 1, 4 (cancelled visibility, NULL end_time, online_url gating), 5 (bilingual 400s); §8 testing → each task's test steps (tier matrix in Task 4, golden ics in Task 2, range/soft-delete/flags in Task 3).
- Department filter: intentionally absent (spec §5) — no task, by design.
- Type consistency: `calmodel` alias used wherever the usecase package name `calendar` would collide with the model package import.
