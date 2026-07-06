---
tags: [concept]
sources: [raw/2026-07-06-bootstrap-go-backend-repositories.md]
updated: 2026-07-06
ingested_at: 91a596d
---

# RBAC Role Derivation

A user's combined roles are not stored directly on the `users` table beyond `user_types` — they're derived at query time by `queryGetRBACByCommunityId`, which unnests a user's `user_types` array and left-joins each type against a `user_types` mapping table to pull in that type's associated `roles`, aggregating the result into `combined_roles`. [internal/repositories/pgsql/user_pg_query.go:34-51](internal/repositories/pgsql/user_pg_query.go#L34-L51)

This means the effective permission set for a user is `user.roles` (directly assigned) unioned with whatever roles are attached to each of the user's `user_types` in the `user_types` table — changing a user type's role mapping retroactively changes what every user of that type can do, without touching individual user rows. This is read by `UserRepository.GetRBAC`, part of the [Repository Layer](repository-layer.md).
