---
tags: [architecture]
sources: [raw/2026-07-06-bootstrap-go-backend-repositories.md]
updated: 2026-07-06
ingested_at: 91a596d
---

# Repository Layer

`internal/repositories/pgsql/` is the data-access layer used by the [Usecase Layer](usecase-layer.md), implemented with GORM over Postgres. Each domain has an interface + struct pair (e.g. `UserRepository`/`userRepository`) constructed via `New*Repository(db *gorm.DB, ...)`.

## Query strategy

Simple CRUD goes through GORM's query builder (`db.Create`, `db.Save`, `db.Where(...).Find`). Anything non-trivial — joins, aggregates, cursor pagination — drops to raw SQL kept as package-level string vars in sibling `*_pg_query.go` files, with separate `Build*` functions assembling dynamic WHERE clauses via `strings.Builder`. [internal/repositories/pgsql/user_pg_query.go:244-309](internal/repositories/pgsql/user_pg_query.go#L244-L309)

`UserRepository` alone lists ~24 methods; every method wraps its body in `defer LogRepository(ctx, err)` for centralized query logging, mirroring the [Usecase Layer](usecase-layer.md)'s `LogService` pattern. [internal/repositories/pgsql/user_pg_repository.go:12-39](internal/repositories/pgsql/user_pg_repository.go#L12-L39)

## Related

- [Cursor Pagination](cursor-pagination.md) — `GetAllWithCursor`'s over-fetch-and-encrypt approach.
- [RBAC Role Derivation](rbac-role-derivation.md) — how combined roles are computed at query time.
- [Competing Transaction Abstractions](competing-transaction-abstractions.md) — `Transaction` vs `Atomic`, and a likely transaction-scoping bug.
