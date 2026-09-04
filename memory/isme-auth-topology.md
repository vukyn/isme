---
name: isme-auth-topology
description: "isme auth facts that contradict stale CLAUDE.md — RS256, medioa2 trusts isme, rainy no auth, SQLite DROP COLUMN works"
metadata: 
  node_type: memory
  type: reference
---

isme auth topology, verified 2026-06-08 (corrects stale root CLAUDE.md):

- **isme access token = RS256** (root CLAUDE.md says HS256 — stale). Refresh token = HS256.
- **medioa2 delegates auth to isme** via `external_auth_service`; `parseIsAdminFromToken` reads only `adm` claim via ParseUnverified — no local verify, ignores `perms`.
- **rainy has NO auth at all** (greenfield).
- **kuery `rbac.RequirePermission` middleware exists but is UNUSED** by medioa2/rainy; admin currently bypasses all checks.
- **SQLite DROP COLUMN IS supported** in isme's pure-Go bun migrate path (used in migrations 003/005/011) — schema drops are plain ALTER.
- **isme HAS `_test.go` files** (root CLAUDE.md "no test files yet" is stale for isme).
- isme RBAC tables: `roles`, `permissions(resource:action)`, `role_permissions`, `user_roles(+app_service_id nullable scope)`. Perms baked into JWT, re-fetched every login+refresh.

Related: [[isme-app-owned-rbac-design]]
