---
name: rbac-no-autoseed-perms
description: isme app-owned RBAC — creating an app no longer auto-seeds permissions; admin role starts empty, perms created manually
metadata:
  type: project
---

isme RBAC is app-owned. As of ~2026-06, creating an app no longer auto-seeds permissions: the app's auto-provisioned `admin` role starts with ZERO permissions. Admins create `resource:action` permissions manually (via an "Add permission" affordance in the permission catalog/matrix), then toggle them onto roles via the existing role×permission matrix.

**Why:** product decision — apps shouldn't get a presumed CRUD catalog they didn't ask for; the catalog is built up explicitly per app.

**How to apply:** any RBAC mock/UI must cover the zero-permission empty state ("Add your first permission" CTA) for new apps. The isme app itself is the exception — it is system-managed/read-only with its seeded perms (lock + amber treatment, Add-permission hidden/locked). Perm strings are NOT app-prefixed (`object:read`, not `medioa2:object:read`); the app namespace is the token `resource_access` key. See `isme/demo/aurora-rbac-roles.html` (Aurora theme) which now mocks the add-permission dialog + empty state. Related: [[demo-mock-source-of-truth]].
