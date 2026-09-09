---
status: accepted
---

# Registration channels are an explicit list, and desk registration gets its own geo policy

Continuation of the 2026-09-09 design session (see ADR 0004, ADR 0005). These decisions cover *through which door* a registration can be created, and whether staff-created advance registrations are location-checked.

## `registration_config.channels` — which doors are open

`registration_config.mode` (`none | required | optional`) says *whether* registration happens; it cannot express *who may create one*. There is currently no way to run an event where the public sign-up link does not exist but staff can still add people at the lobby table.

We add **`channels`**, a list defaulting to both values:

- **`web`** — self-service online registration
- **`desk`** — staff-created advance registration (`source=admin`, the lobby-table endpoint from ADR 0004)

| Intent | Configuration |
|---|---|
| Either route (today's behaviour) | `["web", "desk"]` |
| Self-registration allowed, but only while physically at the venue | `["web", "desk"]` + `geo_config.modes.web_registration.check = "require"` |
| Staff-only; no public sign-up link | `["desk"]` |
| Online only; no lobby table | `["web"]` |

A list rather than a boolean, mirroring `attendance_modes` so the two read consistently, and giving the web-only case for free. An **empty list is rejected** — it would mean nobody can register, which `mode: none` already expresses, and two ways to say the same thing invites drift.

Both "onsite-only" interpretations were wanted: physically-present self-registration is already achievable through the existing `web_registration` geo mode and needs no new field; staff-only registration is the genuinely new capability this adds.

**Frontend consequence:** with `["desk"]`, the public event page must render alternative wording ("sign up at the welcome desk") in place of a sign-up button, or the page will read as broken.

## Advance desk registration and door walk-ins stay independently switchable

`channels` governs the lobby table (in advance); `attendance_modes` governs whether staff can create a registration at the door during check-in (the `manual`/`personal_qr` walk-in path). We deliberately kept these separate rather than folding them into one "staff may register people" switch, because both asymmetric combinations are useful:

- **Door open, table closed** — a members-only event where staff should not be quietly adding people through the preceding month, but an usher should still admit a genuine member who forgot to book.
- **Table open, door closed** — a catered dinner needing final numbers a week ahead: staff add people in advance so the caterer knows, but nobody is added on the night because there is no food for them.

**Admin-UI consequence:** both settings sound like "can staff register people," so they must be labelled by *when* — "in advance" versus "at the door" — or they will be confused for each other.

## Desk registration gets its own geo mode

`geo_config.modes` gains a row for desk/lobby-table registration, with the same `off | warn | require` shape and `staff_override` flag as the existing modes, **defaulting to `off`**.

Hardcoding it off was proposed and rejected: the organizer wants this configurable per event, and anticipates events where registration must happen onsite. But the default matters, because requiring location here breaks the common case — staff signing someone up in the church lobby in November for a December event at a rented hall across town are legitimately nowhere near the event venue, and the phone being checked belongs to a trusted staff member rather than the attendee whose presence the check exists to verify.

**`geo_config` stays event-level.** Per-session geo was considered (a volunteer meeting held at someone's house one month in twelve) and deferred: the config carries the venue's coordinates, so moving it per-session really means "this session is somewhere else," which touches how sessions and venues relate throughout the system. That deserves to be built deliberately rather than as a side effect of adding a geo row.
