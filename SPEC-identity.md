# Spec: identity

Module id `identity` in `CAPABILITY-MAP.md`. Intent: `docs/intent/v3-platform.md`. Depends on `platform`. Re-targets `docs/plans/2026-07-07-user-management-design.md` from `/api/v2` to `/api/v3` and removes everything that now belongs to `permissions`.

## Objective

Give every member one account with a stable community ID. Sign-in needs only a phone or email plus a 6-digit PIN. Admins get a surface to manage accounts. Every other module identifies a person by `community_id` and nothing else.

What this module delivers:

1. Users, carried over from the existing `users` table with the same primary key and `community_id`.
2. Sign-in with OTP, PIN and trusted devices. No passwords for new accounts.
3. JWT access and refresh tokens bound to a device.
4. Admin create, edit, status change, soft delete, dashboard and audit log.
5. A "create on behalf of" path. A leader or core team member makes an account for a guest, and the guest claims it with OTP.
6. A claim path for pre-seeded accounts, replacing the shell-record upgrade inside `userUsecase.Create`.

Users are members (children to seniors, on the web app now and a mobile app later), desk staff and admins. Success: a first-time member signs in with only a phone number, a code and a PIN. An admin never touches `user_types`.

## Tech Stack

- Go 1.26, Echo v4, GORM over PostgreSQL (existing)
- `golang-jwt/jwt` (existing), bcrypt (existing) for PIN and OTP hashes
- `github.com/jeremygprawira/herr` through `internal/pkg/herrx` (from `platform`)
- Platform outbox for OTP delivery, `channel = email` today

## Commands

```bash
go build ./...                                                    # compile
go test ./internal/pkg/otp/ ./internal/usecases/identity/...      # unit tests, no DB
TEST_DATABASE_URL="postgres://postgres:$DB_PASSWORD@localhost:5888/community_db?sslmode=disable" \
  go test ./tests/integration/identity/ -v                        # DB tests, skipped when the URL is unset
make migration name=identity_v3                                   # new migration pair, numbered after 000022
make generate-docs                                                # swagger, after any handler change
```

## Project Structure

```
internal/models/identity/            → entities, DTOs, errors.go (USER_ codes), locale.go (id messages)
internal/pkg/otp/                    → challenge create and verify, pure, clock injected
internal/repositories/pgsql/identity → users (v3 columns), devices, challenges, audit, departments, kom
internal/usecases/identity/          → auth, admin users, audit, dashboard
internal/deliveries/http/v3/identity → users_handler.go, auth_handler.go, me_handler.go, admin_users_handler.go
tests/integration/identity/          → register, login, device, claim and audit flows
tests/integration/db/migrations/000023_identity_v3.up.sql
```

## Code Style

### Entities

Only columns that are new or change. `users` keeps its primary key and `community_id`. `community_id` stays the existing Luhn account number from `generator.LuhnAccountNumber()` and is never regenerated, so `qr` can derive a personal QR from it forever.

| Table | Change |
|---|---|
| `users` | add `pin_hash`, `sub_district`, `district`, `volunteer_status` (`active`, `sabat`, null), `kkj_file_url`, `baptism_file_url`, `deactivated_at`, `deactivated_by`, `claimed_at`. `user_types` and `roles` stay but v3 never reads or writes them. |
| `user_departments` | new: `community_id`, `department_code`, `is_primary`. No `position` column. A title in a department is a grant in `permissions`. |
| `kom_certifications` | new: `community_id`, `level`, `kom_id`, `file_url`, `completed_at`. KOM level of a user is the max level. |
| `user_devices` | new: `id` uuid, `community_id`, `device_id` unique, `platform_name`, `last_seen_at`, `revoked_at`. |
| `auth_challenges` | new: `id` uuid, `community_id`, `purpose`, `channel` (`email`, `sms`, `whatsapp`), `code_hash`, `attempts`, `expires_at`, `used_at`. Purged by `platform.retention_days.auth_challenges`. |
| `user_audit_logs` | new: `id` uuid, `actor`, `target_user`, `action`, `changes` jsonb, `created_at`. Never purged. |

Account status is one of `pending`, `active`, `deactivated`. `pending` means the account exists but nobody has proven ownership with OTP yet. Backfill: existing rows with a password hash become `active`. Existing rows that today match the shell-upgrade branch (a `user_types` category of `internal` or `cool` and no password) become `pending`.

### Sign-in flows

The UI asks for the identifier first. `users/check` tells the client the next step. If no account exists or the account is `pending`, the step is `register`. If the account is `active`, the step is `login`. `users/check` sends nothing and writes nothing.

Register: identifier (phone or email), OTP to that channel, set PIN, device trusted, signed in. Login on a trusted device: identifier plus PIN. Login on an unknown device: OTP step-up, then device trusted. If the user is at `identity.max_active_devices`, the client must confirm before the oldest device is revoked. Forgot PIN: OTP then new PIN. Lost the channel too: admin-assisted one-time code, shown once, audit-logged.

Claim: if `auth/otp` with purpose `register` finds a `pending` account with that identifier, the flow continues on that row instead of creating one. On `POST /v3/users` the row becomes `active` and `claimed_at` is set. This replaces the shell-record upgrade in `userUsecase.Create`. If the identifier belongs to an `active` account, `auth/otp` returns `USER_ALREADY_EXISTS` and sends nothing.

Account existence is not hidden. Decision 2026-10-01: the identifier-first UI needs `users/check`, so the earlier rule that nobody can probe for accounts is dropped. Send and verification endpoints return the real error. `users/check` is limited per client IP.

Legacy password login stays only for rows that already have a password hash, behind `identity.password_login_enabled`. New accounts never get a password.

OTP switch: `identity.otp_enabled` (default `true`) turns OTP off for the whole system. When it is off, `auth/otp` with purpose `register` returns the verification token at once and sends nothing. `login` from an unknown device with the right PIN trusts that device at once. `auth/otp` with purpose `pin_reset` returns `USER_OTP_DISABLED`. `auth/otp/verification` returns `USER_OTP_DISABLED` for every code except an admin-assisted code, so forgot PIN works only through the admin-assisted code. Every endpoint keeps the same request and response shape in both modes, so the web app only changes which screen it shows. The `verification_token` field in the `auth/otp` response is empty while OTP is on. Flipping the switch writes one `user_audit_logs` row with action `otp_switch` and the actor. Turning OTP off means a stolen PIN is enough to sign in on any phone. The switch is for outages and email quota, not for daily use.

### Endpoints

All under `/api/v3`. "Member" means any signed-in user. "Admin" means a user who holds the matching grant in `permissions`. The evaluator lives in `permissions`. Identity only names the action.

A route is scoped by the thing it acts on. `/v3/users` looks up and creates accounts. `/v3/auth` proves who the caller is: codes, PIN and tokens. Decision 2026-10-01.

```
Users, public
POST /v3/users/check              {identifier}                              {next, method}, sends nothing
POST /v3/users                    {verification_token, pin, name, campus_code, device}   account created, tokens

Auth, public
POST /v3/auth/otp                 {identifier, purpose}                     code sent, purpose is register or pin_reset
POST /v3/auth/otp/verification    {identifier, otp, purpose}                short-lived verification token, purpose is register, pin_reset or device
POST /v3/auth/login               {identifier, pin or password, device_id, verification_token?, replace_oldest?}   tokens, or USER_NEW_DEVICE_VERIFICATION_REQUIRED
POST /v3/auth/pin/reset           {verification_token, new_pin}             200
POST /v3/auth/refresh             refresh token in X-Refresh-Token          new token pair
POST /v3/auth/logout              access token                              revokes this device

Me, member
GET    /v3/me                     own profile with departments[] and kom
PATCH  /v3/me                     own editable profile fields
GET    /v3/me/devices             trusted devices
DELETE /v3/me/devices/{id}        sign out one device

Admin
GET    /v3/admin/users                         list, cursor pagination, fields per view grant
GET    /v3/admin/users/{communityId}           detail
POST   /v3/admin/users                         create a pending account on behalf of a person
PATCH  /v3/admin/users/{communityId}           edit profile, departments[], kom, files
DELETE /v3/admin/users/{communityId}           soft delete
PATCH  /v3/admin/users/{communityId}/status    account status or volunteer status
POST   /v3/admin/users/{communityId}/reset     admin-assisted one-time code
GET    /v3/admin/users/dashboard               demographic aggregates
GET    /v3/admin/audit-logs                    user audit log, filter by actor, target, action, date
```

`POST /v3/users/check` answers with the next step for the UI:

| Account | `next` | `method` |
|---|---|---|
| none, or `pending` | `register` | empty |
| `active` with a PIN | `login` | `pin` |
| `active` with a password hash and no PIN, while `identity.password_login_enabled` is `true` | `login` | `password` |
| `deactivated` | 403 `USER_ACCOUNT_DEACTIVATED` | |

`POST /v3/auth/otp` sends one code and is also the "send again" call. Purpose `register` sends for a new identifier and for a `pending` account, and returns `USER_ALREADY_EXISTS` for an `active` account. Purpose `pin_reset` sends for an `active` account, and returns `USER_NOT_FOUND` or `USER_ACCOUNT_PENDING` for the other cases. Purpose `device` is not accepted here. Only `auth/login` sends a `device` code, after a correct PIN, so a stranger cannot start a new-device code for a member. To send a `device` code again, the client calls `auth/login` again.

`POST /v3/auth/otp/verification` takes the code and returns a verification token. The token is short-lived and holds the identifier and the purpose. Each purpose has one consumer: `register` goes to `POST /v3/users`, `pin_reset` goes to `auth/pin/reset`, and `device` goes to `auth/login`.

New device: `auth/login` with the right PIN on an unknown device sends a code and returns `USER_NEW_DEVICE_VERIFICATION_REQUIRED`. The client verifies the code with purpose `device`, then calls `auth/login` again with the `verification_token`. If the device limit is reached, the client also sends `replace_oldest`.

`POST /v3/admin/users` is also the path a COOL leader or core team member uses to create an account for a guest. It needs `name` and one identifier, returns the new `community_id` and status `pending`, and sends no OTP. The person claims it later through the register flow (`auth/otp` with purpose `register`, then `POST /v3/users`). The callers (admin, COOL leader, core team) are defined by a grant in `permissions`.

### JWT

Header `kid` selects the secret from `config.Auth.BearerSecret` or `RefreshSecret`, as today. Claims:

```
sub   community_id
did   device_id
jti   random id
iat   issued at
exp   expiry, identity.access_ttl_minutes or identity.refresh_ttl_days
typ   "access" or "refresh"
```

No `userTypes` and no `roles` claim. The `permissions` module reads grants per request from the database. If the `did` row is revoked or missing, the refresh token is rejected. A deactivated account is rejected on every path and all its devices are revoked at deactivation time.

### Handler

```go
// internal/deliveries/http/v3/identity/auth_handler.go
func (h *AuthHandler) Login(c echo.Context) error {
	var req identity.LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.Fail(c, herrx.ErrInvalidInput.Wrap(err))
	}
	if err := h.validate.Struct(req); err != nil {
		return response.Fail(c, identity.FieldErrors(err))
	}
	out, err := h.auth.Login(c.Request().Context(), req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out, nil)
}
```

Rules:

- Every admin write to `users`, `user_devices` or `user_departments`, and every claim, writes one `user_audit_logs` row. The row is in the same `Atomic` transaction and holds a field-level diff.
- Routine logins are not audit-logged. PIN resets, device revokes, admin resets and status changes are.
- OTP: 6 digits, `identity.otp_expiry_minutes`, hashed at rest, `identity.otp_max_attempts` then cooldown, single use.
- PIN: bcrypt, progressive lockout by `identity.pin_lockout_attempts` and `identity.pin_lockout_minutes`. Never logged.
- `users/check` sends nothing and writes nothing. It allows `identity.check_max_per_minute` calls per client IP through Echo's in-memory rate limiter and returns `USER_TOO_MANY_REQUESTS` past the limit. The count is per app instance, so it slows a probe but is not exact.
- A verification token is accepted only by the consumer of its purpose. Any other use returns `USER_VERIFICATION_TOKEN_INVALID`.
- OTP rows go through the platform outbox. Today the only channel is `email`, so every account must have an email before any OTP flow. `auth/otp` with a phone identifier and no email returns `USER_EMAIL_REQUIRED`. A legacy account with no email can still log in with PIN or password on a trusted device. For that account, `GET /v3/me` returns `email_required: true`. Every other member endpoint returns `USER_EMAIL_REQUIRED` until `PATCH /v3/me` sets an email. A new device for such an account is handled only by the admin-assisted one-time code.
- OTP abuse guard, all keys in configdb. A person can ask for a new code only after `identity.otp_resend_seconds` (3 minutes). A person can ask for at most `identity.otp_max_sends` (3) codes inside `identity.otp_window_hours` (2 hours), counted per identifier and per client IP. A person can enter a wrong code at most `identity.otp_max_attempts` (3) times per challenge. When either limit is hit, the identifier enters a cooldown of `identity.otp_cooldown_hours` (2 hours). During cooldown every send and every verification returns `USER_OTP_COOLDOWN` and sends nothing. The start of a cooldown writes one `user_audit_logs` row with action `otp_cooldown`. The limits count every code, including a `device` code that `auth/login` sends. Codes are compared in constant time.
- Files are upload URLs only. The backend never stores bytes.

### Errors

All through `herrx.Def` with `USER_` prefix, EN message here, ID message in `locale.go`.

| Code | Kind | EN message |
|---|---|---|
| `USER_NOT_FOUND` | not found | We did not find that account. |
| `USER_ALREADY_EXISTS` | conflict | An account with that phone or email already exists. |
| `USER_INVALID_CREDENTIALS` | unauthorized | That PIN does not match. Please try again. |
| `USER_PIN_LOCKED` | forbidden | Too many attempts. Please wait a few minutes and try again. |
| `USER_OTP_INVALID` | unauthorized | That code is not right. Please check and try again. |
| `USER_OTP_EXPIRED` | unauthorized | That code expired. Please request a new one. |
| `USER_OTP_TOO_MANY_ATTEMPTS` | forbidden | Too many wrong codes. Please request a new one later. |
| `USER_OTP_COOLDOWN` | forbidden | Too many tries. Please wait two hours and try again. |
| `USER_OTP_DISABLED` | unprocessable | Code verification is turned off right now. Please ask your leader for help. |
| `USER_NEW_DEVICE_VERIFICATION_REQUIRED` | forbidden | We do not know this device yet. Please enter the code we sent. |
| `USER_DEVICE_LIMIT_REACHED` | conflict | You are signed in on another phone. Sign out there to continue here. |
| `USER_DEVICE_REVOKED` | unauthorized | This device was signed out. Please sign in again. |
| `USER_ACCOUNT_DEACTIVATED` | forbidden | This account is not active. Please contact the office. |
| `USER_ACCOUNT_PENDING` | forbidden | This account has not been claimed yet. Please register with your phone or email. |
| `USER_VERIFICATION_TOKEN_INVALID` | unauthorized | Your verification expired. Please request a new code. |
| `USER_TOO_MANY_REQUESTS` | forbidden | Too many tries. Please wait a minute and try again. |
| `USER_EMAIL_REQUIRED` | unprocessable | Please add your email address first so we can send you a code. |
| `USER_PASSWORD_LOGIN_DISABLED` | forbidden | Password sign-in is off. Please sign in with your PIN. |
| `USER_INVALID_INPUT` | invalid | Something in your submission does not look right. |

### Configdb keys

| Key | Default |
|---|---|
| `identity.max_active_devices` | `1` |
| `identity.otp_enabled` | `true` |
| `identity.otp_expiry_minutes` | `5` |
| `identity.otp_max_attempts` | `3` |
| `identity.pin_lockout_attempts` | `5` |
| `identity.pin_lockout_minutes` | `15` |
| `identity.otp_resend_seconds` | `180` |
| `identity.otp_max_sends` | `3` |
| `identity.otp_window_hours` | `2` |
| `identity.otp_cooldown_hours` | `2` |
| `identity.check_max_per_minute` | `60` |
| `identity.password_login_enabled` | `true` |
| `identity.access_ttl_minutes` | `60` |
| `identity.refresh_ttl_days` | `30` |
| `identity.max_departments_per_user` | `0` (unlimited) |

## Testing Strategy

Unit, no DB:

- `internal/pkg/otp`: expiry, three wrong attempts then cooldown, three sends in two hours then cooldown, three-minute resend gap, single use, with an injected clock.
- PIN lockout counter and cooldown math.
- Audit diff: only changed fields appear, secrets (`pin_hash`) never appear.
- Catalog completeness for every `USER_` class (shared test from `platform`).

Integration, real Postgres via `TEST_DATABASE_URL`:

- Register, login, new device, revoke, refresh rejected after revoke.
- Claim: admin creates pending account, person claims it, status becomes `active`, one audit row each.
- Two devices racing the last device slot: exactly one wins, user row locked in `Atomic`.
- `users/check` returns `register` for an unknown and for a `pending` identifier. It returns `login` with `pin` for an active one and `login` with `password` for a legacy row. It inserts no row in any table.
- `auth/otp` with purpose `register` on an active account returns 409 `USER_ALREADY_EXISTS` and inserts no outbox row. `auth/otp` with purpose `device` is rejected.
- A verification token with purpose `register` is rejected by `auth/pin/reset` and by `auth/login`.
- The fourth `auth/otp` for one email inside two hours returns 403 `USER_OTP_COOLDOWN`, inserts no outbox row, and writes one `otp_cooldown` audit row. An `auth/otp` 2 minutes after the previous one sends nothing. The third wrong code on one challenge starts the same cooldown.
- A legacy account with no email logs in on its trusted device and gets `USER_EMAIL_REQUIRED` on `GET /v3/me/devices`. It sets an email through `PATCH /v3/me`, and the next call succeeds.
- Deactivation revokes all devices and blocks login and refresh.
- With `identity.otp_enabled = false`, register completes with no challenge row and no outbox row, and a new device is trusted on the first correct PIN. Flipping the key writes one `otp_switch` audit row.
- Backfill migration: password users become `active`, shell users become `pending`, `community_id` unchanged for every row.

No database mocks. Every rule in Code Style has one test.

## Boundaries

Always:

- Keep `community_id` immutable. No endpoint changes it.
- Write the audit row in the same transaction as the change.
- Hash OTP codes and PINs before they touch the database.
- Use `Atomic`, never `Transaction`.

Ask first:

- Adding a claim to the JWT.
- Adding a third account status.
- Changing any `identity.*` default.
- Any new column on `users`.

Never:

- Read or write `user_types`, `roles` or the mapping table from v3 code.
- Store passwords for new accounts.
- Log a PIN, OTP, verification token or refresh token.
- Send an email or write a row from `users/check`.
- Send a `device` code from `auth/otp`.
- Accept a verification token for a purpose other than its own.
- Trust a new device without an OTP or an admin-assisted code while `identity.otp_enabled` is `true`.
- Purge `user_audit_logs`.

## Success Criteria

1. A new member with only a phone number and email completes `users/check`, `auth/otp`, `auth/otp/verification` and `POST /v3/users`. The member receives a token pair, and the database holds one `auth_challenges` row and one outbox row.
2. Login with the right PIN on the trusted device returns tokens. Login from a new `device_id` returns 403 `USER_NEW_DEVICE_VERIFICATION_REQUIRED`.
3. With `identity.max_active_devices = 1`, `auth/login` on a second device with a `device` verification token and without `replace_oldest` returns 409 `USER_DEVICE_LIMIT_REACHED`. With it, the first device's refresh token is rejected with `USER_DEVICE_REVOKED`.
4. A COOL leader creates an account for a guest. The response has status `pending`. The guest later registers with the same phone and ends up `active` with the same `community_id`.
5. After migration, every pre-existing `community_id` is unchanged and every legacy password user can still log in while `identity.password_login_enabled` is `true`.
6. The access token decodes to `sub`, `did`, `jti`, `iat`, `exp`, `typ` and nothing else.
7. Deactivating a user returns 403 `USER_ACCOUNT_DEACTIVATED` on the next login and refresh.
8. `GET /v3/admin/audit-logs` shows one row per admin mutation with a field-level diff and no `pin_hash`.
9. Every `USER_` code has an ID message, or `go test` fails.
10. `POST /v3/users/check` returns `register` for a new email and `login` for an active one, and no `auth_challenges` or outbox row appears.

## Open Questions

1. Which existing rows become `pending` at backfill. The rule above uses the same test as the old shell-upgrade branch. Make sure that count against production data before the migration runs.
2. `user_departments.position` from the design is dropped here because a department title is a grant. `permissions` must define a `department` scoped title for leader and core team so the data is not lost. Confirm the backfill target with `permissions`.
3. The old `/v2/users/*` and `/v2/auth/*` endpoints return 410 from day one of v3. The web app must switch sign-in in the same release. Confirm the cut-over order with the frontend.
4. Legacy `forgot-password` and `reset-password` are not in v3. Legacy password users who forget their password use `auth/otp` with purpose `pin_reset` to set a PIN instead. Confirm this is acceptable.
5. Resolved 2026-09-17: no device is ever trusted without proof. Email is required for OTP, and accounts without an email must add one before using member endpoints.
6. Resolved 2026-10-01: routes are scoped by what they act on (`/v3/users` for accounts, `/v3/auth` for codes, PIN and tokens). `users/check` is added for the identifier-first UI. The no-probe rule is dropped. Anyone can test an identifier for membership, limited per IP.
7. No total limit on OTP emails exists. The limits are per identifier and per IP, so 2,000 new members on an event day still means 2,000 emails. Decide a system-wide cap against the email provider quota.
8. The OTP send limit counts per client IP. Members on the same church Wi-Fi share one IP. A low IP count blocks registration on an event day. Decide a separate, higher limit for the IP count. The same question applies to `identity.check_max_per_minute`.
9. No endpoint changes `identity.otp_enabled`. `configdb` has readers only, so the `otp_switch` audit row has no actor. Decide who flips the switch and through which route.
10. With OTP off, `auth/otp` gives the verification token with no proof. Anyone who knows the identifier of a `pending` account can claim it. Decide whether claim stays blocked while OTP is off.
