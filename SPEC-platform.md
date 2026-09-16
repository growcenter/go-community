# Spec: platform

Module id `platform` in `CAPABILITY-MAP.md`. Intent: `docs/intent/v3-platform.md`. No dependencies. Every other v3 module depends on this one.

## Objective

Give every v3 module the same error contract, response envelope, runtime tunables, cleanup and outbound messaging, so no module builds its own.

What this module delivers:

1. One herr error contract with a global, module-prefixed code catalog in English and Indonesian.
2. One success envelope `{ "data", "meta" }` and one error envelope from herr plus a top-level `message`.
3. HTTP 410 with code `COMMON_DEPRECATED_ENDPOINT` from every v1 and v2 route, through one middleware.
4. A configdb read helper with typed getters and boot-time defaults.
5. Retention cleanup of short-lived rows, run from an endpoint that Cloud Scheduler calls.
6. A notification outbox with a `channel` column, email sender now, SMS and WhatsApp reserved.

Users of this module are the v3 module authors and the single web app. Success is that a new v3 handler needs zero error, envelope, configuration or outbox code of its own.

## Tech Stack

- Go 1.26 (bump from 1.23.0, herr requires it per the events plan)
- Echo v4 (existing)
- GORM over PostgreSQL (existing), Supabase free tier
- `github.com/jeremygprawira/herr` and `herr/localizer/mapl` (new)
- Cloud Scheduler free tier for the two internal endpoints

## Commands

```bash
go get github.com/jeremygprawira/herr@latest && go mod tidy       # add herr, then set `go 1.26` in go.mod
go build ./...                                                    # compile
go test ./internal/pkg/... ./internal/platform/...                # unit tests, no DB
TEST_DATABASE_URL="postgres://postgres:$DB_PASSWORD@localhost:5888/community_db?sslmode=disable" \
  go test ./tests/integration/platform/ -v                        # DB tests, skipped when the URL is unset
golangci-lint run ./internal/pkg/... ./internal/platform/...      # lint on defaults
make generate-docs                                                # swagger, after any handler change
```

## Project Structure

```
internal/pkg/herrx/          → catalog registry: def(), AllClasses, completeness helpers
internal/pkg/herrx/common.go → COMMON_* classes shared by all modules
internal/pkg/locale/         → locale.Register(lang, msgs) and locale.Install()
internal/pkg/response/v3.go  → OK(), Created(), Fail() envelope writers (new file, old file untouched)
internal/pkg/configdb/       → Get, GetInt, GetBool, GetDuration, EnsureDefaults
internal/platform/deprecate/ → Echo middleware returning 410
internal/platform/retention/ → Purge(ctx) and the internal handler
internal/platform/outbox/    → outbox repository, Notifier interface, email sender, dispatch handler
internal/models/<module>/errors.go and locale.go → each module's own catalog (pattern only, set here)
tests/integration/platform/  → outbox claim and purge DB tests
tests/integration/db/migrations/0000NN_notification_outbox.up.sql
```

Each module owns its codes in its own `errors.go`. This module owns the registry, the `COMMON_` codes and the rules.

## Code Style

### Catalog entry

Every code is `<MODULE>_<NAME>`. Module prefixes: `COMMON_`, `USER_`, `PERM_`, `QR_`, `EVENT_`, `COOL_`, `CAL_`. A module calls `herrx.Def` for each class, then registers its Indonesian map with `locale.Register`.

```go
// internal/pkg/herrx/common.go
package herrx

import "github.com/jeremygprawira/herr"

var (
	AllClasses      []*herr.Class
	EnglishMessages = map[string]string{}
)

func Def(code string, kind herr.Kind, en string) *herr.Class {
	if _, dup := EnglishMessages[code]; dup {
		panic("duplicate error code " + code)
	}
	c := herr.Define(herr.Class{Code: code, Kind: kind, Public: herr.Public{Message: en}})
	EnglishMessages[code] = en
	AllClasses = append(AllClasses, c)
	return c
}

var (
	ErrDeprecatedEndpoint = Def("COMMON_DEPRECATED_ENDPOINT", herr.KindGone, "This endpoint has been retired. Please update the app.")
	ErrInvalidInput       = Def("COMMON_INVALID_INPUT", herr.KindInvalid, "Something in your submission does not look right. Please review and try again.")
	ErrNotFound           = Def("COMMON_NOT_FOUND", herr.KindNotFound, "We could not find what you are looking for.")
	ErrForbidden          = Def("COMMON_FORBIDDEN", herr.KindForbidden, "You do not have access to do this.")
	ErrUnauthorized       = Def("COMMON_UNAUTHORIZED", herr.KindUnauthorized, "Please sign in to continue.")
	ErrInternal           = Def("COMMON_INTERNAL", herr.KindInternal, "Something went wrong on our side. Please try again.")
)
```

`herr.KindGone` (HTTP 410) does not exist in herr today. Decision 2026-09-17: add `KindGone` to herr before this module starts. The middleware then relies on the class for the status, like every other code.

### Envelope writer

```go
// internal/pkg/response/v3.go
package response

type Meta struct {
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	Total      *int64 `json:"total,omitempty"`
}

type envelope struct {
	Data any   `json:"data"`
	Meta *Meta `json:"meta,omitempty"`
}

func OK(c echo.Context, data any, meta *Meta) error {
	return c.JSON(http.StatusOK, envelope{Data: data, Meta: meta})
}

func Fail(c echo.Context, err error) error {
	he := herr.From(err) // non-herr errors become COMMON_INTERNAL, detail goes to logs
	lang := locale.Pick(c.Request().Header.Get("Accept-Language")) // "id" unless "en" is asked
	status := herr.HTTPStatus(he)
	body := he.Body(lang)
	body = withTopLevelMessage(body, lang) // adds "message" when only "errors" is present
	slog.LogAttrs(c.Request().Context(), slog.LevelError, "request failed", herr.Attrs(he)...)
	return c.JSONBlob(status, body)
}
```

Rules:

- Handlers call `response.OK` or `response.Fail`. Never `c.JSON` directly in v3.
- Usecases return `error`. Match with `Class.Is(err)`, never `==`.
- Internal detail goes through `.Internal()`, `.WithInternal()` or `.Wrap()`. It never reaches the body.
- A success body never has an `errors` key. An error body never has a `data` key.
- `Accept-Language` picks `en` or `id`. Anything else, or nothing, means `id`.
- Validation failures use `FieldError` per field under one class such as `COMMON_INVALID_INPUT`. The body then has `code`, `message` and `errors[]`.
- Never add to `models.ErrorMapping` or the old `response.Error`. v1 and v2 keep them until deleted.

### Error body shape

```json
{ "code": "EVENT_QUOTA_FULL", "message": "Semua kursi sudah terisi." }

{ "code": "COMMON_INVALID_INPUT",
  "message": "Ada isian yang belum benar.",
  "errors": [ { "field": "title", "code": "REQUIRED", "message": "Wajib diisi." } ] }
```

### Deprecation

`internal/platform/deprecate.Middleware()` is applied once in `main_handler.go` to the groups that `v1.NewV1Handler` and `v2.NewV2Handler` receive. It returns 410 and the `COMMON_DEPRECATED_ENDPOINT` body before any handler runs. No handler file under `v1/` or `v2/` is edited. The health group stays live.

### Configdb

Reuse the existing `config` table (`Identifier`, `Key`, `Value`) and `ConfigRepository.GetByKey`. Add typed getters that parse `Value` and fall back to a default. Keys are `<module>.<name>`, for example `platform.retention_days.otp`. `configdb.EnsureDefaults(ctx, map[string]string)` upserts missing keys at boot from `contract.go`. Modules never read admin-tunable values from YAML.

Keys owned by this module, with defaults:

| Key | Default | Meaning |
|---|---|---|
| `platform.retention_days.auth_challenges` | `7` | delete OTP and PIN challenges older than this |
| `platform.retention_days.outbox` | `30` | delete `sent` and `failed` outbox rows older than this |
| `platform.retention_days.scan_logs` | `90` | delete QR scan log rows older than this |
| `platform.outbox.max_attempts` | `5` | after this many failures a row becomes `failed` |
| `platform.outbox.batch_size` | `50` | rows claimed per dispatch call |

### Retention

`retention.Purge(ctx)` runs one `DELETE ... WHERE created_at < now() - interval` per table, each in its own short statement, and returns counts. `POST /api/v3/internal/retention/purge` calls it. The route is protected by the existing `X-API-Key` check and is called by Cloud Scheduler once a day. Attendance, registration, audit and user rows are never in the purge list.

### Outbox

Table `notification_outbox` follows the events design section 8. Columns: `id`, `channel`, `recipient`, `template`, `payload` jsonb, `status`, `attempts`, `scheduled_at`, `sent_at`, `last_error`, `created_at`. Status is one of `pending`, `sent`, `failed`. `channel` is `email` today. `sms` and `whatsapp` are reserved values with no sender yet. Producers insert a row inside their own `Atomic` transaction. `POST /api/v3/internal/notifications/dispatch` claims up to `batch_size` rows with `FOR UPDATE SKIP LOCKED`, sends each through the `Notifier` for its channel, and marks `sent` or bumps `attempts`. A row with an unknown channel is marked `failed` with `last_error = "no sender"`. Cloud Scheduler calls dispatch every 5 minutes. A producer can also call the dispatcher in-process right after commit as a fast path.

```go
type Notifier interface {
	Channel() string
	Send(ctx context.Context, recipient, template string, payload map[string]any) error
}
```

One email implementation, provider and credentials from YAML configuration.

## Testing Strategy

Unit tests, no DB, under `internal/pkg/...` and `internal/platform/...`:

- Catalog completeness: every class in `AllClasses` has an `id` message that is non-empty and differs from English. A missing translation fails `go test`, so it fails CI.
- Duplicate code panics at init.
- Envelope: `OK` never emits `errors`. `Fail` always emits `code` and `message`. `Fail` with field errors emits `errors[]`. Internal detail never appears in the body. `Accept-Language` defaults to `id`.
- Deprecation middleware returns 410 with the right body for a v1 route and does not touch the health route.
- Configdb getters parse ints, bools and durations and fall back on a missing or bad value.

Integration tests, real Postgres via `TEST_DATABASE_URL`, under `tests/integration/platform/`:

- Two concurrent dispatch calls never send the same row twice.
- A row past `max_attempts` becomes `failed`.
- Purge deletes only rows older than the configured days and returns correct counts.

No mocks for the database. Coverage target: every rule in Code Style has one test.

## Boundaries

Always:

- Run `go build ./... && go test ./...` before a commit.
- Prefix every new error code with its module and register its `id` message in the same commit.
- Write outbox rows inside the producer's transaction.
- Use `Atomic`, never `Transaction`, for any new write.

Ask first:

- Adding a Kind or changing a status mapping.
- Adding a second notification channel implementation.
- Changing a retention default or adding a table to the purge list.
- Any change to `go.mod` beyond herr and the Go bump.

Never:

- Add to `models.ErrorMapping` or `response.Error`.
- Put a message string in a handler. Messages live in the catalog only.
- Purge attendance, registration, audit or user rows.
- Log OTP codes, PINs, tokens or full request bodies.
- Run a resident goroutine loop for dispatch or purge.

## Success Criteria

1. `GET /api/v1/campus` and every other v1 and v2 route return 410 with `{ "code": "COMMON_DEPRECATED_ENDPOINT", "message": ... }`, and `/health` still returns 200.
2. A v3 handler that returns `herrx.ErrNotFound.New()` produces 404 and an Indonesian message with no `Accept-Language` header, and English with `Accept-Language: en`.
3. A v3 handler that returns three field errors produces 400 with `code`, `message` and three `errors[]` entries.
4. Removing one `id` translation makes `go test ./internal/pkg/herrx/` fail.
5. Defining two classes with the same code makes the process panic at start.
6. Every success body from a v3 handler has `data` and no `errors`.
7. Two dispatch calls started at the same time against 100 pending rows send exactly 100 emails.
8. After purge, no `auth_challenges` row older than `platform.retention_days.auth_challenges` remains, and no attendance row is gone.
9. `go.mod` says `go 1.26` and `go build ./...` passes.

## Open Questions

1. Resolved: add `KindGone` to herr. Release it before the deprecation middleware is built.
2. The exact names `herr.From` and `herr.HTTPStatus` are taken from the events plan and are not in the README. After `go get`, make sure that the names exist with `go doc github.com/jeremygprawira/herr`. If they differ, adjust the two call sites.
3. Email provider for the free tier is not chosen. The `Notifier` interface does not depend on it, but the YAML configuration keys do.
4. The internal endpoints reuse the static `X-API-Key`. Decide whether Cloud Scheduler gets a separate key.
