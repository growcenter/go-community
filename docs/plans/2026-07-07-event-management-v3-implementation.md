# Event Management v3 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the v3 event management module (spec: `docs/plans/2026-07-07-event-management-v3-design.md`) — flexible events/sessions, 4 check-in modes + checkout, universal QR subsystem, geolocation policy, DB-backed email outbox with PDF tickets, recurrence, templates, CSV/XLSX reports — under `/api/v3` alongside v2.

**Architecture:** Relational spine + JSONB configs. New tables; entities store configs as `json.RawMessage` (jsonb), decoded via typed structs in `internal/models/v3`. Pure validator packages (`eligibility`, `geo`, `form`, `recurrence`, `qr`) are unit-tested without DB. Usecases take the `PostgreRepositories` aggregate and use the `Atomic` tx helper only. HTTP layer mounts v3 beside v1/v2 with existing middleware.

**Tech Stack:** Go 1.23, Echo v4, GORM/Postgres, `excelize` (already in go.mod), new deps: `github.com/skip2/go-qrcode`, `github.com/jung-kurt/gofpdf`.

## Global Constraints

- Module path is `go-community`. All imports use it.
- v3 code NEVER uses `TransactionRepository.Transaction` — only `Atomic` (known scoping bug, see `wiki/entities/competing-transaction-abstractions.md`).
- v3 errors use `github.com/jeremygprawira/herr` end to end (Task 2). Domain errors are `herr.Define` classes in package `v3`; usecases return `error`; matching uses `v3.ErrX.Is(err)`, NEVER `==`. Never add to `models.ErrorMapping`.
- **Bilingual messages:** every error class has a humanized English message (the class `Public.Message`) AND an Indonesian translation in the `mapl` localizer under key `errors.<lowercase_code>.message`. Default locale `id`; `Accept-Language: en` switches to English. Adding an error without its `id` translation is a review-rejectable defect (Task 2's completeness test enforces it).
- **Go toolchain:** herr declares `go 1.26.1`; Task 2 bumps go-community's `go.mod` to `go 1.26` (verify CI/dev machines have Go ≥1.26 before starting).
- Every usecase method body starts with `defer func() { usecases.LogService(ctx, err) }()` pattern is NOT reused in v3; v3 handlers log via existing middleware. Keep v3 free of the double-logging pattern.
- Time zone: store timestamptz; business "today" uses `Asia/Jakarta` via `common.GetLocation()`.
- All timestamps in API JSON are RFC3339.
- DB integration tests read `TEST_DATABASE_URL`; when unset they `t.Skip`. Run them locally with the docker-compose Postgres: `TEST_DATABASE_URL="postgres://postgres:<pw>@localhost:5888/community_db?sslmode=disable"`.
- Run all unit tests: `go test ./internal/... ./tests/...`. Build check: `go build ./...`.
- Config keys added to `config/config.local.yaml` (gitignored) AND documented in `config/config.local.template.yaml`.
- Commit after every green test cycle. Conventional Commits, `feat(v3): ...` scope.

## File Structure

```
tests/integration/db/migrations/000022_v3_event_management.up.sql   all v3 tables
tests/integration/db/migrations/000022_v3_event_management.down.sql
internal/models/v3/errors.go            typed Error + sentinel instances
internal/models/v3/config.go            Eligibility/RegistrationConfig/GeoConfig/Recurrence/Contacts/... + Parse helpers + strict validation
internal/models/v3/entities.go          GORM entities (EventV3, SessionV3, TicketCategory, RegistrationV3, AttendeeV3, AttendanceLog, OutboxMessage, OrganizerV3, TemplateV3)
internal/models/v3/dto.go               request/response DTOs + computed-state builder
internal/pkg/eligibility/eligibility.go pure Check()
internal/pkg/geo/geo.go                 Haversine + Validate()
internal/pkg/form/form.go               Visible()/Validate()/ValidateSchema()
internal/pkg/recurrence/recurrence.go   Expand()
internal/pkg/qr/qr.go                   token codec (HMAC) + action registry
internal/pkg/qr/png.go                  PNG rendering
internal/pkg/notify/notify.go           Notifier interface + SMTP email impl
internal/pkg/pdfticket/ticket.go        PDF generation
internal/repositories/pgsql/v3_repositories.go  all v3 repos (one file, small methods) + wiring into PostgreRepositories
internal/usecases/v3/event_usecase.go   event/session/category CRUD, publish, templates, organizers
internal/usecases/v3/registration_usecase.go   register/cancel/edit answers
internal/usecases/v3/checkin_usecase.go check-in/checkout, walk-in, logs
internal/usecases/v3/qr_usecase.go      resolve/act on top of qr registry
internal/usecases/v3/notification_usecase.go   outbox enqueue/dispatch
internal/usecases/v3/recurrence_usecase.go     session generation + publish-window flips
internal/usecases/v3/report_usecase.go  CSV/XLSX
internal/usecases/v3/usecases.go        V3 aggregate + New()
internal/deliveries/http/v3/*.go        handlers + route mounting
internal/config/config.go               add V3 config block (modify)
internal/usecases/main_usecase.go       add V3 field (modify)
internal/deliveries/http/main_handler.go  mount v3 (modify)
tests/integration/v3/harness_test.go    DB test harness (migration apply + truncate)
tests/integration/v3/*_test.go          integration tests per flow
```

---

### Task 1: v3 database migration

**Files:**
- Create: `tests/integration/db/migrations/000022_v3_event_management.up.sql`
- Create: `tests/integration/db/migrations/000022_v3_event_management.down.sql`
- Create: `tests/integration/v3/harness_test.go`

**Interfaces:**
- Produces: tables `v3_events`, `v3_sessions`, `v3_session_ticket_categories`, `v3_registrations`, `v3_registration_attendees`, `v3_attendance_logs`, `v3_notification_outbox`, `v3_event_organizers`, `v3_event_templates`; test helper `func testDB(t *testing.T) *gorm.DB` (skips without `TEST_DATABASE_URL`, applies migration, truncates v3 tables).

- [ ] **Step 1: Write the up migration**

```sql
-- tests/integration/db/migrations/000022_v3_event_management.up.sql
SET TIME ZONE 'Asia/Jakarta';

CREATE TABLE "v3_events" (
    "id" BIGSERIAL PRIMARY KEY,
    "code" VARCHAR(12) UNIQUE NOT NULL,
    "slug" VARCHAR(255) UNIQUE NOT NULL,
    "title" VARCHAR(255) NOT NULL,
    "description" TEXT NOT NULL DEFAULT '',
    "topics" TEXT[] NOT NULL DEFAULT '{}',
    "terms_and_conditions" TEXT NOT NULL DEFAULT '',
    "image_links" TEXT[] NOT NULL DEFAULT '{}',
    "campus_codes" TEXT[] NOT NULL DEFAULT '{}',
    "status" VARCHAR(10) NOT NULL DEFAULT 'draft',
    "publish_at" TIMESTAMPTZ,
    "unpublish_at" TIMESTAMPTZ,
    "is_visible" BOOLEAN NOT NULL DEFAULT TRUE,
    "content_sections" JSONB NOT NULL DEFAULT '[]',
    "contacts" JSONB NOT NULL DEFAULT '[]',
    "venue_address" JSONB NOT NULL DEFAULT '{}',
    "eligibility" JSONB NOT NULL DEFAULT '{"audience":"everyone"}',
    "registration_config" JSONB NOT NULL DEFAULT '{"mode":"none"}',
    "geo_config" JSONB NOT NULL DEFAULT '{"enabled":false}',
    "recurrence" JSONB,
    "created_by" VARCHAR(15) NOT NULL DEFAULT '',
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMPTZ
);
CREATE INDEX idx_v3_events_status ON v3_events(status);

CREATE TABLE "v3_sessions" (
    "id" BIGSERIAL PRIMARY KEY,
    "code" VARCHAR(20) UNIQUE NOT NULL,
    "event_id" BIGINT NOT NULL REFERENCES v3_events(id),
    "title" VARCHAR(255) NOT NULL,
    "description" TEXT NOT NULL DEFAULT '',
    "start_at" TIMESTAMPTZ NOT NULL,
    "end_at" TIMESTAMPTZ NOT NULL,
    "register_start_at" TIMESTAMPTZ,
    "register_end_at" TIMESTAMPTZ,
    "checkin_open_at" TIMESTAMPTZ,
    "checkin_close_at" TIMESTAMPTZ,
    "location_type" VARCHAR(6) NOT NULL DEFAULT 'onsite',
    "location_name" VARCHAR(255) NOT NULL DEFAULT '',
    "online_url" VARCHAR(512) NOT NULL DEFAULT '',
    "total_seats" INT NOT NULL DEFAULT 0,
    "booked_seats" INT NOT NULL DEFAULT 0,
    "marked_sold_out" BOOLEAN NOT NULL DEFAULT FALSE,
    "attendance_modes" TEXT[] NOT NULL DEFAULT '{}',
    "enable_checkout" BOOLEAN NOT NULL DEFAULT FALSE,
    "generated_from_recurrence" BOOLEAN NOT NULL DEFAULT FALSE,
    "occurrence_key" VARCHAR(10),
    "status" VARCHAR(10) NOT NULL DEFAULT 'scheduled',
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMPTZ,
    UNIQUE ("event_id", "occurrence_key")
);
CREATE INDEX idx_v3_sessions_event ON v3_sessions(event_id);

CREATE TABLE "v3_session_ticket_categories" (
    "id" BIGSERIAL PRIMARY KEY,
    "session_id" BIGINT NOT NULL REFERENCES v3_sessions(id),
    "code" VARCHAR(50) NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "description" TEXT NOT NULL DEFAULT '',
    "total_seats" INT NOT NULL DEFAULT 0,
    "booked_seats" INT NOT NULL DEFAULT 0,
    "marked_sold_out" BOOLEAN NOT NULL DEFAULT FALSE,
    "sort_order" INT NOT NULL DEFAULT 0,
    UNIQUE ("session_id", "code")
);

CREATE TABLE "v3_registrations" (
    "id" UUID PRIMARY KEY,
    "session_id" BIGINT NOT NULL REFERENCES v3_sessions(id),
    "category_id" BIGINT REFERENCES v3_session_ticket_categories(id),
    "registered_by" VARCHAR(15) NOT NULL,
    "party_size" INT NOT NULL DEFAULT 1,
    "status" VARCHAR(10) NOT NULL DEFAULT 'confirmed',
    "source" VARCHAR(8) NOT NULL DEFAULT 'web',
    "registered_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "cancelled_at" TIMESTAMPTZ
);
CREATE INDEX idx_v3_registrations_session ON v3_registrations(session_id);
CREATE INDEX idx_v3_registrations_by ON v3_registrations(registered_by);

CREATE TABLE "v3_registration_attendees" (
    "id" UUID PRIMARY KEY,
    "registration_id" UUID NOT NULL REFERENCES v3_registrations(id),
    "community_id" VARCHAR(15),
    "name" VARCHAR(100) NOT NULL,
    "answers" JSONB NOT NULL DEFAULT '{}',
    "answer_revisions" JSONB NOT NULL DEFAULT '[]',
    "status" VARCHAR(12) NOT NULL DEFAULT 'registered',
    "attended_at" TIMESTAMPTZ,
    "checked_out_at" TIMESTAMPTZ
);
CREATE INDEX idx_v3_attendees_registration ON v3_registration_attendees(registration_id);

CREATE TABLE "v3_attendance_logs" (
    "id" UUID PRIMARY KEY,
    "session_id" BIGINT NOT NULL,
    "attendee_id" UUID,
    "community_id" VARCHAR(15),
    "action" VARCHAR(8) NOT NULL DEFAULT 'checkin',
    "mode" VARCHAR(16) NOT NULL,
    "checked_by" VARCHAR(15) NOT NULL DEFAULT '',
    "lat" NUMERIC, "lng" NUMERIC, "accuracy_m" NUMERIC,
    "geo_result" VARCHAR(12) NOT NULL DEFAULT 'not_required',
    "outcome" VARCHAR(20) NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_v3_attendance_session ON v3_attendance_logs(session_id);

CREATE TABLE "v3_notification_outbox" (
    "id" UUID PRIMARY KEY,
    "channel" VARCHAR(10) NOT NULL DEFAULT 'email',
    "recipient" VARCHAR(255) NOT NULL,
    "template" VARCHAR(50) NOT NULL,
    "payload" JSONB NOT NULL DEFAULT '{}',
    "status" VARCHAR(8) NOT NULL DEFAULT 'pending',
    "attempts" INT NOT NULL DEFAULT 0,
    "scheduled_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "sent_at" TIMESTAMPTZ
);
CREATE INDEX idx_v3_outbox_pending ON v3_notification_outbox(status, scheduled_at);

CREATE TABLE "v3_event_organizers" (
    "id" BIGSERIAL PRIMARY KEY,
    "event_id" BIGINT NOT NULL REFERENCES v3_events(id),
    "community_id" VARCHAR(15) NOT NULL,
    "role" VARCHAR(6) NOT NULL DEFAULT 'staff',
    UNIQUE ("event_id", "community_id")
);

CREATE TABLE "v3_event_templates" (
    "id" BIGSERIAL PRIMARY KEY,
    "name" VARCHAR(255) NOT NULL,
    "description" TEXT NOT NULL DEFAULT '',
    "snapshot" JSONB NOT NULL,
    "is_system" BOOLEAN NOT NULL DEFAULT FALSE,
    "created_by" VARCHAR(15) NOT NULL DEFAULT '',
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

- [ ] **Step 2: Write the down migration**

```sql
-- tests/integration/db/migrations/000022_v3_event_management.down.sql
DROP TABLE IF EXISTS "v3_event_templates";
DROP TABLE IF EXISTS "v3_event_organizers";
DROP TABLE IF EXISTS "v3_notification_outbox";
DROP TABLE IF EXISTS "v3_attendance_logs";
DROP TABLE IF EXISTS "v3_registration_attendees";
DROP TABLE IF EXISTS "v3_registrations";
DROP TABLE IF EXISTS "v3_session_ticket_categories";
DROP TABLE IF EXISTS "v3_sessions";
DROP TABLE IF EXISTS "v3_events";
```

- [ ] **Step 3: Write the DB test harness**

```go
// tests/integration/v3/harness_test.go
package v3_test

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	dbOnce sync.Once
	dbConn *gorm.DB
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	dbOnce.Do(func() {
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		if err != nil {
			t.Fatalf("connect: %v", err)
		}
		sqlBytes, err := os.ReadFile(filepath.Join("..", "db", "migrations", "000022_v3_event_management.up.sql"))
		if err != nil {
			t.Fatalf("read migration: %v", err)
		}
		for _, stmt := range strings.Split(string(sqlBytes), ";") {
			if s := strings.TrimSpace(stmt); s != "" {
				db.Exec(s) // idempotent-ish: errors on existing tables are ignored below via fresh truncate
			}
		}
		dbConn = db
	})
	tables := []string{"v3_attendance_logs", "v3_notification_outbox", "v3_registration_attendees",
		"v3_registrations", "v3_session_ticket_categories", "v3_event_organizers",
		"v3_sessions", "v3_event_templates", "v3_events"}
	for _, tb := range tables {
		if err := dbConn.Exec("TRUNCATE TABLE " + tb + " CASCADE").Error; err != nil {
			t.Fatalf("truncate %s: %v", tb, err)
		}
	}
	return dbConn
}
```

- [ ] **Step 4: Verify** — Run: `go vet ./tests/integration/v3/` → no errors. With docker Postgres up: `TEST_DATABASE_URL="postgres://postgres:$DB_PASSWORD@localhost:5888/community_db?sslmode=disable" go test ./tests/integration/v3/ -run TestNothing -v` → `no tests to run` (harness compiles). Then `psql`-check one table exists after any later test run.

- [ ] **Step 5: Commit** — `git add tests/integration/ && git commit -m "feat(v3): v3 event management schema migration + DB test harness"`

---

### Task 2: herr integration — bilingual humanized error catalog

**Files:**
- Create: `internal/models/v3/errors.go` (herr classes + constructors)
- Create: `internal/models/v3/locale.go` (EN/ID message catalog + localizer install)
- Modify: `go.mod` (add `github.com/jeremygprawira/herr`; bump `go 1.23.0` → `go 1.26`)
- Test: `internal/models/v3/errors_test.go`

**Interfaces:**
- Consumes: `github.com/jeremygprawira/herr` (+ `herr/localizer/mapl`, in the core module).
- Produces (package `v3`) — these names are used by every later task:
  - Classes (`*herr.Class`, immutable, match with `.Is(err)`): `ErrNotFound, ErrNotEligible, ErrRegistrationClosed, ErrQuotaFull, ErrMarkedSoldOut, ErrAlreadyRegistered, ErrGeoRequired, ErrGeoOutOfRange, ErrAlreadyCheckedIn, ErrNotCheckedIn, ErrNotRegistered, ErrCancelled, ErrForbidden, ErrEditWindowClosed, ErrInvalidAccessCode, ErrInvalidToken, ErrExpiredToken`.
  - Constructors returning `error`: `ErrInvalidInput(detail string) error` (Kind Invalid; humanized generic public message, `detail` attached via `WithPublic("detail", ...)` so specifics survive without breaking localization), `ErrPublishValidation(detail string) error` (Kind Unprocessable, same pattern), `FieldErrors(code string, fields ...FieldIssue) error` where `type FieldIssue struct{ Field, Code, Message string }` — renders herr's typed `errors[]`.
  - `func InstallLocale()` — installs the `mapl` localizer with the Indonesian catalog; called once from the composition root (`contract.go`, Task 17) and from `TestMain` in v3 test packages.
  - `func Localize(err error, locale string) (status int, body []byte)` — thin wrapper over herr's `Body(locale)` + HTTP status for the Echo responder (Task 16).

- [ ] **Step 1: Add dependency + toolchain bump**

Run: `go get github.com/jeremygprawira/herr@latest` then edit `go.mod`: `go 1.26`. Run `go mod tidy && go build ./...` → OK (verify local Go ≥1.26 with `go version` first).

- [ ] **Step 2: Write the failing test**

```go
// internal/models/v3/errors_test.go
package v3

import (
	"encoding/json"
	"testing"

	"github.com/jeremygprawira/herr"
)

func TestMain(m *testing.M) { InstallLocale(); m.Run() }

func TestClassesCarryKindAndMatch(t *testing.T) {
	err := ErrQuotaFull.New()
	if !ErrQuotaFull.Is(err) {
		t.Fatal("class must match its instances through Is")
	}
	if herr.HTTPStatus(err) != 409 {
		t.Fatalf("KindConflict must map to 409, got %d", herr.HTTPStatus(err))
	}
	if ErrNotEligible.Is(err) {
		t.Fatal("classes must not cross-match")
	}
}

func TestBilingualBodies(t *testing.T) {
	err := ErrQuotaFull.New()
	var en, id map[string]any
	_, bodyEN := Localize(err, "en")
	_, bodyID := Localize(err, "id")
	json.Unmarshal(bodyEN, &en)
	json.Unmarshal(bodyID, &id)
	if en["message"] == "" || id["message"] == "" || en["message"] == id["message"] {
		t.Fatalf("EN and ID must both exist and differ: %v / %v", en["message"], id["message"])
	}
	if en["code"] != "QUOTA_FULL" || id["code"] != "QUOTA_FULL" {
		t.Fatal("code is language-independent")
	}
}

// Every class must have an Indonesian translation — forgetting one fails CI.
func TestLocaleCatalogComplete(t *testing.T) {
	for _, cls := range AllClasses {
		_, body := Localize(cls.New(), "id")
		var m map[string]any
		json.Unmarshal(body, &m)
		if msg, _ := m["message"].(string); msg == "" || msg == englishMessages[codeOf(cls)] {
			t.Errorf("class %s is missing an Indonesian message", codeOf(cls))
		}
	}
}

func TestInternalNeverLeaks(t *testing.T) {
	err := ErrQuotaFull.New().Internal("shard eu-3 exploded").With("secret", "hunter2")
	_, body := Localize(err, "en")
	if s := string(body); containsAny(s, "eu-3", "hunter2") {
		t.Fatal("internal detail leaked into wire body")
	}
}

func TestFieldErrors(t *testing.T) {
	err := FieldErrors("VALIDATION_FAILED",
		FieldIssue{Field: "email", Code: "INVALID_EMAIL", Message: "Enter a valid email address."})
	_, body := Localize(err, "en")
	var m struct {
		Errors []struct{ Field, Code string } `json:"errors"`
	}
	json.Unmarshal(body, &m)
	if len(m.Errors) != 1 || m.Errors[0].Field != "email" {
		t.Fatalf("field errors must render errors[]: %s", body)
	}
}
```

(`containsAny` and `codeOf` are 3-line test helpers in the same file; `AllClasses` and `englishMessages` are exported by the implementation for the completeness gate.)

- [ ] **Step 3: Run** `go test ./internal/models/v3/ -v` → FAIL.

- [ ] **Step 4: Implement the catalog**

```go
// internal/models/v3/errors.go
package v3

import "github.com/jeremygprawira/herr"

func def(code string, kind herr.Kind, en string) *herr.Class {
	c := herr.Define(herr.Class{Code: code, Kind: kind, Public: herr.Public{Message: en}})
	englishMessages[code] = en
	AllClasses = append(AllClasses, c)
	return c
}

var (
	AllClasses      []*herr.Class
	englishMessages = map[string]string{}

	ErrNotFound           = def("NOT_FOUND", herr.KindNotFound, "We couldn't find what you're looking for.")
	ErrNotEligible        = def("NOT_ELIGIBLE", herr.KindForbidden, "This event isn't open for your account. Reach out to the organizer if you think this is a mistake.")
	ErrRegistrationClosed = def("REGISTRATION_CLOSED", herr.KindForbidden, "Registration isn't open right now. Check the event page for the registration schedule.")
	ErrQuotaFull          = def("QUOTA_FULL", herr.KindConflict, "All seats have been taken. Keep an eye out — a spot may open up if someone cancels.")
	ErrMarkedSoldOut      = def("MARKED_SOLD_OUT", herr.KindConflict, "The organizer has closed registration for this session.")
	ErrAlreadyRegistered  = def("ALREADY_REGISTERED", herr.KindConflict, "You're already registered for this session. Check My Registrations for your ticket.")
	ErrGeoRequired        = def("GEO_REQUIRED", herr.KindInvalid, "We need your location for this step. Please allow location access and try again.")
	ErrGeoOutOfRange      = def("GEO_OUT_OF_RANGE", herr.KindForbidden, "You seem to be away from the venue. Please try again when you've arrived.")
	ErrAlreadyCheckedIn   = def("ALREADY_CHECKED_IN", herr.KindConflict, "You're already checked in — you're all set!")
	ErrNotCheckedIn       = def("NOT_CHECKED_IN", herr.KindConflict, "We can't check you out because you haven't checked in yet.")
	ErrNotRegistered      = def("NOT_REGISTERED", herr.KindNotFound, "We couldn't find a registration for this session.")
	ErrCancelled          = def("REGISTRATION_CANCELLED", herr.KindConflict, "This registration has been cancelled.")
	ErrForbidden          = def("FORBIDDEN", herr.KindForbidden, "You don't have access to do this.")
	ErrEditWindowClosed   = def("EDIT_WINDOW_CLOSED", herr.KindForbidden, "The time window for editing your answers has passed.")
	ErrInvalidAccessCode  = def("INVALID_ACCESS_CODE", herr.KindForbidden, "That access code doesn't look right. Double-check it and try again.")
	ErrInvalidToken       = def("INVALID_QR_TOKEN", herr.KindUnauthorized, "This QR code isn't valid. Please use the QR from your ticket or profile.")
	ErrExpiredToken       = def("EXPIRED_QR_TOKEN", herr.KindUnauthorized, "This QR code has expired. Please open a fresh one.")

	clsInvalidInput = def("INVALID_INPUT", herr.KindInvalid, "Something in your submission doesn't look right. Please review and try again.")
	clsPublish      = def("PUBLISH_VALIDATION", herr.KindUnprocessable, "The event isn't ready to publish yet. Please fix the highlighted issues.")
)

func ErrInvalidInput(detail string) error {
	return clsInvalidInput.New().WithPublic("detail", detail)
}
func ErrPublishValidation(detail string) error {
	return clsPublish.New().WithPublic("detail", detail)
}

type FieldIssue struct{ Field, Code, Message string }

func FieldErrors(code string, fields ...FieldIssue) error {
	e := herr.New(code).Kind(herr.KindUnprocessable)
	for _, f := range fields {
		e = e.FieldError(f.Field, f.Code, f.Message)
	}
	return e
}

func Localize(err error, locale string) (int, []byte) {
	he := herr.From(err) // wraps non-herr errors as KindInternal, per herr docs
	return herr.HTTPStatus(he), he.Body(locale)
}
```

> Check exact helper names against the herr version pulled: the README documents `Body(locale)`, `Define`, `Class.Is`, `FieldError`, `WithPublic`; confirm the exported names for status extraction (`herr.HTTPStatus`) and non-herr wrapping (`herr.From`) in the package docs (`go doc github.com/jeremygprawira/herr`) and adjust the two call sites if they differ.

```go
// internal/models/v3/locale.go
package v3

import (
	"github.com/jeremygprawira/herr"
	"github.com/jeremygprawira/herr/localizer/mapl"
)

// Indonesian catalog — key format errors.<lowercase_code>.message.
var indonesianMessages = map[string]string{
	"errors.not_found.message":              "Kami tidak dapat menemukan yang Anda cari.",
	"errors.not_eligible.message":           "Event ini belum terbuka untuk akun Anda. Hubungi panitia jika menurut Anda ini keliru.",
	"errors.registration_closed.message":    "Pendaftaran belum dibuka saat ini. Silakan cek jadwal pendaftaran di halaman event.",
	"errors.quota_full.message":             "Semua kursi sudah terisi. Pantau terus — kursi bisa tersedia lagi jika ada yang membatalkan.",
	"errors.marked_sold_out.message":        "Panitia telah menutup pendaftaran untuk sesi ini.",
	"errors.already_registered.message":     "Anda sudah terdaftar di sesi ini. Cek tiket Anda di menu Pendaftaran Saya.",
	"errors.geo_required.message":           "Kami memerlukan lokasi Anda untuk langkah ini. Mohon izinkan akses lokasi lalu coba lagi.",
	"errors.geo_out_of_range.message":       "Sepertinya Anda berada jauh dari lokasi acara. Silakan coba lagi setelah tiba di lokasi.",
	"errors.already_checked_in.message":     "Anda sudah check-in — selamat menikmati acaranya!",
	"errors.not_checked_in.message":         "Anda belum check-in, jadi belum bisa check-out.",
	"errors.not_registered.message":         "Kami tidak menemukan pendaftaran Anda untuk sesi ini.",
	"errors.registration_cancelled.message": "Pendaftaran ini sudah dibatalkan.",
	"errors.forbidden.message":              "Anda tidak memiliki akses untuk melakukan ini.",
	"errors.edit_window_closed.message":     "Batas waktu untuk mengubah jawaban Anda sudah lewat.",
	"errors.invalid_access_code.message":    "Kode akses tidak sesuai. Mohon periksa kembali lalu coba lagi.",
	"errors.invalid_qr_token.message":       "Kode QR ini tidak valid. Gunakan QR dari tiket atau profil Anda.",
	"errors.expired_qr_token.message":       "Kode QR ini sudah kedaluwarsa. Silakan buka QR yang baru.",
	"errors.invalid_input.message":          "Ada data yang belum sesuai. Mohon periksa kembali lalu coba lagi.",
	"errors.publish_validation.message":     "Event belum siap dipublikasikan. Mohon perbaiki bagian yang ditandai.",
}

func InstallLocale() {
	herr.SetLocalizer(mapl.New(map[string]map[string]string{"id": indonesianMessages}))
}
```

- [ ] **Step 5: Run** `go test ./internal/models/v3/ -v` → PASS (incl. the leak test and the ID-catalog completeness gate).
- [ ] **Step 6: Commit** — `git add go.mod go.sum internal/models/v3/ && git commit -m "feat(v3): herr error catalog with bilingual EN/ID humanized messages"`

---

> **herr substitution rule for Tasks 3–17.** Later tasks were drafted against a custom `*v3.Error`; with herr they read as follows — apply mechanically, no other changes:
> 1. Every signature `*v3.Error` becomes `error`.
> 2. Every comparison `err == v3.ErrX` / `e != v3.ErrX` becomes `v3.ErrX.Is(err)` / `!v3.ErrX.Is(err)`.
> 3. Every `return v3.ErrX` becomes `return v3.ErrX.New()` (add `.With(...)`/`.Internal(...)` context where the task has useful detail — e.g. `.With("session", code)`).
> 4. `v3.AsError(err)` disappears — return the error as-is; the responder wraps unknown errors via herr.
> 5. Test assertions on `.Code` fields become `v3.ErrX.Is(err)` checks; wire-body assertions decode `Localize(err, locale)` output.
> 6. Form-validation failures in Task 6 may use `v3.FieldErrors("VALIDATION_FAILED", issues...)` when reporting multiple fields; single-field early-exit `v3.ErrInvalidInput(...)` stays valid.
> 7. Task 16's `respond.Err` becomes: resolve locale from `Accept-Language` (values `en`/`id`, default `id`), then `status, body := v3.Localize(err, locale); return ctx.JSONBlob(status, body)`. `InstallLocale()` is called once in `contract.go` (Task 17) and in each integration-test `TestMain`.
---

### Task 3: JSONB config structs + validation

**Files:**
- Create: `internal/models/v3/config.go`
- Test: `internal/models/v3/config_test.go`

**Interfaces:**
- Produces (all in package `v3`, JSON tags snake_case as in spec):
  - `type Eligibility struct{ Audience string; Rules EligibilityRules }`, `type EligibilityRules struct{ Roles, UserTypes, Campuses, CommunityIDs []string }` (tags `user_types`, `community_ids`)
  - `type FormField struct{ Key, Type, Label string; Required bool; Options []string; AnsweredBy string; ShowIf *Condition }` (`answered_by`, `show_if`); `type Condition struct{ Field, Op string; Value any; All, Any []Condition }`
  - `type EditPolicy struct{ Enabled bool; Until string; Fields []string }` (`Fields` nil ⇒ all)
  - `type Phase struct{ Name string; Start, End time.Time; EligibilityOverride *Eligibility; AccessCode string; MaxPerRegistration *int }` (`eligibility_override`, `access_code`, `max_per_registration`)
  - `type RegistrationConfig struct{ Mode string; MaxPerRegistration int; AllowMultiple bool; CompanionDetail string; EditPolicy *EditPolicy; Phases []Phase; Form []FormField }` (`allow_multiple_registrations`, `companion_detail`)
  - `type GeoRule struct{ Check string; StaffOverride bool }`; `type GeoConfig struct{ Enabled bool; Venue Venue; Modes map[string]GeoRule }`; `type Venue struct{ Lat, Lng, RadiusM float64 }` (`radius_m`)
  - `type SessionDefaults struct{ StartTime string; DurationMin int; TotalSeats int }`; `type Recurrence struct{ Freq string; Interval int; ByDay []string; ByMonthDay []int; CustomDates []string; SessionDefaults SessionDefaults; Until string; GenerateAhead int }`
  - `type Contact struct{ Name, Role string; Channels []Channel }`; `type Channel struct{ Type, Value, Label string }`
  - `type ContentSection struct{ Key, Title, Body string; Sort int }`
  - Decoders: `func ParseEligibility(raw []byte) (Eligibility, *Error)` and same pattern `ParseRegistrationConfig`, `ParseGeoConfig`, `ParseRecurrence` (nil raw ⇒ zero value, no error).
  - Strict validation for publish gate: `func (r RegistrationConfig) ValidateStrict() *Error` — checks mode ∈ {none,required,optional}; companion_detail ∈ {full,count_only}; a field with key `name` exists and is required when mode != none; field types ∈ {name,email,phone,nik,text,textarea,number,select,multiselect,checkbox,date}; select/multiselect require options; unique keys; `answered_by` ∈ {primary,everyone,companions} (empty ⇒ primary).

- [ ] **Step 1: Write the failing test**

```go
// internal/models/v3/config_test.go
package v3

import "testing"

func TestParseRegistrationConfig(t *testing.T) {
	raw := []byte(`{"mode":"required","max_per_registration":3,"companion_detail":"full",
	 "form":[{"key":"name","type":"name","label":"Full name","required":true,"answered_by":"everyone"},
	         {"key":"diet","type":"select","label":"Diet","options":["none","allergy"],"required":true,
	          "show_if":{"field":"name","op":"answered"}}]}`)
	rc, err := ParseRegistrationConfig(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if rc.Mode != "required" || rc.MaxPerRegistration != 3 || len(rc.Form) != 2 {
		t.Fatalf("bad parse: %+v", rc)
	}
	if rc.Form[1].ShowIf == nil || rc.Form[1].ShowIf.Op != "answered" {
		t.Fatal("show_if not parsed")
	}
	if e := rc.ValidateStrict(); e != nil {
		t.Fatalf("should be valid: %v", e)
	}
}

func TestValidateStrictRejectsBadConfig(t *testing.T) {
	rc := RegistrationConfig{Mode: "required", CompanionDetail: "full",
		Form: []FormField{{Key: "sel", Type: "select", Label: "x", Required: true}}}
	if e := rc.ValidateStrict(); e == nil {
		t.Fatal("select without options and missing name field must fail")
	}
	rc2 := RegistrationConfig{Mode: "banana"}
	if e := rc2.ValidateStrict(); e == nil {
		t.Fatal("bad mode must fail")
	}
}

func TestParseNilIsZero(t *testing.T) {
	g, err := ParseGeoConfig(nil)
	if err != nil || g.Enabled {
		t.Fatalf("nil raw must yield disabled zero config: %+v %v", g, err)
	}
}
```

- [ ] **Step 2: Run** `go test ./internal/models/v3/ -run TestParse -v` → FAIL.

- [ ] **Step 3: Implement**

```go
// internal/models/v3/config.go
package v3

import (
	"encoding/json"
	"time"
)

type EligibilityRules struct {
	Roles        []string `json:"roles,omitempty"`
	UserTypes    []string `json:"user_types,omitempty"`
	Campuses     []string `json:"campuses,omitempty"`
	CommunityIDs []string `json:"community_ids,omitempty"`
}
type Eligibility struct {
	Audience string           `json:"audience"`
	Rules    EligibilityRules `json:"rules,omitempty"`
}

type Condition struct {
	Field string      `json:"field,omitempty"`
	Op    string      `json:"op,omitempty"`
	Value interface{} `json:"value,omitempty"`
	All   []Condition `json:"all,omitempty"`
	Any   []Condition `json:"any,omitempty"`
}
type FormField struct {
	Key        string     `json:"key"`
	Type       string     `json:"type"`
	Label      string     `json:"label"`
	Required   bool       `json:"required"`
	Options    []string   `json:"options,omitempty"`
	AnsweredBy string     `json:"answered_by,omitempty"`
	ShowIf     *Condition `json:"show_if,omitempty"`
}
type EditPolicy struct {
	Enabled bool     `json:"enabled"`
	Until   string   `json:"until"`
	Fields  []string `json:"fields,omitempty"`
}
type Phase struct {
	Name                string       `json:"name"`
	Start               time.Time    `json:"start"`
	End                 time.Time    `json:"end"`
	EligibilityOverride *Eligibility `json:"eligibility_override,omitempty"`
	AccessCode          string       `json:"access_code,omitempty"`
	MaxPerRegistration  *int         `json:"max_per_registration,omitempty"`
}
type RegistrationConfig struct {
	Mode               string      `json:"mode"`
	MaxPerRegistration int         `json:"max_per_registration"`
	AllowMultiple      bool        `json:"allow_multiple_registrations"`
	CompanionDetail    string      `json:"companion_detail"`
	EditPolicy         *EditPolicy `json:"edit_policy,omitempty"`
	Phases             []Phase     `json:"phases,omitempty"`
	Form               []FormField `json:"form"`
}

type Venue struct {
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	RadiusM float64 `json:"radius_m"`
}
type GeoRule struct {
	Check         string `json:"check"`
	StaffOverride bool   `json:"staff_override"`
}
type GeoConfig struct {
	Enabled bool               `json:"enabled"`
	Venue   Venue              `json:"venue"`
	Modes   map[string]GeoRule `json:"modes,omitempty"`
}

type SessionDefaults struct {
	StartTime   string `json:"start_time"`
	DurationMin int    `json:"duration_min"`
	TotalSeats  int    `json:"total_seats"`
}
type Recurrence struct {
	Freq            string          `json:"freq"`
	Interval        int             `json:"interval"`
	ByDay           []string        `json:"by_day,omitempty"`
	ByMonthDay      []int           `json:"by_month_day,omitempty"`
	CustomDates     []string        `json:"custom_dates,omitempty"`
	SessionDefaults SessionDefaults `json:"session_defaults"`
	Until           string          `json:"until,omitempty"`
	GenerateAhead   int             `json:"generate_ahead"`
}

type Channel struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Label string `json:"label,omitempty"`
}
type Contact struct {
	Name     string    `json:"name"`
	Role     string    `json:"role,omitempty"`
	Channels []Channel `json:"channels"`
}
type ContentSection struct {
	Key   string `json:"key"`
	Title string `json:"title"`
	Body  string `json:"body"`
	Sort  int    `json:"sort"`
}

func parseJSON[T any](raw []byte, out *T) *Error {
	if len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return ErrInvalidInput("invalid config json: " + err.Error())
	}
	return nil
}
func ParseEligibility(raw []byte) (Eligibility, *Error) {
	var v Eligibility
	return v, parseJSON(raw, &v)
}
func ParseRegistrationConfig(raw []byte) (RegistrationConfig, *Error) {
	var v RegistrationConfig
	return v, parseJSON(raw, &v)
}
func ParseGeoConfig(raw []byte) (GeoConfig, *Error) {
	var v GeoConfig
	return v, parseJSON(raw, &v)
}
func ParseRecurrence(raw []byte) (*Recurrence, *Error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var v Recurrence
	if e := parseJSON(raw, &v); e != nil {
		return nil, e
	}
	return &v, nil
}

var fieldTypes = map[string]bool{"name": true, "email": true, "phone": true, "nik": true,
	"text": true, "textarea": true, "number": true, "select": true, "multiselect": true,
	"checkbox": true, "date": true}

func (r RegistrationConfig) ValidateStrict() *Error {
	switch r.Mode {
	case "none":
		return nil
	case "required", "optional":
	default:
		return ErrPublishValidation("registration mode must be none, required or optional")
	}
	if r.CompanionDetail != "full" && r.CompanionDetail != "count_only" && r.CompanionDetail != "" {
		return ErrPublishValidation("companion_detail must be full or count_only")
	}
	seen := map[string]bool{}
	hasName := false
	for _, f := range r.Form {
		if f.Key == "" || f.Label == "" {
			return ErrPublishValidation("form field key and label are required")
		}
		if seen[f.Key] {
			return ErrPublishValidation("duplicate form field key: " + f.Key)
		}
		seen[f.Key] = true
		if !fieldTypes[f.Type] {
			return ErrPublishValidation("unknown form field type: " + f.Type)
		}
		if (f.Type == "select" || f.Type == "multiselect") && len(f.Options) == 0 {
			return ErrPublishValidation("field " + f.Key + " needs options")
		}
		switch f.AnsweredBy {
		case "", "primary", "everyone", "companions":
		default:
			return ErrPublishValidation("field " + f.Key + ": answered_by must be primary, everyone or companions")
		}
		if f.Key == "name" && f.Required {
			hasName = true
		}
	}
	if !hasName {
		return ErrPublishValidation("form must include a required name field")
	}
	return nil
}
```

- [ ] **Step 4: Run** `go test ./internal/models/v3/ -v` → PASS.
- [ ] **Step 5: Commit** — `git add internal/models/v3/ && git commit -m "feat(v3): jsonb config structs, parsers and strict validation"`

---

### Task 4: Eligibility checker (pure)

**Files:**
- Create: `internal/pkg/eligibility/eligibility.go`
- Test: `internal/pkg/eligibility/eligibility_test.go`

**Interfaces:**
- Consumes: `v3.Eligibility` from Task 3.
- Produces: `type User struct{ CommunityID, CampusCode string; Roles, UserTypes []string }`; `func Check(u User, e v3.Eligibility) bool`. Semantics per spec §2.2: `everyone`/`members`/empty ⇒ true (auth handled upstream); `rules` ⇒ match ANY non-empty list.

- [ ] **Step 1: Write the failing test**

```go
// internal/pkg/eligibility/eligibility_test.go
package eligibility

import (
	v3 "go-community/internal/models/v3"
	"testing"
)

func TestCheck(t *testing.T) {
	u := User{CommunityID: "111", CampusCode: "BKS", Roles: []string{"usher"}, UserTypes: []string{"volunteer"}}
	cases := []struct {
		name string
		e    v3.Eligibility
		want bool
	}{
		{"everyone", v3.Eligibility{Audience: "everyone"}, true},
		{"members", v3.Eligibility{Audience: "members"}, true},
		{"empty audience", v3.Eligibility{}, true},
		{"role match", v3.Eligibility{Audience: "rules", Rules: v3.EligibilityRules{Roles: []string{"usher"}}}, true},
		{"campus match", v3.Eligibility{Audience: "rules", Rules: v3.EligibilityRules{Campuses: []string{"BKS"}}}, true},
		{"community id match", v3.Eligibility{Audience: "rules", Rules: v3.EligibilityRules{CommunityIDs: []string{"111"}}}, true},
		{"no match", v3.Eligibility{Audience: "rules", Rules: v3.EligibilityRules{Roles: []string{"admin"}, Campuses: []string{"JKT"}}}, false},
		{"rules all empty", v3.Eligibility{Audience: "rules"}, false},
	}
	for _, c := range cases {
		if got := Check(u, c.e); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}
```

- [ ] **Step 2: Run** `go test ./internal/pkg/eligibility/ -v` → FAIL.

- [ ] **Step 3: Implement**

```go
// internal/pkg/eligibility/eligibility.go
package eligibility

import v3 "go-community/internal/models/v3"

type User struct {
	CommunityID string
	CampusCode  string
	Roles       []string
	UserTypes   []string
}

func Check(u User, e v3.Eligibility) bool {
	if e.Audience != "rules" {
		return true
	}
	if intersects(u.Roles, e.Rules.Roles) || intersects(u.UserTypes, e.Rules.UserTypes) {
		return true
	}
	if contains(e.Rules.Campuses, u.CampusCode) || contains(e.Rules.CommunityIDs, u.CommunityID) {
		return true
	}
	return false
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v && v != "" {
			return true
		}
	}
	return false
}
func intersects(a, b []string) bool {
	for _, x := range a {
		if contains(b, x) {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Run** `go test ./internal/pkg/eligibility/ -v` → PASS.
- [ ] **Step 5: Commit** — `git commit -am "feat(v3): pure eligibility checker"`

---

### Task 5: Geo validation (pure)

**Files:**
- Create: `internal/pkg/geo/geo.go`
- Test: `internal/pkg/geo/geo_test.go`

**Interfaces:**
- Consumes: `v3.GeoConfig` (Task 3), `v3` errors (Task 2).
- Produces: `type Coords struct{ Lat, Lng, AccuracyM float64 }`; `type Result string` with consts `ResultOK, ResultOutOfRange, ResultOverridden, ResultNotProvided, ResultNotRequired`; `func Distance(lat1, lng1, lat2, lng2 float64) float64` (meters, Haversine); `func Validate(cfg v3.GeoConfig, mode string, c *Coords, override bool) (Result, *v3.Error)`.
  Semantics: disabled or mode rule `check=="off"`/missing ⇒ `ResultNotRequired,nil`. `check=="warn"` ⇒ never errors; result OK/OutOfRange/NotProvided recorded. `check=="require"` ⇒ nil coords → `ErrGeoRequired`; outside radius → if `override && rule.StaffOverride` ⇒ `ResultOverridden,nil` else `ErrGeoOutOfRange`.

- [ ] **Step 1: Write the failing test**

```go
// internal/pkg/geo/geo_test.go
package geo

import (
	v3 "go-community/internal/models/v3"
	"math"
	"testing"
)

func cfg(check string, override bool) v3.GeoConfig {
	return v3.GeoConfig{Enabled: true,
		Venue: v3.Venue{Lat: -6.2, Lng: 106.816666, RadiusM: 100},
		Modes: map[string]v3.GeoRule{"personal_qr": {Check: check, StaffOverride: override}}}
}

func TestDistance(t *testing.T) {
	d := Distance(-6.2, 106.816666, -6.2, 106.817566) // ~100m east at this latitude
	if math.Abs(d-99.6) > 2 {
		t.Fatalf("distance ~99.6m, got %f", d)
	}
}

func TestValidate(t *testing.T) {
	inside := &Coords{Lat: -6.2, Lng: 106.81670}
	outside := &Coords{Lat: -6.21, Lng: 106.9}

	if r, e := Validate(v3.GeoConfig{}, "personal_qr", nil, false); r != ResultNotRequired || e != nil {
		t.Fatal("disabled config must be not_required")
	}
	if r, e := Validate(cfg("require", false), "personal_qr", inside, false); r != ResultOK || e != nil {
		t.Fatal("inside radius must be ok")
	}
	if _, e := Validate(cfg("require", false), "personal_qr", nil, false); e != v3.ErrGeoRequired {
		t.Fatal("require + no coords must error GEO_REQUIRED")
	}
	if _, e := Validate(cfg("require", false), "personal_qr", outside, false); e != v3.ErrGeoOutOfRange {
		t.Fatal("require + outside must error GEO_OUT_OF_RANGE")
	}
	if r, e := Validate(cfg("require", true), "personal_qr", outside, true); r != ResultOverridden || e != nil {
		t.Fatal("staff override must pass as overridden")
	}
	if r, e := Validate(cfg("warn", false), "personal_qr", outside, false); r != ResultOutOfRange || e != nil {
		t.Fatal("warn records out_of_range without error")
	}
	if r, _ := Validate(cfg("off", false), "personal_qr", outside, false); r != ResultNotRequired {
		t.Fatal("off must be not_required")
	}
}
```

- [ ] **Step 2: Run** `go test ./internal/pkg/geo/ -v` → FAIL.

- [ ] **Step 3: Implement**

```go
// internal/pkg/geo/geo.go
package geo

import (
	v3 "go-community/internal/models/v3"
	"math"
)

type Coords struct{ Lat, Lng, AccuracyM float64 }

type Result string

const (
	ResultOK          Result = "ok"
	ResultOutOfRange  Result = "out_of_range"
	ResultOverridden  Result = "overridden"
	ResultNotProvided Result = "not_provided"
	ResultNotRequired Result = "not_required"
)

const earthRadiusM = 6371000.0

func Distance(lat1, lng1, lat2, lng2 float64) float64 {
	rad := func(d float64) float64 { return d * math.Pi / 180 }
	dLat, dLng := rad(lat2-lat1), rad(lng2-lng1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rad(lat1))*math.Cos(rad(lat2))*math.Sin(dLng/2)*math.Sin(dLng/2)
	return earthRadiusM * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func Validate(cfg v3.GeoConfig, mode string, c *Coords, override bool) (Result, *v3.Error) {
	if !cfg.Enabled {
		return ResultNotRequired, nil
	}
	rule, ok := cfg.Modes[mode]
	if !ok || rule.Check == "off" || rule.Check == "" {
		return ResultNotRequired, nil
	}
	if c == nil {
		if rule.Check == "require" {
			if override && rule.StaffOverride {
				return ResultOverridden, nil
			}
			return ResultNotProvided, v3.ErrGeoRequired
		}
		return ResultNotProvided, nil
	}
	if Distance(c.Lat, c.Lng, cfg.Venue.Lat, cfg.Venue.Lng) <= cfg.Venue.RadiusM {
		return ResultOK, nil
	}
	if rule.Check == "require" {
		if override && rule.StaffOverride {
			return ResultOverridden, nil
		}
		return ResultOutOfRange, v3.ErrGeoOutOfRange
	}
	return ResultOutOfRange, nil // warn
}
```

- [ ] **Step 4: Run** `go test ./internal/pkg/geo/ -v` → PASS.
- [ ] **Step 5: Commit** — `git commit -am "feat(v3): haversine geo policy validator"`

---

### Task 6: Form engine — show_if evaluator + answer validation

**Files:**
- Create: `internal/pkg/form/form.go`
- Test: `internal/pkg/form/form_test.go`

**Interfaces:**
- Consumes: `v3.FormField`, `v3.Condition`, v3 errors.
- Produces:
  - `func Visible(f v3.FormField, answers map[string]any) bool`
  - `func EvalCondition(c v3.Condition, answers map[string]any) bool` (ops: eq,neq,in,not_in,answered,not_answered,gt,gte,lt,lte,contains; combinators all/any nested)
  - `func Validate(fields []v3.FormField, audience string, answers map[string]any) (map[string]any, *v3.Error)` — audience `"primary"` or `"companion"`; a field applies when its `answered_by` (default primary) includes the audience (`everyone` ⇒ both, `companions` ⇒ companion only, `primary` ⇒ primary only); hidden-field answers are DROPPED from the returned cleaned map; required applies only when visible+applicable; typed format checks: email regex, Indonesian phone (`^\+?62\d{8,13}$` or `^0\d{8,12}$`), nik 16 digits, number numeric, select value ∈ options, multiselect subset, checkbox bool, date `2006-01-02`.
  - `func ValidateSchema(fields []v3.FormField) *v3.Error` — rejects show_if referencing unknown keys and dependency cycles.

- [ ] **Step 1: Write the failing test**

```go
// internal/pkg/form/form_test.go
package form

import (
	v3 "go-community/internal/models/v3"
	"testing"
)

func fields() []v3.FormField {
	return []v3.FormField{
		{Key: "name", Type: "name", Label: "Name", Required: true, AnsweredBy: "everyone"},
		{Key: "phone", Type: "phone", Label: "Phone", Required: false, AnsweredBy: "primary"},
		{Key: "diet", Type: "select", Label: "Diet", Options: []string{"none", "allergy"}, Required: true, AnsweredBy: "everyone"},
		{Key: "allergy_detail", Type: "text", Label: "Detail", Required: true, AnsweredBy: "everyone",
			ShowIf: &v3.Condition{Field: "diet", Op: "eq", Value: "allergy"}},
		{Key: "consent", Type: "checkbox", Label: "Consent", Required: true, AnsweredBy: "companions"},
	}
}

func TestEvalCondition(t *testing.T) {
	ans := map[string]any{"diet": "allergy", "age": float64(10)}
	if !EvalCondition(v3.Condition{Field: "diet", Op: "eq", Value: "allergy"}, ans) {
		t.Fatal("eq")
	}
	if !EvalCondition(v3.Condition{Any: []v3.Condition{
		{Field: "diet", Op: "eq", Value: "none"},
		{All: []v3.Condition{{Field: "age", Op: "lt", Value: float64(12)}, {Field: "diet", Op: "answered"}}},
	}}, ans) {
		t.Fatal("nested any/all")
	}
	if EvalCondition(v3.Condition{Field: "missing", Op: "answered"}, ans) {
		t.Fatal("missing not answered")
	}
}

func TestValidatePrimary(t *testing.T) {
	ans := map[string]any{"name": "Maria", "diet": "allergy", "allergy_detail": "peanuts",
		"consent": true, "smuggled": "x"}
	clean, err := Validate(fields(), "primary", ans)
	if err != nil {
		t.Fatalf("valid: %v", err)
	}
	if _, ok := clean["consent"]; ok {
		t.Fatal("companions-only field must be dropped for primary")
	}
	if _, ok := clean["smuggled"]; ok {
		t.Fatal("unknown keys must be dropped")
	}
	// hidden field answer must be dropped and its required ignored
	ans2 := map[string]any{"name": "Maria", "diet": "none", "allergy_detail": "should vanish"}
	clean2, err := Validate(fields(), "primary", ans2)
	if err != nil {
		t.Fatalf("valid without hidden required: %v", err)
	}
	if _, ok := clean2["allergy_detail"]; ok {
		t.Fatal("hidden answers must be dropped")
	}
	// missing required visible field
	if _, err := Validate(fields(), "primary", map[string]any{"name": "X"}); err == nil {
		t.Fatal("missing required diet must fail")
	}
	// bad select value
	if _, err := Validate(fields(), "primary", map[string]any{"name": "X", "diet": "pizza"}); err == nil {
		t.Fatal("bad option must fail")
	}
}

func TestValidateCompanion(t *testing.T) {
	if _, err := Validate(fields(), "companion", map[string]any{"name": "Kid", "diet": "none"}); err == nil {
		t.Fatal("companion missing required consent must fail")
	}
	if _, err := Validate(fields(), "companion",
		map[string]any{"name": "Kid", "diet": "none", "consent": true}); err != nil {
		t.Fatalf("valid companion: %v", err)
	}
}

func TestValidateSchemaCycles(t *testing.T) {
	bad := []v3.FormField{
		{Key: "a", Type: "text", Label: "A", ShowIf: &v3.Condition{Field: "b", Op: "answered"}},
		{Key: "b", Type: "text", Label: "B", ShowIf: &v3.Condition{Field: "a", Op: "answered"}},
	}
	if ValidateSchema(bad) == nil {
		t.Fatal("cycle must be rejected")
	}
	unknown := []v3.FormField{{Key: "a", Type: "text", Label: "A",
		ShowIf: &v3.Condition{Field: "ghost", Op: "answered"}}}
	if ValidateSchema(unknown) == nil {
		t.Fatal("unknown reference must be rejected")
	}
	if ValidateSchema(fields()) != nil {
		t.Fatal("valid schema must pass")
	}
}
```

- [ ] **Step 2: Run** `go test ./internal/pkg/form/ -v` → FAIL.

- [ ] **Step 3: Implement**

```go
// internal/pkg/form/form.go
package form

import (
	"fmt"
	v3 "go-community/internal/models/v3"
	"regexp"
	"time"
)

var (
	emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	phoneRe = regexp.MustCompile(`^(\+?62\d{8,13}|0\d{8,12})$`)
	nikRe   = regexp.MustCompile(`^\d{16}$`)
)

func EvalCondition(c v3.Condition, answers map[string]any) bool {
	if len(c.All) > 0 {
		for _, sub := range c.All {
			if !EvalCondition(sub, answers) {
				return false
			}
		}
		return true
	}
	if len(c.Any) > 0 {
		for _, sub := range c.Any {
			if EvalCondition(sub, answers) {
				return true
			}
		}
		return false
	}
	val, answered := answers[c.Field]
	switch c.Op {
	case "answered":
		return answered && val != nil && val != ""
	case "not_answered":
		return !answered || val == nil || val == ""
	case "eq":
		return answered && fmt.Sprint(val) == fmt.Sprint(c.Value)
	case "neq":
		return !answered || fmt.Sprint(val) != fmt.Sprint(c.Value)
	case "in", "not_in":
		list, _ := c.Value.([]any)
		found := false
		for _, item := range list {
			if answered && fmt.Sprint(val) == fmt.Sprint(item) {
				found = true
			}
		}
		if c.Op == "in" {
			return found
		}
		return !found
	case "gt", "gte", "lt", "lte":
		a, aok := toFloat(val)
		b, bok := toFloat(c.Value)
		if !answered || !aok || !bok {
			return false
		}
		switch c.Op {
		case "gt":
			return a > b
		case "gte":
			return a >= b
		case "lt":
			return a < b
		default:
			return a <= b
		}
	case "contains":
		return answered && regexp.QuoteMeta("") == "" && // always true guard for lint
			containsString(val, fmt.Sprint(c.Value))
	}
	return false
}

func containsString(val any, needle string) bool {
	switch v := val.(type) {
	case string:
		return regexp.MustCompile(regexp.QuoteMeta(needle)).MatchString(v)
	case []any:
		for _, item := range v {
			if fmt.Sprint(item) == needle {
				return true
			}
		}
	}
	return false
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	}
	return 0, false
}

func Visible(f v3.FormField, answers map[string]any) bool {
	if f.ShowIf == nil {
		return true
	}
	return EvalCondition(*f.ShowIf, answers)
}

func appliesTo(f v3.FormField, audience string) bool {
	ab := f.AnsweredBy
	if ab == "" {
		ab = "primary"
	}
	switch ab {
	case "everyone":
		return true
	case "primary":
		return audience == "primary"
	case "companions":
		return audience == "companion"
	}
	return false
}

func Validate(fields []v3.FormField, audience string, answers map[string]any) (map[string]any, *v3.Error) {
	clean := map[string]any{}
	for _, f := range fields {
		if !appliesTo(f, audience) {
			continue
		}
		if !Visible(f, answers) {
			continue // hidden: answer dropped, required ignored
		}
		val, answered := answers[f.Key]
		if !answered || val == nil || val == "" {
			if f.Required {
				return nil, v3.ErrInvalidInput("field " + f.Key + " is required")
			}
			continue
		}
		if e := checkType(f, val); e != nil {
			return nil, e
		}
		clean[f.Key] = val
	}
	return clean, nil
}

func checkType(f v3.FormField, val any) *v3.Error {
	bad := func(msg string) *v3.Error { return v3.ErrInvalidInput("field " + f.Key + ": " + msg) }
	s, isStr := val.(string)
	switch f.Type {
	case "email":
		if !isStr || !emailRe.MatchString(s) {
			return bad("invalid email format")
		}
	case "phone":
		if !isStr || !phoneRe.MatchString(s) {
			return bad("invalid phone format")
		}
	case "nik":
		if !isStr || !nikRe.MatchString(s) {
			return bad("nik must be 16 digits")
		}
	case "number":
		if _, ok := toFloat(val); !ok {
			return bad("must be a number")
		}
	case "select":
		if !isStr || !inOptions(s, f.Options) {
			return bad("value not in options")
		}
	case "multiselect":
		list, ok := val.([]any)
		if !ok {
			return bad("must be a list")
		}
		for _, item := range list {
			if !inOptions(fmt.Sprint(item), f.Options) {
				return bad("value not in options")
			}
		}
	case "checkbox":
		if _, ok := val.(bool); !ok {
			return bad("must be true or false")
		}
	case "date":
		if !isStr {
			return bad("must be a date string")
		}
		if _, err := time.Parse("2006-01-02", s); err != nil {
			return bad("must be YYYY-MM-DD")
		}
	}
	return nil
}

func inOptions(v string, opts []string) bool {
	for _, o := range opts {
		if o == v {
			return true
		}
	}
	return false
}

func ValidateSchema(fields []v3.FormField) *v3.Error {
	keys := map[string]bool{}
	for _, f := range fields {
		keys[f.Key] = true
	}
	deps := map[string][]string{}
	var collect func(c v3.Condition) []string
	collect = func(c v3.Condition) []string {
		var out []string
		if c.Field != "" {
			out = append(out, c.Field)
		}
		for _, sub := range append(c.All, c.Any...) {
			out = append(out, collect(sub)...)
		}
		return out
	}
	for _, f := range fields {
		if f.ShowIf == nil {
			continue
		}
		for _, ref := range collect(*f.ShowIf) {
			if !keys[ref] {
				return v3.ErrPublishValidation("field " + f.Key + " show_if references unknown field " + ref)
			}
			deps[f.Key] = append(deps[f.Key], ref)
		}
	}
	state := map[string]int{} // 0 unseen, 1 visiting, 2 done
	var visit func(k string) bool
	visit = func(k string) bool {
		if state[k] == 1 {
			return false
		}
		if state[k] == 2 {
			return true
		}
		state[k] = 1
		for _, d := range deps[k] {
			if !visit(d) {
				return false
			}
		}
		state[k] = 2
		return true
	}
	for k := range deps {
		if !visit(k) {
			return v3.ErrPublishValidation("show_if condition cycle involving field " + k)
		}
	}
	return nil
}
```

- [ ] **Step 4: Run** `go test ./internal/pkg/form/ -v` → PASS.
- [ ] **Step 5: Commit** — `git commit -am "feat(v3): form engine with show_if conditions, audiences and typed validation"`

---

### Task 7: Recurrence expansion (pure)

**Files:**
- Create: `internal/pkg/recurrence/recurrence.go`
- Test: `internal/pkg/recurrence/recurrence_test.go`

**Interfaces:**
- Consumes: `v3.Recurrence` (Task 3).
- Produces: `type Occurrence struct{ Key string; StartAt, EndAt time.Time }` (Key = `YYYY-MM-DD`, matches `v3_sessions.occurrence_key` unique constraint); `func Expand(rec v3.Recurrence, now time.Time, loc *time.Location) ([]Occurrence, *v3.Error)` — returns occurrences from `now` up to `GenerateAhead` periods (weeks for weekly, months for monthly; `custom_days` uses the explicit `CustomDates` list), respecting `Until` (inclusive date `2006-01-02`), start time + duration from `SessionDefaults`. Deterministic and idempotent: callers dedupe by Key against existing sessions.

- [ ] **Step 1: Write the failing test**

```go
// internal/pkg/recurrence/recurrence_test.go
package recurrence

import (
	v3 "go-community/internal/models/v3"
	"testing"
	"time"
)

var jkt, _ = time.LoadLocation("Asia/Jakarta")

func TestWeeklyExpand(t *testing.T) {
	rec := v3.Recurrence{Freq: "weekly", Interval: 1, ByDay: []string{"SUN"},
		SessionDefaults: v3.SessionDefaults{StartTime: "09:00", DurationMin: 120},
		GenerateAhead:   3}
	now := time.Date(2026, 7, 6, 8, 0, 0, 0, jkt) // Monday
	occ, err := Expand(rec, now, jkt)
	if err != nil {
		t.Fatal(err)
	}
	if len(occ) != 3 {
		t.Fatalf("want 3 sundays, got %d: %+v", len(occ), occ)
	}
	if occ[0].Key != "2026-07-12" || occ[0].StartAt.Hour() != 9 {
		t.Fatalf("first sunday 09:00: %+v", occ[0])
	}
	if occ[0].EndAt.Sub(occ[0].StartAt) != 2*time.Hour {
		t.Fatal("duration 120min")
	}
}

func TestUntilAndCustom(t *testing.T) {
	rec := v3.Recurrence{Freq: "weekly", Interval: 1, ByDay: []string{"SUN"},
		SessionDefaults: v3.SessionDefaults{StartTime: "09:00", DurationMin: 60},
		Until:           "2026-07-13", GenerateAhead: 8}
	occ, _ := Expand(rec, time.Date(2026, 7, 6, 0, 0, 0, 0, jkt), jkt)
	if len(occ) != 1 {
		t.Fatalf("until caps at 1, got %d", len(occ))
	}
	cust := v3.Recurrence{Freq: "custom_days", CustomDates: []string{"2026-12-24", "2026-12-25"},
		SessionDefaults: v3.SessionDefaults{StartTime: "18:30", DurationMin: 90}, GenerateAhead: 99}
	occ2, _ := Expand(cust, time.Date(2026, 7, 6, 0, 0, 0, 0, jkt), jkt)
	if len(occ2) != 2 || occ2[1].Key != "2026-12-25" {
		t.Fatalf("custom dates: %+v", occ2)
	}
}
```

- [ ] **Step 2: Run** `go test ./internal/pkg/recurrence/ -v` → FAIL.

- [ ] **Step 3: Implement**

```go
// internal/pkg/recurrence/recurrence.go
package recurrence

import (
	v3 "go-community/internal/models/v3"
	"time"
)

type Occurrence struct {
	Key     string
	StartAt time.Time
	EndAt   time.Time
}

var days = map[string]time.Weekday{"SUN": time.Sunday, "MON": time.Monday, "TUE": time.Tuesday,
	"WED": time.Wednesday, "THU": time.Thursday, "FRI": time.Friday, "SAT": time.Saturday}

func Expand(rec v3.Recurrence, now time.Time, loc *time.Location) ([]Occurrence, *v3.Error) {
	startClock, err := time.Parse("15:04", rec.SessionDefaults.StartTime)
	if err != nil {
		return nil, v3.ErrInvalidInput("recurrence start_time must be HH:MM")
	}
	dur := time.Duration(rec.SessionDefaults.DurationMin) * time.Minute
	var until time.Time
	if rec.Until != "" {
		u, err := time.ParseInLocation("2006-01-02", rec.Until, loc)
		if err != nil {
			return nil, v3.ErrInvalidInput("recurrence until must be YYYY-MM-DD")
		}
		until = u.AddDate(0, 0, 1) // inclusive
	}
	interval := rec.Interval
	if interval < 1 {
		interval = 1
	}
	mk := func(date time.Time) Occurrence {
		start := time.Date(date.Year(), date.Month(), date.Day(),
			startClock.Hour(), startClock.Minute(), 0, 0, loc)
		return Occurrence{Key: start.Format("2006-01-02"), StartAt: start, EndAt: start.Add(dur)}
	}
	var out []Occurrence
	within := func(d time.Time) bool { return until.IsZero() || d.Before(until) }

	switch rec.Freq {
	case "weekly":
		limit := now.AddDate(0, 0, 7*interval*rec.GenerateAhead)
		for d := now.In(loc); d.Before(limit); d = d.AddDate(0, 0, 1) {
			for _, wd := range rec.ByDay {
				if days[wd] == d.Weekday() && d.After(now) && within(d) {
					out = append(out, mk(d))
				}
			}
		}
	case "monthly":
		limit := now.AddDate(0, interval*rec.GenerateAhead, 0)
		for d := now.In(loc); d.Before(limit); d = d.AddDate(0, 0, 1) {
			for _, md := range rec.ByMonthDay {
				if d.Day() == md && d.After(now) && within(d) {
					out = append(out, mk(d))
				}
			}
		}
	case "custom_days":
		for _, ds := range rec.CustomDates {
			d, err := time.ParseInLocation("2006-01-02", ds, loc)
			if err != nil {
				return nil, v3.ErrInvalidInput("custom date must be YYYY-MM-DD: " + ds)
			}
			if d.After(now) && within(d) {
				out = append(out, mk(d))
			}
		}
	default:
		return nil, v3.ErrInvalidInput("recurrence freq must be weekly, monthly or custom_days")
	}
	return out, nil
}
```

- [ ] **Step 4: Run** `go test ./internal/pkg/recurrence/ -v` → PASS.
- [ ] **Step 5: Commit** — `git commit -am "feat(v3): deterministic recurrence expansion"`

---

### Task 8: Universal QR package — token codec + action registry

**Files:**
- Create: `internal/pkg/qr/qr.go`, `internal/pkg/qr/png.go`
- Test: `internal/pkg/qr/qr_test.go`
- Modify: `go.mod` (add `github.com/skip2/go-qrcode`)

**Interfaces:**
- Produces:
  - `type Payload struct{ Type string "json:\"t\""; Ref string "json:\"r\""; IssuedAt int64 "json:\"iat\""; ExpiresAt *int64 "json:\"exp,omitempty\"" }`
  - `type Codec struct{ ... }`; `func NewCodec(secret []byte, appDomain string) *Codec`; `(c *Codec) Encode(p Payload) string` (base64url(json)+"."+base64url(hmacSHA256)); `(c *Codec) Decode(token string) (Payload, *v3.Error)` (signature + expiry checks → `ErrInvalidToken`/`ErrExpiredToken`); `(c *Codec) URL(p Payload) string` → `https://<appDomain>/q/<token>`.
  - Registry: `type ActionCtx struct{ Ctx context.Context; Payload Payload; Caller eligibility.User; IsStaffForSession func(sessionCode string) bool; Params map[string]any }`; `type Action struct{ Name string; Allowed func(ActionCtx) bool; Handle func(ActionCtx) (any, *v3.Error) }`; `type Registry struct{...}`; `func NewRegistry() *Registry`; `(r *Registry) Register(qrType string, a Action)`; `(r *Registry) AllowedActions(qrType string, ctx ActionCtx) []string`; `(r *Registry) Dispatch(qrType, action string, ctx ActionCtx) (any, *v3.Error)`.
  - `func RenderPNG(url string, size int) ([]byte, error)` in png.go.
- QR types constants: `TypePersonal = "personal"`, `TypeAttendee = "attendee"`, `TypeSession = "session"`.

- [ ] **Step 1: Add dependency** — Run: `go get github.com/skip2/go-qrcode@latest && go mod tidy`

- [ ] **Step 2: Write the failing test**

```go
// internal/pkg/qr/qr_test.go
package qr

import (
	"context"
	v3 "go-community/internal/models/v3"
	"testing"
	"time"
)

func TestCodecRoundTrip(t *testing.T) {
	c := NewCodec([]byte("secret-key"), "app.example.org")
	p := Payload{Type: TypePersonal, Ref: "1234567890", IssuedAt: time.Now().Unix()}
	tok := c.Encode(p)
	got, err := c.Decode(tok)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Type != TypePersonal || got.Ref != "1234567890" {
		t.Fatalf("roundtrip: %+v", got)
	}
	if c.URL(p) != "https://app.example.org/q/"+tok {
		t.Fatal("url shape")
	}
}

func TestCodecRejectsTamperAndExpiry(t *testing.T) {
	c := NewCodec([]byte("secret-key"), "d")
	p := Payload{Type: TypeSession, Ref: "abc", IssuedAt: time.Now().Unix()}
	tok := c.Encode(p)
	if _, err := c.Decode(tok + "x"); err != v3.ErrInvalidToken {
		t.Fatal("tampered token must be invalid")
	}
	if _, err := NewCodec([]byte("other"), "d").Decode(tok); err != v3.ErrInvalidToken {
		t.Fatal("wrong key must be invalid")
	}
	past := time.Now().Add(-time.Hour).Unix()
	exp := c.Encode(Payload{Type: TypeAttendee, Ref: "x", IssuedAt: past - 10, ExpiresAt: &past})
	if _, err := c.Decode(exp); err != v3.ErrExpiredToken {
		t.Fatal("expired token must error EXPIRED_QR_TOKEN")
	}
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()
	r.Register(TypePersonal, Action{
		Name:    "event_checkin",
		Allowed: func(a ActionCtx) bool { return a.IsStaffForSession("s1") },
		Handle:  func(a ActionCtx) (any, *v3.Error) { return "done", nil },
	})
	staff := ActionCtx{Ctx: context.Background(), IsStaffForSession: func(string) bool { return true }}
	member := ActionCtx{Ctx: context.Background(), IsStaffForSession: func(string) bool { return false }}
	if acts := r.AllowedActions(TypePersonal, staff); len(acts) != 1 || acts[0] != "event_checkin" {
		t.Fatalf("staff sees action: %v", acts)
	}
	if acts := r.AllowedActions(TypePersonal, member); len(acts) != 0 {
		t.Fatal("member sees nothing")
	}
	if _, err := r.Dispatch(TypePersonal, "event_checkin", member); err != v3.ErrForbidden {
		t.Fatal("dispatch enforces Allowed")
	}
	if out, err := r.Dispatch(TypePersonal, "event_checkin", staff); err != nil || out != "done" {
		t.Fatal("staff dispatch works")
	}
	if _, err := r.Dispatch(TypePersonal, "ghost", staff); err != v3.ErrNotFound {
		t.Fatal("unknown action is 404")
	}
}
```

- [ ] **Step 3: Run** `go test ./internal/pkg/qr/ -v` → FAIL.

- [ ] **Step 4: Implement**

```go
// internal/pkg/qr/qr.go
package qr

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	v3 "go-community/internal/models/v3"
	"go-community/internal/pkg/eligibility"
)

const (
	TypePersonal = "personal"
	TypeAttendee = "attendee"
	TypeSession  = "session"
)

type Payload struct {
	Type      string `json:"t"`
	Ref       string `json:"r"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt *int64 `json:"exp,omitempty"`
}

type Codec struct {
	secret    []byte
	appDomain string
}

func NewCodec(secret []byte, appDomain string) *Codec {
	return &Codec{secret: secret, appDomain: appDomain}
}

func (c *Codec) sign(body []byte) string {
	m := hmac.New(sha256.New, c.secret)
	m.Write(body)
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

func (c *Codec) Encode(p Payload) string {
	body, _ := json.Marshal(p)
	return base64.RawURLEncoding.EncodeToString(body) + "." + c.sign(body)
}

func (c *Codec) Decode(token string) (Payload, *v3.Error) {
	var p Payload
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return p, v3.ErrInvalidToken
	}
	body, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || !hmac.Equal([]byte(c.sign(body)), []byte(parts[1])) {
		return p, v3.ErrInvalidToken
	}
	if err := json.Unmarshal(body, &p); err != nil {
		return p, v3.ErrInvalidToken
	}
	if p.ExpiresAt != nil && time.Now().Unix() > *p.ExpiresAt {
		return p, v3.ErrExpiredToken
	}
	return p, nil
}

func (c *Codec) URL(p Payload) string {
	return "https://" + c.appDomain + "/q/" + c.Encode(p)
}

type ActionCtx struct {
	Ctx               context.Context
	Payload           Payload
	Caller            eligibility.User
	IsStaffForSession func(sessionCode string) bool
	Params            map[string]any
}

type Action struct {
	Name    string
	Allowed func(ActionCtx) bool
	Handle  func(ActionCtx) (any, *v3.Error)
}

type Registry struct{ actions map[string][]Action }

func NewRegistry() *Registry { return &Registry{actions: map[string][]Action{}} }

func (r *Registry) Register(qrType string, a Action) {
	r.actions[qrType] = append(r.actions[qrType], a)
}

func (r *Registry) AllowedActions(qrType string, ctx ActionCtx) []string {
	out := []string{}
	for _, a := range r.actions[qrType] {
		if a.Allowed == nil || a.Allowed(ctx) {
			out = append(out, a.Name)
		}
	}
	return out
}

func (r *Registry) Dispatch(qrType, action string, ctx ActionCtx) (any, *v3.Error) {
	for _, a := range r.actions[qrType] {
		if a.Name == action {
			if a.Allowed != nil && !a.Allowed(ctx) {
				return nil, v3.ErrForbidden
			}
			return a.Handle(ctx)
		}
	}
	return nil, v3.ErrNotFound
}
```

```go
// internal/pkg/qr/png.go
package qr

import qrcode "github.com/skip2/go-qrcode"

func RenderPNG(url string, size int) ([]byte, error) {
	return qrcode.Encode(url, qrcode.Medium, size)
}
```

- [ ] **Step 5: Run** `go test ./internal/pkg/qr/ -v` → PASS. Commit: `git add internal/pkg/qr go.mod go.sum && git commit -m "feat(v3): universal QR token codec, action registry and PNG rendering"`

---

### Task 9: Entities + repositories + aggregate wiring

**Files:**
- Create: `internal/models/v3/entities.go`
- Create: `internal/repositories/pgsql/v3_repositories.go`
- Modify: `internal/repositories/pgsql/main_pg_repository.go` (add `V3 *V3Repositories` field, constructed with `NewV3Repositories(db)`)
- Test: `tests/integration/v3/repositories_test.go`

**Interfaces:**
- Produces GORM entities (table names via `TableName()`): `EventV3` (`v3_events`), `SessionV3`, `TicketCategory` (`v3_session_ticket_categories`), `RegistrationV3`, `AttendeeV3` (`v3_registration_attendees`), `AttendanceLog` (`v3_attendance_logs`), `OutboxMessage` (`v3_notification_outbox`), `OrganizerV3` (`v3_event_organizers`), `TemplateV3` (`v3_event_templates`). JSONB columns typed `[]byte` with `gorm:"type:jsonb"`; arrays `pq.StringArray`.
- Produces `type V3Repositories struct` with fields `Event, Session, Category, Registration, Attendee, Attendance, Outbox, Organizer, Template` — interfaces:

```go
type EventV3Repository interface {
	Create(ctx context.Context, e *v3.EventV3) error
	GetByCode(ctx context.Context, code string) (*v3.EventV3, error) // gorm.ErrRecordNotFound when missing
	Update(ctx context.Context, e *v3.EventV3) error
	ListVisible(ctx context.Context, limit, offset int) ([]v3.EventV3, error) // status=published AND is_visible
	ListWithRecurrence(ctx context.Context) ([]v3.EventV3, error)
	ListDuePublishFlips(ctx context.Context, now time.Time) ([]v3.EventV3, error)
}
type SessionV3Repository interface {
	Create(ctx context.Context, s *v3.SessionV3) error
	BulkCreate(ctx context.Context, ss []v3.SessionV3) error
	GetByCode(ctx context.Context, code string) (*v3.SessionV3, error)
	GetByCodeForUpdate(ctx context.Context, code string) (*v3.SessionV3, error) // FOR UPDATE
	ListByEvent(ctx context.Context, eventID int64) ([]v3.SessionV3, error)
	Update(ctx context.Context, s *v3.SessionV3) error
	AddBooked(ctx context.Context, sessionID int64, delta int) error
}
type TicketCategoryRepository interface {
	BulkCreate(ctx context.Context, cs []v3.TicketCategory) error
	GetForUpdate(ctx context.Context, id int64) (*v3.TicketCategory, error)
	ListBySession(ctx context.Context, sessionID int64) ([]v3.TicketCategory, error)
	Update(ctx context.Context, c *v3.TicketCategory) error
	AddBooked(ctx context.Context, id int64, delta int) error
}
type RegistrationV3Repository interface {
	Create(ctx context.Context, r *v3.RegistrationV3) error
	GetByID(ctx context.Context, id uuid.UUID) (*v3.RegistrationV3, error)
	HasConfirmed(ctx context.Context, sessionID int64, communityID string) (bool, error)
	ListByUser(ctx context.Context, communityID string) ([]v3.RegistrationV3, error)
	Update(ctx context.Context, r *v3.RegistrationV3) error
}
type AttendeeV3Repository interface {
	BulkCreate(ctx context.Context, as []v3.AttendeeV3) error
	GetByID(ctx context.Context, id uuid.UUID) (*v3.AttendeeV3, error)
	ListByRegistration(ctx context.Context, regID uuid.UUID) ([]v3.AttendeeV3, error)
	ListBySession(ctx context.Context, sessionID int64) ([]v3.AttendeeV3, error) // join registrations
	FindBySessionAndCommunity(ctx context.Context, sessionID int64, communityID string) (*v3.AttendeeV3, error)
	Update(ctx context.Context, a *v3.AttendeeV3) error
}
type AttendanceLogRepository interface {
	Append(ctx context.Context, l *v3.AttendanceLog) error
	ListBySession(ctx context.Context, sessionID int64) ([]v3.AttendanceLog, error)
	CountHeadcount(ctx context.Context, sessionID int64) (int64, error) // attendee_id IS NULL AND outcome='checked_in'
}
type OutboxRepository interface {
	Enqueue(ctx context.Context, m *v3.OutboxMessage) error
	ClaimBatch(ctx context.Context, limit int) ([]v3.OutboxMessage, error) // FOR UPDATE SKIP LOCKED, status=pending, scheduled_at<=now
	MarkSent(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID, maxAttempts int, backoff time.Duration) error
}
type OrganizerV3Repository interface {
	Add(ctx context.Context, o *v3.OrganizerV3) error
	IsOrganizer(ctx context.Context, eventID int64, communityID string) (bool, error)
	ListEventIDs(ctx context.Context, communityID string) ([]int64, error)
}
type TemplateV3Repository interface {
	Create(ctx context.Context, t *v3.TemplateV3) error
	GetByID(ctx context.Context, id int64) (*v3.TemplateV3, error)
	List(ctx context.Context, communityID string) ([]v3.TemplateV3, error) // is_system OR created_by
}
```

All implementations are thin GORM (Raw SQL only for `FOR UPDATE`/`SKIP LOCKED`). `NewV3Repositories(db *gorm.DB) *V3Repositories`. In `pgsql.New`, add `V3: NewV3Repositories(db)` — **crucially, `New(tx)` inside `Atomic` then yields tx-scoped v3 repos for free.**

- [ ] **Step 1: Write entities** (`internal/models/v3/entities.go`)

```go
package v3

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type EventV3 struct {
	ID                 int64
	Code               string
	Slug               string
	Title              string
	Description        string
	Topics             pq.StringArray `gorm:"type:text[]"`
	TermsAndConditions string
	ImageLinks         pq.StringArray `gorm:"type:text[]"`
	CampusCodes        pq.StringArray `gorm:"type:text[]"`
	Status             string
	PublishAt          *time.Time
	UnpublishAt        *time.Time
	IsVisible          bool
	ContentSections    []byte `gorm:"type:jsonb"`
	Contacts           []byte `gorm:"type:jsonb"`
	VenueAddress       []byte `gorm:"type:jsonb"`
	Eligibility        []byte `gorm:"type:jsonb"`
	RegistrationConfig []byte `gorm:"type:jsonb"`
	GeoConfig          []byte `gorm:"type:jsonb"`
	Recurrence         []byte `gorm:"type:jsonb"`
	CreatedBy          string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          gorm.DeletedAt
}

func (EventV3) TableName() string { return "v3_events" }

type SessionV3 struct {
	ID                      int64
	Code                    string
	EventID                 int64
	Title                   string
	Description             string
	StartAt                 time.Time
	EndAt                   time.Time
	RegisterStartAt         *time.Time
	RegisterEndAt           *time.Time
	CheckinOpenAt           *time.Time
	CheckinCloseAt          *time.Time
	LocationType            string
	LocationName            string
	OnlineURL               string `gorm:"column:online_url"`
	TotalSeats              int
	BookedSeats             int
	MarkedSoldOut           bool
	AttendanceModes         pq.StringArray `gorm:"type:text[]"`
	EnableCheckout          bool
	GeneratedFromRecurrence bool
	OccurrenceKey           *string
	Status                  string
	CreatedAt               time.Time
	UpdatedAt               time.Time
	DeletedAt               gorm.DeletedAt
}

func (SessionV3) TableName() string { return "v3_sessions" }

type TicketCategory struct {
	ID            int64
	SessionID     int64
	Code          string
	Name          string
	Description   string
	TotalSeats    int
	BookedSeats   int
	MarkedSoldOut bool
	SortOrder     int
}

func (TicketCategory) TableName() string { return "v3_session_ticket_categories" }

type RegistrationV3 struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	SessionID    int64
	CategoryID   *int64
	RegisteredBy string
	PartySize    int
	Status       string
	Source       string
	RegisteredAt time.Time
	CancelledAt  *time.Time
}

func (RegistrationV3) TableName() string { return "v3_registrations" }

type AttendeeV3 struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	RegistrationID  uuid.UUID `gorm:"type:uuid"`
	CommunityID     *string
	Name            string
	Answers         []byte `gorm:"type:jsonb"`
	AnswerRevisions []byte `gorm:"type:jsonb"`
	Status          string
	AttendedAt      *time.Time
	CheckedOutAt    *time.Time
}

func (AttendeeV3) TableName() string { return "v3_registration_attendees" }

type AttendanceLog struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	SessionID   int64
	AttendeeID  *uuid.UUID `gorm:"type:uuid"`
	CommunityID *string
	Action      string
	Mode        string
	CheckedBy   string
	Lat         *float64
	Lng         *float64
	AccuracyM   *float64 `gorm:"column:accuracy_m"`
	GeoResult   string
	Outcome     string
	CreatedAt   time.Time
}

func (AttendanceLog) TableName() string { return "v3_attendance_logs" }

type OutboxMessage struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Channel     string
	Recipient   string
	Template    string
	Payload     []byte `gorm:"type:jsonb"`
	Status      string
	Attempts    int
	ScheduledAt time.Time
	SentAt      *time.Time
}

func (OutboxMessage) TableName() string { return "v3_notification_outbox" }

type OrganizerV3 struct {
	ID          int64
	EventID     int64
	CommunityID string
	Role        string
}

func (OrganizerV3) TableName() string { return "v3_event_organizers" }

type TemplateV3 struct {
	ID          int64
	Name        string
	Description string
	Snapshot    []byte `gorm:"type:jsonb"`
	IsSystem    bool
	CreatedBy   string
	CreatedAt   time.Time
}

func (TemplateV3) TableName() string { return "v3_event_templates" }
```

- [ ] **Step 2: Write failing integration test** (`tests/integration/v3/repositories_test.go`)

```go
package v3_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	v3 "go-community/internal/models/v3"
	"go-community/internal/repositories/pgsql"
)

func TestSeatLockAndOutboxClaim(t *testing.T) {
	db := testDB(t)
	repos := pgsql.NewV3Repositories(db)
	ctx := context.Background()

	ev := &v3.EventV3{Code: "evt0001", Slug: "evt-1", Title: "T", Status: "published", IsVisible: true}
	if err := repos.Event.Create(ctx, ev); err != nil {
		t.Fatal(err)
	}
	s := &v3.SessionV3{Code: "ses0001", EventID: ev.ID, Title: "S",
		StartAt: time.Now().Add(time.Hour), EndAt: time.Now().Add(2 * time.Hour), TotalSeats: 1}
	if err := repos.Session.Create(ctx, s); err != nil {
		t.Fatal(err)
	}
	locked, err := repos.Session.GetByCodeForUpdate(ctx, "ses0001")
	if err != nil || locked.ID != s.ID {
		t.Fatalf("lock read: %v", err)
	}
	if err := repos.Session.AddBooked(ctx, s.ID, 1); err != nil {
		t.Fatal(err)
	}

	m := &v3.OutboxMessage{ID: uuid.New(), Recipient: "a@b.c", Channel: "email",
		Template: "registration_confirmed", Payload: []byte(`{}`), Status: "pending",
		ScheduledAt: time.Now().Add(-time.Minute)}
	if err := repos.Outbox.Enqueue(ctx, m); err != nil {
		t.Fatal(err)
	}
	batch, err := repos.Outbox.ClaimBatch(ctx, 10)
	if err != nil || len(batch) != 1 {
		t.Fatalf("claim: %v %d", err, len(batch))
	}
	if err := repos.Outbox.MarkSent(ctx, m.ID); err != nil {
		t.Fatal(err)
	}
	if batch2, _ := repos.Outbox.ClaimBatch(ctx, 10); len(batch2) != 0 {
		t.Fatal("sent messages must not be reclaimed")
	}
}
```

- [ ] **Step 3: Run** `TEST_DATABASE_URL=... go test ./tests/integration/v3/ -run TestSeatLock -v` → FAIL (repos missing).

- [ ] **Step 4: Implement repositories** (`internal/repositories/pgsql/v3_repositories.go`) — thin GORM; key non-trivial methods shown in full, the rest are single-statement GORM calls following the same shape:

```go
package pgsql

import (
	"context"
	"time"

	"github.com/google/uuid"
	v3 "go-community/internal/models/v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type V3Repositories struct {
	Event        EventV3Repository
	Session      SessionV3Repository
	Category     TicketCategoryRepository
	Registration RegistrationV3Repository
	Attendee     AttendeeV3Repository
	Attendance   AttendanceLogRepository
	Outbox       OutboxRepository
	Organizer    OrganizerV3Repository
	Template     TemplateV3Repository
}

func NewV3Repositories(db *gorm.DB) *V3Repositories {
	return &V3Repositories{
		Event:        &eventV3Repo{db}, Session: &sessionV3Repo{db},
		Category:     &categoryRepo{db}, Registration: &registrationV3Repo{db},
		Attendee:     &attendeeRepo{db}, Attendance: &attendanceRepo{db},
		Outbox:       &outboxRepo{db}, Organizer: &organizerRepo{db}, Template: &templateRepo{db},
	}
}

// --- interfaces exactly as declared in the plan's Interfaces block ---
// (paste the interface block from the task header here, unchanged)

type sessionV3Repo struct{ db *gorm.DB }

func (r *sessionV3Repo) GetByCodeForUpdate(ctx context.Context, code string) (*v3.SessionV3, error) {
	var s v3.SessionV3
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("code = ?", code).First(&s).Error
	return &s, err
}

func (r *sessionV3Repo) AddBooked(ctx context.Context, id int64, delta int) error {
	return r.db.WithContext(ctx).Model(&v3.SessionV3{}).Where("id = ?", id).
		UpdateColumn("booked_seats", gorm.Expr("booked_seats + ?", delta)).Error
}

// Create/GetByCode/ListByEvent/Update/BulkCreate: standard one-line GORM calls.

type outboxRepo struct{ db *gorm.DB }

func (r *outboxRepo) Enqueue(ctx context.Context, m *v3.OutboxMessage) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *outboxRepo) ClaimBatch(ctx context.Context, limit int) ([]v3.OutboxMessage, error) {
	var out []v3.OutboxMessage
	err := r.db.WithContext(ctx).Raw(`
		UPDATE v3_notification_outbox SET attempts = attempts + 1
		WHERE id IN (
			SELECT id FROM v3_notification_outbox
			WHERE status = 'pending' AND scheduled_at <= now()
			ORDER BY scheduled_at LIMIT ? FOR UPDATE SKIP LOCKED)
		RETURNING *`, limit).Scan(&out).Error
	return out, err
}

func (r *outboxRepo) MarkSent(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&v3.OutboxMessage{}).Where("id = ?", id).
		Updates(map[string]any{"status": "sent", "sent_at": time.Now()}).Error
}

func (r *outboxRepo) MarkFailed(ctx context.Context, id uuid.UUID, maxAttempts int, backoff time.Duration) error {
	return r.db.WithContext(ctx).Exec(`
		UPDATE v3_notification_outbox
		SET status = CASE WHEN attempts >= ? THEN 'failed' ELSE 'pending' END,
		    scheduled_at = now() + ?::interval
		WHERE id = ?`, maxAttempts, backoff.String(), id).Error
}

// eventV3Repo, categoryRepo, registrationV3Repo, attendeeRepo, attendanceRepo,
// organizerRepo, templateRepo: implement every interface method with standard
// GORM Where/First/Find/Create/Save calls; attendeeRepo.ListBySession and
// FindBySessionAndCommunity join through v3_registrations:
//   r.db.WithContext(ctx).
//     Joins("JOIN v3_registrations reg ON reg.id = v3_registration_attendees.registration_id").
//     Where("reg.session_id = ? AND reg.status = 'confirmed'", sessionID).Find(&out)
```

Then in `main_pg_repository.go` add field `V3 *V3Repositories` to `PostgreRepositories` and `V3: NewV3Repositories(db),` inside `New(db)`.

- [ ] **Step 5: Run** the integration test → PASS. Also `go build ./...` → OK. Commit: `git add internal/ tests/ && git commit -m "feat(v3): entities and repositories with row-lock booking and skip-locked outbox"`

---

### Task 10: Registration usecase — register, cancel, edit answers

**Files:**
- Create: `internal/usecases/v3/registration_usecase.go`, `internal/models/v3/dto.go` (registration DTOs)
- Test: `tests/integration/v3/registration_test.go`

**Interfaces:**
- Consumes: repos (Task 9), `eligibility.Check`, `geo.Validate`, `form.Validate`, `v3` configs/errors, `pgsql.PostgreRepositories.Transaction.Atomic`.
- Produces:

```go
type RegistrationUsecase struct { r *pgsql.PostgreRepositories }
func NewRegistrationUsecase(r *pgsql.PostgreRepositories) *RegistrationUsecase

type RegisterInput struct {
	SessionCode  string
	User         eligibility.User
	Coords       *geo.Coords
	AccessCode   string
	CategoryCode string           // optional
	Primary      map[string]any   // primary answers (must include "name")
	Companions   []map[string]any // per-companion answers; count_only ⇒ [{"name": "..."}]
}
type RegisterOutput struct {
	Registration v3.RegistrationV3
	Attendees    []v3.AttendeeV3
}
func (u *RegistrationUsecase) Register(ctx context.Context, in RegisterInput) (*RegisterOutput, *v3.Error)
func (u *RegistrationUsecase) Cancel(ctx context.Context, regID uuid.UUID, by eligibility.User, isOrganizer bool) *v3.Error
func (u *RegistrationUsecase) EditAnswers(ctx context.Context, attendeeID uuid.UUID, by eligibility.User, isOrganizer bool, answers map[string]any) *v3.Error
func (u *RegistrationUsecase) ListMine(ctx context.Context, communityID string) ([]v3.RegistrationV3, *v3.Error)
```

Register algorithm (all inside one `Atomic` call after pure validation):
1. Load session+event by code (`ErrNotFound`). Event must be `published` and visible; session `scheduled`; `registration_config.mode != "none"` else `ErrRegistrationClosed`.
2. Resolve window: if `Phases` present, find any phase where `now ∈ [Start,End)` AND (no AccessCode or matches case-insensitively → else try next phase; if a matching-window phase requires a code and none matched, `ErrInvalidAccessCode`) AND eligibility (phase override else event eligibility) passes. No active phase ⇒ `ErrRegistrationClosed`. Without phases: `now ∈ [register_start_at, register_end_at]` and event eligibility, else `ErrRegistrationClosed`/`ErrNotEligible`.
3. `marked_sold_out` (session, and category when used) ⇒ `ErrMarkedSoldOut`.
4. Party size = 1+len(Companions); cap = phase override else `MaxPerRegistration` (0 ⇒ unlimited) ⇒ `ErrInvalidInput`.
5. `geo.Validate(cfg, "web_registration", coords, false)` — error passthrough.
6. `form.Validate(fields, "primary", Primary)`; when `CompanionDetail=="full"` validate each companion with audience `"companion"`; `count_only` ⇒ keep only `name` (default `"Guest of <primary name>"` when empty).
7. Duplicate guard: `HasConfirmed(session,user)` unless `AllowMultiple` ⇒ `ErrAlreadyRegistered`.
8. Atomic: lock session (or category) FOR UPDATE → seat check (`TotalSeats>0 && Booked+N>Total` ⇒ `ErrQuotaFull`) → create registration + attendees (primary attendee gets `CommunityID=&user.CommunityID`) → `AddBooked` → enqueue outbox `registration_confirmed` (recipient = primary email answer if provided, else lookup skipped — payload carries registration ID; dispatcher resolves).
Cancel: owner or organizer; already cancelled ⇒ `ErrCancelled`; Atomic: set status/cancelled_at, attendees not `attended` → `cancelled`, decrement seats by count of non-attended, enqueue `registration_cancelled`.
EditAnswers: policy from event config; organizer bypasses window; `Fields` whitelist filters incoming keys; re-run `form.Validate`; append previous answers to `answer_revisions` (JSON array of `{at, answers}`); `until` anchors: `register_end`→session.RegisterEndAt, `checkin_open`→CheckinOpenAt, `session_start`→StartAt, `never`⇒ organizer only.

- [ ] **Step 1: Write the failing integration test** — cover: happy path party of 3 creates 3 attendees + outbox row + booked=3; seat race (session with 1 seat, two goroutines register, exactly one wins, other gets `QUOTA_FULL`); duplicate guard; access-code phase (wrong code fails, right code passes); cancel frees seats; edit answers inside window mutates + appends revision, outside window fails `EDIT_WINDOW_CLOSED`.

```go
// tests/integration/v3/registration_test.go  (core assertions; full file follows this pattern)
func TestRegisterHappyPathAndRace(t *testing.T) {
	db := testDB(t)
	repos := pgsql.New(db) // aggregate incl. V3 + Transaction
	uc := v3uc.NewRegistrationUsecase(repos)
	ctx := context.Background()
	ev, ses := seedEvent(t, db, seedOpts{TotalSeats: 4, Mode: "required"})
	_ = ev
	out, err := uc.Register(ctx, v3uc.RegisterInput{SessionCode: ses.Code,
		User: eligibility.User{CommunityID: "100"}, Primary: map[string]any{"name": "Maria"},
		Companions: []map[string]any{{"name": "A"}, {"name": "B"}}})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if len(out.Attendees) != 3 {
		t.Fatal("3 attendees")
	}
	var booked int
	db.Raw("SELECT booked_seats FROM v3_sessions WHERE id = ?", ses.ID).Scan(&booked)
	if booked != 3 {
		t.Fatalf("booked=3 got %d", booked)
	}
	var pending int64
	db.Raw("SELECT count(*) FROM v3_notification_outbox WHERE status='pending'").Scan(&pending)
	if pending != 1 {
		t.Fatal("outbox row enqueued in same tx")
	}

	// race for the last seat
	_, ses2 := seedEvent(t, db, seedOpts{TotalSeats: 1, Mode: "required"})
	errs := make(chan *v3.Error, 2)
	for i := 0; i < 2; i++ {
		go func(n int) {
			_, e := uc.Register(ctx, v3uc.RegisterInput{SessionCode: ses2.Code,
				User: eligibility.User{CommunityID: fmt.Sprintf("u%d", n)},
				Primary: map[string]any{"name": "X"}})
			errs <- e
		}(i)
	}
	a, b := <-errs, <-errs
	if (a == nil) == (b == nil) {
		t.Fatalf("exactly one must win: %v / %v", a, b)
	}
	winnerless := a
	if winnerless == nil {
		winnerless = b
	}
	if winnerless.Code != "QUOTA_FULL" {
		t.Fatalf("loser gets QUOTA_FULL, got %s", winnerless.Code)
	}
}
```

Include in the same file `seedEvent(t, db, opts)` helper that inserts an `EventV3` (published, visible, registration_config JSON with a required `name` field and given mode) and a `SessionV3` (window open now−1h..now+1h) and tests for duplicate guard / phases / cancel / edit as described above — each one seeds, calls the usecase, asserts the typed error code or DB state.

- [ ] **Step 2: Run** → FAIL. **Step 3: Implement** `registration_usecase.go` exactly per the algorithm above (the `Atomic` closure receives tx-scoped `*PostgreRepositories`, so seat lock + inserts + outbox all share the transaction):

```go
// skeleton of the critical section — the full method implements steps 1–7 before this
err := u.r.Transaction.Atomic(ctx, func(ctx context.Context, tx *pgsql.PostgreRepositories) error {
	if in.CategoryCode != "" {
		cat, err := lockCategory(ctx, tx, session.ID, in.CategoryCode)
		if err != nil { return err }
		if cat.MarkedSoldOut { return v3.ErrMarkedSoldOut }
		if cat.TotalSeats > 0 && cat.BookedSeats+party > cat.TotalSeats { return v3.ErrQuotaFull }
		categoryID = &cat.ID
		defer func() {}()
		if err := tx.V3.Category.AddBooked(ctx, cat.ID, party); err != nil { return err }
	} else {
		locked, err := tx.V3.Session.GetByCodeForUpdate(ctx, session.Code)
		if err != nil { return err }
		if locked.MarkedSoldOut { return v3.ErrMarkedSoldOut }
		if locked.TotalSeats > 0 && locked.BookedSeats+party > locked.TotalSeats { return v3.ErrQuotaFull }
		if err := tx.V3.Session.AddBooked(ctx, locked.ID, party); err != nil { return err }
	}
	if err := tx.V3.Registration.Create(ctx, &reg); err != nil { return err }
	if err := tx.V3.Attendee.BulkCreate(ctx, attendees); err != nil { return err }
	return tx.V3.Outbox.Enqueue(ctx, &outboxMsg)
})
if err != nil { return nil, v3.AsError(err) }
```

- [ ] **Step 4: Run** all registration tests → PASS. **Step 5: Commit** — `git commit -am "feat(v3): registration usecase with seat locking, phases, duplicate guard, cancel and answer editing"`

---

### Task 11: Check-in usecase — 4 modes, walk-in, checkout, logs

**Files:**
- Create: `internal/usecases/v3/checkin_usecase.go`
- Test: `tests/integration/v3/checkin_test.go`

**Interfaces:**
- Produces:

```go
type CheckinUsecase struct{ r *pgsql.PostgreRepositories }
func NewCheckinUsecase(r *pgsql.PostgreRepositories) *CheckinUsecase

type CheckinInput struct {
	SessionCode string
	Mode        string // personal_qr | session_qr | registration_qr | manual | headcount
	Action      string // checkin | checkout
	AttendeeID  *uuid.UUID       // registration_qr
	CommunityID string           // personal_qr target OR session_qr caller
	ManualForm  map[string]any   // manual with info
	Coords      *geo.Coords
	Override    bool
	ActedBy     eligibility.User // staff for staff modes; the member for session_qr
	IsStaff     bool
}
type CheckinResult struct {
	Outcome  string // checked_in | already_checked_in | checked_out | already_checked_out | headcount_recorded
	Attendee *v3.AttendeeV3
}
func (u *CheckinUsecase) Act(ctx context.Context, in CheckinInput) (*CheckinResult, *v3.Error)
```

Rules implemented (per spec §5): mode must be in session `attendance_modes` (headcount rides on `manual`); check-in window `[checkin_open_at, checkin_close_at]` when set; staff modes require `IsStaff`; geo per mode via `geo.Validate(cfg, in.Mode, coords, in.Override)` — on `require` failure append log `outcome=rejected_geo` then return the geo error; **every path appends exactly one `AttendanceLog`**; personal_qr with no registration → walk-in auto-create (registration `source=walk_in` + attendee, seat-locked like Task 10, `ErrQuotaFull` if full and seats finite); idempotency: attendee already `attended` on checkin ⇒ `Outcome=already_checked_in`, log `outcome=already_checked_in`, nil error; checkout requires `enable_checkout` (`ErrForbidden` otherwise) and prior check-in (`ErrNotCheckedIn`), sets `checked_out`/`checked_out_at`; headcount appends log with nil attendee, touches no attendee rows.

- [ ] **Step 1: Failing integration test** — one test per mode: registration_qr happy + double-scan idempotent; personal_qr registered member; personal_qr walk-in auto-creation increments booked_seats; manual headcount creates log with NULL attendee and `CountHeadcount`==1; checkout happy + `NOT_CHECKED_IN` + disabled-checkout `FORBIDDEN`; geo require + outside coords ⇒ error AND a `rejected_geo` log row exists. Reuse `seedEvent` and add `seedRegistered(t, db, uc, ses)` helper that registers one attendee via Task 10's usecase.

- [ ] **Step 2: Run** → FAIL. **Step 3: Implement** per rules (checkout/checkin share one code path switching on `Action`; walk-in creation reuses the same Atomic seat-lock block as Task 10 via a small shared helper `bookSeats(ctx, tx, session, categoryID, n) *v3.Error` — move that helper into `registration_usecase.go` and export within package).

- [ ] **Step 4: Run** → PASS. **Step 5: Commit** — `git commit -am "feat(v3): check-in/checkout usecase with 4 modes, walk-in creation and append-only logs"`

---

### Task 12: Notifications — outbox dispatcher, email notifier, PDF ticket

**Files:**
- Create: `internal/pkg/notify/notify.go`, `internal/pkg/pdfticket/ticket.go`, `internal/usecases/v3/notification_usecase.go`
- Modify: `go.mod` (add `github.com/jung-kurt/gofpdf`), `internal/config/config.go` (add `V3` block)
- Test: `internal/pkg/pdfticket/ticket_test.go`, `tests/integration/v3/outbox_test.go`

**Interfaces:**
- Config addition (in `internal/config/config.go`, `Configuration` gains):

```go
V3 V3Config `mapstructure:"v3"`
// with:
type V3Config struct {
	AppDomain    string            `mapstructure:"app_domain"`
	QRSecret     string            `mapstructure:"qr_secret"`
	Email        EmailConfig       `mapstructure:"email"`
}
type EmailConfig struct {
	Host, Username, Password, From string `mapstructure:"host"...` // host, port, username, password, from
	Port int
}
```

- `notify.Notifier` interface: `Send(ctx context.Context, recipient, subject, htmlBody string, attachments []notify.Attachment) error`; `type Attachment struct{ Filename string; Mime string; Data []byte }`; `func NewSMTPNotifier(cfg config.EmailConfig) Notifier` (net/smtp + MIME multipart, no new deps); channel router `map[string]Notifier`.
- `pdfticket.Generate(in pdfticket.Input) ([]byte, error)` where `Input{EventTitle, SessionTitle, When, Where string; Contacts []v3.Contact; Attendees []pdfticket.AttendeeQR}` and `AttendeeQR{Name string; PNG []byte}` — one page per attendee with embedded QR PNG.
- `NotificationUsecase`:

```go
type NotificationUsecase struct {
	r *pgsql.PostgreRepositories; notifiers map[string]notify.Notifier
	codec *qr.Codec
}
func NewNotificationUsecase(r *pgsql.PostgreRepositories, n map[string]notify.Notifier, codec *qr.Codec) *NotificationUsecase
func (u *NotificationUsecase) Dispatch(ctx context.Context, batchSize int) (sent, failed int, err *v3.Error)
```

Dispatch: `ClaimBatch` → for each message build content by `Template` (`registration_confirmed` ⇒ load registration+attendees+session+event, render simple HTML, generate PDF with per-attendee QR URLs via codec, attach; `registration_cancelled`/`event_cancelled` ⇒ plain HTML) → `Send` → `MarkSent` or `MarkFailed(id, 5, backoff=1<<attempts minutes)`. A message whose registration has no resolvable email (no email answer and empty recipient) is marked sent-with-skip (MarkSent) so it never loops.

- [ ] **Step 1:** `go get github.com/jung-kurt/gofpdf@latest && go mod tidy`
- [ ] **Step 2: PDF unit test** — `Generate` with 2 attendees returns bytes starting `%PDF` and length > 1KB; QR PNG produced by `qr.RenderPNG("https://x/q/abc", 256)`.
- [ ] **Step 3:** implement `pdfticket` + run → PASS.
- [ ] **Step 4: Outbox integration test** — enqueue a `registration_confirmed` for a seeded registration with a **fake notifier** (records calls in-memory, injectable since `NewNotificationUsecase` takes the map); `Dispatch` → notifier called once with a PDF attachment; message `sent`; a notifier that always errors → after `Dispatch` message stays `pending` with `attempts=1` and future `scheduled_at`; after 5 failures status `failed`.
- [ ] **Step 5:** implement `notify` (SMTP impl + fake used only in tests) and `notification_usecase.go`; run → PASS.
- [ ] **Step 6: Commit** — `git commit -am "feat(v3): notification outbox dispatcher, smtp notifier and pdf tickets"`

---

### Task 13: Event usecase — CRUD, publish gate, templates, organizers, recurrence generation

**Files:**
- Create: `internal/usecases/v3/event_usecase.go`, `internal/usecases/v3/recurrence_usecase.go`
- Test: `tests/integration/v3/event_lifecycle_test.go`
- Modify: `internal/models/v3/dto.go` (event/session DTOs + template snapshot type)

**Interfaces:**
- Produces (on `EventUsecase{r *pgsql.PostgreRepositories}`):

```go
CreateDraft(ctx, title, createdBy string) (*v3.EventV3, *v3.Error)        // generates code (7-char, existing generator.GenerateHashCode) + slug
Patch(ctx, code string, patch map[string]any) (*v3.EventV3, *v3.Error)   // whitelisted fields incl. raw config jsons (loose parse check only)
Publish(ctx, code string) *v3.Error                                       // strict gate: ≥1 session, coherent times per session (register_end ≤ start_at when both set, start<end), ValidateStrict on registration config, form.ValidateSchema, geo config check values ∈ {off,warn,require}; error messages name the session/field
Cancel(ctx, code string) *v3.Error                                        // + enqueue event_cancelled for every confirmed registration
Archive/HardDelete(ctx, code string) *v3.Error                            // hard delete drafts only
CreateSessions(ctx, eventCode string, inputs []v3.SessionInput) ([]v3.SessionV3, *v3.Error) // bulk; generates session codes
PatchSession / DeleteSession / CreateCategories / PatchCategory / DeleteCategory (delete only when no confirmed registrations)
AddOrganizer(ctx, eventCode, communityID, role string) *v3.Error
IsOrganizer(ctx, eventCode, communityID string) (bool, *v3.Error)
SaveAsTemplate(ctx, eventCode, name, by string) (*v3.TemplateV3, *v3.Error)   // snapshot = full config set + session defaults
CreateFromTemplate(ctx, templateID int64, title string, firstSessionDate time.Time, by string) (*v3.EventV3, *v3.Error)
Duplicate(ctx, eventCode, by string) (*v3.EventV3, *v3.Error)
SeedSystemTemplates(ctx) *v3.Error   // idempotent: inserts the 4 spec templates when absent (Sunday Service / Conference / Weekly Class / Announcement)
```

- `RecurrenceUsecase.GenerateSessions(ctx) (created int, err *v3.Error)` — for each published event with recurrence: `recurrence.Expand` from now; insert sessions with `occurrence_key`; rely on the `(event_id, occurrence_key)` unique index — insert with `ON CONFLICT DO NOTHING` (GORM `clause.OnConflict{DoNothing: true}`) so re-runs are no-ops and edited sessions are never touched. Also applies publish-window flips: `ListDuePublishFlips` → set `is_visible` accordingly.

- [ ] **Step 1: Failing test** — draft with title only succeeds; publish without sessions fails with message containing "session"; publish with incoherent times fails naming the session code; valid publish flips status; `SeedSystemTemplates` twice yields exactly 4 rows; `CreateFromTemplate` produces a draft whose registration_config equals the snapshot's; recurrence generate on a weekly event creates N sessions, second run creates 0, editing one session then re-running leaves the edit intact.
- [ ] **Step 2: Run** → FAIL. **Step 3: Implement.** **Step 4: Run** → PASS.
- [ ] **Step 5: Commit** — `git commit -am "feat(v3): event lifecycle, templates, organizers and idempotent recurrence generation"`

---

### Task 14: QR usecase + v3 usecase aggregate

**Files:**
- Create: `internal/usecases/v3/qr_usecase.go`, `internal/usecases/v3/usecases.go`
- Test: `tests/integration/v3/qr_flow_test.go`

**Interfaces:**
- `QRUsecase` wires the registry: registers on `TypePersonal` action `event_checkin` (staff) and `event_checkout`; on `TypeAttendee` `event_checkin`/`event_checkout` (staff); on `TypeSession` `event_checkin` (self, caller = member) — handlers adapt `qr.ActionCtx` → `CheckinInput` and call `CheckinUsecase.Act`. Public methods: `Resolve(ctx, token string, caller eligibility.User, staffCheck func(sessionCode string) bool) (ResolveOutput, *v3.Error)` returning `{Type, Ref string; AllowedActions []string}`; `ActToken(ctx, token, action string, caller eligibility.User, staffCheck func(string) bool, params map[string]any) (any, *v3.Error)`.
- `usecases.go`: `type V3 struct{ Event *EventUsecase; Registration *RegistrationUsecase; Checkin *CheckinUsecase; QR *QRUsecase; Notification *NotificationUsecase; Recurrence *RecurrenceUsecase; Report *ReportUsecase }` + `func New(r *pgsql.PostgreRepositories, cfg *config.Configuration) *V3` (builds codec from `cfg.V3.QRSecret`/`AppDomain`, notifier map from `cfg.V3.Email`). Then modify `internal/usecases/main_usecase.go`: add `V3 *v3uc.V3` field, constructed in `New(d Dependencies)` as `v3uc.New(d.Repository, d.Config)`.

- [ ] **Step 1: Failing test** — end-to-end: seed event+session with `personal_qr` mode; encode personal token for community "100" (registered attendee); `Resolve` as staff ⇒ `AllowedActions` contains `event_checkin`; `ActToken` checks the attendee in; `Resolve` as non-staff ⇒ empty actions; session token self check-in works for the registered caller.
- [ ] **Step 2–4:** implement, run → PASS.
- [ ] **Step 5: Commit** — `git commit -am "feat(v3): qr resolve/act usecase wired to check-in via action registry"`

---

### Task 15: Reports — CSV & XLSX

**Files:**
- Create: `internal/usecases/v3/report_usecase.go`
- Test: `tests/integration/v3/report_test.go`

**Interfaces:**
- `func (u *ReportUsecase) SessionReport(ctx, sessionCode, format string) (data []byte, filename, mime string, err *v3.Error)` and `EventReport(ctx, eventCode, format string)`.
- CSV columns: `attendee_name, community_id, category, status, registered_at, attended_at, checked_out_at, registered_by, <one column per form field key in form order>`. XLSX (excelize): sheet `Summary` (per session: registered, attended, no_show, walk-ins [source=walk_in], headcount [from `CountHeadcount`], attendance %) + one sheet per session with the CSV columns. Checkout columns only when `enable_checkout`. Streamed as bytes; caller sets Content-Disposition.

- [ ] **Step 1: Failing test** — seed session + 2 registrations with a custom `diet` answer; CSV contains header with `diet` column and 3 data rows (party of 2 + 1); XLSX opens via excelize, `Summary` sheet cell B2 == registered count.
- [ ] **Step 2–4:** implement with `encoding/csv` + `excelize`, run → PASS.
- [ ] **Step 5: Commit** — `git commit -am "feat(v3): csv and xlsx reports with flattened form answers"`

---

### Task 16: HTTP delivery — handlers, computed state, mounting

**Files:**
- Create: `internal/deliveries/http/v3/v3_handler.go`, `event_handler.go`, `registration_handler.go`, `checkin_handler.go`, `qr_handler.go`, `staff_handler.go`, `internal_handler.go`, `respond.go`
- Create: `internal/models/v3/state.go` (computed state)
- Modify: `internal/deliveries/http/main_handler.go` (add `v3.NewV3Handler(api, u, c)`)
- Test: `internal/models/v3/state_test.go` + `tests/integration/v3/http_smoke_test.go`

**Interfaces:**
- `respond.go`: `func OK(ctx echo.Context, status int, data any) error` (`{"code":"SUCCESS","data":...}`) and `func Err(ctx echo.Context, e *v3.Error) error` (`{"code": e.Code, "message": e.Message}` with `e.Status`).
- Computed state (pure, unit-tested): `func BuildSessionState(ev *EventV3, s *SessionV3, cats []TicketCategory, rc RegistrationConfig, now time.Time) SessionState` returning the spec §7.3 shape: `Availability string; Reasons []string; Seats SeatInfo; Categories []CategoryState; ActivePhase string; ActiveModesNow []string; CheckinWindow WindowState`. Reason constants: `QUOTA_FULL, REGISTRATION_NOT_STARTED, REGISTRATION_ENDED, MARKED_SOLD_OUT, OUTSIDE_PUBLISH_WINDOW, INFO_ONLY, WALK_IN_ONLY`.
- Route table exactly as spec §9. Auth wiring: public group plain; member group `middleware.UserMiddleware(c, u, nil)`; staff endpoints use member middleware + in-handler check `u.V3.Event.IsOrganizer(...) || contains(roles,"event-admin") || superadmin` (helper `requireStaff(ctx, eventCode) *v3.Error` in `v3_handler.go`); admin endpoints `middleware.UserMiddleware(c, u, []string{"event-admin"})`; internal endpoints validate `X-API-Key == cfg.Auth.APIKey`.
- Handler shape identical everywhere: bind → call usecase → `respond.Err(ctx, e)` / `respond.OK`. Caller identity built from the JWT context set by UserMiddleware (`ctx.Get("id")`, `ctx.Get("roles")`, `ctx.Get("userTypes")`) into `eligibility.User` via helper `callerFrom(ctx echo.Context, u *usecases.Usecases) eligibility.User` (loads campus via existing `u.User.GetByCommunityId` once).

- [ ] **Step 1: Unit test computed state** — full session ⇒ `full` + `QUOTA_FULL`; before window ⇒ `opens_soon` + `REGISTRATION_NOT_STARTED`; marked sold out ⇒ reasons contains `MARKED_SOLD_OUT`; mode none ⇒ `info_only`; categories present ⇒ per-category availability array.
- [ ] **Step 2:** implement `state.go`, run → PASS, commit `feat(v3): computed session state with machine-readable reasons`.
- [ ] **Step 3: HTTP smoke integration test** — boot an `echo.New()` with v3 handler against testDB (JWT middleware replaced by a test middleware that injects identity — export `NewV3HandlerWithAuth(g, u, c, authMW echo.MiddlewareFunc)` so tests can inject): `GET /v3/events` lists seeded published event; `POST /v3/sessions/{code}/registrations` returns 201 with attendees; `GET /v3/sessions/{code}/attendees` as organizer returns 1 row; `GET .../report?format=csv` returns `text/csv`.
- [ ] **Step 4:** implement all handlers + mounting in `main_handler.go`; run smoke test + `go build ./...` → PASS.
- [ ] **Step 5: Commit** — `git commit -am "feat(v3): http delivery for events, registration, checkin, qr, staff and internal endpoints"`

---

### Task 17: Config, seeds, docs

**Files:**
- Modify: `config/config.local.template.yaml` (add `v3:` block), `internal/config/config.go` (done in Task 12 — verify), `AGENTS.md` (v3 section pointer), `wiki/raw/` capture note optional
- Create: startup hook — in `internal/contract/contract.go` after `usecases.New`, call `usecase.V3.Event.SeedSystemTemplates(ctx)` (log warn on error, don't crash)

- [ ] **Step 1:** add to `config/config.local.template.yaml`:

```yaml
v3:
  app_domain: "localhost:8080"
  qr_secret: "change-me-32-bytes-min"
  email:
    host: "smtp.example.org"
    port: 587
    username: ""
    password: ""
    from: "GROW Community <no-reply@example.org>"
```

- [ ] **Step 2:** wire seed call in `contract.go`; run `make run_api` locally → server boots, log shows templates seeded.
- [ ] **Step 3:** add a short "Event Management v3" paragraph to `AGENTS.md` architecture section pointing to the spec + this plan.
- [ ] **Step 4:** full check: `go build ./... && go test ./internal/... && TEST_DATABASE_URL=... go test ./tests/integration/v3/...` → all green.
- [ ] **Step 5: Commit** — `git commit -am "feat(v3): config template, system template seeding on boot, docs"`

---

## Deferred to Cloud Scheduler setup (ops, not code)

Two GCP Cloud Scheduler jobs (free tier), both POST with header `X-API-Key: <api_key>`:
- every 5 min → `/api/v3/internal/notifications/dispatch`
- every 30 min → `/api/v3/internal/sessions/generate`

## Self-Review Notes

- **Spec coverage:** eligibility (T4), geo both registration+checkin (T5,T10,T11), form incl. show_if/answered_by/edit_policy (T6,T10), recurrence (T7,T13), QR universal domain (T8,T14), 4 modes+checkout+headcount+walk-in+idempotency (T11), outbox/email/PDF (T12), templates/drafts/publish gate/organizers (T13), phases/access codes/categories/sold-out/reasons (T10,T16), reports (T15), publish window flips (T13), contacts/content sections stored (T1/T9, rendered by frontend; contacts also in PDF T12), dashboard endpoint is `staff_handler.go` in T16 route table.
- **Type consistency:** `eligibility.User` is the single caller-identity type across qr/registration/checkin/handlers; seat booking flows all call `AddBooked` after a `FOR UPDATE` read inside `Atomic`.
- **Known simplification:** event listing uses limit/offset (1k users) rather than v2's cursor pagination — acceptable, noted here deliberately.
