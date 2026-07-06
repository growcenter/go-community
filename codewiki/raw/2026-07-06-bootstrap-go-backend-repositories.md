# Bootstrap: Go Backend — Repositories

_Surveyed 2026-07-06 at commit a0da240._

## Overview

`internal/repositories/pgsql/` implements the data-access layer with GORM over Postgres. Each domain has an interface + struct pair (e.g. `UserRepository`/`userRepository`) constructed via `New*Repository(db *gorm.DB, ...)`. Simple CRUD goes through GORM's query builder; anything non-trivial (joins, aggregates, cursor pagination) drops to raw SQL kept in sibling `*_pg_query.go` files as package-level string vars, with separate `Build*` functions assembling dynamic WHERE clauses.

## Key structures / flows

- [internal/repositories/pgsql/user_pg_repository.go:12-39](internal/repositories/pgsql/user_pg_repository.go#L12-L39) — `UserRepository` interface lists ~24 methods; every method wraps its body in `defer LogRepository(ctx, err)` for centralized query logging.
- [internal/repositories/pgsql/user_pg_repository.go:207-271](internal/repositories/pgsql/user_pg_repository.go#L207-L271) — `GetAllWithCursor` implements cursor-based pagination: fetches `limit+1` rows to detect `hasMore`, reverses the slice for backward pagination, then encrypts `(CreatedAt, ID)` into opaque cursor tokens via `internal/pkg/cursor`.
- [internal/repositories/pgsql/user_pg_query.go:244-309](internal/repositories/pgsql/user_pg_query.go#L244-L309) — `BuildQueryGetAllUser` dynamically appends filter/search/cursor/order/limit clauses to a base query string using `strings.Builder`; the cursor comparison operator (`<` vs `>`) flips based on pagination direction.
- [internal/repositories/pgsql/user_pg_query.go:34-51](internal/repositories/pgsql/user_pg_query.go#L34-L51) — `queryGetRBACByCommunityId` computes combined roles by unnesting `user_types` and left-joining `user_types.roles`, showing RBAC roles are derived at query time from a user-type-to-roles mapping table, not stored directly on the user beyond `user_types`.
- [internal/repositories/pgsql/transaction_pg_repository.go:12-50](internal/repositories/pgsql/transaction_pg_repository.go#L12-L50) — `TransactionRepository.Transaction` begins a GORM tx (`dtx`), recovers panics with rollback, and commits/rolls back based on the callback's error.
- [internal/repositories/pgsql/transaction_pg_repository.go:52-85](internal/repositories/pgsql/transaction_pg_repository.go#L52-L85) — `Atomic` is a second, newer transaction helper that constructs a fresh `*PostgreRepositories` bound to the tx via `New(tx)` and passes it into the callback, so callers get transaction-scoped repositories instead of a raw `*gorm.DB`.

## Open questions

- [internal/repositories/pgsql/user_pg_repository.go:55-57](internal/repositories/pgsql/user_pg_repository.go#L55-L57) — `Create` (and `Update`, `UpdateByEmailPhoneNumber`) call `ur.trx.Transaction(...)` but the closure operates on `ur.db`, not the `dtx` passed into the callback — the transaction wrapper appears to not actually scope these writes to the opened transaction. Worth confirming whether this is intentional (e.g. GORM session sharing) or a bug, especially compared to the newer `Atomic` pattern which does thread the tx-bound repository through.
- Two competing transaction abstractions (`Transaction` vs `Atomic`) coexist — unclear which new code should prefer, or whether `Transaction` is legacy en route to removal.
