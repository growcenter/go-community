# Spec: `http-api`

**Module id:** `http-api` · **Depends on:** all · **Build order:** 7th (last)
**Parent:** [SPEC.md](SPEC.md)

---

## Objective

Expose the seven modules beneath as one coherent API under `/api/v3`, mounted beside v1 and v2 in the existing Echo composition. Handlers bind, call one usecase method, and respond. No logic lives here — if a handler decides something, that decision belongs a layer down.

## Scope

**In:** Route table, handler shape, auth wiring, caller-identity construction, the response envelope, computed-state exposure, the internal Cloud Scheduler endpoints, Swagger annotations.

**Out:** Any business rule. Every rule is specified in the module that owns it.

---

## Route Table

```
── Public (guests may view, never act) ─────────────────────────
GET    /v3/events                        published events, paginated, filters (campus, topic, date)
GET    /v3/events/{code}                 detail + sessions + form schema + computed state

── Authenticated member ────────────────────────────────────────
POST   /v3/sessions/{code}/registrations register self + companions   (channels must include `web`)
GET    /v3/me/registrations              mine, upcoming and past
DELETE /v3/registrations/{id}            cancel own
PATCH  /v3/registrations/{id}/attendees/{attendee_id}   edit answers per edit_policy
GET    /v3/registrations/{id}/ticket.pdf owner, or event.manage
POST   /v3/qr/resolve                    what is this QR + my allowed actions
POST   /v3/qr/act                        perform an action (self check-in via session QR, checkout)

── Staff — requires Can(actor, event.checkin, event:<code>) ────
POST   /v3/sessions/{code}/registrations/desk   NEW — advance desk registration (source=admin)
POST   /v3/sessions/{code}/checkin        personal_qr | registration_qr | manual | headcount
POST   /v3/sessions/{code}/checkout       when enable_checkout
POST   /v3/sessions/{code}/headcount/undo NEW — append a correcting entry
GET    /v3/sessions/{code}/attendees      list, search by name, status filter, live counts
GET    /v3/staff/dashboard                my events, today's sessions, live counts

── Organizer — requires Can(actor, event.manage, event:<code>) ─
GET    /v3/events/{code}/summary          per-session stats
GET    /v3/events/{code}/report           ?format=csv|xlsx
GET    /v3/sessions/{code}/report         ?format=csv|xlsx
POST   /v3/events                         create draft (title is enough)
PATCH  /v3/events/{code}                  partial update
POST   /v3/events/{code}/publish          strict validation gate
POST   /v3/events/{code}/cancel           cancel + notify registrants
DELETE /v3/events/{code}                  archive (hard delete: drafts only)
POST   /v3/events/{code}/sessions         create sessions (single or bulk array)
PATCH  /v3/sessions/{code}                partial update
DELETE /v3/sessions/{code}                cancel/archive session
POST   /v3/sessions/{code}/categories     create ticket categories (single or bulk)
PATCH  /v3/categories/{id}                partial update incl. marked_sold_out
DELETE /v3/categories/{id}                only when no confirmed registrations
GET    /v3/templates                      list (system + own)
POST   /v3/events/from-template/{id}      one-call creation
POST   /v3/events/{code}/save-as-template
POST   /v3/events/{code}/duplicate

── Internal (Cloud Scheduler, X-API-Key) ───────────────────────
POST   /v3/internal/notifications/dispatch    every 5 min
POST   /v3/internal/sessions/generate         every 30 min; also applies publish-window flips
```

**Three changes from the original design §9**, each carried from an ADR:

| Change | Why |
|---|---|
| `POST /v3/events/{code}/organizers` **removed** | ADR 0001 — staff assignment is the generic grants API scoped to `resource=event:<code>` |
| `POST /v3/sessions/{code}/registrations/desk` **added** | ADR 0004 — `admin` exists in the `source` enum but nothing writes it; the lobby table has no endpoint today |
| `POST /v3/sessions/{code}/headcount/undo` **added** | ADR 0005 — a mis-tap otherwise inflates attendance permanently |

---

## Acceptance Criteria

### Handler shape

1. Every handler is: bind → call **one** usecase method → `respond.OK` or `respond.Err`. No branching on business state, no validation beyond binding, no repository access.
2. `respond.OK(ctx, status, data)` emits `{"code":"SUCCESS","data":...}`. `respond.Err(ctx, err)` emits herr's public representation with the class's Kind as the HTTP status, and its field-level `errors[]` for validation failures.
3. **Internal error detail never reaches a response body.** A test asserts a class carrying `Internal(...)` renders without it.
4. `Accept-Language: en` returns English messages; anything else, including absent, returns Indonesian.

### Auth wiring

5. Public group: plain, no middleware. Member group: `middleware.UserMiddleware(c, u, nil)`.
6. Staff and organizer endpoints use the member middleware plus an in-handler `Can(actor, action, "event:<code>")` check against the external permissions engine (ADR 0001). A single helper — `requireGrant(ctx, action, eventCode) error` — is the only place this is expressed, so no endpoint can forget it.
7. **Desk registration's permission requirement is time-dependent** and therefore resolved in the usecase, not the handler: `event.checkin` while registration is open, `event.manage` after `register_end_at` (ADR 0005). The handler passes the actor through; it does not pre-judge.
8. Internal endpoints validate `X-API-Key == cfg.Auth.APIKey` and nothing else.
9. Caller identity is built once per request from the JWT context into `eligibility.User` via `callerFrom(ctx)`, loading campus through the existing user usecase. One helper, used everywhere.

### Behaviour surfaced from below

10. `POST /v3/sessions/{code}/registrations` returns `CHANNEL_CLOSED` when `channels` excludes `web` — and `GET /v3/events/{code}` surfaces the same in `reasons[]`, so the public page can render "sign up at the welcome desk" instead of a broken-looking button (ADR 0006).
11. Check-in endpoints return **404-equivalent inactivity** when `attendance_modes: []` — no check-in endpoint activates for such a session.
12. `POST /v3/qr/resolve` returns `{type, ref, allowed_actions}` computed **per caller**: a staff member scanning a personal QR during a session they can check into sees `event_checkin`; a member scanning a friend's QR sees an empty list.
13. `event_checkout` appears in `allowed_actions` only when the subject is already checked in and the session has `enable_checkout` — this is what makes check-out an explicit offered action rather than a repeat-scan side effect (ADR 0004).
14. `POST /v3/qr/act` on an action absent from that caller's `allowed_actions` returns `FORBIDDEN`.
15. A duplicate check-in returns the `ALREADY_CHECKED_IN` class with its own HTTP status, distinguishable by the scanner from geo and forgery rejections (ADR 0004).
16. Every event and session response embeds the computed state from `event-core` — `availability`, `reasons[]`, `seats`, and `categories`/`active_phase` **only when used**.

### Docs

17. Swagger annotations (`@Summary`, `@Router`, ...) sit above every handler; `make generate-docs` regenerates `docs/` cleanly.
18. The `v3:` config block is documented in `config/config.local.template.yaml`; real values live only in the gitignored `config.local.yaml`.
19. `internal/contract/contract.go` calls `SeedSystemTemplates` after `usecases.New` — logging a warning on failure, never crashing the boot.
20. `AGENTS.md` gains a short "Event Management v3" paragraph pointing at `SPEC.md`.

---

## Verification

```bash
go build ./...
make generate-docs && git diff --exit-code docs/     # annotations are complete and committed
TEST_DATABASE_URL=... go test ./tests/integration/v3/ -run TestHTTPSmoke -v
```

**One smoke test, deliberately thin.** HTTP is not a seam in this program — everything worth asserting is reachable one level down, and handler tests would duplicate assertions across two layers for no additional confidence. The smoke test proves only that routes mount and auth wires up:

`GET /v3/events` lists a seeded published event · `POST /v3/sessions/{code}/registrations` returns 201 with attendees · `POST /v3/sessions/{code}/registrations/desk` succeeds for a granted actor and is `FORBIDDEN` without the grant · `GET /v3/sessions/{code}/attendees` returns the row · `GET .../report?format=csv` returns `text/csv` · an unauthenticated call to a member route is 401.

JWT middleware is replaced in tests by an injected test middleware — export `NewV3HandlerWithAuth(g, u, c, authMW echo.MiddlewareFunc)` so tests can supply identity without minting tokens.

---

## Boundaries (module-specific)

- **Always** keep handlers logic-free. If you are writing an `if` about business state in `deliveries/http/v3/`, it belongs in a usecase.
- **Always** express the permission check through `requireGrant`, never inline.
- **Never** mount v3 in a way that changes v1 or v2 routing.
- **Never** return a generic `INVALID_INPUT` — herr's `FieldError` entries name the field and the problem.
- **Never** log a QR token; it is a bearer credential.
- **Ask first** before adding a route to the table above, or changing the response envelope.

---

## Known simplification

Event listing uses limit/offset rather than v2's cursor pagination. At ~1,000 users this is acceptable, and it is recorded here deliberately so it is not mistaken for an oversight.
