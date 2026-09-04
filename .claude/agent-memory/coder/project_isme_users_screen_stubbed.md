---
name: isme-users-screen-stubbed
description: isme /users UI wired to real backend endpoints; only resetUserPassword remains a stub (blocked on email infra) as of 2026-06-08
metadata:
  type: project
---

The isme Users management screen (`/users`, ported from `demo/cursor-users-design.html` on 2026-06-06) still runs on typed stubs in `ui/src/apis/user.ts` — but as of 2026-06-06 (RBAC phases A–E) the backend endpoints EXIST: `GET /api/v1/users`, `PATCH /users/:id/status`, `DELETE /users/:id`, `GET /users/:id/sessions`, `POST /users/:id/sessions/:sessionId/revoke`, plus a full role domain (`/roles`, `/permissions`). Reset-password and invite remain unimplemented server-side (deferred: no email infra).

**Why:** the RBAC implementation run was explicitly backend-only ("don't touch UI in this run"); the UI wire-up is a separate follow-up task.

**How to apply:** mostly resolved — `ui/src/apis/user.ts` now uses real fetches, and the invite stub was replaced by the real invitation APIs on 2026-06-08 ([[isme-invite-flow]]). Only `resetUserPassword` remains a stub (blocked on email infra; its row action is disabled). The cursor admin shell lives in `ui/src/layouts/AdminShell.tsx`, shared by future cursor-theme screens (Roles & Permissions). Pre-existing lint errors (13) in ProtectedRoute/color-mode/useAuth/button/toaster — not from this work.
