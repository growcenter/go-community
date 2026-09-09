# Spec: `notifications`

**Module id:** `notifications` · **Depends on:** `registration`, `qr` · **Build order:** 5th (parallel with `attendance`)
**Parent:** [SPEC.md](SPEC.md)

---

## Objective

A confirmation email that can never be lost, delivered without a resident worker, on a free tier. The outbox row is written **inside the registration transaction**, so a process dying immediately after commit loses nothing; a scheduler tick and an opportunistic post-request call both drain it.

## Scope

**In:** `notification_outbox` dispatch with retry and backoff; the `Notifier` interface and its SMTP implementation; PDF ticket generation with one QR per attendee.

**Out:** Enqueueing. `registration` and `event-core` enqueue inside their own transactions; this module only drains.

---

## Contracts

```go
// internal/pkg/notify
type Attachment struct{ Filename, Mime string; Data []byte }
type Notifier interface {
	Send(ctx context.Context, recipient, subject, htmlBody string, att []Attachment) error
}
func NewSMTPNotifier(cfg config.EmailConfig) Notifier   // net/smtp + MIME multipart, no new dep

// internal/pkg/pdfticket
type AttendeeQR struct{ Name string; PNG []byte }
type Input struct {
	EventTitle, SessionTitle, When, Where string
	Contacts  []v3.Contact
	Attendees []AttendeeQR
}
func Generate(in Input) ([]byte, error)

// internal/usecases/v3
type NotificationUsecase struct {
	r         *pgsql.PostgreRepositories
	notifiers map[string]notify.Notifier   // keyed by channel — injectable, so tests use a fake
	codec     *qr.Codec
}
func (u *NotificationUsecase) Dispatch(ctx context.Context, batchSize int) (sent, failed int, err error)
```

---

## Acceptance Criteria

### Durability

1. The outbox row is written **in the enqueueing transaction** (`registration`, `event-core`), never after commit. A confirmation cannot be lost even if the process dies the instant after COMMIT.
2. Dispatch is triggered twice, redundantly: opportunistically right after the request (fast path, best-effort, failure ignored) and by Cloud Scheduler hitting the API-key-protected internal endpoint every 5 minutes (safety net).
3. **No resident background worker.** The free tier does not permit one, and nothing here may assume one exists.

### Dispatch

4. `ClaimBatch` uses `FOR UPDATE SKIP LOCKED`, so concurrent dispatchers never double-send. A test runs two concurrent `Dispatch` calls against a batch and asserts each message is sent exactly once.
5. Templates handled: `registration_confirmed` (loads registration + attendees + session + event, renders HTML, generates the PDF with a per-attendee QR URL from the codec, attaches it), `registration_cancelled` and `event_cancelled` (plain HTML, no attachment).
6. Failure marks the message `pending` with `attempts+1` and an exponential backoff `scheduled_at`. After 5 attempts the status becomes `failed` and is visible to admins. Two tests: after one failure the message is `pending` with `attempts=1` and a future `scheduled_at`; after five it is `failed`.
7. **A message with no resolvable recipient is marked sent-with-skip, never retried.** It must never loop. This is the counterpart to `registration` criterion 18 — the desk-registration case where no email was given — and belongs here as a second line of defence, since the first is not to enqueue at all.
8. `Notifier` is a map keyed by channel, injected into the constructor, so tests substitute an in-memory fake recording its calls. **The fake exists only in tests**; nothing in production selects it.

### Email

9. One channel at launch: email, via `net/smtp` and MIME multipart. Provider host, port, credentials and from-address come from the `v3.email` config block.
10. WhatsApp and SMS later mean **a new `Notifier` implementation and a new `channel` value — zero schema change.** The interface must not acquire email-specific parameters.
11. Notifications are sent on: registration confirmed, registration cancelled, event cancelled (to every confirmed registrant).

### PDF ticket

12. Generated in Go at send time, and on demand at `GET /v3/registrations/{id}/ticket.pdf` for the owner or an actor holding `event.manage`. **Never stored.**
13. Contents: event title and image, session date/time/location, party summary, **one QR per attendee** (universal `attendee` tokens), the event's contact persons, and short how-to-use text.
14. Every attendee gets their own row and their own QR, even inside one party, so individual door verification always works.
15. A unit test asserts `Generate` with two attendees returns bytes beginning `%PDF` and longer than 1 KB, with QR PNGs produced by `qr.RenderPNG`.

---

## Verification

```bash
go test ./internal/pkg/pdfticket/ ./internal/pkg/notify/ -v     # pure; fake SMTP
TEST_DATABASE_URL=... go test ./tests/integration/v3/ -run TestOutbox -v
```

The outbox tests use the injected fake notifier — a real SMTP server is never contacted from a test.

---

## Boundaries (module-specific)

- **Always** enqueue inside the caller's transaction; **never** after commit.
- **Always** claim with `FOR UPDATE SKIP LOCKED`.
- **Never** store a generated PDF.
- **Never** create an outbox row that has no chance of delivery (see `registration` criterion 18).
- **Never** introduce a resident worker, a goroutine pool, or a ticker.
- **Never** commit SMTP credentials — they live in the gitignored `config.local.yaml`; only the template is committed.
- **Ask first** before adding a notification template or a channel.
