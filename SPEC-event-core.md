# Spec: `event-core`

**Module id:** `event-core` · **Depends on:** `foundation`, `policy-kernel` · **Build order:** 3rd
**Parent:** [SPEC.md](SPEC.md)

---

## Objective

The organizer's side of the system: create an event as a draft with almost nothing filled in, fill it in gradually without fighting validation, and be stopped only at publish — with an error that names the session and the field, not a generic rejection. Plus the machinery that makes a recurring event configured once rather than fifty-two times, and the computed state that stops every frontend re-deriving "is this open?" for itself.

**Staff can't get lost** is the design principle this module carries. Drafts with late validation, templates, computed human-readable state on every response, specific error messages.

## Scope

**In:** Event/session/ticket-category CRUD; the draft→published lifecycle and its strict publish gate; templates (save-as, create-from, duplicate, system seeds); recurrence generation and publish-window flips; computed session state with machine-readable `reasons[]`.

**Out:** Registrations and attendance (their own modules). Permission checks are consumed from the external `permissions` engine, not defined here.

---

## Contracts

```go
type EventUsecase struct{ r *pgsql.PostgreRepositories; cfg *config.Configuration }

CreateDraft(ctx, title, createdBy string) (*v3.EventV3, error)   // 7-char code + slug
Patch(ctx, code string, patch map[string]any) (*v3.EventV3, error)
Publish(ctx, code string) error                                   // the strict gate
Cancel(ctx, code string) error                                    // + notify every confirmed registrant
Archive(ctx, code string) error
HardDelete(ctx, code string) error                                // drafts only
CreateSessions(ctx, eventCode string, in []v3.SessionInput) ([]v3.SessionV3, error)  // bulk
PatchSession / DeleteSession
CreateCategories / PatchCategory / DeleteCategory                 // delete only when no confirmed registrations
SaveAsTemplate / CreateFromTemplate / Duplicate
SeedSystemTemplates(ctx) error                                    // idempotent

type RecurrenceUsecase struct{ r *pgsql.PostgreRepositories }
GenerateSessions(ctx) (created int, err error)                    // + due publish/unpublish flips

// pure, unit-tested without DB
func BuildSessionState(ev *v3.EventV3, s *v3.SessionV3, cats []v3.TicketCategory,
	rc v3.RegistrationConfig, now time.Time) v3.SessionState
```

---

## Acceptance Criteria

### Lifecycle

1. `CreateDraft` needs **only a title** and produces a draft with a generated code and slug.
2. Draft saves validate loosely — types only. Staff never fight a half-filled form.
3. `Publish` validates strictly and, on failure, returns errors that **name the session and the field**: "session 2: registration closes after the session starts", never `INVALID_INPUT`. A test asserts the session code appears in the message.
4. Publish requires: at least one session; coherent times per session (`register_end_at ≤ start_at` where both are set, `start_at < end_at`, `checkin_open_at < checkin_close_at`); `form.ValidateSchema` passes; `geo_config` mode values ∈ `{off, warn, require}`; `registration_config.ValidateStrict` passes.
5. **Publish rejects `registration_config.channels: []`** — `mode: none` already expresses "nobody registers" (ADR 0006).
6. **Publish accepts `attendance_modes: []`** on any session. This is valid and meaningful: registration happens, attendance is not tracked — a trip sign-up sheet needing bus numbers. A test asserts publish succeeds and that the session activates no check-in endpoint. Forcing a mode here is what produces misleading "0 of 40 attended" statistics.
7. Anything published is deleted **softly** (archived). Hard delete exists only for drafts.
8. `Cancel` enqueues an `event_cancelled` notification for every confirmed registrant, inside the same transaction as the status change.

### Sessions and categories

9. All four time windows are **per session** — session period, registration window, check-in window — and the event's overall period is **derived** (earliest start → latest end), never stored. A test asserts no event-level period column is written.
10. `CreateSessions` accepts a bulk array in one call and generates a code per session.
11. `total_seats: 0` means unlimited, and is ignored entirely when the session has ticket categories — then each category's own count is the pool.
12. `DeleteCategory` succeeds only when the category has no confirmed registrations; otherwise `FORBIDDEN` naming the count.
13. Ticket categories are optional and invisible when unused: a session with zero categories keeps `registrations.category_id` null and gains no category column anywhere.

### Templates

14. `SeedSystemTemplates` is idempotent — running it twice leaves exactly four rows: **Sunday Service** (personal QR, no form), **Conference** (self-registration, per-attendee form, registration QR), **Weekly Class** (recurring weekly, members-only), **Announcement** (info-only).
15. A template snapshot is the **complete** config set — all four JSONB configs plus session defaults **plus `attendance_modes`, `enable_checkout`, `enable_late_marking`**. This is the same failure ADR 0005 identifies for recurrence: a template omitting them silently produces events that record nobody. A test asserts round-trip equality including those three fields.
16. `CreateFromTemplate` takes `{title, first_session_date}` and produces a fully configured draft whose `registration_config` equals the snapshot's.

### Recurrence generation

17. `GenerateSessions` materialises sessions up to `generate_ahead` periods for every published event with a recurrence rule, computed deterministically via `recurrence.Expand`.
18. **Idempotent** — re-running creates nothing, via `UNIQUE(event_id, occurrence_key)` and `ON CONFLICT DO NOTHING`.
19. **Generation never updates an existing row.** Editing a generated session, then re-running generation, leaves the edit intact. Cancelling one occurrence cancels only that session. A test covers both.
20. **Generated sessions carry `attendance_modes`, `enable_checkout` and `enable_late_marking` from `session_defaults`** (ADR 0005). A test generates from a rule specifying `[personal_qr]` and asserts every produced session has it — this is the criterion that prevents 52 weeks of untracked volunteer meetings.
21. Rule changes affect only future, not-yet-generated sessions.
22. The same call applies due `publish_at`/`unpublish_at` visibility flips. Status stays `published`; visibility is computed.

### Computed state

23. Every event and session response embeds derived state so no frontend re-derives it:

```json
{
  "availability": "open | full | opens_soon | closed | walk_in_only | info_only",
  "reasons": ["QUOTA_FULL"],
  "seats": { "total": 500, "booked": 342, "remaining": 158 },
  "categories": [{ "code": "main-hall", "availability": "open", "remaining": 120 }],
  "active_phase": "General",
  "active_modes_now": ["registration_qr", "manual"],
  "checkin_window": { "open": true, "closes_at": "..." }
}
```

24. `reasons[]` is always machine-readable: `QUOTA_FULL`, `REGISTRATION_NOT_STARTED`, `REGISTRATION_ENDED`, `MARKED_SOLD_OUT`, `OUTSIDE_PUBLISH_WINDOW`, `NOT_ELIGIBLE`, `INFO_ONLY`, `WALK_IN_ONLY`, `CHANNEL_CLOSED`.
25. `categories` and `active_phase` appear **only** when the event uses them (design §2.3 — unconfigured capabilities stay invisible).
26. `marked_sold_out` on a session or category closes registration instantly regardless of remaining quota, and surfaces `MARKED_SOLD_OUT` in `reasons[]`.
27. **`channels: ["desk"]` surfaces `CHANNEL_CLOSED`** for the web path, so the public event page can render "sign up at the welcome desk" instead of a sign-up button. Without this the page reads as broken (ADR 0006).
28. `BuildSessionState` is a pure function, unit-tested without a database.

---

## Verification

```bash
go test ./internal/models/v3/ -run TestBuildSessionState -v      # pure state derivation
TEST_DATABASE_URL=... go test ./tests/integration/v3/ -run 'TestEvent|TestRecurrence|TestTemplate' -v
```

Key cases to see fail first: publish with no sessions; publish with incoherent times naming the session; publish with `channels: []`; publish with `attendance_modes: []` **succeeding**; seed-twice-yields-four; generate-twice-creates-zero; edit-then-regenerate preserves the edit; generated session carries `enable_late_marking`.

---

## Boundaries (module-specific)

- **Always** make the publish error name the session and field.
- **Always** keep unconfigured capabilities invisible — no category column, no phase, no lateness on events that never asked.
- **Never** store an event-level period; derive it.
- **Never** let recurrence generation update an existing session row.
- **Never** build an `event_organizers` table or an organizer-assignment endpoint (ADR 0001) — staff assignment is the generic grants API scoped to `resource=event:<code>`.
- **Ask first** before adding a `reasons[]` constant; the frontend switches on them.

---

## Admin-UI wording obligation

This module owns the settings that carry unusual weight, because three pairs are genuinely distinct but sound alike. Whoever writes the admin UI must label them by the distinction, or they will be set wrongly:

| Pair | The distinction to name |
|---|---|
| `channels` vs `attendance_modes` | Staff registering people **in advance** vs **at the door** |
| `max_sessions_per_person` vs `allow_multiple_registrations` | Different sessions of one event vs the **same** session twice |
| `enable_late_marking` vs the check-in window | Whether lateness is **shown** vs how long the door stays **open** |
