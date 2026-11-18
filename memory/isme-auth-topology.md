---
name: isme-auth-topology
description: "isme auth facts that contradict stale CLAUDE.md — RS256, medioa2 trusts isme, rainy no auth"
metadata: 
  node_type: memory
  type: reference
---

isme auth topology, verified 2026-06-08 (corrects stale root CLAUDE.md):

- **isme access token = RS256** (root CLAUDE.md says HS256 — stale). Refresh token = HS256.
- **medioa2 delegates auth to isme** via `external_auth_service`; `parseIsAdminFromToken` reads only `adm` claim via ParseUnverified — no local verify, ignores `perms`.
- **rainy has NO auth at all** (greenfield).
- **kuery `rbac.RequirePermission` middleware exists but is UNUSED** by medioa2/rainy; admin currently bypasses all checks.
- isme is **Postgres-only** since 2026-11-18 (the SQLite path and its incremental history were removed; see `CLAUDE.md` → Database).
- **isme HAS `_test.go` files** (root CLAUDE.md "no test files yet" is stale for isme); its database tests need Docker (kuery `bun/pgtest`).
- isme RBAC tables: `roles`, `permissions(resource:action)`, `role_permissions`, `user_roles(+app_service_id nullable scope)`. Perms baked into JWT, re-fetched every login+refresh.

Related: [[isme-app-owned-rbac-design]]
