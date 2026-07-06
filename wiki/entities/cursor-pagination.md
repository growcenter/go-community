---
tags: [concept]
sources: [raw/2026-07-06-bootstrap-go-backend-repositories.md]
updated: 2026-07-06
ingested_at: 91a596d
---

# Cursor Pagination

`userRepository.GetAllWithCursor` (part of the [Repository Layer](repository-layer.md)) implements keyset/cursor pagination rather than offset-based paging:

1. Fetches `limit + 1` rows to cheaply detect whether more pages exist (`hasMore`), rather than issuing a separate count-ahead query.
2. For backward pagination (`direction=prev`), reverses the fetched slice back into forward order before returning it to the caller.
3. Encrypts `(CreatedAt, ID)` of the boundary record into an opaque cursor token via `internal/pkg/cursor`, returned as `prev`/`next` for the client to pass back on the next request.

[internal/repositories/pgsql/user_pg_repository.go:207-271](internal/repositories/pgsql/user_pg_repository.go#L207-L271)

The query itself is built dynamically: `BuildQueryGetAllUser` appends filter/search/cursor/order/limit clauses to a base query, flipping the cursor comparison operator (`<` vs `>`) depending on pagination direction. [internal/repositories/pgsql/user_pg_query.go:244-309](internal/repositories/pgsql/user_pg_query.go#L244-L309)
