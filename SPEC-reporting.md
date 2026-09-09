# Spec: `reporting`

**Module id:** `reporting` · **Depends on:** `registration`, `attendance` · **Build order:** 6th
**Parent:** [SPEC.md](SPEC.md)

---

## Objective

Answer the organizer's questions after the event: who registered, who came, who was late, how much of the total was people we know nothing about — and, newly, **where the sign-ups came from**, so the lobby table's contribution is visible rather than merged into an undifferentiated "registered" number that hides whether staffing it was worthwhile.

## Scope

**In:** CSV and XLSX generation for an event or a single session; the source split; conditional columns.

**Out:** Deriving the underlying facts. Lateness comes from `attendance`'s rule, the eligibility result is read from the stored log column, headcount netting is `attendance`'s `CountHeadcount`.

---

## Contracts

```go
type ReportUsecase struct{ r *pgsql.PostgreRepositories }

SessionReport(ctx, sessionCode, format string) (data []byte, filename, mime string, err error)
EventReport(ctx, eventCode, format string)   (data []byte, filename, mime string, err error)
```

---

## Acceptance Criteria

### Output

1. `format=csv` produces a flat attendee list for quick import elsewhere. `format=xlsx` produces a Summary sheet plus one detail sheet per session.
2. Both are **streamed in the response, never stored**. At ~1,000 users size is a non-issue.
3. Access requires `Can(actor, "event.manage", "event:<code>")` (ADR 0001). Superadmin bypass applies.

### Detail columns

4. Base columns: `attendee_name`, `community_id`, `category`, `status`, `registered_at`, `attended_at`, `checked_out_at`, `registered_by`, `source`.
5. **Custom form answers flatten into columns** — one column per form field key, in form order.
6. `category` appears only when the session has ticket categories.

### Conditional columns — unconfigured capabilities stay invisible

7. The rule this module inherits: an organizer who did not configure a capability must not see a column for it. Three follow the existing `enable_checkout` precedent:

   | Column | Appears only when |
   |---|---|
   | `checked_out_at`, presence duration | `enable_checkout` |
   | `late` | `enable_late_marking` (ADR 0005) |
   | `eligibility_at_checkin` | `eligibility.recheck_at_checkin != off` (ADR 0008) |

8. A test asserts the header row of a session with all three off contains **none** of those columns — the point being that a Christmas service must not shame late arrivals with a column nobody asked for.
9. `late` is derived at read time as `attended_at > start_at` — never read from a stored flag, because there is none.
10. `eligibility_at_checkin` is read from `attendance_logs.eligibility_result`, **stored at check-in time**, not recomputed. A report run in March asking "did she qualify?" must answer for the December night in question, not for March.

### Summary sheet

11. **Advance registrations split by `source` into three counts** — `web` (online self-registration), `admin` (staff-registered in advance at the lobby table), `walk_in` (registered at the door). This replaces the single undifferentiated "registered" total (ADR 0004). A test seeds one registration of each source and asserts three distinct counts.
12. Per session: the three source counts, attended, no-show, headcount, attendance %, geo violations.
13. **Headcount is reported separately from named attendees**, so an organizer knows how much of the total is people the system knows nothing about. It is the **net** of taps and undos via `CountHeadcount`.
14. Attendance % counts `outcome='checked_in'` only — duplicate scans logged as `already_checked_in` must not inflate it. A test scans a duplicate and asserts the percentage is unchanged.
15. A session with `attendance_modes: []` reports **registration data only** and no attendance figures — never "0 of 40 attended", which is the misleading statistic that motivated allowing empty modes at all.

---

## Verification

```bash
TEST_DATABASE_URL=... go test ./tests/integration/v3/ -run TestReport -v
```

Cases to see fail first: CSV header contains a custom `diet` column and the right row count for a party of 2 plus 1; XLSX opens via `excelize` and the Summary sheet carries three separate source counts; all three conditional columns absent when their settings are off and present when on; duplicate scan leaves attendance % unchanged; empty `attendance_modes` produces no attendance section.

---

## Boundaries (module-specific)

- **Always** derive lateness; **never** read it from a stored field.
- **Always** read the eligibility re-check result from the stored log column; **never** recompute it.
- **Never** show a column for a capability the event did not configure.
- **Never** store a generated report.
- **Never** count `already_checked_in` as attendance.
- **Ask first** before adding a Summary metric — the sheet is what organizers screenshot and circulate, so its shape is a de facto contract.
