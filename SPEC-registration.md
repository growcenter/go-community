# Spec: `registration`

**Module id:** `registration` · **Depends on:** `event-core`, `policy-kernel`, *external* `permissions` · **Build order:** 4th
**Parent:** [SPEC.md](SPEC.md)

---

## Objective

Three doors into an event, all enforcing the same rules: a person signing themselves up online, staff signing someone up in advance at a desk, and — created by `attendance` calling into this module — a registration made on the spot when somebody scans in without a booking. Organizers choose which doors are open.

Two of the problems this solves are not features but **bugs in already-approved work**: staff cannot sign anyone up at all today, and the eligibility rules an organizer configures are enforced on exactly one of the paths people actually use.

## Scope

**In:** `registration_config.channels`; online self-registration; advance desk registration (`source=admin`) with its permission split; the shared atomic seat-booking primitive; eligibility on every creation path; anonymous registration rules; `max_sessions_per_person`; cancellation; answer editing.

**Out:** Check-in of any kind, including the walk-in creation *decision* — `attendance` decides to create and calls `BookAndCreate` here. Everything below applies to that call too.

---

## Contracts

```go
type RegistrationUsecase struct{ r *pgsql.PostgreRepositories; cfg *config.Configuration }

type RegisterInput struct {
	SessionCode  string
	User         eligibility.User   // zero value = anonymous (desk only)
	Source       string             // web | admin | walk_in
	ActedBy      *eligibility.User  // staff actor for admin/walk_in; nil for web
	Coords       *geo.Coords
	AccessCode   string
	CategoryCode string             // required when the session has categories
	Primary      map[string]any
	Companions   []map[string]any
}

Register(ctx, in RegisterInput) (*RegisterOutput, error)
Cancel(ctx, regID uuid.UUID, by eligibility.User, canManage bool) error
EditAnswers(ctx, attendeeID uuid.UUID, by eligibility.User, canManage bool, answers map[string]any) error
ListMine(ctx, communityID string) ([]v3.RegistrationV3, error)

// The shared seat primitive. `attendance` calls this for walk-in creation so that
// capacity is one pool across all three doors and cannot drift between them.
BookAndCreate(ctx context.Context, tx *pgsql.PostgreRepositories, in RegisterInput) (*RegisterOutput, error)
```

---

## Acceptance Criteria

### Channels — which doors are open (ADR 0006)

1. `registration_config.channels` is a list of `web` and `desk`, defaulting to **both**. `mode` says *whether* registration happens; `channels` says *through which door*.
2. `channels: ["desk"]` → online self-registration returns `CHANNEL_CLOSED`; desk registration still works.
3. `channels: ["web"]` → desk registration returns `CHANNEL_CLOSED`; online still works.
4. `channels: []` is rejected at publish (`event-core` criterion 5), never at registration time.
5. **`channels` and `attendance_modes` are independent**, and both asymmetric combinations must work. Two tests, because both are real:
   - *Door open, table closed* — a members-only event where staff should not be quietly adding people through the preceding month, but an usher should still admit a genuine member who forgot to book.
   - *Table open, door closed* — a catered dinner needing final numbers a week ahead: staff add people in advance so the caterer knows, but nobody is added on the night because there is no food for them.

### Online self-registration

6. Auth is **always** required. Guests may view any event page but must sign in to register.
7. Window resolution: with `phases`, the active phase is resolved server-side by current time, honouring its `eligibility_override`, `max_per_registration`, and `access_code` (compared case-insensitively). Without phases, `now ∈ [register_start_at, register_end_at]`. No active window → `REGISTRATION_CLOSED`.
8. `marked_sold_out` on the session, or on the targeted category, → `MARKED_SOLD_OUT` regardless of remaining quota.
9. Party size = 1 + companions, capped by the phase override else `max_per_registration` (`0` = unlimited).
10. `geo.Validate(cfg, "web_registration", coords, false)` runs; a `require` failure blocks. Self-registrants cannot override their own geo check.
11. Form validation runs per attendee: `primary` audience for the registrant, `companion` for each companion when `companion_detail: full`; `count_only` keeps names only, defaulting to `"Guest of <primary name>"` when blank.
12. Duplicate guard: one `confirmed` registration per `(session_id, registered_by)` unless `allow_multiple_registrations`. Companions are never dedup-checked — guest companions have no identity.

### Advance desk registration — `source=admin` (ADR 0004, 0005)

13. This is a **capability on this usecase, not a new module**, deliberately, so the test seam count stays at one. Do not create a separate desk-registration usecase.
14. Governed by the same `desk_registration` setting as the door: `account_required` (staff look up an existing member), `anonymous_only`, `anonymous_allowed` (ADR 0002).
15. **Permission split across the `register_end_at` boundary**, evaluated via `Can(actor, action, "event:<code>")`:

    | When | Required grant | Why |
    |---|---|---|
    | While registration is open | `event.checkin` | The lobby table is typically staffed by volunteers, doing for someone what that person could have done online |
    | After `register_end_at` | `event.manage` | This overrides a cutoff the organizer set and silently changes catering and seating numbers they believed final |

    Four tests: `event.checkin` before the boundary succeeds; `event.checkin` after it is `FORBIDDEN`; `event.manage` succeeds in both.
16. Allowed up until **check-in closes**. Blocking it would not prevent the outcome — staff would tell the person to arrive on the day and be registered as a walk-in, which admits them anyway while giving the organizer *less* warning and filing them under the wrong `source`.
17. Seat limits apply throughout — desk registration cannot overfill a venue.
18. **A lobby-table registrant receives nothing physical.** An email confirmation is enqueued when an address was given, and **silently skipped otherwise** — no outbox row is created that can never be delivered. They are found by name at the door via `attendance`'s lookup check-in. A test asserts zero outbox rows for a registration with no email.
19. `desk_registration_outcome` was considered and **dropped** — do not add it. It made configuration out of what is simply which action a staff member took, and `source` already distinguishes the two cases (ADR 0004).

### Anonymous registration (ADR 0002, 0007)

20. Anonymous means `registrations.registered_by IS NULL`. Only ever reachable through desk registration — online self-registration and QR registration always require an account.
21. **Never auto-creates a user account.** If that person wants one later, it is a separate, self-initiated, consent-driven signup, unlinked from the registration record. A test asserts no `users` row is created.
22. Anonymous registrations are exempt from the same-session duplicate guard, the way guest companions already are.
23. **When `eligibility.audience != everyone`, anonymous registration is refused** with `ANONYMOUS_NOT_ALLOWED` — at the lobby table and at the door alike. There is no account for the rules to evaluate, and anyone legitimately eligible has one by definition. Open events are unaffected.

### Eligibility on every creation path (ADR 0007)

24. `eligibility.Check` runs wherever a registration is **created**: online self-registration, advance desk registration, staff walk-in creation at the door, and QR-triggered walk-in auto-creation. Because `attendance` creates through `BookAndCreate`, putting the check inside that primitive is the way to make it structurally impossible to skip.
25. A refusal returns `NOT_ELIGIBLE` **naming the reason**, never a generic failure.
26. A test drives the volunteers-only bypass directly: a non-volunteer presenting a personal QR at the door is refused. This is the most serious of the three gaps because it needs no staff involvement at all.

### Cross-session limit — `max_sessions_per_person` (ADR 0004)

27. Lives in `registration_config`. `0` = unlimited (default, and correct for a weekly service where every week is a session of one event); `1` = one session of that event only; `N` = up to N.
28. **Distinct from `allow_multiple_registrations`**, which governs registering twice for the *same* session. Both exist; they answer different questions.
29. Exceeding it is a hard reject **naming the session already registered for** — never a silent auto-cancel-and-switch, which would be a surprising side effect of pressing register.
30. **Anonymous registrations are exempt.** With no account the only identifiers are a name and possibly a phone; name matching fails in both directions (two real "Budi Santoso"s get one wrongly turned away; one person entering "Budi" then "Budi S." passes twice). Enforcing it would require making phone mandatory for people who may have none to give.
31. Only `confirmed` registrations count toward the limit — cancelling frees the slot. *(Parent Open Question 3; proceeding with this reading.)*

### Seat integrity — one pool, three doors

32. Capacity is **one shared pool per session** across every path. No per-path quotas. Enforced once, at the moment of registration — including the auto-registration moment triggered by a QR scan.
33. Check-and-increment happen inside one `Atomic` transaction with `SELECT ... FOR UPDATE` on the session row, or on the ticket-category row when categories are used.
34. **The race test is mandatory and runs under `-race`:** a session with one seat, two concurrent `Register` calls, exactly one success and exactly one `QUOTA_FULL`. Repeated for a categorised session racing for the last Family Room seat, and for a walk-in creation racing an online registration — because a mocked repository cannot demonstrate that `FOR UPDATE` serialises anything, which is the property under test.
35. Full is a **hard reject**, never a waitlist.

### Cancellation and editing

36. `Cancel` sets status and `cancelled_at`, moves non-attended attendees to `cancelled`, decrements `booked_seats` by the count of non-attended party members atomically, and enqueues `registration_cancelled`. Already-cancelled → error.
37. `EditAnswers` honours `edit_policy`: `until` anchors to `register_end` / `checkin_open` / `session_start` / `never`; `fields` whitelists editable keys; an actor holding `event.manage` bypasses the window.
38. Edits re-run full form validation including `show_if` — changing a controlling answer drops now-hidden answers and may newly require others.
39. Every edit appends the previous `answers` snapshot to `answer_revisions`. **Party size and session are not editable** — cancel and re-register instead, which keeps seat accounting trivial.

---

## Verification

```bash
TEST_DATABASE_URL=... go test ./tests/integration/v3/ -run TestRegistration -race -v
```

Tests assert **observable outcomes**: a seat was consumed, a registration exists with the expected `source`, a refusal carries the right error class. Never call order, never internal helper invocation. Seeding helpers build the event and session through the real `event-core` usecases, not by inserting rows.

---

## Boundaries (module-specific)

- **Always** route every registration creation through `BookAndCreate`, so eligibility and the seat lock cannot be bypassed by a new caller.
- **Always** use `Atomic`; the seat lock, the inserts and the outbox row share one transaction.
- **Never** create a separate desk-registration usecase or module.
- **Never** auto-create a user account from an anonymous registration.
- **Never** enforce `max_sessions_per_person` against an anonymous registration.
- **Never** silently cancel-and-switch on a limit breach.
- **Ask first** before adding a `source` value or a `channels` value.
