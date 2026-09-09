# Spec: `qr`

**Module id:** `qr` · **Depends on:** `foundation` · **Build order:** 2nd (parallel with `policy-kernel`)
**Parent:** [SPEC.md](SPEC.md)

---

## Objective

QR is a bounded domain of its own, not an event feature. One URL shape for every QR the system will ever emit, so that a code printed on a poster or saved in someone's phone in 2026 still works when COOL joining ships in 2027. Event check-in is its first consumer; future consumers register handlers without touching the core, and **the format, the endpoint and the scanner never change**.

## Scope

**In:** `internal/pkg/qr` — token codec, HMAC signing and verification, expiry, PNG rendering, the action-handler registry. Token type registry.

**Out:** What a check-in *does*. `attendance` owns that and registers a handler here.

---

## Contracts

```go
type Codec struct{ /* HMAC keys by kid, app domain */ }
func NewCodec(secrets map[string]string, appDomain string) *Codec

type Payload struct {
	T   string     // "personal" | "attendee" | "session"
	R   string     // reference
	Iat int64
	Exp *int64     // nil = never expires
}

func (c *Codec) Encode(p Payload) (token string, err error)
func (c *Codec) URL(token string) string          // https://<app-domain>/q/<token>
func (c *Codec) Decode(token string) (Payload, error)   // verifies signature AND expiry
func RenderPNG(url string, size int) ([]byte, error)

// Action registry — consumers register, they are never imported by the core.
type ActionCtx struct {
	Caller  eligibility.User
	Params  map[string]any
	Payload Payload
}
type Handler struct {
	Action  string
	Allowed func(ctx context.Context, a ActionCtx) bool
	Do      func(ctx context.Context, a ActionCtx) (any, error)
}
func (r *Registry) Register(tokenType string, h Handler)
func (r *Registry) AllowedActions(ctx context.Context, a ActionCtx) []string
func (r *Registry) Dispatch(ctx context.Context, action string, a ActionCtx) (any, error)
```

---

## Acceptance Criteria

### Format

1. Every QR the system emits is `https://<app-domain>/q/<token>`, where `token` is `base64url(payload)` plus an HMAC-SHA256 signature.
2. HMAC keys come from config as a per-`kid` map, following the existing auth-secret pattern, so keys can rotate without invalidating every printed code at once.
3. A tampered payload fails `Decode` with `INVALID_TOKEN`. A test mutates one byte of a valid token and asserts the failure.
4. An expired token fails with `TOKEN_EXPIRED`, distinctly from a forged one — a scanner needs to tell "this poster is old" from "this code is fake."

### Token types and lifetimes

5. `personal` → `R` is the community ID. **Never expires.** Generated once at account registration, permanent, non-rotatable — no self-service or staff-initiated regeneration exists or is planned (ADR 0003).
6. `attendee` → `R` is the attendee UUID. Expires after the session ends.
7. `session` → `R` is the session code (the venue poster). **Expiry follows `enable_checkout`** (ADR 0009):

   | `enable_checkout` | `session` token expires at |
   |---|---|
   | `false` (default) | `checkin_close_at` — unchanged |
   | `true` | `end_at` |

   The failure this fixes: a volunteer meeting whose check-in closes at 20:00 and ends at 21:00 has a full hour where checking out via the poster is impossible because the poster is dead. Tying the extension to the setting preserves what the short expiry is *for* — limiting how long a photographed poster stays useful — for the majority of sessions that have no use for it after the doors close. A test covers both branches at a timestamp between the two.

8. New token types are added to the registry. Adding one changes no format, no endpoint, and no scanner behaviour. A test registers a synthetic type and asserts resolve still works for the existing three.

### Registry and `allowed_actions`

9. `AllowedActions` is computed **per caller**. A staff member scanning a personal QR during a session they can check people into sees `event_checkin`; a member scanning a friend's QR sees nothing actionable. A test asserts the empty result for the unprivileged caller.
10. `Dispatch` on an action not in that caller's `AllowedActions` returns `FORBIDDEN` — the allowed list is an authorization boundary, not a UI hint.
11. `event_checkout` appears in `AllowedActions` only when the subject is already checked in **and** the session has `enable_checkout`. This is what makes check-out an explicit, offered action rather than a consequence of scanning twice (ADR 0004).
12. The core imports no consumer package. `attendance` registers its handlers; `qr` never imports `attendance`.

### Rendering

13. `RenderPNG` produces a scannable PNG at the requested size, used by both the PDF ticket and API responses. A test asserts a PNG magic-number prefix and non-trivial length.

### Explicitly not built

14. **No anti-replay protection on the session poster.** It stays a static, non-rotating code; the trust boundary is accepted for this audience. If abuse appears, `session_qr` is disabled for that session in favour of `personal_qr`/`registration_qr` (ADR 0003). Do not add nonces, rotation, or one-time codes.
15. **No personal QR rotation.** See criterion 5.

---

## Verification

```bash
go test ./internal/pkg/qr/ -v     # codec sign/verify/expiry/tamper, registry, PNG
```

Pure unit tests only — no database. The registry is tested with synthetic handlers; the real check-in wiring is verified in `http-api`'s end-to-end QR flow test.

---

## Boundaries (module-specific)

- **Never** import an event, registration, or attendance package from `internal/pkg/qr`. Dependencies point inward.
- **Never** change the URL shape or the payload field names. Printed codes and saved screenshots outlive deployments.
- **Never** log a full token — it is a bearer credential.
- **Ask first** before adding a token type, or before making `session` tokens rotate.
- **Always** verify signature *and* expiry in `Decode`. There is no "decode without verifying" variant, so no caller can accidentally skip it.
