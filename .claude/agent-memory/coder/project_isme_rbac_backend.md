---
name: isme-rbac-backend
description: isme RBAC backend landed 2026-06-06 (kuery rbac pkg + role domain + migrations 006-011); kuery v1.15.0 tagged + isme on it; /auth/me no longer returns is_admin (adm lives in JWT claims)
metadata:
  type: project
---

RBAC backend for isme implemented 2026-06-06 (phases A–E): kuery gained `rbac` package + claims/ctx/jwt perms-roles-adm extensions (additive — medioa2/rainy verified compiling); isme gained `role` domain, user/user_session admin endpoints, migrations 006–011 (18-permission catalog, rol_admin/rol_member/rol_viewer system roles), and login/refresh now embed perms/roles/adm in the RS256 access token.

**Why:** plan said the kuery release would be "v1.13.0" but that tag already existed (security fixes), so isme keeps `replace github.com/vukyn/kuery => ../kuery` in go.mod until the committer tags the NEXT version (v1.14.0+) and swaps the replace for a real require. Also: root CLAUDE.md says `make migrate-up DB=app` for isme but `db/migrate.go` only accepts `sqlite` — use `DB=sqlite`.

**How to apply:** deferred items if asked to extend: reset-password endpoint (needs email infra; `user:reset_password` perm row exists in catalog). The invite flow landed 2026-06-08 (see [[isme-invite-flow]]) and CLOSED the signup-role gap: public signup is removed entirely; invite-accept assigns the invited role via AddMembers. `[[isme-users-screen-stubbed]]` covers the pending UI wire-up.

Update 2026-06-07: app_service admin endpoints added (GET list + PATCH status, RBAC-guarded). Behavior change user approved: `VerifyApp` now fails for non-active apps and `RefreshApp` rejects terminated apps — marking an app inactive/terminated in isme breaks that app's SSO handshake (medioa2's `my-lab-medioa2` row was active at change time). "Last rotated" UI label was accepted as `updated_at` (no `secret_rotated_at` column, decided option a).

Update 2026-06-07 (later): kuery is tagged through v1.15.0 and isme's go.mod requires it directly (replace directive gone). Cross-repo gotcha resolved during medioa2 share work: isme's internal `/auth/me` response dropped `is_admin` (only id/name/email; adm/perms moved into the RS256 token claims via isme #19), but isme's own `external/auth/models` client STILL declares the IsAdmin field — consumers parsing the body silently get false. medioa2 was fixed by bumping kuery to v1.15.0 and decoding `adm` from the (isme-validated) access token via `jwt.NewParser().ParseUnverified` + `claims.GetIsAdmin()` in its auth usecase (`parseIsAdminFromToken`, unit-tested). Any other isme consumer with the same shim needs the same fix; cleaner long-term: drop IsAdmin from isme's external client model or re-add it to the /auth/me body.
