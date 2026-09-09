# Spec: Event Management v3

**Status:** Authoritative. Supersedes the statements listed in §9.
**Scope:** `/api/v3` only. v2 is untouched and stays read-only for its existing events.
**Refined from:** `docs/plans/2026-07-07-event-management-v3-design.md`, its implementation plan, `docs/plans/2026-09-09-checkin-and-registration-paths-spec.md`, `CONTEXT.md`, and ADRs 0001–0009.

This document and the `SPEC-<module>.md` files beside it are what you implement against. The plan documents in `docs/plans/` remain in the repository as the historical record of how these decisions were reached; where they disagree with this spec, **this spec wins**, and §9 lists every such disagreement explicitly so nobody has to diff them by hand.

---

## 1. Objective

Replace v2 event management with one model flexible enough to express everything this church actually runs — a single gathering, a conference, a Christmas event with three sessions, a weekly recurring service, a monthly volunteer meeting, or a pure information page — with per-event control over who may attend, what is asked of them, how they are checked in, and where they must physically be.

**Who uses it**

| User | What they need |
|---|---|
| **Attendee / member** | Find an event, register themselves and a party, receive a ticket, check themselves in at the venue, check out |
| **Welcome-desk volunteer** | Sign someone up at the lobby table weeks in advance — including a first-time visitor with no account, no email and no smartphone |
| **Usher at the door** | Scan people in, find a registered person by name when their ticket never arrived, add a genuine walk-up, tap an anonymous headcount and undo a mis-tap |
| **Event organizer** | Configure who may register and through which doors, see who came, who was late, and where the sign-ups came from |
| **Volunteer coordinator** | Run a small recurring meeting where the question is attendance and punctuality, not gatekeeping |

**What success looks like**

Every everyday situation that has no correct path today has exactly one. A person can be registered through three doors — online self-service, staff at a desk in advance, and QR-triggered creation at the door — and **all three enforce the same rules**. Nobody who registered legitimately can be turned away for want of a working QR. A restriction an organizer configures is actually enforced everywhere, not just on the path they happened to look at. Attendance is optional, and an event that never asked for it never reports "0 of 40 attended."

**Scale:** ~1,000 users total. Correctness and clarity beat premature optimization everywhere.

**Non-goals** — no payments ever; no migration of v2 data; email is the only notification channel at launch; API only, frontend is a separate repo. Plus the deferrals in §8.

---

## 2. Capability Map

| Module id | Responsibility | Depends on | Spec |
|---|---|---|---|
| `foundation` | Migration, Go 1.26 bump, herr bilingual error catalog, `models/v3` entities + JSONB config structs, repository aggregate, config block | — | [SPEC-foundation.md](SPEC-foundation.md) |
| `policy-kernel` | Pure validators: eligibility, geo, form (`show_if`), recurrence expansion | `foundation` | [SPEC-policy-kernel.md](SPEC-policy-kernel.md) |
| `qr` | Universal token codec + HMAC, PNG render, action registry, per-caller `allowed_actions` | `foundation` | [SPEC-qr.md](SPEC-qr.md) |
| `event-core` | Event/session/category CRUD, publish gate, templates, recurrence generation, computed state | `foundation`, `policy-kernel` | [SPEC-event-core.md](SPEC-event-core.md) |
| `registration` | Three doors, `channels`, eligibility on every creation path, anonymous rules, `max_sessions_per_person`, cancel, answer editing | `event-core`, `policy-kernel` | [SPEC-registration.md](SPEC-registration.md) |
| `attendance` | Check-in/checkout, 4 modes, walk-in creation, headcount + undo, lateness, eligibility re-check, append-only logs | `registration`, `qr`, `policy-kernel` | [SPEC-attendance.md](SPEC-attendance.md) |
| `notifications` | Outbox + dispatcher, SMTP notifier, PDF ticket | `registration`, `qr` | [SPEC-notifications.md](SPEC-notifications.md) |
| `reporting` | CSV/XLSX, source split, conditional columns | `registration`, `attendance` | [SPEC-reporting.md](SPEC-reporting.md) |
| `http-api` | Handlers, route table, auth wiring, internal endpoints | all | [SPEC-http-api.md](SPEC-http-api.md) |

**Build order:** `foundation` → (`policy-kernel`, `qr`) → `event-core` → `registration` → (`attendance`, `notifications`) → `reporting` → `http-api`

Module ids are stable and never renamed. Parenthesised pairs may be built in parallel.

### External precondition — `permissions`

Per ADR 0001, v3 authorizes staff through `permission_grants` and `Can(actor, action, resource)`, **not** a standalone `event_organizers` table. That engine is owned by `docs/plans/2026-07-07-cool-and-unified-permissions-design.md` (§2.1) and its implementation plan (Tasks 2–3, migrations `000023`/`000024`). It is **not** a module of this spec and is not re-specified here.

v3 consumes exactly this contract and nothing else:

```go
// Provided by the permissions engine. v3 depends on this signature only.
Can(ctx context.Context, actor Actor, action string, resource string) (bool, error)
```

| Action | Resource | Used by |
|---|---|---|
| `event.checkin` | `event:<code>` | Desk registration while registration is open; all staff-actuated check-in modes; headcount and headcount undo |
| `event.manage` | `event:<code>` | Desk registration after `register_end_at`; event/session CRUD; reports; grants administration |

Superadmin bypass applies as it does everywhere else in the system.

**Build implication:** `registration` and `attendance` cannot pass their permission acceptance criteria until the permissions engine exists. Everything up to and including `event-core` can be built and verified without it.

---

## 3. Tech Stack

| Concern | Choice |
|---|---|
| Language | **Go 1.26.1** — `go.mod` and the Dockerfile build stage were bumped from `1.23.0` on 2026-09-10; herr declares `go 1.26.1`. |
| HTTP | Echo v4 (existing) |
| Persistence | GORM over PostgreSQL (Supabase free tier; pooled connection string, small pool) |
| Errors | `github.com/jeremygprawira/herr` — v3 only, bilingual EN/ID |
| Config | Viper, `config/config.<env>.yaml` |
| XLSX | `github.com/xuri/excelize` (already in `go.mod`) |
| PDF | `github.com/jung-kurt/gofpdf` (new) |
| QR PNG | `github.com/skip2/go-qrcode` (new) |
| Email | `net/smtp` + MIME multipart — no new dependency |
| Async | No resident worker. DB-backed outbox, triggered opportunistically after requests and by GCP Cloud Scheduler (free tier, 3 jobs). |

All new dependencies are pure Go, no CGO, free.

---

## 4. Commands

```bash
# First-time setup
cp config/config.local.template.yaml config/config.local.yaml
make docker-start                 # Postgres + app via docker-compose
make database_up                  # create community_db

# Build & run
go build ./...
ENV=DEV make run_api              # tidy modules, run ./cmd/api/main.go
make run                          # generate swagger docs, tidy, run

# Test
go test ./internal/... ./tests/...                          # unit — no DB, always runnable
TEST_DATABASE_URL="postgres://postgres:<pw>@localhost:5888/community_db?sslmode=disable" \
  go test ./tests/integration/v3/...                        # integration — skips when unset
go test ./internal/... ./tests/... -race                    # race detector; required for seat-contention tests
go test ./internal/pkg/... -cover                           # pure-validator coverage

# Lint & docs
golangci-lint run                 # no .golangci.yml checked in — runs on defaults
make generate-docs                # swag init -g cmd/api/main.go

# Migrations
make migration name=v3_event_management     # creates the next up/down pair
make migration_up                            # edit the DB connection args first
make migration_down
```

There is no `make test` or `make lint` target despite the README mentioning them. Use the commands above directly; adding those targets is in scope for `foundation`.

---

## 5. Project Structure

```
internal/
  models/v3/
    errors.go          herr.Define classes + bilingual catalog
    config.go          Eligibility / RegistrationConfig / GeoConfig / Recurrence / Contacts + Parse + ValidateStrict
    entities.go        GORM entities (EventV3, SessionV3, TicketCategory, RegistrationV3,
                       AttendeeV3, AttendanceLog, OutboxMessage, TemplateV3)
    dto.go             request/response DTOs
    state.go           computed session state + reasons[]
  pkg/
    eligibility/       pure Check()
    geo/               Haversine + Validate()
    form/              Visible() / Validate() / ValidateSchema()
    recurrence/        Expand()
    qr/                token codec (HMAC), PNG render, action registry
    notify/            Notifier interface + SMTP implementation
    pdfticket/         PDF generation
  repositories/pgsql/
    v3_repositories.go all v3 repos + wiring into PostgreRepositories
  usecases/v3/
    event_usecase.go registration_usecase.go checkin_usecase.go qr_usecase.go
    notification_usecase.go recurrence_usecase.go report_usecase.go usecases.go
  deliveries/http/v3/  handlers + route mounting

tests/
  integration/db/migrations/000025_v3_event_management.{up,down}.sql
  integration/v3/
    harness_test.go    migrated DB + truncate between tests
    *_test.go          one file per flow

docs/plans/            historical record — superseded by this spec
docs/adr/              accepted decisions — binding
SPEC.md                this file
SPEC-<module>.md       per-module specs
```

**Placement rules.** Anything decidable without a database or an HTTP request goes in `internal/pkg/` as a pure function. Anything that touches rows goes in a usecase. Handlers bind, call one usecase method, and respond — no logic. No v3 code lives outside `internal/*/v3/`, `internal/pkg/`, and the single migration.

---

## 6. Code Style

Match the surrounding repository, with the four v3-specific conventions below. One real example carries them all:

```go
// internal/usecases/v3/registration_usecase.go

// RegistrationUsecase takes the repository aggregate, not seven interfaces — v3 has
// exactly one dependency style (convention 3).
type RegistrationUsecase struct {
	r   *pgsql.PostgreRepositories
	cfg *config.Configuration
}

func NewRegistrationUsecase(r *pgsql.PostgreRepositories, cfg *config.Configuration) *RegistrationUsecase {
	return &RegistrationUsecase{r: r, cfg: cfg}
}

func (u *RegistrationUsecase) Register(ctx context.Context, in RegisterInput) (*RegisterOutput, error) {
	session, event, err := u.load(ctx, in.SessionCode)
	if err != nil {
		return nil, err
	}

	// Pure validators run first, outside the transaction (convention 4).
	if !eligibility.Check(event.Eligibility, in.User) {
		// herr classes carry a humanized public message; Internal detail never reaches
		// the client (convention 1). Match with .Is(), never ==.
		return nil, v3.ErrNotEligible.
			Internal("audience=%s community_id=%s", event.Eligibility.Audience, in.User.CommunityID)
	}
	if err := form.Validate(event.RegistrationConfig.Form, "primary", in.Primary); err != nil {
		return nil, err
	}

	// Atomic gives the closure a tx-scoped repository set, so the seat lock, the inserts
	// and the outbox row all share one transaction. Transaction is banned in v3 (convention 2).
	var out RegisterOutput
	err = u.r.Transaction.Atomic(ctx, func(ctx context.Context, tx *pgsql.PostgreRepositories) error {
		locked, err := tx.V3.Session.GetByCodeForUpdate(ctx, session.Code)
		if err != nil {
			return err
		}
		if locked.TotalSeats > 0 && locked.BookedSeats+in.PartySize > locked.TotalSeats {
			return v3.ErrQuotaFull
		}
		...
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}
```

**The four conventions, stated plainly:**

1. **herr end to end.** Every domain error is an immutable `herr.Define` class in package `v3`; Kind drives the HTTP status. Match with `v3.ErrX.Is(err)`, **never** `==`. Every class has a humanized English `Public.Message` **and** an Indonesian translation under `errors.<lowercase_code>.message`; default locale `id`, `Accept-Language: en` switches. **An error class without its `id` translation is a review-rejectable defect.** Never add to v2's `models.ErrorMapping`.
2. **`Atomic` only.** `TransactionRepository.Transaction` has a known scoping bug (`wiki/entities/competing-transaction-abstractions.md`) and is banned in v3.
3. **One dependency style.** Every v3 usecase takes `*pgsql.PostgreRepositories` plus config. No multi-interface constructors.
4. **Validators are pure functions.** `eligibility.Check`, `form.Validate`, `geo.Validate`, `recurrence.Expand` take values and return values — no DB, no HTTP, no clock reads except an injected `now`.

**Also:** module path is `go-community`; store `timestamptz`, business "today" uses `Asia/Jakarta` via `common.GetLocation()`; all API timestamps are RFC3339; v3 does **not** reuse v2's `defer usecases.LogService(ctx, err)` double-logging pattern — v3 logs via middleware. Validation errors name the exact field and problem, never a generic `INVALID_INPUT`. Conventional Commits, scope `feat(v3):`, commit after every green cycle.

---

## 7. Testing Strategy

The repository has **no test suite today**. v3 introduces one; its scope is v3 code only.

### What a good test asserts here

**Observable outcomes, never mechanism.** That a seat was consumed, that a registration exists with the expected `source`, that an `attendance_logs` row was appended with the expected `outcome`, that a refusal carries the right error class. Never call order, never internal helper invocation, never the shape of intermediate values.

### Seams

| Seam | What it covers | Tooling |
|---|---|---|
| **Primary — usecase against a real Postgres** | Every behaviour in this spec is reachable by driving the registration and check-in usecases directly, with migrations applied and tables truncated between tests | `tests/integration/v3/`, shared harness, `TEST_DATABASE_URL` |
| **Secondary — pure functions** | Eligibility matrix, geo policy, `show_if` evaluator, recurrence expansion, QR token codec | `internal/pkg/*/`, table-driven, no DB, no HTTP |
| **Computed state** | Availability + `reasons[]` derivation | `internal/models/v3/state_test.go`, pure |

**A real database is not negotiable for the seat tests.** Several decisions here are specifically about concurrent seat integrity — two people racing for the last Family Room seat, walk-in creation booking atomically under load. A mocked repository cannot demonstrate that `SELECT ... FOR UPDATE` serialises anything, which is the property under test. Race tests run under `-race`.

**Deliberately not added: HTTP-level tests**, beyond one smoke test proving routes mount and auth wires up. Handler tests are not a seam here; everything worth asserting is reachable one level down, and adding them would duplicate assertions across two layers for no additional confidence.

### Test-first order

Each module spec lists its acceptance criteria as testable statements. Write the failing test, watch it fail, implement, watch it pass, commit. A task is not done because it looks right — it is done when the test that would have caught the mistake passes.

### Harness

`tests/integration/v3/harness_test.go` provides a migrated database and truncation between tests. Seeding helpers build an Event and Session **through the real usecases**, not by inserting rows directly, so seeds cannot drift from production behaviour. Tests `t.Skip` when `TEST_DATABASE_URL` is unset, so `go test ./...` stays green on a machine with no database.

---

## 8. Boundaries

### Always

- Run `go build ./...` and the unit suite before every commit; run the integration suite before every push.
- Write the failing test first. Commit only on green.
- Use `Atomic` for anything touching seats, registrations, attendance, or the outbox.
- Give every new error class both an English public message and an Indonesian translation.
- Append an `attendance_logs` row on **every** check-in path — success and rejection alike. The log records that a scan was attempted; it is not a claim of attendance.
- Enforce eligibility wherever a registration is **created** — all four paths (ADR 0007).
- Keep unconfigured capabilities invisible: no badge, no column, no endpoint (design §2.3).
- Name the field and the problem in every validation error.

### Ask first

- Any change to the JSONB config shapes (`eligibility`, `registration_config`, `geo_config`, `recurrence`) once `foundation` has shipped — these are the frontend's contract.
- Adding a dependency beyond the three named in §3.
- Adding a fifth `attendance_modes` value, or a new `registrations.source` value.
- Changing the migration number, or adding a second v3 migration.
- Anything that would make v3 depend on more of the permissions engine than the `Can()` contract in §2.
- Renaming a module id, or moving responsibility between modules.
- Touching v2 code, v2 tables, or `models.ErrorMapping`.

### Never

- Use `TransactionRepository.Transaction` in v3 code.
- Match a v3 error with `==` instead of `.Is()`.
- Update or delete an `attendance_logs` row. It is append-only audit truth; corrections are new rows.
- Auto-create a user account from an anonymous registration (ADR 0002).
- Guess a ticket category for a walk-in (ADR 0009).
- Store lateness, or allow a late mark to be cleared (ADR 0005).
- Retroactively invalidate an existing registration when eligibility rules change (ADR 0008).
- Auto-toggle check-out on a repeat scan (ADR 0004).
- Commit secrets — the QR HMAC key and SMTP credentials live in `config.local.yaml`, which is gitignored; only the template is committed.
- Build an `event_organizers` table (ADR 0001).

---

## 9. Superseded Statements

The plan documents remain in `docs/plans/` as history. These specific statements in them are **wrong** and must not be implemented. Every replacement is an accepted ADR.

| Where | Stale statement | Replaced by |
|---|---|---|
| design §1, §2.1, §9 | `event_organizers` table; `POST /v3/events/{code}/organizers`; "lightweight per-event organizer list" | **ADR 0001** — `permission_grants` + `Can()`; the generic grants API scoped to `resource=event:<code>` |
| design §2.1 | `registrations.registered_by` — "community_id, always set" | **ADR 0002** — nullable; anonymous registrations have none |
| design §2.1 | `attendance_logs.mode` includes `headcount` as a fifth value | **CONTEXT.md / ADR 0005** — the enum is the four modes; headcount is a `manual` sub-case |
| design §2.1 | `attendance_logs.outcome` = `checked_in\|already_checked_in\|rejected_geo\|rejected_other` | **ADR 0005 / 0007 / 0009** — adds `headcount_undone`, `rejected_not_eligible`, `rejected_no_category`; **ADR 0008** adds a stored eligibility-recheck result column |
| design §2.1 | `event_sessions` has no `enable_late_marking` | **ADR 0005** — per-session, default off |
| design §2.2 | `registration_config` has no `channels`, no `max_sessions_per_person` | **ADR 0006**, **ADR 0004** |
| design §2.2 | `geo_config.modes` has five rows, none for desk | **ADR 0006** — adds `desk_registration`, default `off` |
| design §2.2 | `eligibility` has no re-check setting | **ADR 0008** — event-level `off\|note\|block`, default `off` |
| design §2.2 | `recurrence.session_defaults` = `start_time`/`duration_min`/`total_seats` | **ADR 0005** — must also carry `attendance_modes`, `enable_checkout`, `enable_late_marking`; templates likewise |
| design §5.2, §11; impl Task 11 | Duplicate check-in is idempotent and returns **no error** | **ADR 0004** — distinct `ALREADY_CHECKED_IN` error class; the log row is still appended |
| design §5.2; impl Task 11 | Walk-in branch checks only "mode enabled" and "seats available" | **ADR 0007** — eligibility also runs; **ADR 0009** — category required on categorised sessions |
| design §5.2 | Staff authorization = event organizer or event-admin RBAC role | **ADR 0001** — `Can(actor, event.checkin\|event.manage, event:<code>)` |
| design §5.3; impl Task 11 | Check-out shares the check-in window | **ADR 0004** — check-out is bounded by the session's `end_at` |
| design §6.1 | `session` token expires at `checkin_close_at` | **ADR 0009** — expires at `end_at` when `enable_checkout` is true; unchanged otherwise |
| impl Task 11 | `manual` mode always **creates** a walk-in | **ADR 0004** — `manual` accepts an optional existing attendee id and checks that person in without creating a registration or booking a seat |
| impl Task 15 | Summary sheet reports one undifferentiated "registered" total | **ADR 0004** — split into `web`, `admin`, `walk_in` |
| impl plan, File Structure | Migration `000022_v3_event_management` | **Collision** — `000022_configs_setup` exists; the COOL plan claims `000023`/`000024`. v3 uses **`000025`**. |
| impl plan §Tech Stack | "Go 1.23" | **§3** — herr requires Go 1.26.1. Already done: `go.mod` and the Dockerfile were bumped on 2026-09-10. |
| an earlier decision, cited in the 2026-09-09 spec | `attendance_modes` must be non-empty | **Reversed** — empty is valid and meaningful: registration happens, attendance is not tracked. Distinct from `registration_config.mode: none`. |

---

## 10. Success Criteria

The program is done when all of the following are true and demonstrated by a passing test.

**Correctness of the three doors**

1. A registration created online, at a desk in advance, and by QR at the door all draw from **one shared seat pool** per session; two concurrent attempts on a one-seat session produce exactly one success and one `QUOTA_FULL` (under `-race`).
2. All three creation paths **plus** QR walk-in auto-creation run `eligibility.Check`; a non-volunteer presenting a personal QR at a volunteers-only event is refused with `NOT_ELIGIBLE` and an `attendance_logs` row with `outcome=rejected_not_eligible`.
3. On an event where `eligibility.audience != everyone`, an anonymous desk registration is refused; on `everyone`, it succeeds with `registered_by = NULL` and no user account is created.
4. Desk registration succeeds for an actor holding `event.checkin` while registration is open, is refused for that same actor after `register_end_at`, and succeeds for an actor holding `event.manage` in both windows.

**Nobody legitimate is turned away**

5. An anonymous lobby-table registrant with no email and no QR is found by name through `manual` mode and checked in **without** creating a second registration or consuming a second seat.
6. A second scan of an already-checked-in attendee returns the `ALREADY_CHECKED_IN` error class **and** appends a log row with `outcome=already_checked_in`; a report counting `outcome='checked_in'` is unchanged by it.
7. A walk-in scan on a session with ticket categories is refused with `CATEGORY_REQUIRED` when no category is supplied, and books into the named category's pool when one is.

**Attendance mechanics**

8. `attended_at > start_at` renders as late **only** when `enable_late_marking` is on; with it off, lateness appears in no response and no report column.
9. A headcount tap followed by an undo nets to zero, both rows remain in `attendance_logs`, and a second undo with nothing to cancel is refused.
10. Check-out succeeds after `checkin_close_at` and before `end_at`; the session poster token still resolves in that window when `enable_checkout` is true, and expires at `checkin_close_at` when it is false.
11. With `eligibility` re-check set to `note`, a person who no longer qualifies is admitted and the result is **stored** on the log row; with `block` they are refused; with `off` (default) no re-check runs and nothing is stored.

**Configuration behaves**

12. `channels: ["desk"]` produces no active public registration endpoint; `channels: []` is rejected at publish.
13. `attendance_modes: []` activates no check-in endpoint and reports registration data only — never "0 of 40 attended".
14. `max_sessions_per_person: 1` rejects a second session of the same event, naming the session already registered for; anonymous registrations are exempt.
15. A recurring event generates sessions carrying `attendance_modes`, `enable_checkout` and `enable_late_marking` from `session_defaults`; re-running generation creates nothing new and does not touch an edited session.

**Cross-cutting**

16. Every error class returns a humanized message in both English and Indonesian, selected by `Accept-Language`, defaulting to `id`.
17. `go build ./...`, `go test ./internal/... ./tests/...` and the integration suite are all green; `golangci-lint run` is clean.
18. No v3 code path calls `TransactionRepository.Transaction`; no v3 error is matched with `==`.

---

## 11. Out of Scope

Deliberate deferrals, recorded so nobody re-litigates them:

- **Per-session geo policy** — `geo_config` carries venue coordinates, so moving it per-session really means "this session is somewhere else," which deserves its own design (ADR 0006).
- **Personal QR rotation** — generated once at account registration, permanent, non-rotatable. A deliberate trade for a congregation that skews older and less technical (ADR 0003).
- **Anti-replay on the session poster QR** — it stays a static, non-rotating code; the trust boundary is accepted. If abuse appears, disable `session_qr` for that session (ADR 0003).
- **Clearing a late mark** — lateness is a fact, not a judgment; a reason belongs in a note beside the attendance (ADR 0005).
- **Enforcing `max_sessions_per_person` for anonymous registrations** — would require making phone mandatory for people who may have none, and name matching fails in both directions (ADR 0004).
- **Printing at the lobby table** — the attendee-lookup check-in solves the actual problem without printers, paper, or a slip likely lost over five weeks.
- **Revising how user types and roles are modelled** — flagged as wanting attention, explicitly left as-is.
- **Payments, v2 data migration, additional notification channels, assigned seat numbers, per-field i18n, SEO blocks, captcha.**

---

## 12. Open Questions

Each names the module that will be blocked if it stays unanswered.

1. **Where does the eligibility re-check setting physically live?** ADR 0008 says "event-level, alongside `eligibility` itself." This spec places it **inside** the `eligibility` JSONB as `recheck_at_checkin: off|note|block`, so one config object owns both who qualifies and how strictly. The alternative is a sibling column. Blocks: `foundation` (schema), `attendance`. — *Proceeding with the JSONB field unless told otherwise.*
2. **How is a headcount undo linked to its tap?** ADR 0005 says one undo cancels exactly one tap and the total is the net. This spec nets by **counting outcomes** (`checked_in` minus `headcount_undone` among headcount rows) and refuses an undo that would take the net below zero. A `undoes_log_id` self-reference would be stricter but adds a column and a lookup. Blocks: `foundation` (schema), `attendance`.
3. **Does the cross-session limit count cancelled registrations?** `max_sessions_per_person` presumably counts only `confirmed` ones, so cancelling frees the slot. Not stated in ADR 0004. Blocks: `registration`. — *Proceeding with confirmed-only.*
4. **What is `Actor` in the `Can()` signature?** v3 builds caller identity as `eligibility.User`. Whether the permissions engine accepts that type or its own must be settled at the boundary before `registration` is built. Blocks: `registration`, `attendance`.
5. ~~**Is Go 1.26 available on CI and every dev machine?**~~ **Resolved 2026-09-10.** Go 1.26.1 is installed and first on PATH (Homebrew); `go.mod` now declares `go 1.26.1` and the Dockerfile build stage uses `golang:1.26.1-alpine`. `go build ./...` passes. Two caveats for whoever picks this up:
   - A second, stale Go 1.25.5 lives at `/usr/local/go/bin/go`, shadowed by Homebrew's. Anything invoking that path explicitly will still see 1.25.5.
   - There are **no CI workflows in this repository** (`.github/workflows` does not exist), so the Dockerfile build stage is the only pinned toolchain besides `go.mod`. When CI is added, pin it to 1.26.1 too.
   - `go vet ./...` exits 1 on three pre-existing findings in v2 code (`internal/pkg/database/postgre/postgre.go:13`, `internal/repositories/pgsql/transaction_pg_repository.go:38,65`). **Not caused by the upgrade** — verified by re-running vet with the `go 1.23.0` directive, which fails identically. Left untouched as out of scope; worth a separate cleanup before `golangci-lint run` becomes a gate.
