# Spec: qr

Module id `qr` in `CAPABILITY-MAP.md`. Intent: `docs/intent/v3-platform.md`. Depends on `platform` and `identity`. Replaces section 6 of `docs/plans/2026-07-07-event-management-v3-design.md` and Task 8 of its implementation plan as the source of truth for QR. Consumers are `events` and `cool`.

## Objective

Give the whole system one QR format, one scanner endpoint and one signing key set. A QR never carries logic. It only says what it points at. The module that owns the target decides what a scan means.

What this module delivers:

1. One signed token format with four kinds: `personal`, `session`, `meeting`, `attendee`.
2. A static personal QR for every member, derived from `community_id`, valid for life.
3. A resolve endpoint that tells a client what it scanned and which actions are open to it.
4. An act endpoint that runs one action through a handler that the owning module registered.
5. PNG rendering for posters, phone screens and PDF tickets.
6. A scan log, purged by `platform.retention_days.scan_logs`.

Users are members, desk staff, COOL leaders, and the `events` and `cool` modules. Members show or scan a QR. Desk staff scan members. Leaders show a meeting QR. The two modules register handlers. Success: a member's printed QR from 2026 still works in 2036, and adding a new QR use never changes the format, the scanner or the endpoints.

## Tech Stack

- Go 1.26, Echo v4, GORM (existing)
- `crypto/hmac`, `crypto/sha256`, `encoding/base64`, `encoding/json` from the standard library for the token
- `github.com/skip2/go-qrcode` for PNG. It is not in `go.mod` today. See Boundaries, Ask first.
- `internal/pkg/herrx` and `internal/pkg/response` from `platform`

## Commands

```bash
go get github.com/skip2/go-qrcode@latest && go mod tidy       # once, after approval
go build ./...
go test ./internal/pkg/qr/                                    # unit tests, no DB
TEST_DATABASE_URL="postgres://postgres:$DB_PASSWORD@localhost:5888/community_db?sslmode=disable" \
  go test ./tests/integration/qr/ -v                          # scan log tests, skipped when unset
make migration name=qr_scan_logs                              # numbered after 000023 (identity)
make generate-docs
```

## Project Structure

```
internal/pkg/qr/codec.go            → Payload, Codec (Encode, Decode, URL), key map with kid
internal/pkg/qr/registry.go         → Action, ActionCtx, Registry (Register, AllowedActions, Dispatch)
internal/pkg/qr/png.go              → RenderPNG(url, size)
internal/models/qr/                 → DTOs, errors.go (QR_ codes), locale.go
internal/repositories/pgsql/qr/     → scan log insert
internal/usecases/qr/               → Resolve, Act (verify, log, dispatch)
internal/deliveries/http/v3/qr/     → resolve, act, me/qr.png handlers
tests/integration/qr/               → scan log rows and purge
tests/integration/db/migrations/000024_qr_scan_logs.up.sql
```

`events` and `cool` call `qr.Registry.Register` from their own packages at boot in `contract.go`. The `qr` package imports neither of them.

## Code Style

### Token wire format

Every QR encodes one URL:

```
https://<public_url>/q/<kid>.<payload>.<sig>
```

- `kid` is the key id in `config.QR.Secret` (`map[string]string`, same pattern as `Auth.BearerSecret`). New tokens use `config.QR.ActiveKid`. Old kids stay in the map so old posters still verify. The personal QR key is never removed.
- `payload` is base64url, no padding, of compact JSON: `{"t":"personal","r":"1234567890123456","iat":1751850000}`. If the kind expires, the JSON also has `"exp":1751900000`.
- `sig` is base64url, no padding, of HMAC-SHA256 over `kid + "." + payload`.

| Kind | `r` | `exp` | Who mints it | Where it shows |
|---|---|---|---|---|
| `personal` | `community_id` | never | `qr`, on demand from `community_id` | member's phone or printed card |
| `session` | session code | set by `events`. Check-in close time, or session end time if checkout is on (ADR 0009) | `events` | poster or staff phone or shared screen |
| `meeting` | meeting code | set by `cool` (meeting end plus a grace the `cool` spec defines) | `cool` | poster or leader phone or shared screen |
| `attendee` | attendee row id | session end | `events`, only for events with a registration window | attendee's ticket, printed or on phone, shown to desk staff |

Rules:

- A personal token has no `exp` and no random part. The same `community_id` always yields the same token, so nothing needs storing and a printed card never goes stale.
- A `session` or `meeting` token is one token per session or meeting. The printed poster and the on-screen QR are the same token.
- An `attendee` token is one per attendee row, so a family of four gets four codes, never one code for four people. It is shown by the attendee and scanned by desk staff. It is never a self-scan token. `events` mints it only for events that have a registration window.
- `qr` exposes only `kind`, `ref`, `issued_at` and `expires_at` to a consumer. Whether a scan from home counts is the consumer's rule (events: protection level per event, cool: the leader marks attendance by hand).
- Size budget: the URL must stay under 160 characters. That fits QR version 7 at error level M. A cheap phone camera then reads it from a printed A4 poster. With a 30 character public URL, a 16 digit ref and an `exp`, the token is about 115 characters. A test asserts the bound for every kind.
- The payload is signed, not encrypted. Anyone can read `t` and `r` from it without the key. This is on purpose: an offline desk reads `community_id` or the attendee id from a scanned QR and matches its downloaded roster. Only the server can tell a real token from a forged one, and it does so when the desk syncs (`SPEC-events.md` delta 15).
- Decode fails closed. Wrong part count, unknown `kid`, bad base64, bad signature or unknown kind all return `QR_INVALID_TOKEN`. Only a good signature with `exp` in the past returns `QR_EXPIRED_TOKEN`. Signature comparison uses `hmac.Equal`.

### Handler registry

```go
type ActionCtx struct {
	Ctx      context.Context
	Payload  Payload           // Kind, Ref, IssuedAt, ExpiresAt
	Caller   string            // community_id of the signed-in scanner
	Params   map[string]any    // free-form, defined by the action, for example session_code, lat, lng, ticket_category
}

type Action struct {
	Name    string                          // for example "event_checkin", "cool_mark_present"
	Allowed func(ActionCtx) bool            // consumer asks permissions here
	Handle  func(ActionCtx) (any, error)    // returns herr errors from the consumer's catalog
}

type Registry interface {
	Register(kind string, a Action)
	AllowedActions(kind string, c ActionCtx) []string
	Dispatch(kind, action string, c ActionCtx) (any, error)
}
```

A consumer registers at boot:

```go
// internal/usecases/events/qr_actions.go, called from contract.go
func RegisterQRActions(r qr.Registry, ev EventUsecase, perm permissions.Evaluator) {
	r.Register(qr.KindPersonal, qr.Action{
		Name: "event_checkin",
		Allowed: func(c qr.ActionCtx) bool {
			code, _ := c.Params["session_code"].(string)
			return perm.Can(c.Ctx, c.Caller, "event.checkin", "session", code) // the caller may run the desk for this session
		},
		Handle: func(c qr.ActionCtx) (any, error) {
			return ev.CheckInByCommunityID(c.Ctx, c.Payload.Ref, c.Params)
		},
	})
}
```

Rules:

- `Dispatch` runs `Allowed` first. If `Allowed` is false, it returns `QR_ACTION_NOT_ALLOWED` and never calls `Handle`.
- Registering the same `kind` and `Name` twice panics at boot.
- `qr` never inspects `Params`. It passes them through.
- Errors from `Handle` pass through `response.Fail` unchanged, so the client sees the consumer's `EVENT_` or `COOL_` code.

### Endpoints

All under `/api/v3`. Every endpoint needs a signed-in member. There is no anonymous scan.

```
POST /v3/qr/resolve   {token}                       → {kind, ref, issued_at, expires_at, actions: ["event_checkin", ...]}
POST /v3/qr/act       {token, action, params{}}     → whatever the handler returns, in data
GET  /v3/me/qr.png    ?size=256                     → own personal QR as image/png
```

- `resolve` verifies the token, writes one scan log row, and returns the actions that are open to the caller right now. A member who scans a friend's personal QR gets an empty `actions` list and no error.
- `act` verifies again, checks `Allowed`, dispatches, writes one scan log row with the result, and returns the handler output as `data`.
- `me/qr.png` renders the caller's own personal token. Size is clamped to 128 to 1024 pixels. The response sets `Cache-Control: private, max-age=86400`, because the image never changes.
- Poster images for a session or meeting are served by `events` and `cool` from their own endpoints. They call `qr.RenderPNG` with the token they minted. The `qr` module has no endpoint that renders someone else's token.

### Scan log

Table `qr_scan_logs`:

| Column | Type | Meaning |
|---|---|---|
| `id` | uuid | |
| `kind` | text | token kind. If decode failed, `invalid` |
| `ref` | text | payload ref. If decode failed, empty |
| `scanner` | text | caller `community_id` |
| `action` | text | `resolve`, or the action name |
| `result` | text | `ok` or the error code |
| `client_ip` | inet | from the request |
| `offline` | boolean | true when the scan was made offline and arrived through the events batch sync |
| `device_scanned_at` | timestamptz | nullable, the device's scan time for an offline scan |
| `created_at` | timestamptz | |

Index on `(ref, created_at)` and `(scanner, created_at)`. The events batch sync decodes each offline token through `qr.Codec` and writes one row per item with `offline = true`, so the trail is complete even for scans made with no signal. Written on every `resolve` and `act`, inside the same `Atomic` transaction as the handler when there is one. Purged by `platform.retention_days.scan_logs` (default 90). This is the trail an admin reads when someone asks "who scanned this poster from home".

### Errors

All through `herrx.Def`, `QR_` prefix, EN here, ID in `locale.go`.

| Code | Kind | EN message |
|---|---|---|
| `QR_INVALID_TOKEN` | unauthorized | This QR code is not valid. Please use the QR from your ticket or profile. |
| `QR_EXPIRED_TOKEN` | unauthorized | This QR code has expired. Please ask for a fresh one. |
| `QR_ACTION_NOT_FOUND` | not found | That action does not exist for this QR code. |
| `QR_ACTION_NOT_ALLOWED` | forbidden | You cannot do that with this QR code right now. |
| `QR_RENDER_FAILED` | internal | We could not draw the QR code. Please try again. |

### Configuration

Static YAML, not configdb, because a key change is a deploy:

```yaml
qr:
  active_kid: "k1"
  secret:
    k1: "<32 random bytes, base64>"
application:
  public_url: "https://app.example.org"
```

`config.QR.Secret` must have at least one key and `active_kid` must be in it, or the app refuses to start.

## Testing Strategy

Unit, no DB, `internal/pkg/qr/`:

- Round trip for every kind. Personal token for one `community_id` is byte-identical across calls.
- Tamper, wrong key, unknown kid, missing part, unknown kind: all `QR_INVALID_TOKEN`. Past `exp` with a good signature: `QR_EXPIRED_TOKEN`.
- Key rotation: a token signed with `k1` still verifies after `active_kid` moves to `k2`.
- URL length under 160 for every kind with realistic refs.
- Registry: `AllowedActions` filters by `Allowed`, `Dispatch` refuses a not-allowed action, unknown action is `QR_ACTION_NOT_FOUND`, duplicate registration panics.
- `RenderPNG` output decodes back to the same URL with a decoder in the test only.

Integration, real Postgres, `tests/integration/qr/`:

- `resolve` and `act` each insert exactly one scan log row, with `result` equal to the error code on failure.
- Purge removes rows older than the configured days and nothing newer.

No database mocks. Every rule in Code Style has one test.

## Boundaries

Always:

- Verify the signature before reading any payload field.
- Write a scan log row for every resolve and act, success or failure.
- Keep `personal` tokens deterministic and unexpiring.

Ask first:

- Adding `github.com/skip2/go-qrcode` to `go.mod`. It is the only new dependency in this module.
- Adding a fifth token kind.
- Removing a `kid` from `config.QR.Secret`.
- Any change to the URL shape or payload field names. Printed posters depend on them.

Never:

- Put a name, phone, email or any personal data in a payload. Only ids.
- Let `qr` decide whether a scan counts as attendance.
- Rotate or expire a `personal` token.
- Serve an unauthenticated resolve or act.
- Log the full token in application logs. Log `kind` and `ref` only.

## Success Criteria

1. `GET /v3/me/qr.png` for one member returns the same bytes on two calls a day apart, and the decoded URL verifies.
2. A desk staff member scans a personal QR during a session they may run and `resolve` lists `event_checkin`. A plain member scanning the same QR gets an empty list.
3. A session poster scanned after its `exp` returns 401 `QR_EXPIRED_TOKEN`, and a scan log row shows `result = QR_EXPIRED_TOKEN`.
4. Changing `active_kid` to a new key does not break any poster printed with the old key.
5. Registering a `meeting` handler in `cool` requires no change in `internal/pkg/qr` or its handlers.
6. Every personal, session, meeting and attendee URL is under 160 characters in the length test.
7. `act` with an action that `Allowed` rejects returns 403 `QR_ACTION_NOT_ALLOWED` and never runs `Handle`.
8. After purge, no `qr_scan_logs` row older than `platform.retention_days.scan_logs` remains.
9. Every `QR_` code has an ID message, or `go test` fails.

## Open Questions

1. Resolved 2026-09-17: `attendee` stays as the fourth kind, one per attendee row, desk-scanned only, minted only for events with a registration window.
2. SVG rendering is left out. The web app can draw the URL itself with a browser QR library, and `skip2/go-qrcode` has no SVG output. Add SVG only if a printed poster needs vector output.
3. `application.public_url` is a new YAML key. Confirm the domain the web app will use for `/q/<token>`, and that the web app owns the `/q/` route and calls `resolve` from there.
4. The `meeting` token grace period after meeting end is defined in `SPEC-cool.md`, not here. Make sure that spec sets it.
