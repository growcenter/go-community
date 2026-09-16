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
internal/deliveries/http/v3/identity → auth_handler.go, me_handler.go, admin_users_handler.go
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

Register: identifier (phone or email), OTP to that channel, set PIN, device trusted, signed in. Login on a trusted device: identifier plus PIN. Login on an unknown device: OTP step-up, then device trusted. If the user is at `identity.max_active_devices`, the client must confirm before the oldest device is revoked. Forgot PIN: OTP then new PIN. Lost the channel too: admin-assisted one-time code, shown once, audit-logged.

Claim: if `register/start` finds a `pending` account with that identifier, the flow continues on that row instead of creating one. On `register/complete` the row becomes `active` and `claimed_at` is set. This replaces the shell-record upgrade in `userUsecase.Create`. If the identifier belongs to an `active` account, `register/start` still returns 200 and sends nothing, so nobody can probe for accounts.

Legacy password login stays only for rows that already have a password hash, behind `identity.password_login_enabled`. New accounts never get a password.

OTP switch: `identity.otp_enabled` (default `true`) turns OTP off for the whole system. When it is off, `register/start` returns the setup token at once and sends nothing, `login` from an unknown device with the right PIN trusts that device at once, `device/verify` and `pin/reset` return `USER_OTP_DISABLED`, and forgot PIN works only through the admin-assisted code. Every endpoint keeps the same request and response shape in both modes, so the web app only changes which screen it shows. Flipping the switch writes one `user_audit_logs` row with action `otp_switch` and the actor. Turning OTP off means a stolen PIN is enough to sign in on any phone, so the switch is for outages and email quota, not for daily use.

### Endpoints

All under `/api/v3`. "Member" means any signed-in user. "Admin" means a user who holds the matching grant in `permissions`. The evaluator lives in `permissions`. Identity only names the action.

```
Auth, public
POST /v3/auth/register/start      {identifier}                              OTP sent, or claim continues, 200 always
POST /v3/auth/register/verify     {identifier, otp}                         short-lived setup token
POST /v3/auth/register/complete   {setup_token, pin, name, campus_code, device}   tokens
POST /v3/auth/login               {identifier, pin, device_id}              tokens, or USER_NEW_DEVICE_VERIFICATION_REQUIRED
POST /v3/auth/device/verify       {identifier, otp, device_id, replace_oldest}   tokens
POST /v3/auth/pin/forgot          {identifier}                              200 always
POST /v3/auth/pin/reset           {identifier, otp, new_pin}                200
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

`POST /v3/admin/users` is also the path a COOL leader or core team member uses to create an account for a guest. It needs `name` and one identifier, returns the new `community_id` and status `pending`, and sends no OTP. The person claims it later through `register/start`. The callers (admin, COOL leader, core team) are defined by a grant in `permissions`.

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
- Forgot and start endpoints return 200 on the same code path whether or not the account exists.
- OTP rows go through the platform outbox. Today the only channel is `email`, so every account must have an email before any OTP flow. `register/start` with a phone identifier and no email returns `USER_EMAIL_REQUIRED`. A legacy account with no email can still log in with PIN or password on a trusted device, but `GET /v3/me` returns `email_required: true` and every other member endpoint returns `USER_EMAIL_REQUIRED` until `PATCH /v3/me` sets an email. A new device for such an account is handled only by the admin-assisted one-time code.
- OTP abuse guard, all keys in configdb. A person can ask for a new code only after `identity.otp_resend_seconds` (3 minutes). A person can ask for at most `identity.otp_max_sends` (3) codes inside `identity.otp_window_hours` (2 hours), counted per identifier and per client IP. A person can enter a wrong code at most `identity.otp_max_attempts` (3) times per challenge. When either limit is hit, the identifier enters a cooldown of `identity.otp_cooldown_hours` (2 hours). During cooldown every send returns 200 and sends nothing, every verify returns `USER_OTP_COOLDOWN`, and one `user_audit_logs` row with action `otp_cooldown` is written when the cooldown starts. Codes are compared in constant time. The response time of `register/start` and `pin/forgot` does not depend on whether the account exists or the limit was hit.
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
| `USER_SETUP_TOKEN_INVALID` | unauthorized | Your registration session expired. Please start again. |
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
- `register/start` on an active account returns 200 and inserts no outbox row.
- The fourth `register/start` for one email inside two hours returns 200, inserts no outbox row, and writes one `otp_cooldown` audit row. A `register/start` 2 minutes after the previous one sends nothing. The third wrong code on one challenge starts the same cooldown.
- A legacy account with no email logs in on its trusted device, gets `USER_EMAIL_REQUIRED` on `GET /v3/me/devices`, sets an email through `PATCH /v3/me`, and the next call succeeds.
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
- Log a PIN, OTP, setup token or refresh token.
- Return a different status or timing from `register/start` or `pin/forgot` based on account existence or rate limit state.
- Trust a new device without an OTP or an admin-assisted code while `identity.otp_enabled` is `true`.
- Purge `user_audit_logs`.

## Success Criteria

1. A new member with only a phone number and email completes register start, verify, complete and receives a token pair, with one `auth_challenges` row and one outbox row.
2. Login with the right PIN on the trusted device returns tokens. Login from a new `device_id` returns 403 `USER_NEW_DEVICE_VERIFICATION_REQUIRED`.
3. With `identity.max_active_devices = 1`, verifying a second device without `replace_oldest` returns 409 `USER_DEVICE_LIMIT_REACHED`. With it, the first device's refresh token is rejected with `USER_DEVICE_REVOKED`.
4. A COOL leader creates an account for a guest. The response has status `pending`. The guest later registers with the same phone and ends up `active` with the same `community_id`.
5. After migration, every pre-existing `community_id` is unchanged and every legacy password user can still log in while `identity.password_login_enabled` is `true`.
6. The access token decodes to `sub`, `did`, `jti`, `iat`, `exp`, `typ` and nothing else.
7. Deactivating a user returns 403 `USER_ACCOUNT_DEACTIVATED` on the next login and refresh.
8. `GET /v3/admin/audit-logs` shows one row per admin mutation with a field-level diff and no `pin_hash`.
9. Every `USER_` code has an ID message, or `go test` fails.

## Open Questions

1. Which existing rows become `pending` at backfill. The rule above uses the same test as the old shell-upgrade branch. Make sure that count against production data before the migration runs.
2. `user_departments.position` from the design is dropped here because a department title is a grant. `permissions` must define a `department` scoped title for leader and core team so the data is not lost. Confirm the backfill target with `permissions`.
3. The old `/v2/users/*` and `/v2/auth/*` endpoints return 410 from day one of v3. The web app must switch sign-in in the same release. Confirm the cut-over order with the frontend.
4. Legacy `forgot-password` and `reset-password` are not in v3. Legacy password users who forget their password use `pin/forgot` to set a PIN instead. Confirm this is acceptable.
5. Resolved 2026-09-17: no device is ever trusted without proof. Email is required for OTP, and accounts without an email must add one before using member endpoints.
