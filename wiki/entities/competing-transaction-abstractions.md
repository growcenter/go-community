---
tags: [tech-debt]
sources: [raw/2026-07-06-bootstrap-go-backend-repositories.md]
updated: 2026-07-06
ingested_at: 91a596d
---

# Competing Transaction Abstractions

Two transaction helpers coexist in the [Repository Layer](repository-layer.md)'s `TransactionRepository`:

- **`Transaction(fc func(dtx *gorm.DB) error) error`** — the older helper. Begins a GORM tx (`dtx`), recovers panics with rollback, and commits/rolls back based on the callback's returned error. [internal/repositories/pgsql/transaction_pg_repository.go:12-50](internal/repositories/pgsql/transaction_pg_repository.go#L12-L50)
- **`Atomic(ctx, fc func(ctx, *PostgreRepositories) error) error`** — the newer helper. Begins a tx, then constructs a fresh `*PostgreRepositories` bound to that tx via `New(tx)` and passes the tx-scoped repository set into the callback, so all repository calls inside `fc` are automatically transaction-scoped. [internal/repositories/pgsql/transaction_pg_repository.go:52-85](internal/repositories/pgsql/transaction_pg_repository.go#L52-L85)

## Likely transaction-scoping issue in `Transaction` callers

`userRepository.Create`, `Update`, and `UpdateByEmailPhoneNumber` all call `ur.trx.Transaction(func(dtx *gorm.DB) error { ... })`, but the closure body operates on `ur.db` (the repository's original, non-transactional connection) rather than the `dtx` parameter the callback receives. [internal/repositories/pgsql/user_pg_repository.go:55-57](internal/repositories/pgsql/user_pg_repository.go#L55-L57)

This means these writes likely execute outside the transaction that `Transaction` opens — the `Begin`/`Commit`/`Rollback` bracketing runs, but the actual `Create`/`Save`/`Updates` calls go through a separate connection/session. Whether this is a real bug (writes aren't atomic with whatever else runs in the same `dtx`) or a no-op in practice (e.g. because these calls are never combined with other statements in the same transaction) is unconfirmed — worth verifying before relying on `Transaction`-wrapped multi-statement atomicity anywhere that follows this pattern.

No apparent guidance on which of `Transaction` or `Atomic` new code should use; `Atomic` appears to be the more correct pattern given it actually threads the tx-bound repository through.
